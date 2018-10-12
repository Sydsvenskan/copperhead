// Package copperhead provides a configuration loader that can load
// configuration from environment, files, or byte slices.
package copperhead

import (
	"encoding"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// Config encapsulates configuration loading.
type Config struct {
	obj reflect.Value
}

// Option configures our... inception!
type Option func(c *Config) error

// URL is an TextUnmarshaler-aware URL
type URL struct {
	url.URL
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (u *URL) UnmarshalText(text []byte) error {
	var u2 url.URL
	if err := u2.UnmarshalBinary(text); err != nil {
		return err
	}
	u.URL = u2
	return nil
}

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

// New creates a new configuration that populates conf.
func New(conf interface{}, opts ...Option) (*Config, error) {
	if conf == nil {
		return nil, errors.New("conf cannot be nil")
	}

	v := reflect.ValueOf(conf)
	if v.Type().Kind() != reflect.Ptr {
		return nil, errors.New(
			"conf must be a pointer to a struct")
	}

	v = reflect.Indirect(reflect.ValueOf(conf))
	if v.Kind() != reflect.Struct {
		return nil, errors.New(
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

// Environment populates our configuration with environment variables.
func (c *Config) Environment(envMap map[string]string) error {
	for name, envName := range envMap {
		v, err := c.resolve(name)
		if err != nil {
			return errors.Wrapf(err,
				"could not resolve %q", name)
		}

		eVal, ok := os.LookupEnv(envName)
		if !ok {
			continue
		}

		if err := c.assign(*v, eVal); err != nil {
			return errors.Wrapf(err,
				"could not assign the value of %q to %q",
				envName, name,
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
		return errors.Errorf(
			"missing configuration file %q",
			filename,
		)
	} else if err != nil {
		return errors.Wrap(err,
			"failed to read configuration file")
	}

	err = unm.Unmarshal(data, c.obj.Addr().Interface())
	return errors.Wrapf(err,
		"failed to unmarshal configuration file %q",
		filename,
	)
}

// Data reads the provided configuration data.
func (c *Config) Data(data []byte, unm Unmarshaler) error {
	if unm == nil {
		unm = UnmarshalerFunc(json.Unmarshal)
	}
	err := unm.Unmarshal(data, c.obj.Addr().Interface())
	return errors.Wrap(err, "failed to unmarshal configuration data")
}

var urlType = reflect.TypeOf(url.URL{})

func (c *Config) assign(target reflect.Value, val string) error {
	v := reflect.ValueOf(val)

	if !target.CanSet() {
		return errors.New("cannot set the value")
	}

	// Direct assignment
	if v.Type().AssignableTo(target.Type()) {
		target.Set(v)
		return nil
	}

	iface := target.Addr().Interface()

	// Special handling of URLs, because it's so common and why
	// doesn't it implement TextUnmarshaler.
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
	err := json.Unmarshal([]byte(val), iface)
	return errors.Wrap(err, "failed to decode value as JSON")
}

func (c *Config) resolve(name string) (*reflect.Value, error) {
	path := strings.Split(name, ".")

	n := c.obj
	for len(path) > 0 {
		head := path[0]
		path = path[1:]

		if n.Kind() != reflect.Struct {
			return nil, errors.Errorf(
				"cannot get field %q from a %q value",
				head, n.Kind().String(),
			)
		}

		field := n.FieldByName(head)
		if !field.IsValid() {
			return nil, errors.Errorf(
				"%q doesn't have a field %q",
				n.Type().Name(), head,
			)
		}

		// We attempt to populate nil pointers with zero
		// values.
		if field.Kind() == reflect.Ptr && field.IsNil() {
			e := field.Type().Elem()
			if e.Kind() == reflect.Ptr {
				return nil, errors.Errorf(
					"pointers to pointers (as in %q being a %q) are unsupported",
					head, field.Type().String(),
				)
			}

			zero := reflect.New(e)
			field.Set(zero)
		}

		if field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		n = field
	}

	return &n, nil
}
