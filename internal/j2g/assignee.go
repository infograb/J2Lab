package j2g

import (
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
)

func convertJiraUserToGitLabUser(gl *gitlab.Client, jiraUser *jira.User) (*gitlab.User, error) {
	if jiraUser == nil {
		return nil, nil
	}

	cfg := config.GetConfig()

	jiraUserEmail := jiraUser.EmailAddress
	if jiraUserEmail == "" {
		log.Fatalf("Error getting Jira user email: %s", jiraUser.DisplayName)
	}
	gitlabUserEmail := cfg.Users[jiraUserEmail]
	if gitlabUserEmail == "" {
		log.Fatalf("Error getting GitLab user email from config.yaml: %s", jiraUserEmail)
	}

	users, _, err := gl.Users.ListUsers(&gitlab.ListUsersOptions{
		Username: &gitlabUserEmail,
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if len(users) == 0 {
		log.Fatalf("No User found, gitlab: %s, jira: %s", gitlabUserEmail, jiraUserEmail)
		return nil, err
	}

	return users[0], nil
}
