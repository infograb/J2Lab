/*
 * This file is part of the InfoGrab project.
 *
 * Copyright (C) 2023 InfoGrab
 *
 * This program is free software: you can redistribute it and/or modify it
 * it is available under the terms of the GNU Lesser General Public License
 * by the Free Software Foundation, either version 3 of the License or by the Free Software Foundation
 * (at your option) any later version.
 */

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
		Short:   "Create the config.yaml and users.csv file",
		Long:    "Create the config.yaml and users.csv file on the current working directory",
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

	srcPath, _ = filepath.Abs("./internal/config/example/users.csv")
	destPath = filepath.Join(pwd, "users.csv")
	err = utils.CopyFile(io.Out, srcPath, destPath)
	if err != nil {
		return errors.Wrap(err, "error copying file")
	}

	return nil
}
