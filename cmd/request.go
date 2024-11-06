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
	Use:     "request user_name repo justification manager_name",
	Aliases: []string{"r"},
	Short:   "Request access to repository",
	Long:    "Request access to selected repository with user ID, repository name, business justification and manager name",
	Args:    cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {

		githubClient := github.NewClient(nil).WithAuthToken(GITHUB_PAT)

        isCollaborator, _, err := githubClient.Repositories.IsCollaborator(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, args[3])
        if err != nil {
            fmt.Println("Error: ",err)
        }

        if !isCollaborator {
            fmt.Printf("Invalid manager name")
        }

		_, _, err = githubClient.Users.Get(context.Background(), args[0])

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		_, _, err = githubClient.Repositories.Get(context.Background(), strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1])

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		id := uuid.New().String()

		newRequest := RequestData{
			ID:    id,
			Added: false,
            Justification: args[2],
            WhenRequested: time.Now().Unix(),
		}

		content, err := yaml.Marshal(&newRequest)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Check if branch exists
		branchName := fmt.Sprintf("%s/%s/%s", strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1], args[0]) 
		

		commitMessage := "Created a file with request data"
		options := &github.RepositoryContentFileOptions{
			Message: github.String(commitMessage),
			Content: content,
			Branch:  github.String(fmt.Sprintf("%s/%s/%s", strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1], args[0])),
		}

		newPR := &github.NewPullRequest{
			Title: github.String(fmt.Sprintf("Request access - %s", id)),
			Head:  github.String(fmt.Sprintf("%s/%s/%s", strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1], args[0])), 
			Base:  github.String("main"),                                                                                           
			Body:  github.String(fmt.Sprintf("User %s requests access to repository %s. Business justification: %s", args[0],args[1], args[2])),
		}

		_, _, _, err = githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, strings.Split(args[1], "/")[0], nil)

		if err != nil {

            
            _, _, err = githubClient.Repositories.GetBranch(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, branchName, 0)
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

			_, _, err := githubClient.Repositories.CreateFile(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s/%s.yaml", strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1], args[0]), options)
			if err != nil {
				fmt.Println("Error:", err)
			}
			pr, _, err := githubClient.PullRequests.Create(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, newPR)
			if err != nil {
				fmt.Println("Error creating pull request:", err)
				return
			}

            reviewers := github.ReviewersRequest{
                Reviewers: []string{args[3]},
            }
            _, _, err = githubClient.PullRequests.RequestReviewers(context.Background(), "praktykanci-IBM", "user-access-records", pr.GetNumber(), reviewers)
            if err != nil {
                fmt.Println("Error adding reviewers:", err)
                return
            }


			fmt.Println("Request added successfully.")
			fmt.Println("ID of your request:", id)

			return
		}

		_, userDirContent, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s", strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1]), nil)

		if err != nil {
            _, _, err = githubClient.Repositories.GetBranch(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, branchName, 0)
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

			_, _, err := githubClient.Repositories.CreateFile(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s/%s.yaml", strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1], args[0]), options)
			if err != nil {
				fmt.Println("Error:", err)
			}

			pr, _, err := githubClient.PullRequests.Create(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, newPR)
			if err != nil {
				fmt.Println("Error creating pull request:", err)
				return
			}

            reviewers := github.ReviewersRequest{
                Reviewers: []string{args[3]},
            }
            _, _, err = githubClient.PullRequests.RequestReviewers(context.Background(), "praktykanci-IBM", "user-access-records", pr.GetNumber(), reviewers)
            if err != nil {
                fmt.Println("Error adding reviewers:", err)
                return
            }
			fmt.Println("Request added successfully.")
			fmt.Println("ID of your request:", id)

			return
		}

		for _, content := range userDirContent {
			if content.Type != nil {
				if *content.Name == fmt.Sprintf("%v.yaml", args[0]) {
					fmt.Println("Request for this user and repository already exists!")
					return
				}
			}
		}


        _, _, err = githubClient.Repositories.GetBranch(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, branchName, 0)
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

		_, _, err = githubClient.Repositories.CreateFile(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, fmt.Sprintf("%s/%s/%s.yaml", strings.Split(args[1], "/")[0], strings.Split(args[1], "/")[1], args[0]), options)
		if err != nil {
			fmt.Println("Error:", err)
		}

		pr, _, err := githubClient.PullRequests.Create(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, newPR)
			if err != nil {
				fmt.Println("Error creating pull request:", err)
				return
			}

            reviewers := github.ReviewersRequest{
                Reviewers: []string{args[3]},
            }
            _, _, err = githubClient.PullRequests.RequestReviewers(context.Background(), "praktykanci-IBM", "user-access-records", pr.GetNumber(), reviewers)
            if err != nil {
                fmt.Println("Error adding reviewers:", err)
                return
            }

		fmt.Println("Request added successfully.")
		fmt.Println("ID of your request:", id)

		os.Exit(0)

	},
}

func init() {
	rootCmd.AddCommand(requestCmd)
}
