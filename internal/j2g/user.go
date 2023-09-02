package j2g

import (
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func newUserMap(gl *gitlab.Client, jiraIssues []*jira.Issue, users map[string]int) UserMap {
	jiraUsers := GetJiraUsersFromIssues(jiraIssues)

	userMap := make(UserMap)
	for _, jiraUser := range jiraUsers {
		gitlabID, ok := users[jiraUser.AccountID]
		if !ok {
			log.Fatalf("No GitLab user found for Jira account ID %s (%s)", jiraUser.AccountID, jiraUser.DisplayName)
		}

		user, _, err := gl.Users.GetUser(gitlabID, gitlab.GetUsersOptions{}) //! 병렬
		if err != nil {
			log.Fatalf("Error getting GitLab user: %s", err)
		}

		userMap[jiraUser.AccountID] = user
	}

	return userMap
}

func GetJiraUsersFromIssues(issues []*jira.Issue) []*jira.User {
	jiraAccountIds := make(map[string]*jira.User)
	for _, jiraIssue := range issues {
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

	users := make([]*jira.User, 0)
	for _, user := range jiraAccountIds {
		users = append(users, user)
	}

	return users
}
