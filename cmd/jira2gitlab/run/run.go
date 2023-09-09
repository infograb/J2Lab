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

package run

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/j2g"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

type Options struct {
	*utils.IOStreams
}

func NewOptions(ioStreams *utils.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

func NewCmdRun(ioStreams *utils.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:   "run [options]",
		Short: "Run the application",
		Long:  "Run the application",
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(o.complete(cmd, args))
			utils.CheckErr(o.validate())
			utils.CheckErr(o.run())
		},
	}

	return cmd
}

func (o *Options) complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *Options) validate() error {
	return nil
}

func (o *Options) run() error {
	cfg, err := config.GetConfig()
	if err != nil {
		return errors.Wrap(err, "Error getting config")
	}

	gl := config.GetGitLabClient(cfg)
	jr := config.GetJiraClient(cfg)
	return j2g.ConvertByProject(gl, jr)
}
