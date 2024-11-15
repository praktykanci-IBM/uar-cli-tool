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
	Use:     "grant",
	Short:   "Grant a request",
	Aliases: []string{"g"},
	Run: func(cmd *cobra.Command, args []string) {
		uarID, _ := cmd.Flags().GetString("uar-id")
		user, _ := cmd.Flags().GetString("user")
		repo, _ := cmd.Flags().GetString("repo")

		if uarID == "" && (user == "" || repo == "") {
			fmt.Println("Error: You must provide either a UAR ID or both a user name and a repo.")
			cmd.Help()
			os.Exit(1)
		}

		if uarID != "" && user != "" {
			fmt.Println("Error: You cannot provide both a UAR ID and a user name.")
			cmd.Help()
			os.Exit(1)
		}

		if user != "" && repo == "" {
			fmt.Println("Error: You must provide both a user name and a repo.")
			cmd.Help()
			os.Exit(1)
		}

		if repo != "" && !strings.Contains(repo, "/") {
			fmt.Println("Error: Invalid repository name. Repo name should be in format owner/repo.")
			cmd.Help()
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if uarID != "" {
			grantByUarID(uarID, githubClient)
		} else if user != "" && repo != "" {
			grantByUserAndRepo(user, repo, githubClient)
		}

		os.Exit(0)
	},
}

func grantByUarID(uarID string, githubClient *github.Client) {
	branches, _, err := githubClient.Repositories.ListBranches(context.Background(), configData.ORG_NAME, configData.DB_NAME, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var requestBranch *github.Branch

Outer:
	for _, branch := range branches {
		pullRequests, _, err := githubClient.PullRequests.ListPullRequestsWithCommit(context.Background(), configData.ORG_NAME, configData.DB_NAME, *branch.Commit.SHA, nil)
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
	branches, _, err := githubClient.Repositories.ListBranches(context.Background(), configData.ORG_NAME, configData.DB_NAME, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var requestBranch *github.Branch

	for _, branch := range branches {
		if *branch.Name == fmt.Sprintf("user-access-records/%s/%s", repo, user) {
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
	pullRequests, _, err := githubClient.PullRequests.ListPullRequestsWithCommit(context.Background(), configData.ORG_NAME, configData.DB_NAME, commitSHA, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_, _, err = githubClient.PullRequests.CreateReview(context.Background(), configData.ORG_NAME, configData.DB_NAME, *pullRequests[0].Number, &github.PullRequestReviewRequest{
		Body:  github.String("access granted"),
		Event: github.String("APPROVE"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	_, _, err = githubClient.PullRequests.Merge(context.Background(), configData.ORG_NAME, configData.DB_NAME, *pullRequests[0].Number, "access granted", &github.PullRequestOptions{
		MergeMethod: "squash",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	grantCmd.Flags().StringP("uar-id", "i", "", "UAR ID to grant access")
	grantCmd.Flags().StringP("user", "u", "", "GitHub username requesting access")
	grantCmd.Flags().StringP("repo", "r", "", "Repository name (owner/repo)")

	grantCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", "", "GitHub personal access token")

	grantCmd.Flags().SortFlags = false
	rootCmd.AddCommand(grantCmd)
}
