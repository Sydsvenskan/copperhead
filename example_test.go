package copperhead_test

import (
	"fmt"
	"os"

	"github.com/Sydsvenskan/copperhead"
	yaml "gopkg.in/yaml.v2"
)

// Configuration is an example configuration struct
type Configuration struct {
	Name    string
	WD      string
	URL     *copperhead.URL
	FileURL *copperhead.URL
	Birdie  *ExampleNest
}

// ExampleNest is an exampe of a nested configuration struct.
type ExampleNest struct {
	Name       string
	Value      int64
	YamlIT     string `yaml:"yaml_it"`
	ComplexEnv int
}

// ExampleConfig is a testable example for Copperhead
func ExampleConfig() {
	os.Setenv("APP_NAME", "example-app")
	os.Setenv("APP_URL", "https://www.example.com")
	os.Setenv("APP_BIRD", "12")
	os.Setenv("DUMB_USAGE", `{"ComplexEnv":42}`)

	cfg := Configuration{
		WD: "with defaults",
	}
	ch, err := copperhead.New(&cfg,
		copperhead.WithConfigurationFile(
			"./test-data/example.conf", copperhead.FileRequired, nil,
		),
		copperhead.WithConfigurationFile(
			"./test-data/example.conf.yaml",
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

	// Load additional config file.
	err = ch.File("./test-data/file-url.json",
		copperhead.FileRequired, nil)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Println("File URL (json):", cfg.FileURL.String())

	err = ch.File("./test-data/file-url.yaml",
		copperhead.FileRequired,
		copperhead.UnmarshalerFunc(yaml.Unmarshal))
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	fmt.Println("File URL (yaml):", cfg.FileURL.String())

	// Output:
	// Name: example-app
	// WD: with defaults
	// URL: https://www.example.com
	// URL proto: https
	// Birdie.Name: Heron
	// Birdie.Value: 12
	// Birdie.YamlIT: Hello from YAML
	// Birdie.ComplexEnv: 42
	// File URL (json): http://www.example.com/json
	// File URL (yaml): http://www.example.com/yaml
}
