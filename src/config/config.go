package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func LoadConfig(path string, log *logrus.Logger) *AutoCloserConfig {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("gitlab")

	err := viper.BindEnv("gitlabApiToken", "GITLAB_API_TOKEN")
	if err != nil {
		log.Fatalf("Unable to bind GITLAB_API_TOKEN env var!")
	}

	if gt := viper.Get("gitlabApiToken"); gt == nil {
		log.Fatal("GITLAB_API_TOKEN env var is not set!")
	}

	var cfg AutoCloserConfig

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Unable to read config.yaml, %s", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Unable to unmarshal AutoCloserConfig struct, %v", err)
	}

	return &cfg
}
