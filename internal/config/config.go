package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config contains all the settings parsed from the config file.
type Config struct {
	Telegram struct {
		ApiKey string `yaml:"apikey"`
	} `yaml:"telegram"`

	Twitch struct {
		ClientID     string `yaml:"client-id"`
		ClientSecret string `yaml:"client-secret"`
	} `yaml:"twitch"`

	General struct {
		PollingInterval int `yaml:"polling-interval"`
	} `yaml:"general"`
}

// GetConfig parses a config.yml file placed in the root execution path containing credentials and settings for the application.
// It returns an object containing the parsed settings.
func GetConfig() (*Config, error) {

	config := &Config{}

	// open config file
	p := filepath.FromSlash("./config.yml")
	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// init new YAML decode
	d := yaml.NewDecoder(file)

	// start YAML decoding from file
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	if (&Config{}) == config {
		return nil, errors.New("config is empty")
	}

	return config, err
}
