package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var gitlabClient *gitlab.Client

func GetGitLabClient(gitlabConfig GitLab) *gitlab.Client {
	if gitlabClient != nil {
		return gitlabClient
	}

	client, err := gitlab.NewClient(gitlabConfig.Token, gitlab.WithBaseURL(gitlabConfig.Host))
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
