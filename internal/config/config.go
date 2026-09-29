package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// HTTPServer holds HTTP server specific configurations.
// NOTE: Field names MUST be capitalized (exported in Go) so external packages
// like cleanenv / yaml unmarshaler can read and write to them using reflection.
type HTTPServer struct {
	// yaml:"address" maps this field to the 'address' key in YAML.
	// env-default specifies a fallback value if not provided in config or env.
	Addr string `yaml:"address" env-default:"localhost:8082"`
}

// Config represents the overall application configuration.
// Struct tags define how each field is mapped and validated:
//   - yaml:"...": maps the field to a key in the YAML file.
//   - env:"...": allows overriding this value via an environment variable.
//   - env-required:"true": cleanenv will throw an error if this field is missing.
//   - env-default:"...": sets a default value if not specified.
type Config struct {
	Env         string     `yaml:"env" env:"ENV" env-required:"true" env-default:"production"`
	StoragePath string     `yaml:"storage_path" env-required:"true"`
	HTTPServer  HTTPServer `yaml:"http_server"`
}

// MustLoad reads the configuration from either an environment variable or a CLI flag,
// parses the YAML file, and populates the Config struct.
// The "Must" prefix is a Go convention meaning: this function will terminate the app
// (log.Fatal/panic) if it fails, instead of returning an error.
func MustLoad() *Config {
	var configPath string

	// 1. Try to read the path from the environment variable CONFIG_PATH.
	configPath = os.Getenv("CONFIG_PATH")
	

	// 2. If no environment variable was found, check command line flags: e.g. -config=config/local.yml
	if configPath == "" {
		flags := flag.String("config", "", "path to the configuration file")
		flag.Parse()
		configPath = *flags

		// If neither env var nor flag was provided, we cannot proceed.
		if configPath == "" {
			log.Fatal("Config path is not set")
		}
	}

	// 3. Verify that the file actually exists at the given path.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file does not exist: %s", configPath)
	}

	var cfg Config

	// 4. cleanenv reads the YAML file, parses the values, fills the struct fields,
	// and validates required fields based on the struct tags.
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("Cannot read config file: %s", err.Error())
	}

	return &cfg
}
