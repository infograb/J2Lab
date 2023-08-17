package j2g

import (
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	gitlab "github.com/xanzy/go-gitlab"
)

func ConvertJiraIssueToGitLabIssue(jiraIssue *jira.Issue) *gitlab.Issue {
	gitlabIssue := &gitlab.Issue{
		Title:       jiraIssue.Fields.Summary,
		Description: jiraIssue.Fields.Description,
	}

	return gitlabIssue
}
