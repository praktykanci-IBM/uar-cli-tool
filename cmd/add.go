package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"praktykanci/uar/configData"
	"praktykanci/uar/types"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var addCommand = &cobra.Command{
	Use:     "add",
	Short:   "Add a user as a collaborator",
	Aliases: []string{"a"},
	Run: func(cmd *cobra.Command, args []string) {
		uarID, err := cmd.Flags().GetString("uar-id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		user, err := cmd.Flags().GetString("user")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		repo, err := cmd.Flags().GetString("repo")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if repo != "" && !strings.Contains(repo, "/") {
			fmt.Println("Error: Invalid repository name. Repo name should be in format owner/repo.")
			cmd.Help()
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if uarID != "" {
			addByUarID(uarID, githubClient)
		} else if user != "" && repo != "" {
			addByUserAndRepo(user, repo, githubClient)
		}

	},
}

func addByUarID(uarID string, githubClient *github.Client) {
	_, ownersDirs, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, "user-access-records", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	foundUser := false

Outer:
	for _, ownerDir := range ownersDirs {
		_, reposDirs, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s", *ownerDir.Name), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for _, repoDir := range reposDirs {
			_, requestFiles, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s", *ownerDir.Name, *repoDir.Name), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			for _, requestFileWithouData := range requestFiles {
				requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf(("user-access-records/%s/%s/%s"), *ownerDir.Name, *repoDir.Name, *requestFileWithouData.Name), nil)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				requestFileMarshaled, err := requestFile.GetContent()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				var requestFileContent types.RequestData
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
	requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s.yaml", repo, user), nil)
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

	var requestFileContent types.RequestData
	err = yaml.Unmarshal([]byte(requestFileMarshaled), &requestFileContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if requestFileContent.State == types.Completed {
		fmt.Printf("User %s is already a collaborator of %s\n", user, repo)
		os.Exit(0)
	}

	orgName, repoName := strings.Split(repo, "/")[0], strings.Split(repo, "/")[1]
	_, _, err = githubClient.Repositories.AddCollaborator(context.Background(), orgName, repoName, user, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

	completedBy, _, err := githubClient.Users.Get(context.Background(), "")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	requestCompleted := types.RequestDataCompleted{
		ID:            requestFileContent.ID,
		State:         types.Completed,
		Justification: requestFileContent.Justification,
		RequestedOn:   requestFileContent.RequestedOn,
		RequestedBy:   requestFileContent.RequestedBy,
		CompletedOn:   formattedTime,
		CompletedBy:   *completedBy.Login,
	}

	// requestFileContent.State = Completed
	resFileMarshaled, err := yaml.Marshal(requestCompleted)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_, _, err = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s.yaml", repo, user), &github.RepositoryContentFileOptions{
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
	addCommand.Flags().StringP("uar-id", "i", "", "UAR ID to add as a collaborator")
	addCommand.Flags().StringP("user", "u", "", "GitHub username requesting access")
	addCommand.Flags().StringP("repo", "r", "", "Repository name (owner/repo)")

	addCommand.MarkFlagsMutuallyExclusive("uar-id", "user")
	addCommand.MarkFlagsMutuallyExclusive("uar-id", "repo")
	addCommand.MarkFlagsRequiredTogether("user", "repo")
	addCommand.MarkFlagsOneRequired("uar-id", "user")

	addCommand.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", configData.GITHUB_PAT, "GitHub personal access token")

	addCommand.Flags().SortFlags = false
	rootCmd.AddCommand(addCommand)
}
