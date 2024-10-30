package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var AddCommand = &cobra.Command{
	Use:     "add your_id {uar_id | user repo}",
	Short:   "Add a user as a collaborator",
	Aliases: []string{"d"},
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
		acceptedUrl := "https://api.github.com/repos/praktykanci-IBM/user-access-records/contents/granted.json"
		res, err := http.Get(acceptedUrl)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not fetch granted requests\n")
			os.Exit(1)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			fmt.Fprint(os.Stderr, "Could not fetch granted requests\n")
			os.Exit(1)
		}

		resBodyGranted, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not read response body\n")
			os.Exit(1)
		}

		var gitResponse GitResponse
		err = json.Unmarshal(resBodyGranted, &gitResponse)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not unmarshal response body\n")
			os.Exit(1)
		}

		grantedSha := gitResponse.Sha
		decoded, err := base64.StdEncoding.DecodeString(gitResponse.Content)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not decode content\n")
			os.Exit(1)
		}

		var grantedRequests GrantedRequests
		err = json.Unmarshal(decoded, &grantedRequests)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not unmarshal granted requests\n")
			os.Exit(1)
		}

		var completedRequest GrantedRequest
		if len(args) == 2 {
			for i, request := range grantedRequests.Requests {
				if request.ID == args[1] {
					completedRequest = grantedRequests.Requests[i]

					grantedRequests.Requests[i] = GrantedRequest{
						Name:          completedRequest.Name,
						WhenRequested: completedRequest.WhenRequested,
						WhenAccepted:  completedRequest.WhenAccepted,
						Justification: completedRequest.Justification,
						Repo:          completedRequest.Repo,
						ID:            completedRequest.ID,
						ApproverID:    completedRequest.ApproverID,
						AdminID:       args[0],
						WhenCompleted: time.Now().Unix(),
					}

					break
				}
			}
		} else {
			for i, request := range grantedRequests.Requests {
				if request.Name == args[1] && request.Repo == args[2] {
					completedRequest = grantedRequests.Requests[i]

					grantedRequests.Requests[i] = GrantedRequest{
						Name:          completedRequest.Name,
						WhenRequested: completedRequest.WhenRequested,
						WhenAccepted:  completedRequest.WhenAccepted,
						Justification: completedRequest.Justification,
						Repo:          completedRequest.Repo,
						ID:            completedRequest.ID,
						ApproverID:    completedRequest.ApproverID,
						AdminID:       args[0],
						WhenCompleted: time.Now().Unix(),
					}

					break
				}
			}
		}

		if completedRequest.ID == "" {
			fmt.Fprint(os.Stderr, "Could not find request\n")
			os.Exit(1)

		}

		grantAccessUrl := fmt.Sprintf("https://api.github.com/repos/%s/collaborators/%s", completedRequest.Repo, completedRequest.Name)
		req, err := http.NewRequest("PUT", grantAccessUrl, nil)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not create request\n")
			os.Exit(1)
		}

		githubToken := os.Getenv("GITHUB_PAT")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		res, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not send request\n")
			os.Exit(1)
		}
		defer res.Body.Close()

		if res.StatusCode != 201 && res.StatusCode != 204 {
			fmt.Fprint(os.Stderr, "Could not grant access\n")
			os.Exit(1)
		}

		if res.StatusCode == 204 {
			fmt.Fprint(os.Stderr, "User already has access\n")
			os.Exit(1)
		}

		grantedJson, err := json.Marshal(grantedRequests)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not marshal granted requests\n")
			os.Exit(1)
		}

		encodedGranted := base64.StdEncoding.EncodeToString(grantedJson)

		reqBody, err := json.Marshal(map[string]string{
			"message": "Approve request",
			"content": encodedGranted,
			"sha":     grantedSha,
		})
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not marshal request body\n")
			os.Exit(1)
		}

		req, err = http.NewRequest("PUT", acceptedUrl, bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not create request\n")
			os.Exit(1)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		res, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not send request\n")
			os.Exit(1)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			fmt.Fprint(os.Stderr, "Could not add access\n")
			os.Exit(1)
		}

		fmt.Print("Access granted successfully\n")
		fmt.Printf("ID of the request: %s\n", completedRequest.ID)
	},
}

func init() {
	rootCmd.AddCommand(AddCommand)
}
