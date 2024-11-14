package cbn

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"
	"praktykanci/uar/types"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var useRepoNameInsteadOfCbnIdExtract bool
var adminName, cbnID, repoName string

var extractCmd = &cobra.Command{
	Use:     "extract",
	Short:   "Extract data for the CBN",
	Aliases: []string{"e"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if adminName == "" {
			fmt.Println("Error: admin_name is required.")
			os.Exit(1)
		}

		if cbnID == "" && repoName == "" {
			fmt.Println("Error: Either cbn_id or repo name is required.")
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

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

		_, usersWithAccess, res, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, cbnContent.Repo, nil)
		if err != nil {
			if res.StatusCode == 404 {
				fmt.Fprintf(os.Stderr, "No access records for such repository\n")
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}

		cbnContent.Users = []types.CbnUser{}
		for _, user := range usersWithAccess {
			cbnContent.Users = append(cbnContent.Users, types.CbnUser{
				Name:   strings.Split(*user.Name, ".")[0],
				Status: types.Unset,
			})
		}

		resCbnMarshaled, err := yaml.Marshal(cbnContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		_, _, err = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, *cbnOriginalFile.Name, &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Extract data for the CBN %s", cbnIDToUse)),
			Content: resCbnMarshaled,
			SHA:     cbnOriginalFile.SHA,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Data extracted for CBN with ID: %s\n", cbnIDToUse)
	},
}

func init() {
	extractCmd.Flags().StringVarP(&adminName, "admin", "a", "", "The admin name who is extracting the CBN data")
	extractCmd.Flags().StringVarP(&cbnID, "cbn-id", "c", "", "The CBN ID to extract data for")
	extractCmd.Flags().StringVarP(&repoName, "repo", "r", "", "The repository name ")

	extractCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", "", "GitHub personal access token")

	CbnCommand.AddCommand(extractCmd)
}
