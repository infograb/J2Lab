package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	configCmd "gitlab.com/infograb/team/devops/toy/j2lab/cmd/jira2gitlab/config"
	runCmd "gitlab.com/infograb/team/devops/toy/j2lab/cmd/jira2gitlab/run"
	"gitlab.com/infograb/team/devops/toy/j2lab/cmd/jira2gitlab/version"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
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

	rootCmd.PersistentFlags().StringP("config", "c", "", "config.yaml file")
	rootCmd.PersistentFlags().StringP("user", "u", "", "user.csv file")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")
	viper.BindPFlag("CONFIG_FILE", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("CONFIG_USER_FILE", rootCmd.PersistentFlags().Lookup("user"))
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
