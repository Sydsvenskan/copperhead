// Package copperhead provides a configuration loader that can load
// configuration from environment, files, or byte slices.
package copperhead

import (
	"encoding"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"strings"
)

// Config encapsulates configuration loading.
type Config struct {
	obj reflect.Value
}

// Option configures our... inception!
type Option func(c *Config) error

// WithEnvironment bootstraps our configuration with environment
// variables.
func WithEnvironment(envMap map[string]string) Option {
	return func(c *Config) error {
		return c.Environment(envMap)
	}
}

// Unmarshaler is something that can unmarshal a configuration.
type Unmarshaler interface {
	Unmarshal(data []byte, v interface{}) error
}

// UnmarshalerFunc is an unmarshal function
type UnmarshalerFunc func(data []byte, v interface{}) error

// Unmarshal the config data.
func (uf UnmarshalerFunc) Unmarshal(
	data []byte, v interface{},
) error {
	return uf(data, v)
}

// WithConfigurationData reads the provided configuration data.
func WithConfigurationData(data []byte, unm Unmarshaler) Option {
	return func(c *Config) error {
		return c.Data(data, unm)
	}
}

// FileMode controls file loading behaviour.
type FileMode string

// The file loading modes
const (
	FileRequired FileMode = "required"
	FileOptional          = "optional"
)

// WithConfigurationFile reads configuration from a file.
func WithConfigurationFile(filename string, mode FileMode, unm Unmarshaler) Option {
	return func(c *Config) error {
		return c.File(filename, mode, unm)
	}
}

// Require verifies that configuration values are set.
func Require(names ...string) Option {
	return func(c *Config) error {
		return c.Require(names...)
	}
}

// Configure populates conf.
func Configure(conf interface{}, opts ...Option) error {
	_, err := New(conf, opts...)
	return err
}

// New creates a new configuration that populates conf.
func New(conf interface{}, opts ...Option) (*Config, error) {
	if conf == nil {
		return nil, fmt.Errorf("conf cannot be nil")
	}

	v := reflect.ValueOf(conf)
	if v.Type().Kind() != reflect.Ptr {
		return nil, fmt.Errorf(
			"conf must be a pointer to a struct")
	}

	v = reflect.Indirect(reflect.ValueOf(conf))
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"conf must be a pointer to a struct")
	}

	c := &Config{
		obj: v,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Getenv reads a single environment variable.
func (c *Config) Getenv(field, env string) error {
	return c.Environment(map[string]string{
		field: env,
	})
}

// Environment populates our configuration with environment variables.
func (c *Config) Environment(envMap map[string]string) error {
	for name, envName := range envMap {
		v, err := c.resolve(name)
		if err != nil {
			return fmt.Errorf("could not resolve %q: %w",
				name, err)
		}

		eVal, ok := os.LookupEnv(envName)
		if !ok {
			continue
		}

		if err := c.assign(v, eVal); err != nil {
			return fmt.Errorf("could not assign the value of %q to %q: %w",
				envName, name, err,
			)
		}
	}
	return nil
}

// File reads configuration from a file.
func (c *Config) File(filename string, mode FileMode, unm Unmarshaler) error {
	if unm == nil {
		unm = UnmarshalerFunc(json.Unmarshal)
	}

	data, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) && mode == FileOptional {
		return nil
	}

	if os.IsNotExist(err) {
		return fmt.Errorf("missing configuration file %q: %w",
			filename, err,
		)

	} else if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	err = unm.Unmarshal(data, c.obj.Addr().Interface())
	if err != nil {
		return fmt.Errorf("failed to unmarshal configuration file %q: %w",
			filename, err,
		)
	}

	return nil
}

// Data reads the provided configuration data.
func (c *Config) Data(data []byte, unm Unmarshaler) error {
	if unm == nil {
		unm = UnmarshalerFunc(json.Unmarshal)
	}
	err := unm.Unmarshal(data, c.obj.Addr().Interface())
	if err != nil {
		return fmt.Errorf("failed to unmarshal configuration data: %w", err)
	}
	return nil
}

var urlType = reflect.TypeOf(url.URL{})

func (c *Config) assign(target reflect.Value, val string) error {
	v := reflect.ValueOf(val)

	zero, err := ensureZero("target", target)
	if err != nil {
		return err
	}
	target = *zero

	if !target.CanSet() {
		return fmt.Errorf("cannot set the value")
	}

	// Direct assignment
	if v.Type().AssignableTo(target.Type()) {
		target.Set(v)
		return nil
	}

	iface := target.Addr().Interface()

	// DEPRECATED: Special handling of URLs, because it's so
	// common and why doesn't it implement TextUnmarshaler.
	if target.Type().ConvertibleTo(urlType) {
		ub := iface.(encoding.BinaryUnmarshaler)
		err := ub.UnmarshalBinary([]byte(val))
		if err != nil {
			return err
		}
		return nil
	}

	// Generic text unmarshaling
	if tx, ok := iface.(encoding.TextUnmarshaler); ok {
		err := tx.UnmarshalText([]byte(val))
		if err != nil {
			return err
		}
		return nil
	}

	// Fall back to JSON unmarshalling
	err = json.Unmarshal([]byte(val), iface)
	if err != nil {
		return fmt.Errorf("failed to decode value as JSON: %w", err)
	}
	return nil
}

// Require checks if congiguration values are set.
func (c *Config) Require(names ...string) error {
	for _, name := range names {
		v, err := c.resolve(name)
		if err != nil {
			return fmt.Errorf("failed to resolve %q: %w",
				name, err,
			)
		}

		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				return fmt.Errorf("%q is nil", name)
			}
			continue
		}

		if v.Kind() == reflect.Bool {
			// Required isn't a meaningful concept for
			// booleans.
			continue
		}

		zero := reflect.New(v.Type()).Elem()

		if zero.Interface() == v.Interface() {
			return fmt.Errorf(
				"%q is empty", name)
		}

	}
	return nil
}

func (c *Config) resolve(name string) (reflect.Value, error) {
	path := strings.Split(name, ".")

	n := c.obj
	for len(path) > 0 {
		head := path[0]
		path = path[1:]

		if n.Kind() != reflect.Struct {
			return n, fmt.Errorf(
				"cannot get field %q from a %q value",
				head, n.Kind().String(),
			)
		}

		field := n.FieldByName(head)
		if !field.IsValid() {
			return n, fmt.Errorf(
				"%q doesn't have a field %q",
				n.Type().Name(), head,
			)
		}

		if len(path) > 0 {
			z, err := ensureZero(head, field)
			if err != nil {
				return n, err
			}
			field = *z
		}

		n = field
	}

	return n, nil
}

func ensureZero(name string, field reflect.Value) (*reflect.Value, error) {
	// We attempt to populate nil pointers with zero
	// values.
	if field.Kind() == reflect.Ptr && field.IsNil() {
		e := field.Type().Elem()
		if e.Kind() == reflect.Ptr {
			return nil, fmt.Errorf(
				"pointers to pointers (as in %q being a %q) are unsupported",
				name, field.Type().String(),
			)
		}

		zero := reflect.New(e)
		field.Set(zero)
	}

	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	return &field, nil
}
