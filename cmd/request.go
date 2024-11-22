package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"praktykanci/uar/configData"
	"praktykanci/uar/types"

	"github.com/google/go-github/v66/github"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var requestCmd = &cobra.Command{
	Use:     "request",
	Aliases: []string{"r"},
	Short:   "Request access to repository",
	Long:    "Request access to selected repository with user ID, repository name, business justification and manager name",
	Run: func(cmd *cobra.Command, args []string) {
		userName, err := cmd.Flags().GetString("user")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		repo, _ := cmd.Flags().GetString("repo")
		org, _ := cmd.Flags().GetString("org")
		team, _ := cmd.Flags().GetString("team")

		justification, err := cmd.Flags().GetString("justification")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		managerName, err := cmd.Flags().GetString("manager")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		requestedByUser, _, err := githubClient.Users.Get(context.Background(), "")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if userName == "" {
			userName = *requestedByUser.Login
		}

		_, _, err = githubClient.Users.Get(context.Background(), userName)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if repo != "" {
			_, _, err = githubClient.Repositories.Get(context.Background(), org, repo)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
		}

		isCollaborator, _, err := githubClient.Repositories.IsCollaborator(context.Background(), configData.ORG_NAME, configData.DB_NAME, managerName)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if !isCollaborator {
			fmt.Printf("Invalid manager name\n")
			return
		}

		currentTime := time.Now()
		formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

		id := uuid.New().String()
		newRequest := types.RequestData{
			ID:            id,
			State:         types.Granted,
			Justification: justification,
			RequestedOn:   formattedTime,
			RequestedBy:   *requestedByUser.Login,
			Manager:       managerName,
		}

		content, err := yaml.Marshal(&newRequest)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		var branchName, targetPath string

		if repo != "" {
			branchName = fmt.Sprintf("user-access-records/%s/%s/%s", org, repo, userName)
			targetPath = fmt.Sprintf("user-access-records/%s/%s/%s.yaml", org, repo, userName)
		} else if org != "" && team == "" {
			branchName = fmt.Sprintf("org-access-records/%s/%s", org, userName)
			targetPath = fmt.Sprintf("org-access-records/%s/%s.yaml", org, userName)
		} else if org != "" && team != "" {
			branchName = fmt.Sprintf("team-access-records/%s/%s/%s", org, team, userName)
			targetPath = fmt.Sprintf("team-access-records/%s/%s/%s.yaml", org, team, userName)
		}

		commitMessage := "Created a file with request data"

		options := &github.RepositoryContentFileOptions{
			Message: github.String(commitMessage),
			Content: content,
			Branch:  github.String(branchName),
		}

		// _, _, _, err = githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s", strings.Split(repo, "/")[0]), nil)
		// if err != nil {
		// 	createRequestFile(githubClient, branchName, newRequest, targetPath, userName, options, managerName, id)
		// 	return
		// }

		// _, userDirContent, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s", strings.Split(repo, "/")[0], strings.Split(repo, "/")[1]), nil)
		// if err != nil {
		// 	createRequestFile(githubClient, branchName, newRequest, targetPath, userName, options, managerName, id)
		// 	return
		// }

		// for _, content := range userDirContent {
		// 	if content.Type != nil {
		// 		if *content.Name == fmt.Sprintf("%v.yaml", userName) {
		// 			fmt.Println("Request for this user and repository already exists!")
		// 			return
		// 		}
		// 	}
		// }

		createRequestFile(githubClient, branchName, newRequest, targetPath, userName, options, managerName, id)
		os.Exit(0)
	},
}

func createRequestFile(githubClient *github.Client, branchName string, newRequest types.RequestData, targetPath string, userName string, options *github.RepositoryContentFileOptions, reviewer string, id string) {
	_, _, err := githubClient.Repositories.GetBranch(context.Background(), configData.ORG_NAME, configData.DB_NAME, branchName, 0)
	if err != nil {
		baseRef, _, err := githubClient.Git.GetRef(context.Background(), configData.ORG_NAME, configData.DB_NAME, "refs/heads/main")
		if err != nil {
			fmt.Println("Error fetching base reference:", err)
			return
		}

		newRef := &github.Reference{
			Ref:    github.String("refs/heads/" + branchName),
			Object: &github.GitObject{SHA: baseRef.Object.SHA},
		}
		_, _, err = githubClient.Git.CreateRef(context.Background(), configData.ORG_NAME, configData.DB_NAME, newRef)
		if err != nil {
			fmt.Println("Error creating branch:", err)
			return
		}
		fmt.Println("Branch created successfully!")
	} else {
		fmt.Println("This request already exists!")
		return
	}

	_, _, err = githubClient.Repositories.CreateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, targetPath, options)
	if err != nil {
		fmt.Println("Request for this user already exists.")
		return
	}

	newPR := &github.NewPullRequest{
		Title: github.String(fmt.Sprintf("Request access - %s", id)),
		Head:  github.String(branchName),
		Base:  github.String("main"),
		Body:  github.String(fmt.Sprintf("User %s requests access. Business justification: %s", userName, newRequest.Justification)),
	}

	pr, _, err := githubClient.PullRequests.Create(context.Background(), configData.ORG_NAME, configData.DB_NAME, newPR)
	if err != nil {
		fmt.Println("Error creating pull request:", err)
		return
	}

	reviewers := github.ReviewersRequest{
		Reviewers: []string{reviewer},
	}
	_, _, err = githubClient.PullRequests.RequestReviewers(context.Background(), configData.ORG_NAME, configData.DB_NAME, pr.GetNumber(), reviewers)
	if err != nil {
		fmt.Println("Error adding reviewers:", err)
		return
	}

	fmt.Println("Request added successfully.")
	fmt.Println("ID of your request:", id)
}

func init() {
	requestCmd.Flags().StringP("user", "u", "", "User's GitHub handle for whom access is requested")
	requestCmd.Flags().StringP("repo", "r", "", "Repository name (owner/repo)")
	requestCmd.Flags().StringP("justification", "j", "", "Business justification for access")
	requestCmd.Flags().StringP("manager", "m", "", "Users's manager's GitHub handle")
	requestCmd.Flags().StringP("org", "o", "", "Organization")
	requestCmd.Flags().StringP("team", "", "", "Team name")

	requestCmd.MarkFlagsMutuallyExclusive("repo", "team")
	requestCmd.MarkFlagRequired("org")
	requestCmd.MarkFlagRequired("justification")
	requestCmd.MarkFlagRequired("manager")

	requestCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", configData.GITHUB_PAT, "GitHub personal access token")

	requestCmd.Flags().SortFlags = false
	rootCmd.AddCommand(requestCmd)
}
