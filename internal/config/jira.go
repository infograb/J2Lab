/*
 * This file is part of the InfoGrab project.
 *
 * Copyright (C) 2023 InfoGrab
 *
 * This program is free software: you can redistribute it and/or modify it
 * it is available under the terms of the GNU Lesser General Public License
 * by the Free Software Foundation, either version 3 of the License or by the Free Software Foundation
 * (at your option) any later version.
 */

package config

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	log "github.com/sirupsen/logrus"
)

var jiraClient *jira.Client

func GetJiraClient(jiraConfig Jira) *jira.Client {
	if jiraClient != nil {
		return jiraClient
	}

	tp := jira.BearerAuthTransport{
		Token: jiraConfig.Token,
	}

	client, err := jira.NewClient(jiraConfig.Host, tp.Client())
	if err != nil {
		log.Fatalf("Error creating Jira client: %s", err)
	}

	currnetUser, _, err := client.User.GetSelf(context.Background())
	if err != nil {
		// log.Fatalf("Error getting current user for Jira: %s", err)
		log.Fatalf("Error getting current user for Jira")
	}

	log.Infof("Jira client created for user: %s", currnetUser.EmailAddress)

	jiraClient = client
	return jiraClient
}
