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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/j2g"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/jirax"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

func NewCmdNewUser(ioStreams *utils.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Short:   "List the Jira User Account Id to user.csv file",
		Long:    "List the Jira User Account Id to user.csv file",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(runConfigNewUser(ioStreams))
		},
	}

	cmd.SetHelpFunc(func(command *cobra.Command, strings []string) {
		command.Flags().MarkHidden("config")
		command.Parent().HelpFunc()(command, strings)
	})

	return cmd
}

func runConfigNewUser(io *utils.IOStreams) error {
	var cfg *config.Config
	config.InitConfig()
	err := viper.Unmarshal(&cfg)
	if err != nil {
		errors.Wrap(err, "Error unmarshalling config")
	}

	jr := config.GetJiraClient(cfg)

	jiraEpics, jiraIssues, err := j2g.GetJiraIssues(jr, cfg.Jira.Name, cfg.Jira.Jql)
	if err != nil {
		return errors.Wrap(err, "Error getting Jira issues")
	}

	usernames, err := j2g.GetJiraUsernamesFromIssues(append(jiraEpics, jiraIssues...))
	if err != nil {
		return errors.Wrap(err, "Error getting Jira users")
	}

	//* Check if the file already exists
	fileExists := false
	if _, err := os.Stat("user.csv"); err == nil {
		fileExists = true
	}

	//* Ask for confirmation to overwrite the file if it already exists
	if fileExists {
		fmt.Print("The 'user.csv' file already exists. Do you want to overwrite it? (y/n): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.ToLower(scanner.Text())
		if answer != "y" {
			logrus.Debugf("Exiting without overwriting the 'user.csv' file")
			return nil
		}
	}

	file, err := os.Create("user.csv")
	if err != nil {
		return errors.Wrap(err, "Error creating file")
	}

	if _, err = file.WriteString("Jira User Name,GitLab User ID\n"); err != nil {
		return errors.Wrap(err, "Error writing to file")
	}

	for _, username := range usernames {
		options := &jirax.UserQueryOptions{Username: username}

		user, _, err := jirax.GetUser(jr, options)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error getting user %s", username))
		}

		if _, err = file.WriteString(user.Name + ",\n"); err != nil {
			return errors.Wrap(err, "Error writing to file")
		}
	}

	return nil
}
