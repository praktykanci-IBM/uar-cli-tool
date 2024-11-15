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

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if cbnID == "" && repoName == "" {
			fmt.Println("Error: Either cbn_id or repo name is required.")
			cmd.Help()
			os.Exit(1)
		}

		if cbnID != "" && repoName != "" {
			fmt.Println("Error: You cannot provide both a CBN ID and a repo name.")
			cmd.Help()
			os.Exit(1)
		}

		if cbnID == "" {
			cbnID = getCbnID(repoName, githubClient)
		}

		cbnOriginalFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("CBN/%s.yaml", cbnID), nil)
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

		executedBy, _, err := githubClient.Users.Get(context.Background(), "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		currentTime := time.Now()
		formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

		cbnContentUpdated := types.CbnData{
			StartedBy:    cbnContent.StartedBy,
			StartedOn:    cbnContent.StartedOn,
			Repo:         cbnContent.Repo,
			Type:         cbnContent.Type,
			ExtractedBy:  cbnContent.ExtractedBy,
			ExtractedOn:  cbnContent.ExtractedOn,
			Users:        cbnContent.Users,
			ExecutedBy:   *executedBy.Login,
			ExecutedOn:   formattedTime,
			UsersChanged: []types.CbnUser{},
		}

		for _, user := range cbnContent.Users {

			if (cbnContent.Type == "positive" && user.State == types.Pending) || (user.State == types.Rejected) {
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

		_, _, err = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("CBN/%s", *cbnOriginalFile.Name), &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Update collaborators list, complete CBN %s", cbnID)),
			Content: resCbnMarshaled,
			SHA:     cbnOriginalFile.SHA,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("CBN with ID: %s completed\n", cbnID)

	},
}

func init() {
	updateCmd.Flags().StringP("cbn-id", "i", "", "The CBN ID to extract data for")
	updateCmd.Flags().StringP("repo", "r", "", "The repository name ")

	updateCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", "", "GitHub personal access token")

	updateCmd.Flags().SortFlags = false
	CbnCommand.AddCommand(updateCmd)
}
