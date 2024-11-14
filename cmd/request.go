package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"praktykanci/uar/configData"
	. "praktykanci/uar/configData"
	. "praktykanci/uar/types"

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

		userName, _ := cmd.Flags().GetString("user")
		repo, _ := cmd.Flags().GetString("repo")
		justification, _ := cmd.Flags().GetString("justification")
		managerName, _ := cmd.Flags().GetString("manager")

		fmt.Printf("token: %s\n", GITHUB_PAT)

		if userName == "" || repo == "" || justification == "" || managerName == "" {
			fmt.Println("Error: All flags are required.")
			return
		}

		githubClient := github.NewClient(nil).WithAuthToken(GITHUB_PAT)

		user, _, err := githubClient.Users.Get(context.Background(), "")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		isCollaborator, _, err := githubClient.Repositories.IsCollaborator(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, managerName)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if !isCollaborator {
			fmt.Printf("Invalid manager name\n")
			return
		}

		_, _, err = githubClient.Users.Get(context.Background(), userName)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		_, _, err = githubClient.Repositories.Get(context.Background(), strings.Split(repo, "/")[0], strings.Split(repo, "/")[1])
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		currentTime := time.Now()
		formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

		id := uuid.New().String()
		newRequest := RequestData{
			ID:            id,
			State:         Granted,
			Justification: justification,
			RequestedOn:   formattedTime,
			RequestedBy:   *user.Login,
		}

		content, err := yaml.Marshal(&newRequest)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		branchName := fmt.Sprintf("%s/%s/%s", strings.Split(repo, "/")[0], strings.Split(repo, "/")[1], userName)
		commitMessage := "Created a file with request data"

		options := &github.RepositoryContentFileOptions{
			Message: github.String(commitMessage),
			Content: content,
			Branch:  github.String(branchName),
		}

		_, _, _, err = githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, strings.Split(repo, "/")[0], nil)
		if err != nil {
			createRequestFile(githubClient, branchName, newRequest, repo, userName, options, managerName, id)
			return
		}

		_, userDirContent, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s", strings.Split(repo, "/")[0], strings.Split(repo, "/")[1]), nil)
		if err != nil {
			createRequestFile(githubClient, branchName, newRequest, repo, userName, options, managerName, id)
			return
		}

		for _, content := range userDirContent {
			if content.Type != nil {
				if *content.Name == fmt.Sprintf("%v.yaml", userName) {
					fmt.Println("Request for this user and repository already exists!")
					return
				}
			}
		}

		createRequestFile(githubClient, branchName, newRequest, repo, userName, options, managerName, id)
		os.Exit(0)
	},
}

func createRequestFile(githubClient *github.Client, branchName string, newRequest RequestData, repo string, userName string, options *github.RepositoryContentFileOptions, reviewer string, id string) {
	_, _, err := githubClient.Repositories.GetBranch(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, branchName, 0)
	if err != nil {
		baseRef, _, err := githubClient.Git.GetRef(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, "refs/heads/main")
		if err != nil {
			fmt.Println("Error fetching base reference:", err)
			return
		}

		newRef := &github.Reference{
			Ref:    github.String("refs/heads/" + branchName),
			Object: &github.GitObject{SHA: baseRef.Object.SHA},
		}
		_, _, err = githubClient.Git.CreateRef(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, newRef)
		if err != nil {
			fmt.Println("Error creating branch:", err)
			return
		}
		fmt.Println("Branch created successfully!")
	} else {
		fmt.Println("Request for this user and repository already exists!")
		return
	}

	_, _, err = githubClient.Repositories.CreateFile(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s/%s.yaml", strings.Split(repo, "/")[0], strings.Split(repo, "/")[1], userName), options)
	if err != nil {
		fmt.Println("Error:", err)
	}

	newPR := &github.NewPullRequest{
		Title: github.String(fmt.Sprintf("Request access - %s", id)),
		Head:  github.String(branchName),
		Base:  github.String("main"),
		Body:  github.String(fmt.Sprintf("User %s requests access to repository %s. Business justification: %s", userName, repo, newRequest.Justification)),
	}

	pr, _, err := githubClient.PullRequests.Create(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, newPR)
	if err != nil {
		fmt.Println("Error creating pull request:", err)
		return
	}

	reviewers := github.ReviewersRequest{
		Reviewers: []string{reviewer},
	}
	_, _, err = githubClient.PullRequests.RequestReviewers(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, pr.GetNumber(), reviewers)
	if err != nil {
		fmt.Println("Error adding reviewers:", err)
		return
	}

	fmt.Println("Request added successfully.")
	fmt.Println("ID of your request:", id)
}

func init() {
	rootCmd.AddCommand(requestCmd)

	requestCmd.Flags().StringP("user", "u", "", "GitHub username requesting access")
	requestCmd.Flags().StringP("repo", "r", "", "Repository name (owner/repo)")
	requestCmd.Flags().StringP("justification", "j", "", "Business justification for access")
	requestCmd.Flags().StringP("manager", "m", "", "Manager's GitHub username")

	requestCmd.MarkFlagRequired("user")
	requestCmd.MarkFlagRequired("repo")
	requestCmd.MarkFlagRequired("justification")
	requestCmd.MarkFlagRequired("manager")

	requestCmd.Flags().StringVarP(&GITHUB_PAT, "token", "t", GITHUB_PAT, "GitHub personal access token")
}
