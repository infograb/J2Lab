package config

import (
	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

func newCmdConfigNew(ioStreams utils.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "new",
		Short:   "Create the config.yml file",
		Long:    "Create the config.yml file on the current working directory",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("config")
		command.Parent().HelpFunc()(command, strings)
	})

	return cmd
}
