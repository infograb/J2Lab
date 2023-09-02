package jirax

import (
	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

func Unpaginate[T any](
	jr *jira.Client,
	jiraAPIFunction func(searchOptions *jira.SearchOptions) ([]T, *jira.Response, error),
) ([]T, error) {
	var result []T
	searchOptions := &jira.SearchOptions{
		StartAt:    0,
		MaxResults: 100,
	}

	for {
		items, _, err := jiraAPIFunction(searchOptions)
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		result = append(result, items...)
		searchOptions.StartAt += len(items)
	}

	return result, nil
}
