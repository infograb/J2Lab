package config

import (
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/config"
	"gitlab.com/infograb/team/devops/toy/gos/boilerplate/internal/utils"
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
	// Config 가져오기
	var cfg config.Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		return err
	}

	// Syntax Validation - 단순 형식 검사
	v := validator.New()
	err = v.Struct(&cfg)
	if err != nil {
		log.Fatalf("Error validating config: %s", err)
	}

	// Semantic Validation - 의미 검사
	// TODO: Semantic Validation

	log.Info("Config file is valid")
	return nil
}
