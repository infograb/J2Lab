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
