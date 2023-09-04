package j2g

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/jirax"
)

func newUserMap(gl *gitlab.Client, jiraIssues []*jirax.Issue, users map[string]int) (UserMap, error) {
	jiraUsers, err := GetJiraUsersFromIssues(jiraIssues)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Jira users from issues")
	}

	userMap := make(UserMap)
	for _, jiraUser := range jiraUsers {
		gitlabID, ok := users[jiraUser.AccountID]
		if !ok {
			return nil, errors.New(fmt.Sprintf("No GitLab user found for Jira account ID %s (%s)", jiraUser.AccountID, jiraUser.DisplayName))
		}

		user, _, err := gl.Users.GetUser(gitlabID, gitlab.GetUsersOptions{}) //! 병렬
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error getting GitLab user %d", gitlabID))
		}

		userMap[jiraUser.AccountID] = user
	}

	return userMap, nil
}

func GetJiraUsersFromIssues(issues []*jirax.Issue) ([]*jira.User, error) {
	jiraAccountIds := make(map[string]*jira.User)
	for _, jiraIssue := range issues {
		// TODO: API를 분석해서 User를 판단할 구석을 만들어야 함
		// TODO: Check User Mention
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

	return users, nil
}
