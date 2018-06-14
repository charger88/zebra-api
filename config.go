package main

import (
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
	"time"
)

var config Config

type Config struct {
	Version string `yaml:"version"`
	Redis RedisConfig `yaml:"redis"`
	HttpPort string `yaml:"http-port"`
	PasswordPolicy string `yaml:"password-policy"`
	MinimalKeyLength int `yaml:"minimal-key-length"`
	ConfigReloadTime int64 `yaml:"config-reload-time"`
	ExpectedStripesPerHour int `yaml:"expected-stripes-per-hour"`
	AllowedBadAttempts int `yaml:"allowed-bad-attempts"`
	AppropriateChanceToGuess int `yaml:"appropriate-chance-to-guess"`
	RequireApiKey bool `yaml:"require-api-key"`
	RequireApiKeyForPostOnly bool `yaml:"require-api-key-for-post-only"`
	AllowedApiKeys []string `yaml:"allowed-api-keys"`
	AllowedSharesPeriod int `yaml:"allowed-shares-period"`
	AllowedSharesNumberInPeriod int `yaml:"allowed-shares-number-in-period"`
	PublicName string `yaml:"public-name"`
	PublicColor string `yaml:"public-color"`
	PublicURL string `yaml:"public-url"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func loadConfig() {
	var localConfig = Config{}
	addFileDataToConfig("default", &localConfig)
	addFileDataToConfig("config", &localConfig)
	config = localConfig
}

func reloadConfig() {
	ticker := time.NewTicker(time.Duration(config.ConfigReloadTime) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <- ticker.C:
				loadConfig()
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func addFileDataToConfig(name string, localConfig *Config) {
	dat, err := ioutil.ReadFile("config/"  + name + ".yaml")
	if err != nil {
		log.Fatal("Can't load config file config/"  + name + ".yaml: " + err.Error())
	}
	err = yaml.Unmarshal(dat, localConfig)
	if err != nil {
		log.Fatal("Can't parse config file config/"  + name + ".yaml: " + err.Error())
	}
}

func (c Config) isApiKeyEnabled(apiKey string) bool {
	for _, b := range c.AllowedApiKeys {
		if b == apiKey {
			return true
		}
	}
	return false
}