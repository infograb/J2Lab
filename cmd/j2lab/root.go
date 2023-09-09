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

package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	configCmd "gitlab.com/infograb-public/j2lab/cmd/j2lab/config"
	runCmd "gitlab.com/infograb-public/j2lab/cmd/j2lab/run"
	"gitlab.com/infograb-public/j2lab/cmd/j2lab/version"
	"gitlab.com/infograb-public/j2lab/internal/utils"
)

var (
	// Used for flags.

	rootCmd = &cobra.Command{
		Use:   "j2lab",
		Short: "The Jira miration tool for Gitlab",
		Long:  "This command is the Jira miration tool for Gitlab",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func init() {

	rootCmd.PersistentFlags().StringP("config", "c", "", "config.yaml file")
	rootCmd.PersistentFlags().StringP("user", "u", "", "user.csv file")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")
	viper.BindPFlag("CONFIG_FILE", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("USER_FILE", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("DEBUG", rootCmd.PersistentFlags().Lookup("debug"))

	ioStreams := utils.NewStdIOStreams()
	log.SetOutput(ioStreams.ErrOut)
	log.SetLevel(log.DebugLevel) // TODO Set log level from flag

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		debug := viper.GetBool("DEBUG")
		if debug {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	}

	io := utils.NewStdIOStreams()
	rootCmd.AddCommand(
		version.NewCmdVersion(io),
		runCmd.NewCmdRun(io),
		configCmd.NewCmdConfig(io),
	)
}

func Execute() error {
	return rootCmd.Execute()
}
