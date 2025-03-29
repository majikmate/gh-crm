package initialize

import (
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/github/gh-classroom/cmd/gh-classroom/shared"
	"github.com/github/gh-classroom/pkg/classroom"
	"github.com/majikmate/gh-crm/pkg/crm"
	"github.com/spf13/cobra"
)

func NewCmdInit(f *cmdutil.Factory) *cobra.Command {
	var cId int

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initializes the local repository for GitHub Classroom",
		Long: heredoc.Doc(`
		
			Initializes the local repository for GitHub Classroom using a list of accounts.

			The accounts are read from an Excel file in the current directory that matches 
			the filename pattern [Aa]ccounts*.xlsx. It must contain a header in the first 
			row with following fields:

			- Name         ... Full name of the student
			- Email        ... Email address of the student
			- GitHub User  ... GitHub username of the student

			If the classroom-id is known, it can be passed as an argument. Otherwise, the 
			user will be prompted to select a classroom.`),
		Example: `$ gh crm init`,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := api.DefaultRESTClient()
			if err != nil {
				crm.Fatal(fmt.Errorf("failed to create gh client: %v", err))
			}

			as, err := crm.ReadAccounts()
			if err != nil {
				crm.Fatal(fmt.Errorf("failed to read accounts: %v", err))
			}

			c, err := crm.LoadClassroom()
			if err != nil {
				if errors.Is(err, crm.ClassroomNotFound) {
					c, err := shared.PromptForClassroom(client)
					if err != nil {
						crm.Fatal(fmt.Errorf("failed to get classroom: %v", err))
					}

					cId = c.Id
				} else {
					crm.Fatal(err)
				}
			} else {
				isClassroomFolder, err := crm.IsClassroomFolder()
				if err != nil {
					crm.Fatal(err)
				} else if !isClassroomFolder {
					crm.Fatal(fmt.Errorf("Classroom folder exists in the folder hierarchy above, but the current folder is not a classroom folder. Change to the classroom folder."))
				} else {
					cId = c.Classroom.Id
				}
			}

			cls, err := classroom.GetClassroom(client, cId)
			if err != nil {
				crm.Fatal(fmt.Errorf("failed to get classroom: %v", err))
			}

			c = crm.NewClassroom()
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
