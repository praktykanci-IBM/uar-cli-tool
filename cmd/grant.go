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
		org, err := cmd.Flags().GetString("org")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		team, err := cmd.Flags().GetString("team")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if uarID != "" {
			grantByUarID(uarID, githubClient)
		} else if user != "" && repo != "" {
			grantByUserAndRepo(user, org, repo, githubClient)
		} else if user != "" && org != "" && team == "" {
			grantByUsernameOrg(user, org, githubClient)
		} else if user != "" && org != "" && team != "" {
			grantByUserAndTeam(user, org, team, githubClient)
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

func grantByUserAndRepo(user string, org string, repo string, githubClient *github.Client) {
	branches, _, err := githubClient.Repositories.ListBranches(context.Background(), configData.ORG_NAME, configData.DB_NAME, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var requestBranch *github.Branch

	for _, branch := range branches {
		if *branch.Name == fmt.Sprintf("user-access-records/%s/%s/%s", org, repo, user) {
			requestBranch = branch
			break
		}
	}

	if requestBranch == nil {
		fmt.Fprintf(os.Stderr, "Error: branch not found\n")
		os.Exit(1)
	}

	grantAccess(*requestBranch.Commit.SHA, githubClient)

	fmt.Printf("Access granted to %s\n", user)
}

func grantByUsernameOrg(user string, org string, githubClient *github.Client) {
	branches, _, err := githubClient.Repositories.ListBranches(context.Background(), configData.ORG_NAME, configData.DB_NAME, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var requestBranch *github.Branch

	for _, branch := range branches {
		if *branch.Name == fmt.Sprintf("org-access-records/%s/%s", org, user) {
			requestBranch = branch
			break
		}
	}

	if requestBranch == nil {
		fmt.Fprintf(os.Stderr, "Error: branch not found\n")
		os.Exit(1)
	}

	grantAccess(*requestBranch.Commit.SHA, githubClient)

	fmt.Printf("Access granted to %s\n", user)
}

func grantByUserAndTeam(user string, org string, team string, githubClient *github.Client) {
	branches, _, err := githubClient.Repositories.ListBranches(context.Background(), configData.ORG_NAME, configData.DB_NAME, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var requestBranch *github.Branch

	for _, branch := range branches {
		if *branch.Name == fmt.Sprintf("team-access-records/%s/%s/%s", org, team, user) {
			requestBranch = branch
			break
		}
	}

	if requestBranch == nil {
		fmt.Fprintf(os.Stderr, "Error: branch not found\n")
		os.Exit(1)
	}

	grantAccess(*requestBranch.Commit.SHA, githubClient)

	fmt.Printf("Access granted to %s\n", user)
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
	grantCmd.Flags().StringP("repo", "r", "", "Repository name")
	grantCmd.Flags().StringP("team", "e", "", "Team name")
	grantCmd.Flags().StringP("org", "o", "", "Organization")

	grantCmd.MarkFlagsMutuallyExclusive("uar-id", "user")
	grantCmd.MarkFlagsMutuallyExclusive("uar-id", "repo", "team")
	grantCmd.MarkFlagsOneRequired("uar-id", "user")

	grantCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", configData.GITHUB_PAT, "GitHub personal access token")

	grantCmd.Flags().SortFlags = false
	rootCmd.AddCommand(grantCmd)
}
