package j2g

import (
	log "github.com/sirupsen/logrus"
	gitlab "github.com/xanzy/go-gitlab"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

func createIssueNoteFromFile(gl *gitlab.Client, pid interface{}, gitlabIssue *gitlab.Issue, url string, fileName string) {
	fileReader, err := utils.DownloadFile(url)
	if err != nil {
		log.Fatalf("Error downloading file: %s", err)
	}

	// Upload image to GitLab and retreive a URL
	gitlabUploadedFile, _, err := gl.Projects.UploadFile(pid, fileReader, fileName, nil)
	if err != nil {
		log.Fatalf("Error uploading file: %s", err)
	}

	gl.Notes.CreateIssueNote(pid, gitlabIssue.IID, &gitlab.CreateIssueNoteOptions{
		Body: &gitlabUploadedFile.Markdown,
	})
}
