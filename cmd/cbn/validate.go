package cbn

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"
	"praktykanci/uar/types"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var managerName, userName, action string
var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Validate the CBN",
	Aliases: []string{"v"},
	Run:     runValidateCmd,
}

func init() {
	validateCmd.Flags().StringVarP(&managerName, "manager", "m", "", "Name of the manager (required)")
	validateCmd.Flags().StringVarP(&userName, "user", "u", "", "Name of the user to validate (required)")
	validateCmd.Flags().StringVarP(&cbnID, "cbn-id", "c", "", "CBN ID (use either this or --repo)")
	validateCmd.Flags().StringVarP(&repoName, "repo", "r", "", "Repository name (use either this or --cbn-id)")
	validateCmd.Flags().StringVarP(&action, "action", "a", "", "Validation action: approve or reject (required)")

	validateCmd.MarkFlagRequired("manager")
	validateCmd.MarkFlagRequired("user")
	validateCmd.MarkFlagRequired("action")

	CbnCommand.AddCommand(validateCmd)
}

func runValidateCmd(cmd *cobra.Command, args []string) {
	if (cbnID == "" && repoName == "") || (cbnID != "" && repoName != "") {
		fmt.Fprintln(os.Stderr, "Error: must specify either --cbn-id or --repo")
		os.Exit(1)
	}

	if action != "approve" && action != "reject" {
		fmt.Fprintln(os.Stderr, "Error: --action must be either 'approve' or 'reject'")
		os.Exit(1)
	}

	githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

	finalCbnID := cbnID
	if repoName != "" {
		finalCbnID = getCbnID(true, repoName, githubClient)
	}

	cbnOriginalFile, _, _, err := githubClient.Repositories.GetContents(
		context.Background(),
		configData.ORG_NAME,
		configData.CBN_DB_NAME,
		fmt.Sprintf("%s.yaml", finalCbnID),
		nil,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving CBN file: %v\n", err)
		os.Exit(1)
	}

	cbnContentMarshalled, err := cbnOriginalFile.GetContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving file content: %v\n", err)
		os.Exit(1)
	}

	var cbnContent types.CbnData
	err = yaml.Unmarshal([]byte(cbnContentMarshalled), &cbnContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling YAML content: %v\n", err)
		os.Exit(1)
	}

	userFound := false
	for i, user := range cbnContent.Users {
		if user.Name == userName {
			userFound = true
			if action == "approve" {
				cbnContent.Users[i].Status = types.Aproved
			} else {
				cbnContent.Users[i].Status = types.Rejected
			}
			break
		}
	}

	if !userFound {
		fmt.Fprintln(os.Stderr, "Error: specified user does not have access to this CBN")
		os.Exit(1)
	}

	resCbnMarshalled, err := yaml.Marshal(cbnContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling updated content: %v\n", err)
		os.Exit(1)
	}

	_, _, err = githubClient.Repositories.UpdateFile(
		context.Background(),
		configData.ORG_NAME,
		configData.CBN_DB_NAME,
		*cbnOriginalFile.Name,
		&github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Validate user %s for the CBN %s", userName, finalCbnID)),
			Content: resCbnMarshalled,
			SHA:     cbnOriginalFile.SHA,
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating file on GitHub: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User %s has been %sed for the CBN %s\n", userName, action, finalCbnID)
}
