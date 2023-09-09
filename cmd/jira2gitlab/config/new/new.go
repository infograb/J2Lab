package new

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

func NewCmdNew(ioStreams *utils.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "new",
		Short:   "Create the config.yaml and user.csv file",
		Long:    "Create the config.yaml and user.csv file on the current working directory",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(runConfigNew(ioStreams))
		},
	}

	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("config")
		command.Parent().HelpFunc()(command, strings)
	})

	cmd.AddCommand(
		NewCmdNewUser(ioStreams),
	)

	return cmd
}

func runConfigNew(io *utils.IOStreams) error {
	pwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "error getting current working directory")
	}

	srcPath, _ := filepath.Abs("./internal/config/example/config.yaml")
	destPath := filepath.Join(pwd, "config.yaml")
	err = utils.CopyFile(io.Out, srcPath, destPath)
	if err != nil {
		return errors.Wrap(err, "error copying file")
	}

	srcPath, _ = filepath.Abs("./internal/config/example/user.csv")
	destPath = filepath.Join(pwd, "user.csv")
	err = utils.CopyFile(io.Out, srcPath, destPath)
	if err != nil {
		return errors.Wrap(err, "error copying file")
	}

	return nil
}
