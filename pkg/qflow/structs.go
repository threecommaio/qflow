package qflow

import (
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	HTTP struct {
		Concurrency int           `yaml:"concurrency"`
		Timeout     time.Duration `yaml:"timeout"`
	}
	Queue struct {
		MaxMessageSize int32 `yaml:"maxMsgSize"`
	}

	Endpoints []struct {
		Name  string   `yaml:"name"`
		Hosts []string `yaml:"hosts"`
	}
}

// type UnmarshalingTimeout time.Duration

// ParseConfig handles mapping the filename to config struct
func ParseConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config := Config{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
