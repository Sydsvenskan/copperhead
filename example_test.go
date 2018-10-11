package config_test

import (
	"fmt"
	"net/url"
	"os"

	"github.com/Sydsvenskan/config"
	yaml "gopkg.in/yaml.v2"
)

// Configuration is an example configuration struct
type Configuration struct {
	Name   string
	WD     string
	URL    *url.URL
	Birdie *ExampleNest
}

type ExampleNest struct {
	Name       string
	Value      int64
	YamlIT     string `yaml:"yaml_it"`
	ComplexEnv int
}

func ExampleUsage() {
	os.Setenv("APP_NAME", "example-app")
	os.Setenv("APP_URL", "https://www.example.com")
	os.Setenv("APP_BIRD", "12")
	os.Setenv("DUMB_USAGE", `{"ComplexEnv":42}`)

	cfg := Configuration{
		WD: "with defaults",
	}
	_, err := config.New(&cfg,
		config.WithConfigurationFile(
			"example.conf", config.FileRequired, nil,
		),
		config.WithConfigurationFile(
			"example.conf.yaml",
			config.FileRequired,
			config.UnmarshalerFunc(yaml.Unmarshal),
		),
		config.WithEnvironment(map[string]string{
			"Name":         "APP_NAME",
			"URL":          "APP_URL",
			"Birdie.Value": "APP_BIRD",
			"Birdie":       "DUMB_USAGE",
		}),
	)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Name:", cfg.Name)
	fmt.Println("WD:", cfg.WD)
	fmt.Println("URL:", cfg.URL.String())
	fmt.Println("URL proto:", cfg.URL.Scheme)
	fmt.Println("Birdie.Name:", cfg.Birdie.Name)
	fmt.Println("Birdie.Value:", cfg.Birdie.Value)
	fmt.Println("Birdie.YamlIT:", cfg.Birdie.YamlIT)
	fmt.Println("Birdie.ComplexEnv:", cfg.Birdie.ComplexEnv)

	// Output:
	// Name: example-app
	// WD: with defaults
	// URL: https://www.example.com
	// URL proto: https
	// Birdie.Name: Heron
	// Birdie.Value: 12
	// Birdie.YamlIT: Hello from YAML
	// Birdie.ComplexEnv: 42
}
