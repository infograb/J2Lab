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
	"net/http"
	"time"

	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
)

type CreateEpicOptions struct {
	Title            *string         `url:"title,omitempty" json:"title,omitempty"`
	Description      *string         `url:"description,omitempty" json:"description,omitempty"`
	Labels           *gitlab.Labels  `url:"labels,comma,omitempty" json:"labels,omitempty"`
	StartDateIsFixed *bool           `url:"start_date_is_fixed,omitempty" json:"start_date_is_fixed,omitempty"`
	StartDateFixed   *gitlab.ISOTime `url:"start_date_fixed,omitempty" json:"start_date_fixed,omitempty"`
	DueDateIsFixed   *bool           `url:"due_date_is_fixed,omitempty" json:"due_date_is_fixed,omitempty"`
	DueDateFixed     *gitlab.ISOTime `url:"due_date_fixed,omitempty" json:"due_date_fixed,omitempty"`

	//* 라이브러리에서 지원하지 않는 추가 옵션
	Color        *string    `url:"color,omitempty" json:"color,omitempty"`
	Confidential *bool      `url:"confidential,omitempty" json:"confidential,omitempty"`
	CreatedAt    *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
	// ParentID ...
}

func CreateEpic(gl *gitlab.Client, gid interface{}, opt *CreateEpicOptions) (*gitlab.Epic, *gitlab.Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error parsing ID")
	}
	u := fmt.Sprintf("groups/%s/epics", gitlab.PathEscape(group))

	req, err := gl.NewRequest(http.MethodPost, u, opt, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error creating request")
	}

	e := new(gitlab.Epic)
	resp, err := gl.Do(req, e)
	if err != nil {
		return nil, resp, errors.Wrap(err, "Error making request")
	}

	return e, resp, nil
}
