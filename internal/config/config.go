package config

import (
	"fmt"
	"io"
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

	Project map[string]string `yaml:"project"`
	User    map[string]string `yaml:"user"`
}

func GetConfig() *Config {
	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %s", err)
	}

	return &cfg
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

	if configFile := viper.GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debugf("Config file not found, creating... : %s", err)
			if err := runConfigNew(""); err != nil {
				log.Fatalf("Error creating config file: %s", err)
			}

			if err := viper.ReadInConfig(); err != nil {
				log.Fatalf("Error reading config file: %s", err)
			}
		} else {
			log.Fatalf("Error reading config file: %s", err)
		}
	}

	log.Debugf("Using config file: %s", viper.ConfigFileUsed())
}

func runConfigNew(configFile string) error {
	srcFilePathAbs, err := filepath.Abs("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Error getting absolute path of source file: %s", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %s", err)
		return err
	}

	var dstFilePathAbs string
	if configFile != "" {
		dstFilePathAbs, err = filepath.Abs(configFile)
		if err != nil {
			log.Fatalf("Error getting absolute path of destination file: %s", err)
		}
	} else {
		dstFilePathAbs, err = filepath.Abs(filepath.Join(pwd, "config.yml"))
		if err != nil {
			log.Fatalf("Error getting absolute path of destination file: %s", err)
		}
	}

	// Check if the source file exists
	if _, err := os.Stat(srcFilePathAbs); os.IsNotExist(err) {
		log.Fatalf("Source file does not exist: %s", srcFilePathAbs)
	}

	// Open the source file for reading
	srcFile, err := os.Open(srcFilePathAbs)
	if err != nil {
		log.Fatalf("Error opening source file: %s", err)
	}
	defer srcFile.Close()

	// Ask to overwrite the destination file if it already exists
	if _, err := os.Stat(dstFilePathAbs); err == nil {
		log.Warnf("Destination file already exists: %s", dstFilePathAbs)
		log.Warn("Do you want to overwrite it? (y/n)")
		var overwrite string
		_, err := fmt.Scanln(&overwrite)
		// ignore unexpected newline error
		if err != nil && err.Error() != "unexpected newline" {
			log.Fatalf("Error reading input: %s", err)
		}

		if overwrite != "y" {
			log.Info("Exiting...")
			return nil
		}
	}

	// Create or open the destination file for writing
	dstFile, err := os.Create(dstFilePathAbs)
	if err != nil {
		log.Fatalf("Error creating destination file: %s", err)
	}
	defer dstFile.Close()

	// Copy the contents from the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Fatalf("Error copying file contents: %s", err)
	}

	log.Info("File copied successfully.")

	return nil
}
