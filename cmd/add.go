package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"praktykanci/uar/configData"
	. "praktykanci/uar/types"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var AddCommand = &cobra.Command{
	Use:     "add admin_name {uar_id | user_name repo}",
	Short:   "Add a user as a collaborator",
	Aliases: []string{"a"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("your github ID is required")
		}
		if len(args) == 1 {
			return fmt.Errorf("either UAR ID or user and repo name are required")
		}
		if len(args) > 3 {
			return fmt.Errorf("too many arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// validate admin

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if len(args) == 2 {
			addByUarID(args[1], githubClient)
		} else {
			addByUserAndRepo(args[1], args[2], githubClient)
		}

	},
}

func addByUarID(uarID string, githubClient *github.Client) {
	_, ownersDirs, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, "", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	foundUser := false

Outer:
	for _, ownerDir := range ownersDirs {
		_, reposDirs, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, *ownerDir.Name, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for _, repoDir := range reposDirs {
			_, requestFiles, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s", *ownerDir.Name, *repoDir.Name), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			for _, requestFileWithouData := range requestFiles {
				requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf(("%s/%s/%s"), *ownerDir.Name, *repoDir.Name, *requestFileWithouData.Name), nil)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				requestFileMarshaled, err := requestFile.GetContent()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				var requestFileContent RequestData
				err = yaml.Unmarshal([]byte(requestFileMarshaled), &requestFileContent)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				if requestFileContent.ID == uarID {
					username := strings.Split(*requestFile.Name, ".")[0]
					repo := fmt.Sprintf("%s/%s", *ownerDir.Name, *repoDir.Name)
					foundUser = true

					addByUserAndRepo(username, repo, githubClient)

					break Outer
				}
			}
		}

		if !foundUser {
			fmt.Fprintf(os.Stderr, "Request with ID %s not found\n", uarID)
			os.Exit(1)
		}
	}
}

func addByUserAndRepo(user string, repo string, githubClient *github.Client) {
	requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s.yaml", repo, user), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err) // no such request
		os.Exit(1)
	}

	requestFileSHA := *requestFile.SHA
	requestFileMarshaled, err := requestFile.GetContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var requestFileContent RequestData
	err = yaml.Unmarshal([]byte(requestFileMarshaled), &requestFileContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if requestFileContent.Added {
		fmt.Printf("User %s is already a collaborator of %s\n", user, repo)
		os.Exit(0)
	}

	orgName, repoName := strings.Split(repo, "/")[0], strings.Split(repo, "/")[1]
	_, _, err = githubClient.Repositories.AddCollaborator(context.Background(), orgName, repoName, user, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	requestFileContent.Added = true
	resFileMarshaled, err := yaml.Marshal(requestFileContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_, _, err = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s.yaml", repo, user), &github.RepositoryContentFileOptions{
		Message: github.String("Add collaborator"),
		Content: resFileMarshaled,
		SHA:     &requestFileSHA,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User %s added as a collaborator of %s\n", user, repo)
}

func init() {
	rootCmd.AddCommand(AddCommand)
}
