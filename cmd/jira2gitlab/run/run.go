package run

import (
	jira "github.com/andygrunwald/go-jira/v2/cloud"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/j2g"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

type Options struct {
	utils.IOStreams
}

func NewOptions(ioStreams utils.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

func NewCmdRun(ioStreams utils.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:   "run [options]",
		Short: "Run the application",
		Long:  "Run the application",
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(o.complete(cmd, args))
			utils.CheckErr(o.validate())
			utils.CheckErr(o.run())
		},
	}

	return cmd
}

func (o *Options) complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *Options) validate() error {
	return nil
}

func (o *Options) run() error {
	jiraIssue := &jira.Issue{
		Key: "TEST-1",
		Fields: &jira.IssueFields{
			Summary:     "Test issue",
			Description: "Test issue description",
		},
	}

	gitlabIssue := j2g.ConvertJiraIssueToGitLabIssue(jiraIssue)
	log.Debugf("GitLab issue: %+v", gitlabIssue)
	return nil
}
