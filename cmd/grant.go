package cmd

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"

	// . "praktykanci/uar/configData"
	// . "praktykanci/uar/types"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
)

var grantCmd = &cobra.Command{
	Use:     "grant manager_name {uar_id | user_name repo}",
	Short:   "Grant a request",
	Aliases: []string{"g"},
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
		// validate manager

		if len(args) == 2 {
			grantByUarID(args[1])
		} else {
			grantByUserAndRepo(args[1], args[2])
		}

		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(grantCmd)
}

func grantByUarID(uarID string) {
	fmt.Printf("Granting request with UAR ID %s\n", uarID)
}

func grantByUserAndRepo(user string, repo string) {
	fmt.Printf("Granting request for user %s and repo %s\n", user, repo)

	githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

	// repoOwner, repoName := strings.Split(repo, "/")[0], strings.Split(repo, "/")[1]

	branches, _, err := githubClient.Repositories.ListBranches(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var requestBranch *github.Branch

	for _, branch := range branches {
		if *branch.Name == fmt.Sprintf("%s/%s", repo, user) {
			requestBranch = branch
			break
		}
	}

	if requestBranch == nil {
		fmt.Fprintf(os.Stderr, "Error: branch not found\n")
		os.Exit(1)
	}

	pullRequests, _, err := githubClient.PullRequests.ListPullRequestsWithCommit(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, *requestBranch.Commit.SHA, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_, _, err = githubClient.PullRequests.CreateReview(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, *pullRequests[0].Number, &github.PullRequestReviewRequest{
		Body:  github.String("access granted"),
		Event: github.String("APPROVE"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_, _, err = githubClient.PullRequests.Merge(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, *pullRequests[0].Number, "access granted", &github.PullRequestOptions{
		MergeMethod: "merge",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_, err = githubClient.Git.DeleteRef(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("heads/%s", *requestBranch.Name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
