package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

func NewCmdNew(ioStreams *utils.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "new",
		Short:   "Create the config.yml file",
		Long:    "Create the config.yml file on the current working directory",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(runConfigNew(ioStreams))
		},
	}

	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("config")
		command.Parent().HelpFunc()(command, strings)
	})

	return cmd
}

func runConfigNew(io *utils.IOStreams) error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %s", err)
	}

	srcPath, _ := filepath.Abs("./internal/config/example/config.yaml")
	destPath := filepath.Join(pwd, "config.yaml")
	err = utils.CopyFile(io.Out, srcPath, destPath)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	srcPath, _ = filepath.Abs("./internal/config/example/users.csv")
	destPath = filepath.Join(pwd, "users.csv")
	err = utils.CopyFile(io.Out, srcPath, destPath)
	if err != nil {
		return fmt.Errorf("error copying file: %s", err)
	}

	return nil
}
