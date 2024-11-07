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
var extractCmd = &cobra.Command{
	Use:     "extract admin_name {cbn_id | --repo repo_name}",
	Short:   "Extract data for the CBN",
	Aliases: []string{"e"},
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
			Message: github.String(fmt.Sprintf("Extract data for the CBN %s", cbnID)),
			Content: resCbnMarshaled,
			SHA:     cbnOriginalFile.SHA,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Data extracted for CBN with ID: %s\n", cbnID)
	},
}

func init() {
	extractCmd.Flags().BoolVarP(&useRepoNameInsteadOfCbnIdExtract, "repo", "r", false, "Use the repo name instead of the CBN ID")
	CbnCommand.AddCommand(extractCmd)
}
