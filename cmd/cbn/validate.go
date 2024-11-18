package cbn

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"
	"praktykanci/uar/types"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var managerName, userName, action string
var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Validate the CBN",
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		userName, err := cmd.Flags().GetString("user")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cbnID, err := cmd.Flags().GetString("cbn-id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		repoName, err := cmd.Flags().GetString("repo")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		action, err := cmd.Flags().GetString("action")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if action != "approve" && action != "reject" {
			fmt.Fprintln(os.Stderr, "Error: Action must be either 'approve' or 'reject'")
			cmd.Help()
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if cbnID == "" {
			cbnID = getCbnID(repoName, githubClient)
		}

		cbnOriginalFile, _, _, err := githubClient.Repositories.GetContents(
			context.Background(),
			configData.ORG_NAME,
			configData.DB_NAME,
			fmt.Sprintf("CBN/%s.yaml", cbnID),
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
					cbnContent.Users[i].State = types.Aproved
				} else {
					cbnContent.Users[i].State = types.Rejected
				}

				currentTime := time.Now()
				formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

				validatedBy, _, err := githubClient.Users.Get(context.Background(), "")
				if err != nil {
					fmt.Println("Error:", err)
					return
				}

				cbnContent.Users[i].ValidatedBy = *validatedBy.Login
				cbnContent.Users[i].ValidatedOn = formattedTime

				break
			}
		}

		if !userFound {
			fmt.Fprintln(os.Stderr, "Error: specified user does not have access to repo in this CBN")
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
			configData.DB_NAME,
			fmt.Sprintf("CBN/%s", *cbnOriginalFile.Name),
			&github.RepositoryContentFileOptions{
				Message: github.String(fmt.Sprintf("Validate user %s for the CBN %s", userName, cbnID)),
				Content: resCbnMarshalled,
				SHA:     cbnOriginalFile.SHA,
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating file on GitHub: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("User %s has been %sed for the CBN %s\n", userName, action, cbnID)
	},
}

func init() {
	validateCmd.Flags().StringP("user", "u", "", "Name of the user to validate")
	validateCmd.Flags().StringP("cbn-id", "i", "", "CBN ID")
	validateCmd.Flags().StringP("repo", "r", "", "Repository name")
	validateCmd.Flags().StringP("action", "a", "", "Validation action: approve or reject")

	validateCmd.MarkFlagRequired("user")
	validateCmd.MarkFlagRequired("action")
	validateCmd.MarkFlagsMutuallyExclusive("cbn-id", "repo")
	validateCmd.MarkFlagsOneRequired("cbn-id", "repo")

	validateCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", "", "GitHub personal access token")

	validateCmd.Flags().SortFlags = false
	CbnCommand.AddCommand(validateCmd)
}
