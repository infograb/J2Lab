package run

import (
	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/j2g"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

type Options struct {
	*utils.IOStreams
}

func NewOptions(ioStreams *utils.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

func NewCmdRun(ioStreams *utils.IOStreams) *cobra.Command {
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
	cfg := config.GetConfig()

	gl := config.GetGitLabClient(cfg.GitLab)
	jr := config.GetJiraClient(cfg.Jira)
	j2g.ConvertByProject(gl, jr)

	return nil
}
