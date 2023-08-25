package j2g

import (
	"testing"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	gitlab "github.com/xanzy/go-gitlab"
)

type args struct {
	jiraComment *jira.Comment
}

var tests = []struct {
	name string
	args args
	want *gitlab.CreateIssueNoteOptions
}{
	{
		name: "Empty Jira Comment",
		args: args{
			jiraComment: &jira.Comment{},
		},
		want: &gitlab.CreateIssueNoteOptions{
			Body: nil,
		},
	},
	{
		name: "Empty Jira Comment",
		args: args{
			jiraComment: &jira.Comment{},
		},
		want: &gitlab.CreateIssueNoteOptions{
			Body: nil,
		},
	},
}

func TestConvertToGitLabComment(t *testing.T) {

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToGitLabComment(tt.args.jiraComment); got == nil {
				t.Errorf("convertToGitLabComment() = %v, want %v", got, tt.want)
			}
		})
	}
}
