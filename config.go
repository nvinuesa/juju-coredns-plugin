package juju

import (
	"errors"
	"fmt"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

const defaultTTL = uint32(60) // 60 seconds TTL.

type Config struct {
	Ttl         *uint32               `yaml:"ttl"`
	Controllers map[string]Controller `yaml:"controllers"`
}

type Controller struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	address  string `yaml:"address"`
}

func (c *Config) Validate() error {
	if c.Ttl == nil {
		return errors.New("ttl is required")
	}
	for controllerName, controller := range c.Controllers {
		if controller.Username == "" {
			return errors.New(fmt.Sprintf("username is required in controller %q", controllerName))
		}
		if controller.Password == "" {
			return errors.New(fmt.Sprintf("password is required in controller %q", controllerName))
		}
		if controller.address == "" {
			return errors.New(fmt.Sprintf("address is required in controller %q", controllerName))
		}
	}
	return nil
}

func FromConfigFile() (*Config, error) {
	configFilePath := os.Getenv("JUJU_DNS_PLUGIN_CONF_PATH")
	if configFilePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		configFilePath = path.Join(home, ".local/share/juju-dns", "juju-dns-plugin.yaml")
	}
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	config := &Config{}
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	ttl := defaultTTL
	if config.Ttl == nil {
		config.Ttl = &ttl
	}

	return config, config.Validate()
}
