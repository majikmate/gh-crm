package initialize

import (
	"fmt"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/go-gh"
	"github.com/github/gh-classroom/cmd/gh-classroom/shared"
	"github.com/github/gh-classroom/pkg/classroom"
	"github.com/scalarion/gh-crm/pkg/crm"
	"github.com/spf13/cobra"
)

func NewCmdInit(f *cmdutil.Factory) *cobra.Command {
	var cId int

	cmd := &cobra.Command{
		Use:     "init",
		Example: `$ gh crm init`,
		Short:   "Initialize the local repository for GitHub Classroom",
		Long: `Initialize the local repository for GitHub Classroom using an "Accounts*.xlsx" file in the current directory.
		The "Accounts*.xlsx" file should contain the following columns:
		- Name
		- Email
		- GitHub User`,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := gh.RESTClient(nil)
			if err != nil {
				crm.Fatal(fmt.Errorf("failed to create gh client: %v", err))
			}

			if cId == 0 {
				c, err := shared.PromptForClassroom(client)
				if err != nil {
					crm.Fatal(fmt.Errorf("failed get classroom: %v", err))
				}

				cId = c.Id
			}
			cls, err := classroom.GetClassroom(client, cId)
			if err != nil {
				crm.Fatal(fmt.Errorf("failed get classroom: %v", err))
			}

			as, err := crm.ReadAccounts()
			if err != nil {
				crm.Fatal(fmt.Errorf("failed to read accounts: %v", err))
			}

			c := crm.NewClassroom()
			c.SetOrganization(cls.Organization.Id, cls.Organization.Login)
			c.SetClassroom(cls.Id, cls.Name)
			for _, a := range as {
				c.AddStudent(a.Name, a.Email, a.GithubUser)
			}
			err = c.Save(".")
			if err != nil {
				crm.Fatal(fmt.Errorf("failed to save classroom: %v", err))
			}
		},
	}

	cmd.Flags().IntVarP(&cId, "classroom-id", "c", 0, "ID of the classroom")
	return cmd
}
