# Copperhead

[![GoDoc](https://godoc.org/github.com/Sydsvenskan/copperhead?status.svg)](https://godoc.org/github.com/Sydsvenskan/copperhead) [![Goreport](https://goreportcard.com/badge/github.com/Sydsvenskan/copperhead)](https://goreportcard.com/report/github.com/Sydsvenskan/copperhead) [![Build Status](https://travis-ci.org/Sydsvenskan/copperhead.svg?branch=master)](https://travis-ci.org/Sydsvenskan/copperhead) [![codecov](https://codecov.io/gh/Sydsvenskan/copperhead/branch/master/graph/badge.svg)](https://codecov.io/gh/Sydsvenskan/copperhead)

Configuration loader that can load configuration from environment, files, or byte slices.

Copperhead was written to match the features that we actually used in viper (https://github.com/spf13/viper). Configuration is always loaded into a struct. Unless you pass an `UnmarshalerFunc` the only supported format is JSON. The predecence of configuration sources is completely controlled by the order in which you load them. URLs can be parsed as a part of the configuration loading step.

Copperhead supports the "option function"-style shown below, which has the advantage of just giving you one place to error check. You can also call `func (c *Config) Environment`, `func (c *Config) File`, and `func (c *Config) Data` to load configuration sources one by one.

Example usage:

```
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
	_, err := copperhead.New(&cfg,
		copperhead.WithConfigurationFile(
			"example.conf", copperhead.FileRequired, nil,
		),
		copperhead.WithConfigurationFile(
			"example.conf.yaml",
			copperhead.FileRequired,
			copperhead.UnmarshalerFunc(yaml.Unmarshal),
		),
		copperhead.WithEnvironment(map[string]string{
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
```
