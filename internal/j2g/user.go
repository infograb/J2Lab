package j2g

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
)

// TODO
// Jira Username -> GitLab ID
type UserMap map[string]*gitlab.User

func newUserMap(gl *gitlab.Client, jiraIssues []*jira.Issue, users map[string]int) (UserMap, error) {
	jiraUsernames, err := GetJiraUsernamesFromIssues(jiraIssues)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting Jira users from issues")
	}

	userMap := make(UserMap)
	for _, jiraUsername := range jiraUsernames {
		gitlabID, ok := users[jiraUsername]
		if !ok {
			return nil, errors.New(fmt.Sprintf("No GitLab user found for Jira account ID %s", jiraUsername))
		}

		gitlabUser, _, err := gl.Users.GetUser(gitlabID, gitlab.GetUsersOptions{}) // TODO 병렬
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Error getting GitLab user %d", gitlabID))
		}

		userMap[jiraUsername] = gitlabUser
	}

	return userMap, nil
}

// @Ouput: Jira User List
func GetJiraUsernamesFromIssues(issues []*jira.Issue) ([]string, error) {
	usernameArray := make([]string, 0)
	for _, issue := range issues {
		// TODO: API를 분석해서 User를 판단할 구석을 만들어야 함
		assignee := issue.Fields.Assignee
		reporter := issue.Fields.Reporter

		//* Assignee
		if assignee != nil {
			usernameArray = append(usernameArray, assignee.Name)
		}

		//* Reporter
		if reporter != nil {
			usernameArray = append(usernameArray, reporter.Name)
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

	usernameMap := make(map[string]bool)
	for _, username := range usernameArray {
		usernameMap[username] = true
	}

	result := make([]string, len(usernameMap))
	idx := 0
	for userId := range usernameMap {
		result[idx] = userId
		idx++
	}

	return result, nil
}
