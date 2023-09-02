package config

import (
	"github.com/spf13/cobra"
	newCmd "gitlab.com/infograb/team/devops/toy/j2lab/cmd/jira2gitlab/config/new"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

func NewCmdConfig(ioStreams *utils.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config SUBCOMMAND [options]",
		Short: "Modify config files",
		Long:  "Modify config files",
	}

	cmd.AddCommand(
		newCmd.NewCmdNew(ioStreams),
		newCmdConfigLint(ioStreams),
	)

	return cmd
}
