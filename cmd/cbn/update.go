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

// var useRepoNameInsteadOfCbnIdExtract bool
var updateCmd = &cobra.Command{
	Use:     "update admin_name {cbn_id | --repo repo_name}",
	Short:   "Complete the CBN, removing all rejected users from the repo",
	Aliases: []string{"u"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("requires owner_name and repo")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		cbnID := getCbnID(useRepoNameInsteadOfCbnIdExtract, args[1], githubClient)

		cbnOriginalFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, fmt.Sprintf("%s.yaml", cbnID), nil)
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

		cbnContentUpdated := types.CbnDataCompleted{
			Owner:        cbnContent.Owner,
			Repo:         cbnContent.Repo,
			Type:         cbnContent.Type,
			Users:        cbnContent.Users,
			ExecutedBy:   args[0],
			ExecutedOn:   time.Now().Unix(),
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
	updateCmd.Flags().BoolVarP(&useRepoNameInsteadOfCbnIdExtract, "repo", "r", false, "Use the repo name instead of the CBN ID")
	CbnCommand.AddCommand(updateCmd)
}
