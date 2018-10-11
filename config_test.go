package config_test

import (
	"os"
	"testing"

	"github.com/Sydsvenskan/config"
)

type someIFace interface {
	SomeMethod()
}

type mixConf struct {
	hidden string

	EmptyInterface interface{}
	Interface      someIFace
	DoublePointer  **string
}

func TestEmptyInterfaceStringValue(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}
	_, err := config.New(v,
		config.WithEnvironment(map[string]string{
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

func TestUnknownField(t *testing.T) {
	os.Setenv("FUBAR", "foo")

	v := &mixConf{}
	_, err := config.New(v,
		config.WithEnvironment(map[string]string{
			"No.Exist": "FUBAR",
		}),
	)
	if err == nil {
		t.Error("expected mapping to unknown field to fail")
	}
}
