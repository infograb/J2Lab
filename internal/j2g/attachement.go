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
	Name      string
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
		Name:      attachement.Filename,
		CreatedAt: attachement.Created,
	}, nil
}
