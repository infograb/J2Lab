package j2g

import (
	"context"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
)

func convertJiraAttachementToMarkdown(gl *gitlab.Client, jr *jira.Client, id interface{}, attachement *jira.Attachment) string {
	res, err := jr.Issue.DownloadAttachment(context.Background(), attachement.ID)
	if err != nil {
		log.Fatalf("Error downloading file: %s", err)
	}

	fileReader := res.Body
	defer fileReader.Close()

	// Upload image to GitLab and retreive a URL
	gitlabUploadedFile, _, err := gl.Projects.UploadFile(id, fileReader, attachement.Filename, nil)
	if err != nil {
		log.Fatalf("Error uploading file: %s", err)
	}

	return gitlabUploadedFile.Markdown
}
