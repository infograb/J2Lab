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
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var gitlabClient *gitlab.Client

func GetGitLabClient(cfg *Config) *gitlab.Client {
	if gitlabClient != nil {
		return gitlabClient
	}

	client, err := gitlab.NewClient(cfg.GitLab.Token, gitlab.WithBaseURL(cfg.GitLab.Host))
	if err != nil {
		log.Fatalf("Error creating GitLab client: %s", err)
	}

	currnetUser, _, err := client.Users.CurrentUser()
	if err != nil {
		log.Fatalf("Error getting current user for GitLab: %s", err)
	}

	log.Infof("GitLab client created for user: %s", currnetUser.Username)

	gitlabClient = client
	return gitlabClient
}
