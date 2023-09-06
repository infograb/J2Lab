package j2g

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
)

func newUserMap(gl *gitlab.Client, jiraIssues []*jira.Issue, users map[string]int) (UserMap, error) {
	jiraUserKeys, err := GetJiraUsersFromIssues(jiraIssues)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Jira users from issues")
	}

	userMap := make(UserMap)
	for _, jiraUserKey := range jiraUserKeys {
		gitlabID, ok := users[jiraUserKey]
		if !ok {
			return nil, errors.New(fmt.Sprintf("No GitLab user found for Jira account ID %s", jiraUserKey))
		}

		user, _, err := gl.Users.GetUser(gitlabID, gitlab.GetUsersOptions{}) //! 병렬
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error getting GitLab user %d", gitlabID))
		}

		userMap[jiraUserKey] = user
	}

	return userMap, nil
}

// @Ouput: Jira User List
func GetJiraUsersFromIssues(issues []*jira.Issue) ([]string, error) {
	userKeyArray := make([]string, 0)
	for _, issue := range issues {
		// TODO: API를 분석해서 User를 판단할 구석을 만들어야 함
		assignee := issue.Fields.Assignee
		reporter := issue.Fields.Reporter

		//* Assignee
		if assignee != nil {
			userKeyArray = append(userKeyArray, assignee.Key)
		}

		//* Reporter
		if reporter != nil {
			userKeyArray = append(userKeyArray, reporter.Key)
		}

		//* Description
		// TODO
		// userIds = append(userIds, newUserAccountIds...)

		//* Comment
		// TODO
		// for _, comment := range issue.Fields.Comments.Comments {
		// 	newUserIds := adf.FindMentionIDs(comment.Body)
		// 	userIds = append(userIds, newUserIds...)
		// }
	}

	userKeyMap := make(map[string]bool)
	for _, key := range userKeyArray {
		userKeyMap[key] = true
	}

	result := make([]string, len(userKeyMap))
	idx := 0
	for userId := range userKeyMap {
		result[idx] = userId
		idx++
	}

	return result, nil
}
