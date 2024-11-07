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

var useRepoNameInsteadOfCbnIdValidate bool
var validateCmd = &cobra.Command{
	Use:     "validate manager_name user_name {cbn_id | --repo repo_name} {approve | reject}",
	Short:   "Validate the CBN",
	Aliases: []string{"v"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 4 {
			return fmt.Errorf("requires manager_name, user_name, cbn_id or repo_name and approve or reject")
		}

		if args[3] != "approve" && args[3] != "reject" {
			return fmt.Errorf("requires approve or reject as the last argument")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		// TODO validate manager

		cbnID := getCbnID(useRepoNameInsteadOfCbnIdValidate, args[2], githubClient)

		cbnOriginalFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, fmt.Sprintf("%s.yaml", cbnID), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cbnContentMarshalled, err := cbnOriginalFile.GetContent()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var cbnContent types.CbnData
		err = yaml.Unmarshal([]byte(cbnContentMarshalled), &cbnContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		userFound := false
		for i, user := range cbnContent.Users {
			if user.Name == args[1] {
				userFound = true

				if args[3] == "approve" {
					cbnContent.Users[i].Status = types.Aproved
				} else {
					cbnContent.Users[i].Status = types.Rejected
				}

				break
			}
		}

		if !userFound {
			fmt.Fprintf(os.Stderr, "This user doesn't have acces to this repo\n")
			os.Exit(1)
		}

		resCbnMarshalled, err := yaml.Marshal(cbnContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		_, _, err = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.CBN_DB_NAME, *cbnOriginalFile.Name, &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Validate user %s for the CBN %s", args[1], cbnID)),
			Content: resCbnMarshalled,
			SHA:     cbnOriginalFile.SHA,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("User %s has been %sed for the CBN %s\n", args[1], args[3], cbnID)
	},
}

func init() {
	validateCmd.Flags().BoolVarP(&useRepoNameInsteadOfCbnIdValidate, "repo", "r", false, "Use repo name instead of cbn id")
	CbnCommand.AddCommand(validateCmd)
}
