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

	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
)

type CreateEpicLinkOptions struct {
	TargetGroupID *string `json:"target_group_id"`
	TargetEpicIID *string `json:"target_epic_iid"`
	LinkType      *string `json:"link_type"`
}

type EpicLink struct {
	SourceEpic *gitlab.Epic `json:"source_epic"`
	TargetEpic *gitlab.Epic `json:"target_epic"`
	LinkType   string       `json:"link_type"`
}

func CreateEpicLink(gl *gitlab.Client, gid interface{}, epic int, opt *CreateEpicLinkOptions) (*EpicLink, *gitlab.Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error parsing ID")
	}
	u := fmt.Sprintf("groups/%s/epics/%d/related_epics", gitlab.PathEscape(group), epic)
	req, err := gl.NewRequest(http.MethodPost, u, opt, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error creating request")
	}

	i := new(EpicLink)
	resp, err := gl.Do(req, &i)
	if err != nil {
		return nil, resp, errors.Wrap(err, "Error making request")
	}

	return i, resp, nil
}
