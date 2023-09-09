/*
 * This file is part of the InfoGrab project.
 *
 * Copyright (C) 2023 InfoGrab
 *
 * This program is free software: you can redistribute it and/or modify it
 * it is available under the terms of the GNU Lesser General Public License
 * by the Free Software Foundation, either version 3 of the License or by the Free Software Foundation
 * (at your option) any later version.
 */

package jirax

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
)

func UnpaginateIssue(
	jr *jira.Client,
	jql string,
) ([]*jira.Issue, error) {

	var result []*jira.Issue

	searchOptions := &jira.SearchOptions{
		StartAt:    0,
		MaxResults: 100,
		Fields:     []string{"*all"},
	}

	for {
		itemsV2, r, err := jr.Issue.Search(context.Background(), jql, searchOptions)
		if err != nil {
			return nil, errors.Wrap(err, "Error getting Jira issues V2")
		}

		//* Mapping Media
		for _, itemV2 := range itemsV2 {
			item := itemV2
			result = append(result, &item)
		}

		if err != nil {
			return nil, errors.Wrap(err, "Error getting Jira issues")
		}

		if r.StartAt+r.MaxResults >= r.Total {
			break
		}

		searchOptions.StartAt += len(itemsV2)
	}

	return result, nil
}
