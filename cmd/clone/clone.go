package clone

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/github/gh-classroom/cmd/gh-classroom/shared"
	"github.com/github/gh-classroom/pkg/classroom"
	"github.com/majikmate/gh-crm/cmd/clone/utils"
	"github.com/majikmate/gh-crm/pkg/crm"
	"github.com/spf13/cobra"
)

func NewCmdClone(f *cmdutil.Factory) *cobra.Command {
	var aId int
	var starterFolder string
	var isAssignmentFolder bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clones starter repo and student repos for an assignment",
		Long: heredoc.Doc(`
		
			Clones the starter repo and the student repos for an assignment.

			The repos are cloned into the current directory in a directory named after the
			assignment.

			If the student repos are individual assignements the cloned directories will be
			named after the student email address as lastname.firstname. If the student repos 
			are group assignments the cloned directories will be named after the repo name.
			
			The starter repo is cloned into a directory named ".main"`),
		Example: `$ gh crm clone`,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := api.DefaultRESTClient()
			if err != nil {
				crm.Fatal(err)
			}

			c, err := crm.LoadClassroom()
			if err != nil {
				crm.Fatal(err)
			}

			isClassroomFolder, err := crm.IsClassroomFolder()
			if err != nil {
				crm.Fatal(err)
			}

			if isAssignmentFolder, err = crm.IsAssignmentFolder(); err == nil && isAssignmentFolder {
				a, err := crm.LoadAssignment()
				if err != nil {
					crm.Fatal(err)
				}
				aId = a.Id
			}
			if err != nil {
				crm.Fatal(err)
			}

			if !isClassroomFolder && !isAssignmentFolder {
				crm.Fatal("No classroom or assignment found. `gh crm clone` should either be run from within a classroom folder or from within an assignment folder. Run `gh crm init` to initialize a classroom folder or change to an initialized classroom folder.")
			}

			if aId == 0 {
				a, err := shared.PromptForAssignment(client, c.Classroom.Id)
				if err != nil {
					crm.Fatal(err)
				}

				aId = a.Id
			}

			assignment, err := classroom.GetAssignment(client, aId)
			if err != nil {
				crm.Fatal(err)
			}

			var assignmentPath string
			if isAssignmentFolder {
				assignmentPath, err = os.Getwd()
			} else {
				assignmentPath, err = filepath.Abs(assignment.Slug)
			}
			if err != nil {
				fmt.Println("Error getting absolute path for directory: ", err)
				return
			}

			if !isAssignmentFolder {
				if _, err := os.Stat(assignmentPath); os.IsNotExist(err) {
					fmt.Println("Creating directory: ", assignmentPath)
					err = os.MkdirAll(assignmentPath, 0755)
					if err != nil {
						crm.Fatal(err)
					}
				}

				a := crm.NewAssignment()
				a.Set(assignment.Id, assignment.Slug)
				err = a.Save(assignmentPath)
				if err != nil {
					crm.Fatal(err)
				}
			}

			totalCloned := 0
			cloneErrors := []string{}

			if assignment.StarterCodeRepository.Id != 0 {
				if starterFolder == "" {
					starterFolder = ".main"
				}
				starterPath := filepath.Join(assignmentPath, starterFolder)
				err = utils.CloneRepository(starterPath, assignment.StarterCodeRepository.FullName, gh.Exec)
				if err != nil {
					errMsg := fmt.Sprintf("Error cloning %s: %v", assignment.StarterCodeRepository.FullName, err)
					cloneErrors = append(cloneErrors, errMsg)
				} else {
					totalCloned++
				}
			}

			acceptedAssignmentList, err := shared.ListAllAcceptedAssignments(client, aId, 15)
			if err != nil {
				crm.Fatal(err)
			}

			for _, acceptedAssignment := range acceptedAssignmentList.AcceptedAssignments {
				repoName := acceptedAssignment.Repository.Name
				if len(acceptedAssignment.Students) == 1 {
					if name, err := c.GetRepoName(acceptedAssignment.Students[0].Login); err == nil {
						repoName = name
					}
				}
				clonePath := filepath.Join(assignmentPath, repoName)
				err := utils.CloneRepository(clonePath, acceptedAssignment.Repository.FullName, gh.Exec)
				if err != nil {
					errMsg := fmt.Sprintf("Error cloning %s: %v", acceptedAssignment.Repository.FullName, err)
					cloneErrors = append(cloneErrors, errMsg)
					continue // Continue with the next iteration
				}
				totalCloned++
			}
			if len(cloneErrors) > 0 {
				fmt.Println("Some repositories failed to clone.")
				if !verbose {
					fmt.Println("Run with --verbose flag to see more details")
				} else {
					for _, errMsg := range cloneErrors {
						fmt.Println(errMsg)
					}
				}
			}
			fmt.Printf("Cloned %v repos.\n", totalCloned)
		},
	}

	cmd.Flags().IntVarP(&aId, "assignment-id", "a", 0, "ID of the assignment")
	cmd.Flags().StringVarP(&starterFolder, "starter-folder", "s", "", "name of the folder the starter code shall be cloned to")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose error output")

	return cmd
}
