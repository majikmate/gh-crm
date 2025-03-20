package teams

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/go-gh"
	"github.com/github/gh-classroom/cmd/gh-classroom/shared"
	"github.com/github/gh-classroom/pkg/classroom"
	"github.com/spf13/cobra"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache"

	graph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func NewCmdTeams(f *cmdutil.Factory) *cobra.Command {
	var teamId string
	var crmId int

	cmd := &cobra.Command{
		Use:     "teams",
		Example: `$ gh crm teams`,
		Run: func(cmd *cobra.Command, args []string) {
			client, err := gh.RESTClient(nil)

			if err != nil {
				log.Fatal(err)
				return
			}

			if crmId == 0 {
				crm, err := shared.PromptForClassroom(client)
				crmId = crm.Id

				if err != nil {
					log.Fatal(err)
					return
				}
			}

			crm, err := classroom.GetClassroom(client, crmId)
			if err != nil {
				log.Fatal(err)
			}

			team, err := promptForTeam()
			if err != nil {
				log.Fatalf("Failed to get teams: %v", err)
			}

			assignemets, err := getAssignments(team.ID)
			if err != nil {
				log.Fatalf("Failed to get assignments: %v", err)
			}

			for _, assignment := range assignemets {
				log.Printf("assignment: %s", *assignment.GetDisplayName())
			}

			log.Printf("organization    : %s", crm.Organization.Login)
			log.Printf("organization url: %s", crm.Organization.HtmlUrl)
			log.Printf("classroom       : %s", crm.Name)
			log.Printf("classroom url   : %s", crm.Url)
			log.Printf("team            : %s", team.Name)
		},
	}

	cmd.Flags().StringVarP(&teamId, "team-id", "t", "", "ID of the team")
	cmd.Flags().IntVarP(&crmId, "classroom-id", "c", 0, "ID of the classroom")

	return cmd
}

const (
	tenantID = "9987787c-b131-4ef8-bdad-35e3b7c323ec"
	clientID = "86a60d51-eb6f-405c-b727-59d1d3222cdc"
)

var (
	permissions = []string{
		"User.Read",
		"Team.ReadBasic.All",
		// TODO: "TeamMember.Read.All",
	}
)

func authRecordPath() string {
	return os.TempDir() + "/" + clientID
}

func getAuthRec() (azidentity.AuthenticationRecord, error) {
	record := azidentity.AuthenticationRecord{}
	b, err := os.ReadFile(authRecordPath())
	if err == nil {
		err = json.Unmarshal(b, &record)
	}
	return record, nil
}

func putAuthRec(record azidentity.AuthenticationRecord) error {
	b, err := json.Marshal(record)
	if err == nil {
		err = os.WriteFile(authRecordPath(), b, 0600)
	}
	return err
}

// Struct to parse team response
type Team struct {
	ID   string
	Name string
}

var client *graph.GraphServiceClient

func getClient() (*graph.GraphServiceClient, error) {
	if client != nil {
		return client, nil
	}

	record, err := getAuthRec()
	if err != nil {
		return nil, err
	}

	c, err := cache.New(nil)
	if err != nil {
		return nil, err
	}

	cred, err := azidentity.NewInteractiveBrowserCredential(&azidentity.InteractiveBrowserCredentialOptions{
		ClientID:             clientID,
		TenantID:             tenantID,
		AuthenticationRecord: record,
		Cache:                c,
	})
	if err != nil {
		return nil, err
	}

	if record == (azidentity.AuthenticationRecord{}) {
		// No stored record; call Authenticate to acquire one.
		// This will prompt the user to authenticate interactively.
		record, err = cred.Authenticate(context.Background(), &policy.TokenRequestOptions{
			Scopes: permissions,
		})
		if err != nil {
			return nil, err
		}
		err = putAuthRec(record)
		if err != nil {
			return nil, err
		}
	}

	client, err = graph.NewGraphServiceClientWithCredentials(cred, permissions)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// returns a list of non-archived teams
func getTeams() ([]Team, error) {
	ctx := context.Background()

	// get the ms graph client
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	// get objects owned by the user
	ownedObjects, err := client.Me().OwnedObjects().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	// get teams
	teams, err := client.Me().JoinedTeams().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	// compile owned teams list
	teamList := []Team{}
	for _, team := range teams.GetValue() {
		for _, ownedObject := range ownedObjects.GetValue() {
			if *team.GetId() == *ownedObject.GetId() {
				// Skip archived teams
				if *team.GetIsArchived() {
					continue
				}

				// Append team to list
				teamList = append(teamList, Team{
					ID:   *team.GetId(),
					Name: *team.GetDisplayName(),
				})
			}
		}
	}

	// Sort the teamList by name
	sort.Slice(teamList, func(i, j int) bool {
		return teamList[i].Name < teamList[j].Name
	})

	return teamList, nil
}

func promptForTeam() (team *Team, err error) {
	teams, err := getTeams()

	if err != nil {
		return nil, err
	}

	if len(teams) == 0 {
		return nil, errors.New("No teams found.")
	}

	optionMap := make(map[string]*Team)
	options := make([]string, 0, len(teams))

	for _, team := range teams {
		optionMap[team.Name] = &team
		options = append(options, team.Name)
	}

	var qs = []*survey.Question{
		{
			Name: "team",
			Prompt: &survey.Select{
				Message: "Select a team:",
				Options: options,
			},
		},
	}

	answer := struct {
		Team string
	}{}

	err = survey.Ask(qs, &answer)

	if err != nil {
		return nil, err
	}

	return optionMap[answer.Team], nil
}

func getAssignments(classID string) ([]models.EducationAssignmentable, error) {
	ctx := context.Background()

	assignmentsRequest := client.Education().Classes().ByEducationClassId(classID).Assignments()
	assignments, err := assignmentsRequest.Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting assignments: %v", err)
	}

	return assignments.GetValue(), nil
}
