package cbn

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"
	"praktykanci/uar/types"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Complete the CBN, removing all rejected users from the repo",
	Aliases: []string{"u"},
	Run: func(cmd *cobra.Command, args []string) {
		id, err := cmd.Flags().GetString("cbn-id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		org, err := cmd.Flags().GetString("org")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if id == "" {
			id = getCbnID(org, githubClient)
		}

		cbnOriginalFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("CBN/%s.yaml", id), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cbnContentMarshaled, err := cbnOriginalFile.GetContent()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var cbnContent types.CbnData
		err = yaml.Unmarshal([]byte(cbnContentMarshaled), &cbnContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if org == "" {
			org = cbnContent.Org
		}

		executedBy, _, err := githubClient.Users.Get(context.Background(), "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		currentTime := time.Now()
		formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

		cbnContentUpdated := types.CbnData{
			StartedBy:    cbnContent.StartedBy,
			StartedOn:    cbnContent.StartedOn,
			Org:          cbnContent.Org,
			Type:         cbnContent.Type,
			ExtractedBy:  cbnContent.ExtractedBy,
			ExtractedOn:  cbnContent.ExtractedOn,
			Users:        cbnContent.Users,
			ExecutedBy:   *executedBy.Login,
			ExecutedOn:   formattedTime,
			UsersChanged: []types.CbnUser{},
		}

		for _, user := range cbnContent.Users {
			if (cbnContent.Type == "positive" && user.State == types.Pending) || (user.State == types.Rejected) {
				for _, access := range user.ListOfAccesses {
					switch access.AccessType {
					case types.Repo:
						_, err = githubClient.Repositories.RemoveCollaborator(context.Background(), org, access.AccessTo, user.Name)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s/%s.yaml", org, access.AccessTo, user.Name), nil)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestFileSHA := *requestFile.SHA
						requestFileMarshaled, err := requestFile.GetContent()
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						var requestFileContent types.RequestDataCompleted
						err = yaml.Unmarshal([]byte(requestFileMarshaled), &requestFileContent)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestCompleted := types.RequestDataCompleted{
							ID:            requestFileContent.ID,
							State:         types.Removed,
							Justification: requestFileContent.Justification,
							RequestedOn:   requestFileContent.RequestedOn,
							RequestedBy:   requestFileContent.RequestedBy,
							CompletedOn:   requestFileContent.CompletedOn,
							CompletedBy:   requestFileContent.CompletedBy,
						}

						resFileMarshaled, err := yaml.Marshal(requestCompleted)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						_, _, _ = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s/%s.yaml", org, access.AccessTo, user.Name), &github.RepositoryContentFileOptions{
							Message: github.String("Removed user"),
							Content: resFileMarshaled,
							SHA:     &requestFileSHA,
						})

					case types.Org:
						_, err := githubClient.Organizations.RemoveMember(context.Background(), access.AccessTo, user.Name)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("org-access-records/%s/%s.yaml", access.AccessTo, user.Name), nil)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestFileSHA := *requestFile.SHA
						requestFileMarshaled, err := requestFile.GetContent()
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						var requestFileContent types.RequestDataCompleted
						err = yaml.Unmarshal([]byte(requestFileMarshaled), &requestFileContent)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestCompleted := types.RequestDataCompleted{
							ID:            requestFileContent.ID,
							State:         types.Removed,
							Justification: requestFileContent.Justification,
							RequestedOn:   requestFileContent.RequestedOn,
							RequestedBy:   requestFileContent.RequestedBy,
							CompletedOn:   requestFileContent.CompletedOn,
							CompletedBy:   requestFileContent.CompletedBy,
						}

						resFileMarshaled, err := yaml.Marshal(requestCompleted)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						_, _, _ = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("org-access-records/%s/%s.yaml", access.AccessTo, user.Name), &github.RepositoryContentFileOptions{
							Message: github.String("Removed user"),
							Content: resFileMarshaled,
							SHA:     &requestFileSHA,
						})

					case types.Team:
						_, err := githubClient.Teams.RemoveTeamMembershipBySlug(context.Background(), org, access.AccessTo, user.Name)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("team-access-records/%s/%s/%s.yaml", org, access.AccessTo, user.Name), nil)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestFileSHA := *requestFile.SHA
						requestFileMarshaled, err := requestFile.GetContent()
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						var requestFileContent types.RequestDataCompleted
						err = yaml.Unmarshal([]byte(requestFileMarshaled), &requestFileContent)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						requestCompleted := types.RequestDataCompleted{
							ID:            requestFileContent.ID,
							State:         types.Removed,
							Justification: requestFileContent.Justification,
							RequestedOn:   requestFileContent.RequestedOn,
							RequestedBy:   requestFileContent.RequestedBy,
							CompletedOn:   requestFileContent.CompletedOn,
							CompletedBy:   requestFileContent.CompletedBy,
						}

						resFileMarshaled, err := yaml.Marshal(requestCompleted)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error: %v\n", err)
							os.Exit(1)
						}

						_, _, _ = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("team-access-records/%s/%s/%s.yaml", org, access.AccessTo, user.Name), &github.RepositoryContentFileOptions{
							Message: github.String("Removed user"),
							Content: resFileMarshaled,
							SHA:     &requestFileSHA,
						})
					}

					cbnContentUpdated.UsersChanged = append(cbnContentUpdated.UsersChanged, user)
				}
			}
		}

		resCbnMarshaled, err := yaml.Marshal(cbnContentUpdated)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		_, _, err = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("CBN/%s", *cbnOriginalFile.Name), &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Update collaborators list, complete CBN %s", id)),
			Content: resCbnMarshaled,
			SHA:     cbnOriginalFile.SHA,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		pullRequests, _, err := githubClient.PullRequests.List(context.Background(), configData.ORG_NAME, configData.DB_NAME, &github.PullRequestListOptions{
			State: "open",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for _, pr := range pullRequests {
			if *pr.Title == fmt.Sprintf("Validate CBN - %s", id) {
				_, _, err = githubClient.PullRequests.Edit(context.Background(), configData.ORG_NAME, configData.DB_NAME, *pr.Number, &github.PullRequest{
					State: github.String("closed"),
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}

				_, err = githubClient.Git.DeleteRef(context.Background(), configData.ORG_NAME, configData.DB_NAME, *pr.Base.Ref)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			}
		}

		fmt.Printf("CBN with ID: %s completed\n", id)
	},
}

func init() {
	updateCmd.Flags().StringP("cbn-id", "i", "", "The CBN ID to extract data for")
	updateCmd.Flags().StringP("org", "o", "", "The organisation name ")

	updateCmd.MarkFlagsMutuallyExclusive("cbn-id", "org")
	updateCmd.MarkFlagsOneRequired("cbn-id", "org")

	updateCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", "", "GitHub personal access token")

	updateCmd.Flags().SortFlags = false
	CbnCommand.AddCommand(updateCmd)
}
