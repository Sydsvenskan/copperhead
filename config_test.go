package copperhead_test

import (
	"net/url"
	"os"
	"testing"
	"time"

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

	TextFail  textFail
	URL       *url.URL
	CopperURL *copperhead.URL

	Time     copperhead.Time
	Duration copperhead.Duration
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

func TestTimeConfig(t *testing.T) {
	var conf mixConf

	var ts = "2018-10-12T13:47:05Z"
	var ds = "10s"

	os.Setenv("TEST_TIME", ts)
	os.Setenv("TEST_DURATION", ds)

	_, err := copperhead.New(&conf,
		copperhead.WithEnvironment(map[string]string{
			"Time":     "TEST_TIME",
			"Duration": "TEST_DURATION",
		}))
	if err != nil {
		t.Error(err.Error())
	}

	if conf.Time.Format(time.RFC3339) != ts {
		t.Errorf("unexpected Time value %q",
			conf.Time.Format(time.RFC3339))
	}

	if conf.Duration.String() != ds {
		t.Errorf("unexpected Duration value %q",
			conf.Duration.String())
	}
}

func TestBadTimeConfig(t *testing.T) {
	var conf mixConf

	os.Setenv("TEST_TIME", "NEED MORE INPUT")
	os.Setenv("TEST_DURATION", "UNTIL THE END OF TIME!")

	cop, err := copperhead.New(&conf)
	if err != nil {
		t.Error(err.Error())
	}

	if err := cop.Getenv("Time", "TEST_TIME"); err == nil {
		t.Error("expected time parsing to fail")
	}

	if err := cop.Getenv("Duration", "TEST_DURATION"); err == nil {
		t.Error("expected duration parsing to fail")
	}
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

func TestRequire(t *testing.T) {
	v := &mixConf{}
	c, err := copperhead.New(v)
	if err != nil {
		t.Error("failed to create config: " + err.Error())
		return
	}

	if err := c.Require("Text"); err == nil {
		t.Error("Text should be missing")
	}

	if err := c.Require("Is___NotAField"); err == nil {
		t.Error("Is___NotAField should not resolve")
	}

	if err := c.Require("EmptyInterface"); err == nil {
		t.Error("EmptyInterface should be missing")
	}

	if err := c.Require("CopperURL"); err == nil {
		t.Error("CopperURL should be missing")
	}

	if err := c.Require("Nested"); err == nil {
		t.Error("Nested struct should be missing")
	}

	v.Nested.Value = "Hello!"

	if err := c.Require("Nested"); err != nil {
		t.Error("Nested struct should not be missing")
	}

	if err := c.Require("Nested", "Text"); err == nil {
		t.Error("Text should still be missing")
	}

	v.CopperURL = copperhead.MustParseURL("https://example.com")

	if err := c.Require("CopperURL"); err != nil {
		t.Error("CopperURL should not be missing")
	}
}

func TestRequireOnStatic(t *testing.T) {
	conf := struct {
		Env    string
		Port   int
		Consul *copperhead.URL

		BulkQueue          *copperhead.URL
		TransactionalQueue *copperhead.URL
		TriggerQueue       *copperhead.URL
	}{
		Env:  "test",
		Port: 6194,
		Consul: copperhead.MustParseURL(
			"http://127.0.0.1:8500",
		),
	}

	c, err := copperhead.New(&conf)
	if err != nil {
		t.Error("failed to create config: " + err.Error())
		return
	}

	err = c.Require(
		"Env", "Port", "Consul",
		"BulkQueue", "TransactionalQueue", "TriggerQueue",
	)
	if err == nil {
		t.Error("BulkQueue should be missing")
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

func TestURLOK(t *testing.T) {
	os.Setenv("FUBAR", "https://www.example.com")

	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"URL": "FUBAR",
		}),
	)
	if err != nil {
		t.Error(err.Error())
	}
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

func TestCopperheadURLFail(t *testing.T) {
	os.Setenv("FUBAR", ":")

	v := &mixConf{}

	_, err := copperhead.New(v,
		copperhead.WithEnvironment(map[string]string{
			"CopperURL": "FUBAR",
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

func TestMustParseURLPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected MustParseURL to panic")
		}
	}()

	_ = copperhead.MustParseURL(":")
}

func TestMustParseURL(t *testing.T) {
	rawURL := "https://example.com"
	u := copperhead.MustParseURL(rawURL)

	if u.String() != rawURL {
		t.Errorf("expected parsed URL to be %q, got %q",
			rawURL, u.String())
	}
}
