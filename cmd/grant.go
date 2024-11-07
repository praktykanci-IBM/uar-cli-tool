package cmd

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"
	"strings"

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

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if len(args) == 2 {
			grantByUarID(args[1], githubClient)
		} else {
			grantByUserAndRepo(args[1], args[2], githubClient)
		}

		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(grantCmd)
}

func grantByUarID(uarID string, githubClient *github.Client) {
	branches, _, err := githubClient.Repositories.ListBranches(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var requestBranch *github.Branch

Outer:
	for _, branch := range branches {
		pullRequests, _, err := githubClient.PullRequests.ListPullRequestsWithCommit(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, *branch.Commit.SHA, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for _, pullRequest := range pullRequests {
			splittedTitle := strings.Split(*pullRequest.Title, " - ")
			if len(splittedTitle) != 2 {
				continue
			}
			id := splittedTitle[1]

			if id == uarID {
				requestBranch = branch
				break Outer
			}
		}
	}

	if requestBranch == nil {
		fmt.Fprintf(os.Stderr, "Error: branch not found\n")
		os.Exit(1)
	}

	grantAccess(*requestBranch.Commit.SHA, githubClient)

	fmt.Printf("Access granted to UAR ID %s\n", uarID)
}

func grantByUserAndRepo(user string, repo string, githubClient *github.Client) {
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

	grantAccess(*requestBranch.Commit.SHA, githubClient)

	fmt.Printf("Access granted to %s on %s\n", user, repo)
}

func grantAccess(commitSHA string, githubClient *github.Client) {
	pullRequests, _, err := githubClient.PullRequests.ListPullRequestsWithCommit(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, commitSHA, nil)
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
		MergeMethod: "squash",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
