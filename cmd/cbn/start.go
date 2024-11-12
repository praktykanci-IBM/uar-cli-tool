package cbn

import (
	"context"
	"fmt"
	"os"
	"strings"

	"praktykanci/uar/configData"
	"praktykanci/uar/types"

	"github.com/google/go-github/v66/github"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var ownerName, repo, cbnType string

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the CBN",
	Aliases: []string{"s"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if ownerName == "" || repo == "" || cbnType == "" {
			fmt.Println("Error: owner_name, repo, and type are required flags.")
			os.Exit(1)
		}

		if cbnType != "positive" && cbnType != "negative" {
			fmt.Println("Error: type of CBN must be 'positive' or 'negative'.")
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		_, currentCbns, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, "", nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for _, cbn := range currentCbns {
			cbnFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, *cbn.Name, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			cbnContentMarshaled, err := cbnFile.GetContent()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			var cbnContent types.CbnData
			err = yaml.Unmarshal([]byte(cbnContentMarshaled), &cbnContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if cbnContent.Repo == repo {
				fmt.Fprintf(os.Stderr, "CBN for this repository already exists\n")
				os.Exit(1)
			}
		}

		repoOrg, repoName := strings.Split(repo, "/")[0], strings.Split(repo, "/")[1]
		collaborators, _, err := githubClient.Repositories.ListCollaborators(context.Background(), repoOrg, repoName, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "No such repository\n")
			os.Exit(1)
		}

		isMaintainer := false
		for _, collaborator := range collaborators {
			if *collaborator.Login == ownerName && collaborator.Permissions["maintain"] {
				isMaintainer = true
				break
			}
		}

		if !isMaintainer {
			fmt.Fprintf(os.Stderr, "User %s is not a maintainer of %s\n", ownerName, repo)
			os.Exit(1)
		}

		newCbnID := uuid.New().String()
		newCbn := types.CbnData{
			Owner: ownerName,
			Repo:  repo,
			Type:  cbnType,
			Users: []types.CbnUser{},
		}

		newCmdMarshaled, err := yaml.Marshal(newCbn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		_, _, err = githubClient.Repositories.CreateFile(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, fmt.Sprintf("%s.yaml", newCbnID), &github.RepositoryContentFileOptions{
			Content: []byte(newCmdMarshaled),
			Message: github.String(fmt.Sprintf("Start CBN %s", newCbnID)),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("CBN %s started\n", newCbnID)
	},
}

func init() {
	startCmd.Flags().StringVarP(&ownerName, "owner", "o", "", "The owner of the repository")
	startCmd.Flags().StringVarP(&repo, "repo", "r", "", "The repository in the format owner/repo")
	startCmd.Flags().StringVarP(&cbnType, "type", "t", "", "The type of CBN (positive or negative)")
	CbnCommand.AddCommand(startCmd)
}
