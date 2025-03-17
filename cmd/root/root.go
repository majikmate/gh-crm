package root

import (
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/scalarion/gh-crm/cmd/clone"
	"github.com/scalarion/gh-crm/cmd/initialize"
	"github.com/spf13/cobra"
)

func NewRootCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crm <command>",
		Short: "A GitHub Classroom CLI",
	}

	cmd.AddCommand(initialize.NewCmdInit(f))
	// cmd.AddCommand(teams.NewCmdTeams(f))
	cmd.AddCommand(clone.NewCmdClone(f))

	return cmd
}
