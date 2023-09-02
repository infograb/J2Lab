package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	configCmd "gitlab.com/infograb/team/devops/toy/gos/boilerplate/cmd/jira2gitlab/config"
	runCmd "gitlab.com/infograb/team/devops/toy/gos/boilerplate/cmd/jira2gitlab/run"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/cmd/jira2gitlab/version"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
)

var (
	// Used for flags.

	rootCmd = &cobra.Command{
		Use:   "jira2gitlab",
		Short: "The Jira miration tool for Gitlab",
		Long:  "This command is the Jira miration tool for Gitlab",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

func init() {
	ioStreams := utils.NewStdIOStreams()
	log.SetOutput(ioStreams.ErrOut)
	log.SetLevel(log.DebugLevel) // TODO Set log level from flag

	io := utils.NewStdIOStreams()

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	rootCmd.AddCommand(
		version.NewCmdVersion(io),
		runCmd.NewCmdRun(io),
		configCmd.NewCmdConfig(io),
	)
}

func Execute() error {
	return rootCmd.Execute()
}
