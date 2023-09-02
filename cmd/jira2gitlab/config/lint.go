package config

import (
	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

func newCmdConfigLint(ioStreams *utils.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lint",
		Short:   "Lint the config.yml file",
		Long:    "Lint the config.yml file",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(runConfigLint(ioStreams))
		},
	}

	return cmd
}

func runConfigLint(ioStreams *utils.IOStreams) error {
	_ = config.GetConfig()

	return nil
}
