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

package j2g

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/onpremise"
	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
)

type AttachmentMap map[string]*Attachment

type Attachment struct {
	Markdown  string
	Filename  string
	Alt       string
	URL       string
	CreatedAt string
}

func convertJiraAttachmentToMarkdown(gl *gitlab.Client, jr *jira.Client, id interface{}, attachement *jira.Attachment) (*Attachment, error) {
	res, err := jr.Issue.DownloadAttachment(context.Background(), attachement.ID)
	if err != nil {
		return nil, errors.Wrap(err, "Error downloading file")
	}

	fileReader := res.Body
	defer fileReader.Close()

	// Upload image to GitLab and retreive a URL
	gitlabUploadedFile, _, err := gl.Projects.UploadFile(id, fileReader, attachement.Filename, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Error uploading file")
	}

	return &Attachment{
		Markdown:  gitlabUploadedFile.Markdown,
		Filename:  attachement.Filename,
		CreatedAt: attachement.Created,
		Alt:       gitlabUploadedFile.Alt,
		URL:       gitlabUploadedFile.URL,
	}, nil
}
