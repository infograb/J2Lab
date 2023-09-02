package jirax

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

type JiraIssue struct {
	*jira.Issue
	*jira.IssueFields
}

func UnpaginateIssue(
	jr *jira.Client,
	jql string,
) ([]*jira.Issue, *jira.Response, error) {
	var result []*jira.Issue
	searchOptions := &jira.SearchOptions{
		StartAt:    0,
		MaxResults: 100,
		Fields:     []string{"*all"},
	}

	var res *jira.Response
	for {
		items, r, err := jr.Issue.Search(context.Background(), jql, searchOptions)
		if err != nil {
			return nil, nil, err
		}

		for _, item := range items {
			result = append(result, &item)
		}

		searchOptions.StartAt += len(items)

		if r.StartAt+r.MaxResults >= r.Total {
			break
		}

		res = r
	}

	return result, res, nil
}
