package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/infograb/team/devops/toy/j2lab/internal/utils"
)

const (
	// Version is the current version of the CLI
	Version = "0.0.1"
)

type Options struct {
	*utils.IOStreams
}

func NewOptions(ioStreams *utils.IOStreams) *Options {
	return &Options{
		IOStreams: ioStreams,
	}
}

// NewCmdVersion returns a cobra command for fetching versions
func NewCmdVersion(ioStreams *utils.IOStreams) *cobra.Command {
	o := NewOptions(ioStreams)
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the client and server version information",
		Long:    "Print the client and server version information for the current context.",
		Example: "",
		Run: func(cmd *cobra.Command, args []string) {
			utils.CheckErr(o.complete(cmd, args))
			utils.CheckErr(o.validate())
			utils.CheckErr(o.run())
		},
	}

	// cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "One of 'yaml' or 'json'.")
	return cmd
}

func (o *Options) complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *Options) validate() error {
	return nil
}

func (o *Options) run() error {
	fmt.Fprintf(o.Out, "%s\n", Version)
	return nil
}
