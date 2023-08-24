package config

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
)

var jiraClient *jira.Client

func GetJiraClient() *jira.Client {
	if jiraClient != nil {
		return jiraClient
	}

	cfg := GetConfig()

	tp := jira.BasicAuthTransport{
		Username: cfg.Jira.Email,
		APIToken: cfg.Jira.Token,
	}

	client, err := jira.NewClient(cfg.Jira.Host, tp.Client())
	if err != nil {
		log.Fatalf("Error creating Jira client: %s", err)
	}

	currnetUser, _, err := client.User.GetCurrentUser(context.Background())
	if err != nil {
		// log.Fatalf("Error getting current user for Jira: %s", err)
		log.Fatalf("Error getting current user for Jira")
	}

	log.Infof("Jira client created for user: %s", currnetUser.EmailAddress)

	jiraClient = client
	return jiraClient
}
