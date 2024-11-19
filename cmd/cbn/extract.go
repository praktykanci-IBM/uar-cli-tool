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

		repoName, err := cmd.Flags().GetString("repo")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		githubClient := github.NewClient(nil).WithAuthToken(configData.GITHUB_PAT)

		if cbnID == "" {
			cbnID = getCbnID(repoName, githubClient)
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

		_, usersWithAccess, res, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s", cbnContent.Repo), nil)
		if err != nil {
			if res.StatusCode == 404 {
				fmt.Fprintf(os.Stderr, "No access records for such repository\n")
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}

		cbnContent.Users = []types.CbnUser{}
		for _, user := range usersWithAccess {

			requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s/%s", strings.Split(cbnContent.Repo, "/")[0], strings.Split(cbnContent.Repo, "/")[1], *user.Name), nil)
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
				cbnContent.Users = append(cbnContent.Users, types.CbnUser{
					Name:        strings.Split(*user.Name, ".")[0],
					State:       types.Pending,
					ValidatedOn: "",
					ValidatedBy: "",
				})
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

		for _, user := range usersWithAccess {

			requestFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("user-access-records/%s/%s/%s", strings.Split(cbnContent.Repo, "/")[0], strings.Split(cbnContent.Repo, "/")[1], *user.Name), nil)
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
				// fmt.Println(*user.Name, userInCbn.Name)
				if *user.Name == fmt.Sprintf("%s.yaml", userInCbn.Name) {
					if cbnContent.Type == "positive" {
						prCbnContent.Users[i].State = types.Aproved
					} else {
						prCbnContent.Users[i].State = types.Rejected
					}
					fmt.Println("zmiana")
					fmt.Println(prCbnContent.Users)
					fmt.Println(cbnContent.Users)
					// currentTime := time.Now()
					// formattedTime := currentTime.Format("02.01.2006, 15:04 MST")

					// validatedBy, _, err := githubClient.Users.Get(context.Background(), "")
					// if err != nil {
					// 	fmt.Println("Error:", err)
					// 	return
					// }

					// cbnContent.Users[i].ValidatedBy = *validatedBy.Login
					// cbnContent.Users[i].ValidatedOn = formattedTime

					break
				}
			}

			branchName := fmt.Sprintf("CBN/%s/%s", cbnContent.Repo, *user.Name)

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

				fmt.Println(prCbnContent.Users)
				fmt.Println(cbnContent.Users)
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
	extractCmd.Flags().StringP("repo", "r", "", "The repository name ")

	extractCmd.MarkFlagsMutuallyExclusive("cbn-id", "repo")
	extractCmd.MarkFlagsOneRequired("cbn-id", "repo")

	extractCmd.Flags().StringVarP(&configData.GITHUB_PAT, "token", "t", "", "GitHub personal access token")

	extractCmd.Flags().SortFlags = false
	CbnCommand.AddCommand(extractCmd)
}
