package j2g

import (
	"fmt"

	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/adf"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/jirax"
)

func newUserMap(gl *gitlab.Client, jiraIssues []*jirax.Issue, users map[string]int) (UserMap, error) {
	jiraUserAccountIds, err := GetJiraUsersFromIssues(jiraIssues)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Jira users from issues")
	}

	userMap := make(UserMap)
	for _, jiraUserAccountId := range jiraUserAccountIds {
		gitlabID, ok := users[jiraUserAccountId]
		if !ok {
			return nil, errors.New(fmt.Sprintf("No GitLab user found for Jira account ID %s", jiraUserAccountId))
		}

		user, _, err := gl.Users.GetUser(gitlabID, gitlab.GetUsersOptions{}) //! 병렬
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error getting GitLab user %d", gitlabID))
		}

		userMap[jiraUserAccountId] = user
	}

	return userMap, nil
}

// @Ouput: Jira User List
func GetJiraUsersFromIssues(issues []*jirax.Issue) ([]string, error) {
	userIds := make([]string, 0)
	for _, issue := range issues {
		// TODO: API를 분석해서 User를 판단할 구석을 만들어야 함
		assignee := issue.Fields.Assignee
		reporter := issue.Fields.Reporter

		//* Assignee
		if assignee != nil {
			userIds = append(userIds, assignee.AccountID)
		}

		//* Reporter
		if reporter != nil {
			userIds = append(userIds, reporter.AccountID)
		}

		//* Description
		newUserAccountIds := adf.FindMentionIDs(issue.Fields.Description)
		userIds = append(userIds, newUserAccountIds...)

		//* Comment
		for _, comment := range issue.Fields.Comments.Comments {
			newUserIds := adf.FindMentionIDs(comment.Body)
			userIds = append(userIds, newUserIds...)
		}
	}

	jiraAccountIds := make(map[string]bool)
	for _, userAccountId := range userIds {
		jiraAccountIds[userAccountId] = true
	}

	result := make([]string, len(jiraAccountIds))
	idx := 0
	for userId := range jiraAccountIds {
		result[idx] = userId
		idx++
	}

	return result, nil
}
