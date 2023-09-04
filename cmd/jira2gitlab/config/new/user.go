package new

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/config"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/j2g"
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

	users, err := j2g.GetJiraUsersFromIssues(append(jiraEpics, jiraIssues...))
	if err != nil {
		return errors.Wrap(err, "Error getting Jira users")
	}

	file, err := os.Create("users.csv")
	if err != nil {
		return err
	}

	if _, err = file.WriteString("Jira Account ID,Jira Display Name,GitLab User ID\n"); err != nil {
		return err
	}

	for _, user := range users {
		if _, err = file.WriteString(user.AccountID + "," + user.DisplayName + ",\n"); err != nil {
			return err
		}
	}

	return nil
}
