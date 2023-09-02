package j2g

import (
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
)

func newUserMap(gl *gitlab.Client, jiraIssues []jira.Issue) UserMap {
	cfg := config.GetConfig()

	jiraAccountIds := make(map[string]*jira.User)
	for _, jiraIssue := range jiraIssues {
		// TODO: API를 분석해서 User를 판단할 구석을 만들어야 함
		assignee := jiraIssue.Fields.Assignee
		reporter := jiraIssue.Fields.Reporter

		if assignee != nil {
			jiraAccountIds[assignee.AccountID] = assignee
		}
		if reporter != nil {
			jiraAccountIds[reporter.AccountID] = reporter
		}
	}

	userMap := make(UserMap)
	for jiraAccountID, jiraUser := range jiraAccountIds {
		gitlabID, ok := cfg.Users[jiraAccountID]
		if !ok {
			log.Fatalf("No GitLab user found for Jira account ID %s (%s)", jiraAccountID, jiraUser.DisplayName)
		}

		user, _, err := gl.Users.GetUser(gitlabID, gitlab.GetUsersOptions{}) //! 병렬
		if err != nil {
			log.Fatalf("Error getting GitLab user: %s", err)
		}

		userMap[jiraAccountID] = user
	}

	return userMap
}
