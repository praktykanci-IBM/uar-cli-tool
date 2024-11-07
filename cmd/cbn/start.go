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

var negativeRevalidation bool
var startCmd = &cobra.Command{
	Use:     "start owner_name repo",
	Short:   "Start the CBN",
	Aliases: []string{"s"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("requires owner_name and repo")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
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

			if cbnContent.Repo == args[1] {
				fmt.Fprintf(os.Stderr, "CBN for this repository already exists\n")
				os.Exit(1)
			}
		}

		repoOrg, repoName := strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1]
		collaborators, _, err := githubClient.Repositories.ListCollaborators(context.Background(), repoOrg, repoName, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "No such repository\n")
			os.Exit(1)
		}

		isMaintainer := false
		for _, collaborator := range collaborators {
			if *collaborator.Login == args[0] && collaborator.Permissions["maintain"] {
				isMaintainer = true
				break
			}
		}

		if !isMaintainer {
			fmt.Fprintf(os.Stderr, "User %s is not a maintainer of %s\n", args[0], args[1])
			os.Exit(1)
		}

		var cbnType types.CbnType
		if negativeRevalidation {
			cbnType = types.Negative
		} else {
			cbnType = types.Positive
		}

		newCbnID := uuid.New().String()
		newCbn := types.CbnData{
			Owner: args[0],
			Repo:  args[1],
			Type:  cbnType,
			Users: []string{},
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
	startCmd.Flags().BoolVarP(&negativeRevalidation, "negative-revalidation", "n", false, "Use negative revalidation insted of positive")
	CbnCommand.AddCommand(startCmd)
}
