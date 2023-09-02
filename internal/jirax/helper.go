package jirax

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

func UnpaginateIssue(
	jr *jira.Client,
	jql string,
) ([]*Issue, error) {

	issueService := IssueService{client: jr}

	var result []*Issue

	searchOptions := &jira.SearchOptions{
		StartAt:    0,
		MaxResults: 100,
		Fields:     []string{"*all"},
	}

	for {
		items, r, err := issueService.Search(context.Background(), jql, searchOptions)
		if err != nil {
			return nil, err
		}

		for _, item := range items {
			result = append(result, &item)
		}

		if err != nil {
			return nil, err
		}

		if r.StartAt+r.MaxResults >= r.Total {
			break
		}

		searchOptions.StartAt += len(items)
	}

	return result, nil
}
