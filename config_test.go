package copperhead_test

import (
	"net/url"
	"os"
	"testing"

	"github.com/Sydsvenskan/copperhead"
	"github.com/pkg/errors"
)

type someIFace interface {
	SomeMethod()
}

type mixConf struct {
	hidden string

	Text string

	EmptyInterface interface{}
	Interface      someIFace
	DoublePointer  **nested
	NestedPtr      *nested
	Nested         nested

	TextFail textFail
	URL      *url.URL
}

type nested struct {
	Value    string
	ValuePtr *string
}

type textFail struct{}

// UnmarshalText that always fails
func (*textFail) UnmarshalText(data []byte) error {
	return errors.New("born to fail")
}

func TestNilConfiguration(t *testing.T) {
	_, err := copperhead.New(nil)
	if err == nil {
		t.Error("expected nil config value to fail")
		return
	}
	t.Log(err.Error())
}

func TestNonStructConfiguration(t *testing.T) {
	var conf string
	_, err := copperhead.New(&conf)
	if err == nil {
		t.Error("expected non-struct config value to fail")
		return
	}
	t.Log(err.Error())
}

func TestValueConfiguration(t *testing.T) {
	_, err := copperhead.New(mixConf{})
	if err == nil {
		t.Error("expected config passed by value to fail")
		return
	}
	t.Log(err.Error())
}

func TestOptionalFile(t *testing.T) {
	_, err := copperhead.New(&mixConf{},
		copperhead.WithConfigurationFile(
			"foobar.json",
			copperhead.FileOptional, nil,
		),
	)
	if err != nil {
		t.Error("failed when optional file was missing: " + err.Error())
	}
}

func TestReadFileFailure(t *testing.T) {
	_, err := copperhead.New(&mixConf{},
		copperhead.WithConfigurationFile(
			"./",
			copperhead.FileRequired, nil,
		),
	)
	if err == nil {
		t.Error("expected file that is a directory to cause a failure")
	}
	t.Log(err.Error())
}

func TestRequiredFileFailure(t *testing.T) {
	_, err := copperhead.New(&mixConf{},
		copperhead.WithConfigurationFile(
			"foobar.json",
			copperhead.FileRequired, nil,
		),
	)
	if err == nil {
		t.Error("expected missing required file to cause a failure")
	}
	t.Log(err.Error())
}

func TestConfigurationData(t *testing.T) {
	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithConfigurationData(
			[]byte(`{"Text":"hello"}`), nil,
		),
	)
	if err != nil {
		t.Error("failed to read environment: " + err.Error())
	}

	if v.Text != "hello" {
		t.Errorf("unexpected 'Text' value %#v", v.Text)
	}
}

func TestEmptyInterfaceStringValue(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}
	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"EmptyInterface": "FUBAR",
		}),
	)
	if err != nil {
		t.Error("failed to read environment: " + err.Error())
	}

	str, ok := v.EmptyInterface.(string)
	if !ok {
		t.Error("'EmptyInterface' value isn't a string")
	}

	if str != "foo" {
		t.Error("'EmptyInterface' value wasn't \"foo\"")
	}
}

func TestPointerToPointer(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}
	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"DoublePointer.Value": "FUBAR",
		}),
	)
	if err == nil {
		t.Error("expected mapping through double pointer to fail")
		return
	}
	t.Log(err.Error())
}

func TestUnknownField(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}
	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"No.Exist": "FUBAR",
		}),
	)
	if err == nil {
		t.Error("expected mapping to unknown field to fail")
		return
	}
	t.Log(err.Error())
}

func TestMissingEnv(t *testing.T) {
	v := &mixConf{
		Text: "default",
	}
	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"Text": "__TEST_MISSING_ENV_VAR",
		}),
	)
	if err != nil {
		t.Error("failed to create config: " + err.Error())
		return
	}

	if v.Text != "default" {
		t.Error("`Text` should not have been changed")
	}
}

func TestEnvTypeMismatch(t *testing.T) {
	os.Setenv("FUBAR", "0")

	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"Nested": "FUBAR",
		}),
	)
	if err == nil {
		t.Error("should have failed to assing an int to a struct")
		return
	}
	t.Log(err.Error())

}

func TestUnmarshalTextFail(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"TextFail": "FUBAR",
		}),
	)
	if err == nil {
		t.Error("should have failed unmarshal text")
		return
	}
	t.Log(err.Error())
}

func TestURLFail(t *testing.T) {
	os.Setenv("FUBAR", ":")

	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"URL": "FUBAR",
		}),
	)
	if err == nil {
		t.Error("should have failed to parse URL")
		return
	}
	t.Log(err.Error())
}

func TestNested(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"NestedPtr.Value":    "FUBAR",
			"NestedPtr.ValuePtr": "FUBAR",
		}),
	)
	if err != nil {
		t.Error("failed to create config: " + err.Error())
		return
	}

	if v.NestedPtr == nil {
		t.Error("'NestedPtr' should not be nil")
		return
	}

	if v.NestedPtr.Value != "foo" {
		t.Errorf("unexpected value %q for 'NestedPtr.Value'",
			v.NestedPtr.Value,
		)
	}

	if v.NestedPtr.ValuePtr == nil {
		t.Error("'ValuePtr' should not be nil")
		return
	}

	if *v.NestedPtr.ValuePtr != "foo" {
		t.Errorf("unexpected value %q for 'NestedPtr.ValuePtr'",
			*v.NestedPtr.ValuePtr,
		)
	}
}

func TestBadNesting(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"Nested.Value.WeDugTooDeep": "FUBAR",
		}),
	)
	if err == nil {
		t.Error("should have failed to assign to field on string")
		return
	}
	t.Log(err.Error())
}

func TestAssignUnexported(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"hidden": "FUBAR",
		}),
	)
	if err == nil {
		t.Error("should have failed to assign to unexported field")
		return
	}
	t.Log(err.Error())
}
