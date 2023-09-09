package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-playground/validator"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/spf13/viper"
)

// Config is the struct for the config file
// The following features are supported:
// 1. Mapping between config file and struct
// 2. Configuration syntax validation

type GitLab struct {
	Host  string `yaml:"host" validate:"required,url"`
	Token string `yaml:"token" validate:"required"`
}

type Jira struct {
	Host  string `yaml:"host" validate:"required,url"`
	Email string `yaml:"email" validate:"required,email"`
	Token string `yaml:"token" validate:"required"`
}

type Config struct {
	GitLab GitLab `yaml:"gitlab" validate:"required"`
	Jira   Jira   `yaml:"jira" validate:"required"`

	Project struct {
		Jira struct {
			Name        string `yaml:"name" validate:"required"`
			Jql         string `yaml:"jql"`
			CustomField struct {
				StoryPoint    string `yaml:"story_point" mapstructure:"story_point"`
				EpicStartDate string `yaml:"epic_start_date" mapstructure:"epic_start_date"`
				ParentEpic    string `yaml:"parent_epic" mapstructure:"parent_epic"`
			} `yaml:"custom_field" mapstructure:"custom_field"`
		} `yaml:"jira"`
		GitLab struct {
			Issue string `yaml:"issue" validate:"required" mapstructure:"issue"`
			Epic  string `yaml:"epic" validate:"required" mapstructure:"epic"`
		} `yaml:"gitlab"`
	} `yaml:"project"`

	Users map[string]int `yaml:"users" validate:"required" mapstructure:"users"`
}

var cfg *Config

func capitalizeJiraProject(cfg *Config) {
	jiraProjectID := cfg.Project.Jira.Name
	caser := cases.Upper(language.English)
	cfg.Project.Jira.Name = caser.String(jiraProjectID)
}

func GetConfig() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	err := InitConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Error initializing config")
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshalling config")
	}

	envs := os.Environ()
	for _, env := range envs {
		parts := strings.Split(env, "=")
		key := parts[0]
		value := strings.Join(parts[1:], "=")

		//* Put Environment Variables here
		switch key {
		case "GITLAB_TOKEN":
			cfg.GitLab.Token = value
		case "JIRA_TOKEN":
			cfg.Jira.Token = value
		}
	}

	cfg.Users, err = parseUserCSVs()
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing user.csv")
	}

	capitalizeJiraProject(cfg)

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, errors.Wrap(err, "Error validating config")
	}

	return cfg, nil
}

// config file is read by yaml format
// You can add --config option to specify the config file
// If you don't specify the config file, the default config file is used
// - $HOME/.config/jira2gitlab/config.yaml
// - $PWD/config.yaml

func InitConfig() error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Error getting current working directory")
	}

	// Search config in home directory with name ".cobra" (without extension).
	viper.AddConfigPath(pwd)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if configFile := viper.GetString("CONFIG_FILE"); configFile != "" {
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return errors.Wrap(err, "Config file not found: %s\nConfig file must be in the format of conf")
		} else {
			return errors.Wrap(err, "Error reading config file")
		}
	}

	log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	return nil
}

func parseUserCSVs() (map[string]int, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "Error getting home directory")
	}

	var file *os.File

	path := viper.GetString("CONFIG_USER_FILE")
	if path == "" {
		path = filepath.Join(pwd, "user.csv")
		file, err = os.Open(path)
		if err != nil {
			return nil, errors.Wrap(err, "Error opening file")
		}
	} else {
		path, err = filepath.Abs(path)
		if err != nil {
			return nil, errors.Wrap(err, "Error getting absolute path")
		}

		file, err = os.Open(path)
		if err != nil {
			return nil, errors.Wrap(err, "Error opening file")
		}
	}
	defer file.Close()

	users := make(map[string]int)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	scanner.Scan() // skip the first line
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			username := strings.TrimSpace(parts[0]) //* Jira Username
			valueStr := strings.TrimSpace(parts[1]) //* GitLab User ID

			gitlabUserId, err := strconv.Atoi(valueStr)
			if err != nil {
				log.Fatal("Error parsing user ID: user.csv must be in the format of <Jira Account ID>,<Jira Display Name>,<GitLab User ID>")
			}

			users[username] = gitlabUserId
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "Error reading users file")
	}

	return users, nil
}
