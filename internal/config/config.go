package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

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
	} `yaml:"gitlab" validate:"required"`

	Jira struct {
		Host  string `yaml:"host" validate:"required,url"`
		Email string `yaml:"email" validate:"required,email"`
		Token string `yaml:"token"`
	} `yaml:"jira"`

	Project struct {
		Jira struct {
			Name        string `yaml:"name" validate:"required"`
			Jql         string `yaml:"jql"`
			CustomField struct {
				StoryPoint    string `yaml:"story_point" mapstructure:"story_point"`
				EpicStartDate string `yaml:"epic_start_date" mapstructure:"epic_start_date"`
			} `yaml:"custom_field" mapstructure:"custom_field"`
		} `yaml:"jira"`
		GitLab struct {
			Issue string `yaml:"issue" validate:"required"`
			Epic  string `yaml:"epic"`
		} `yaml:"gitlab"`
	} `yaml:"project"`

	Users map[string]int `yaml:"users"`
}

var cfg *Config

func capitalizeJiraProject(cfg *Config) {
	jiraProjectID := cfg.Project.Jira.Name
	caser := cases.Upper(language.English)
	cfg.Project.Jira.Name = caser.String(jiraProjectID)

}

func GetConfig() *Config {
	if cfg != nil {
		return cfg
	}

	InitConfig()

	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %s", err)
	}

	cfg.Users = parseUsers()
	capitalizeJiraProject(cfg)

	return cfg
}

// config file is read by yaml format
// You can add --config option to specify the config file
// If you don't specify the config file, the default config file is used
// - $HOME/.config/jira2gitlab/config.yaml
// - $PWD/config.yaml

func InitConfig() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %s", err)
	}

	// Search config in home directory with name ".cobra" (without extension).
	viper.AddConfigPath(pwd)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if configFile := viper.GetString("config"); configFile != "" {
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("Config file not found: %s\nConfig file must be in the format of config.yaml", err)
		} else {
			log.Fatalf("Error reading config file: %s", err)
		}
	}

	log.Debugf("Using config file: %s", viper.ConfigFileUsed())
}

func parseUsers() map[string]int {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting home directory: %s", err)
	}

	file, err := os.Open(filepath.Join(pwd, "users.csv"))
	if err != nil {
		log.Fatalf("Error opening users file: %s", err)
	}
	defer file.Close()

	userMap := make(map[string]int)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	scanner.Scan() // skip the first line
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			valueStr := strings.TrimSpace(parts[1])

			value, err := strconv.Atoi(valueStr)
			if err != nil {
				log.Fatal("Error parsing user ID: users.csv must be in the format of <Jira Account ID>,<GitLab User ID>")
			}

			userMap[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading users file: %s", err)
	}

	return userMap
}
