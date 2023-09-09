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

package gitlabx

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

func parseID(id interface{}) (string, error) {
	switch v := id.(type) {
	case int:
		return strconv.Itoa(v), nil
	case string:
		return v, nil
	default:
		return "", errors.New(fmt.Sprintf("invalid ID type %#v, the ID must be an int or a string", id))
	}
}

func Unpaginate[T any](
	gl *gitlab.Client,
	gitlabAPIFunction func(opt *gitlab.ListOptions) ([]*T, *gitlab.Response, error),
) ([]*T, error) {
	var result []*T
	page := 1
	perPage := 100

	for {
		opt := &gitlab.ListOptions{
			Page:    page,
			PerPage: perPage,
		}

		items, resp, err := gitlabAPIFunction(opt)
		if err != nil {
			return nil, errors.Wrap(err, "Error making request")
		}

		result = append(result, items...)

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		page = resp.NextPage
	}

	return result, nil
}
