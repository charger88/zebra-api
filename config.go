package main

import (
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
	"time"
	"reflect"
	"strings"
	"os"
	"strconv"
)

var config Config

type Config struct {
	Version string `yaml:"version"`
	RedisHost string `yaml:"redis-host"`
	RedisPort string `yaml:"redis-port"`
	RedisPassword string `yaml:"redis-password"`
	HttpInterface string `yaml:"http-interface"`
	HttpPort string `yaml:"http-port"`
	TrustedProxy []string `yaml:"trusted-proxy"`
	MaxExpirationTime int `yaml:"max-expiration-time"`
	PasswordPolicy string `yaml:"password-policy"`
	EncryptionPasswordPolicy string `yaml:"encryption-password-policy"`
	MinimalKeyLength int `yaml:"minimal-key-length"`
	ConfigReloadTime int64 `yaml:"config-reload-time"`
	ExpectedStripesPerHour int `yaml:"expected-stripes-per-hour"`
	AllowedBadAttempts int `yaml:"allowed-bad-attempts"`
	AppropriateChanceToGuess int `yaml:"appropriate-chance-to-guess"`
	RequireApiKey bool `yaml:"require-api-key"`
	RequireApiKeyForPostOnly bool `yaml:"require-api-key-for-post-only"`
	MaxTextLength int `yaml:"max-text-length"`
	AllowedApiKeys []string `yaml:"allowed-api-keys"`
	AllowedSharesPeriod int `yaml:"allowed-shares-period"`
	AllowedSharesNumberInPeriod int `yaml:"allowed-shares-number-in-period"`
	PublicName string `yaml:"public-name"`
	PublicColor string `yaml:"public-color"`
	PublicURL string `yaml:"public-url"`
	PublicEmail string `yaml:"public-email"`
	ExtendedLogs bool `yaml:"extended-logs"`
}

func loadConfig() {
	var localConfig = Config{}
	addFileDataToConfig("default", &localConfig, true)
	buildEnvironmentConfig(&localConfig)
	addFileDataToConfig("env", &localConfig, false)
	addFileDataToConfig("config", &localConfig, false)
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

func addFileDataToConfig(name string, localConfig *Config, strict bool) {
	dat, err := ioutil.ReadFile("./config/"  + name + ".yaml")
	if err != nil {
		if strict {
			log.Fatal("Can't load config file config/" + name + ".yaml: " + err.Error())
		} else {
			log.Print("Skip config file config/" + name + ".yaml")
		}
		return
	}
	err = yaml.Unmarshal(dat, localConfig)
	if err != nil {
		log.Fatal("Can't parse config file config/"  + name + ".yaml: " + err.Error())
	}
}

func buildEnvironmentConfig(localConfig *Config){
	v := reflect.ValueOf(*localConfig)
	var key string
	var envVal string
	var fType string
	var yamlKey string
	var envValBool bool
	var envValStrings []string
	configString := ""
	for i := 0; i < v.NumField(); i++ {
		yamlKey = v.Type().Field(i).Tag.Get("yaml")
		key = "ZEBRA_" + strings.ToUpper(strings.Replace(yamlKey,"-", "_", -1))
		envVal = os.Getenv(key)
		if envVal != "" {
			fType = v.Type().Field(i).Type.String()
			configString += yamlKey + ": "
			if strings.Index(fType, "[]") == 0 {
				envValStrings = strings.Split(envVal, ",")
				for _, envValStringsItem := range envValStrings {
					configString += "\n  - \"" + envValStringsItem + "\""
				}
			} else if fType == "int" {
				configString += envVal
			} else if fType == "bool" {
				envValBool, _ = strconv.ParseBool(envVal)
				if envValBool {
					configString += "true"
				} else {
					configString += "false"
				}
			} else {
				configString += "\"" + envVal + "\""
			}
			configString += "\n"
		}
	}
	err := ioutil.WriteFile("./config/env.yaml", []byte(configString), 0644)
	if err != nil {
		log.Fatal("Can't create config/env.yaml")
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