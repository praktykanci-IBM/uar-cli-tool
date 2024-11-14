package cbn

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"
	"praktykanci/uar/types"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var updateCmd = &cobra.Command{
	Use:     "update admin_name {cbn_id | --repo repo_name}",
	Short:   "Complete the CBN, removing all rejected users from the repo",
	Aliases: []string{"u"},
	Run: func(cmd *cobra.Command, args []string) {
		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if adminName == "" {
			fmt.Println("Error: admin_name is required.")
			os.Exit(1)
		}

		if cbnID == "" && repoName == "" {
			fmt.Println("Error: Either cbn_id or repo name is required.")
			os.Exit(1)
		}

		if cbnID == "" && repoName != "" {
			useRepoNameInsteadOfCbnIdExtract = true
		} else {
			useRepoNameInsteadOfCbnIdExtract = false
			repoName = cbnID
		}

		cbnIDToUse := getCbnID(useRepoNameInsteadOfCbnIdExtract, repoName, githubClient)

		cbnOriginalFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, fmt.Sprintf("%s.yaml", cbnIDToUse), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cbnContentMarshaled, err := cbnOriginalFile.GetContent()
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

		currentTime := time.Now()
		formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

		cbnContentUpdated := types.CbnDataCompleted{
			StartedBy:    cbnContent.StartedBy,
			StartedOn:    cbnContent.StartedOn,
			Repo:         cbnContent.Repo,
			Type:         cbnContent.Type,
			ExtractedBy:  cbnContent.ExtractedBy,
			ExtractedOn:  cbnContent.ExtractedOn,
			Users:        cbnContent.Users,
			ExecutedBy:   adminName,
			ExecutedOn:   formattedTime,
			UsersChanged: []types.CbnUser{},
		}

		for _, user := range cbnContent.Users {

			if (cbnContent.Type == "positive" && user.Status == types.Unset) || (user.Status == types.Rejected) {
				_, err = githubClient.Repositories.RemoveCollaborator(context.Background(), strings.Split(cbnContent.Repo, "/")[0], strings.Split(cbnContent.Repo, "/")[1], user.Name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				cbnContentUpdated.UsersChanged = append(cbnContentUpdated.UsersChanged, user)
			}
		}

		resCbnMarshaled, err := yaml.Marshal(cbnContentUpdated)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		_, _, err = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, *cbnOriginalFile.Name, &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Update collaborators list, complete CBN %s", cbnIDToUse)),
			Content: resCbnMarshaled,
			SHA:     cbnOriginalFile.SHA,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("CBN with ID: %s completed\n", cbnIDToUse)

	},
}

func init() {
	// updateCmd.Flags().BoolVarP(&useRepoNameInsteadOfCbnIdExtract, "repo", "r", false, "Use the repo name instead of the CBN ID")
	CbnCommand.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&adminName, "admin", "a", "", "The admin name who is extracting the CBN data")
	updateCmd.Flags().StringVarP(&cbnID, "cbn-id", "c", "", "The CBN ID to extract data for")
	updateCmd.Flags().StringVarP(&repoName, "repo", "r", "", "The repository name ")

	updateCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", "", "GitHub personal access token")
}
