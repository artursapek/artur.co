package config

import (
	"io/ioutil"
	"log"

	"launchpad.net/goyaml"
)

type AppConfig struct {
	RawRoot     string `yaml:"raw_root"`
	ResizedRoot string `yaml:"resized_root"`

	RawURLPrefix     string `yaml:"raw_url_prefix"`
	ResizedURLPrefix string `yaml:"resized_url_prefix"`
}

var Config AppConfig

func init() {
	configData, readErr := ioutil.ReadFile("config.yml")
	if readErr != nil {
		log.Fatal(readErr)
	}
	parseErr := goyaml.Unmarshal(configData, &Config)
	if parseErr != nil {
		log.Fatal(parseErr)
	}
}
