package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

var (
	configFile string
)

func newCmdConfigNew(ioStreams utils.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "new",
		Short:   "Create the config.yml file",
		Long:    "Create the config.yml file on the current working directory",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(runConfigNew(o, cmd, args))
		},
	}

	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("config")
		command.Parent().HelpFunc()(command, strings)
	})

	return cmd
}

func runConfigNew(o *Options, cmd *cobra.Command, args []string) error {
	srcFilePathAbs, err := filepath.Abs("./internal/config/sample.yaml")
	if err != nil {
		log.Fatalf("Error getting absolute path of source file: %s", err)
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %s", err)
		return err
	}

	var dstFilePathAbs string
	if configFile != "" {
		dstFilePathAbs, err = filepath.Abs(configFile)
		if err != nil {
			log.Fatalf("Error getting absolute path of destination file: %s", err)
		}
	} else {
		dstFilePathAbs, err = filepath.Abs(filepath.Join(pwd, "config.yml"))
		if err != nil {
			log.Fatalf("Error getting absolute path of destination file: %s", err)
		}
	}

	// Check if the source file exists
	if _, err := os.Stat(srcFilePathAbs); os.IsNotExist(err) {
		log.Fatalf("Source file does not exist: %s", srcFilePathAbs)
	}

	// Open the source file for reading
	srcFile, err := os.Open(srcFilePathAbs)
	if err != nil {
		log.Fatalf("Error opening source file: %s", err)
	}
	defer srcFile.Close()

	// Ask to overwrite the destination file if it already exists
	if _, err := os.Stat(dstFilePathAbs); err == nil {
		log.Warnf("Destination file already exists: %s", dstFilePathAbs)
		log.Warn("Do you want to overwrite it? (y/n)")
		var overwrite string
		_, err := fmt.Scanln(&overwrite)
		// ignore unexpected newline error
		if err != nil && err.Error() != "unexpected newline" {
			log.Fatalf("Error reading input: %s", err)
		}

		if overwrite != "y" {
			log.Info("Exiting...")
			return nil
		}
	}

	// Create or open the destination file for writing
	dstFile, err := os.Create(dstFilePathAbs)
	if err != nil {
		log.Fatalf("Error creating destination file: %s", err)
	}
	defer dstFile.Close()

	// Copy the contents from the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Fatalf("Error copying file contents: %s", err)
	}

	log.Info("File copied successfully.")

	return nil
}
