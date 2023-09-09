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
	_, err := config.GetConfig()
	return err
}
