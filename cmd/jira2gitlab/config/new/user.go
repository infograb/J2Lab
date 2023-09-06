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
		Short:   "List the Jira User Account Id to users.csv file",
		Long:    "List the Jira User Account Id to users.csv file",
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

	jr := config.GetJiraClient(cfg.Jira)

	jiraEpics, jiraIssues, err := j2g.GetJiraIssues(jr, cfg.Project.Jira.Name, cfg.Project.Jira.Jql)
	if err != nil {
		return errors.Wrap(err, "Error getting Jira issues")
	}

	userKeys, err := j2g.GetJiraUsersFromIssues(append(jiraEpics, jiraIssues...))
	if err != nil {
		return errors.Wrap(err, "Error getting Jira users")
	}

	//* Check if the file already exists
	fileExists := false
	if _, err := os.Stat("users.csv"); err == nil {
		fileExists = true
	}

	//* Ask for confirmation to overwrite the file if it already exists
	if fileExists {
		fmt.Print("The 'users.csv' file already exists. Do you want to overwrite it? (yes/no): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.ToLower(scanner.Text())
		if answer != "yes" {
			logrus.Debugf("Exiting without overwriting the 'users.csv' file")
			return nil
		}
	}

	file, err := os.Create("users.csv")
	if err != nil {
		return errors.Wrap(err, "Error creating file")
	}

	if _, err = file.WriteString("Jira User Key,Jira Display Name,GitLab User ID\n"); err != nil {
		return errors.Wrap(err, "Error writing to file")
	}

	for _, userKey := range userKeys {

		options := &jirax.UserQueryOptions{
			Key: userKey,
		}

		user, _, err := jirax.GetUser(jr, options)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Error getting user %s", userKey))
		}

		if _, err = file.WriteString(user.Key + "," + user.DisplayName + ",\n"); err != nil {
			return errors.Wrap(err, "Error writing to file")
		}
	}

	return nil
}
