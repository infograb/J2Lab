package config

import (
	"github.com/spf13/cobra"
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

func NewCmdConfig(ioStreams utils.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config SUBCOMMAND [options]",
		Short: "Modify config files",
		Long:  "Modify config files",
	}

	cmd.AddCommand(
		newCmdConfigNew(ioStreams),
		newCmdConfigLint(ioStreams),
	)

	return cmd
}
