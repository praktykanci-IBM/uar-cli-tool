package cbn

import (
	"context"
	"fmt"
	"os"

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
