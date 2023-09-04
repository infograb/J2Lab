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
		// Add DescriptionPlain and BodyPlain to itemV3
		itemsV2, _, err := jr.Issue.Search(context.Background(), jql, searchOptions)
		if err != nil {
			return nil, err
		}

		itemsV3, r, err := issueService.Search(context.Background(), jql, searchOptions)
		if err != nil {
			return nil, err
		}

		for issueIdx, issue := range itemsV3 {
			issue.Fields.DescriptionPlain = itemsV2[issueIdx].Fields.Description
			for commentIdx := range issue.Fields.Comments.Comments {
				issue.Fields.Comments.Comments[commentIdx].BodyPlain = itemsV2[issueIdx].Fields.Comments.Comments[commentIdx].Body
			}

			result = append(result, &issue)
		}

		if err != nil {
			return nil, err
		}

		if r.StartAt+r.MaxResults >= r.Total {
			break
		}

		searchOptions.StartAt += len(itemsV3)
	}

	return result, nil
}
