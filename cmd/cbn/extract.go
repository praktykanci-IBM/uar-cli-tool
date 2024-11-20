package cbn

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"
	"praktykanci/uar/types"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var extractCmd = &cobra.Command{
	Use:     "extract",
	Short:   "Extract data for the CBN",
	Aliases: []string{"e"},
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cbnID, err := cmd.Flags().GetString("cbn-id")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		orgName, err := cmd.Flags().GetString("org")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if cbnID == "" {
			cbnID = getCbnID(orgName, githubClient)
		}

		cbnOriginalFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("CBN/%s.yaml", cbnID), nil)
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

		cbnContent.Users = []types.CbnUser{}
		_, repos, res, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s", cbnContent.Org), nil)
		if err != nil {
			if res.StatusCode == 404 {
				fmt.Fprintf(os.Stderr, "No access records for such repository\n")
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}

		for _, repo := range repos {
			_, usersWithAccess, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s", cbnContent.Org, *repo.Name), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err) // no such request
				os.Exit(1)
			}

			for _, user := range usersWithAccess {

				requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s/%s", cbnContent.Org, *repo.Name, *user.Name), nil)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err) // no such request
					os.Exit(1)
				}
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

				if requestFileContent.State != types.Removed {
					userInList := false
					for i, u := range cbnContent.Users {
						if fmt.Sprintf("%s.yaml", u.Name) == *requestFile.Name {
							userInList = true
							cbnContent.Users[i].ListOfAccesses = append(cbnContent.Users[i].ListOfAccesses, types.UserAccess{
								AccessType:    types.Repo,
								AccessTo:      *repo.Name,
								Justification: requestFileContent.Justification,
							})
						}
					}
					if !userInList {
						newUser := types.CbnUser{
							Name:           strings.Split(*user.Name, ".")[0],
							State:          types.Pending,
							ListOfAccesses: []types.UserAccess{},
							ValidatedOn:    "",
							ValidatedBy:    "",
						}
						newUser.ListOfAccesses = append(newUser.ListOfAccesses, types.UserAccess{
							AccessType:    types.Repo,
							AccessTo:      *repo.Name,
							Justification: requestFileContent.Justification,
						})
						cbnContent.Users = append(cbnContent.Users, newUser)
					}

				}

			}

		}

		_, usersWithAccessOrg, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("org-access-records/%s", cbnContent.Org), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err) // no such request
			os.Exit(1)
		}

		for _, user := range usersWithAccessOrg {

			requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("org-access-records/%s/%s", cbnContent.Org, *user.Name), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err) // no such request
				os.Exit(1)
			}
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

			if requestFileContent.State != types.Removed {
				userInList := false
				for i, u := range cbnContent.Users {
					if fmt.Sprintf("%s.yaml", u.Name) == *requestFile.Name {
						userInList = true
						cbnContent.Users[i].ListOfAccesses = append(cbnContent.Users[i].ListOfAccesses, types.UserAccess{
							AccessType:    types.Org,
							AccessTo:      cbnContent.Org,
							Justification: requestFileContent.Justification,
						})
					}
				}
				if !userInList {
					newUser := types.CbnUser{
						Name:           strings.Split(*user.Name, ".")[0],
						State:          types.Pending,
						ListOfAccesses: []types.UserAccess{},
						ValidatedOn:    "",
						ValidatedBy:    "",
					}
					newUser.ListOfAccesses = append(newUser.ListOfAccesses, types.UserAccess{
						AccessType:    types.Org,
						AccessTo:      cbnContent.Org,
						Justification: requestFileContent.Justification,
					})
					cbnContent.Users = append(cbnContent.Users, newUser)
				}

			}

		}

		_, teams, res, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("team-access-records/%s", cbnContent.Org), nil)
		if err != nil {
			if res.StatusCode == 404 {
				fmt.Fprintf(os.Stderr, "No access records for such repository\n")
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}

		for _, team := range teams {
			_, usersWithAccess, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("team-access-records/%s/%s", cbnContent.Org, *team.Name), nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err) // no such request
				os.Exit(1)
			}

			for _, user := range usersWithAccess {

				requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("team-access-records/%s/%s/%s", cbnContent.Org, *team.Name, *user.Name), nil)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err) // no such request
					os.Exit(1)
				}
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

				if requestFileContent.State != types.Removed {
					userInList := false
					for i, u := range cbnContent.Users {
						if fmt.Sprintf("%s.yaml", u.Name) == *requestFile.Name {
							userInList = true
							cbnContent.Users[i].ListOfAccesses = append(cbnContent.Users[i].ListOfAccesses, types.UserAccess{
								AccessType:    types.Team,
								AccessTo:      *team.Name,
								Justification: requestFileContent.Justification,
							})
						}
					}
					if !userInList {
						newUser := types.CbnUser{
							Name:           strings.Split(*user.Name, ".")[0],
							State:          types.Pending,
							ListOfAccesses: []types.UserAccess{},
							ValidatedOn:    "",
							ValidatedBy:    "",
						}
						newUser.ListOfAccesses = append(newUser.ListOfAccesses, types.UserAccess{
							AccessType:    types.Team,
							AccessTo:      *team.Name,
							Justification: requestFileContent.Justification,
						})
						cbnContent.Users = append(cbnContent.Users, newUser)
					}

				}

			}

		}

		currentTime := time.Now()
		formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

		extractedBy, _, err := githubClient.Users.Get(context.Background(), "")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		cbnContent.ExtractedBy = *extractedBy.Login
		cbnContent.ExtractedOn = formattedTime

		resCbnMarshaled, err := yaml.Marshal(cbnContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		a, _, err := githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("CBN/%s", *cbnOriginalFile.Name), &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Extract data for the CBN %s", cbnID)),
			Content: resCbnMarshaled,
			SHA:     cbnOriginalFile.SHA,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		sha := a.Content.SHA

		for _, user := range cbnContent.Users {
			var prCbnContent types.CbnData
			tmp, err := yaml.Marshal(cbnContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			err = yaml.Unmarshal(tmp, &prCbnContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			for i, userInCbn := range prCbnContent.Users {
				if user.Name == userInCbn.Name {
					if cbnContent.Type == "positive" {
						prCbnContent.Users[i].State = types.Aproved
					} else {
						prCbnContent.Users[i].State = types.Rejected
					}

					break
				}
			}

			branchName := fmt.Sprintf("CBN/%s/%s", cbnContent.Org, user.Name)

			_, _, err = githubClient.Repositories.GetBranch(context.Background(), configData.ORG_NAME, configData.DB_NAME, branchName, 0)
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

			prCbnMarshaled, err := yaml.Marshal(prCbnContent)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			_, _, err = githubClient.Repositories.UpdateFile(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("CBN/%s", *cbnOriginalFile.Name), &github.RepositoryContentFileOptions{
				Message: github.String(fmt.Sprintf("Extract data for the CBN %s", cbnID)),
				Content: prCbnMarshaled,
				SHA:     sha,
				Branch:  github.String(branchName),
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			newPR := &github.NewPullRequest{
				Title: github.String(fmt.Sprintf("Validate CBN - %s", cbnID)),
				Head:  github.String(branchName),
				Base:  github.String("main"),
				Body:  github.String("Validate user"),
			}

			_, _, err = githubClient.PullRequests.Create(context.Background(), configData.ORG_NAME, configData.DB_NAME, newPR)
			if err != nil {
				fmt.Println("Error creating pull request:", err)
				return
			}

		}

		fmt.Printf("Data extracted for CBN with ID: %s\n", cbnID)
	},
}

func init() {
	extractCmd.Flags().StringP("cbn-id", "i", "", "The CBN ID to extract data for")
	extractCmd.Flags().StringP("org", "o", "", "The organization name ")

	extractCmd.MarkFlagsMutuallyExclusive("cbn-id", "org")
	extractCmd.MarkFlagsOneRequired("cbn-id", "org")

	extractCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", "", "GitHub personal access token")

	extractCmd.Flags().SortFlags = false
	CbnCommand.AddCommand(extractCmd)
}
