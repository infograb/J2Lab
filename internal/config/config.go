package config

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// Config is the struct for the config file
// The following features are supported:
// 1. Mapping between config file and struct
// 2. Configuration syntax validation

type Config struct {
	GitLab struct {
		Host  string `yaml:"host" validate:"required,url"`
		Token string `yaml:"token"`
	} `yaml:"gitlab"`

	Jira struct {
		Host  string `yaml:"host"`
		Token string `yaml:"token"`
	} `yaml:"jira"`
}

// config file is read by yaml format
// You can add --config option to specify the config file
// If you don't specify the config file, the default config file is used
// - $HOME/.config/jira2gitlab/config.yaml
// - $PWD/config.yaml

func InitConfig() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Error getting home directory: %s", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %s", err)
	}

	// Search config in home directory with name ".cobra" (without extension).
	viper.AddConfigPath(filepath.Join(home, ".config/jira2gitlab"))
	viper.AddConfigPath(pwd)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if configEnv := viper.GetString("config"); configEnv != "" {
		viper.SetConfigFile(configEnv)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	log.Debugf("Using config file: %s", viper.ConfigFileUsed())
}
