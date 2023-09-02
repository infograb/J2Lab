package new

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/j2g"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
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
		log.Fatalf("Error unmarshalling config: %s", err)
	}

	jr := config.GetJiraClient(cfg.Jira)
	jiraEpics, jiraIssues := j2g.GetJiraIssues(jr, cfg.Project.Jira.Name, cfg.Project.Jira.Jql)
	users := j2g.GetJiraUsersFromIssues(append(jiraEpics, jiraIssues...))

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
