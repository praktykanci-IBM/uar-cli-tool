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

type GitResponse struct {
	Content string `json:"content"`
	Sha     string `json:"sha"`
}

var idInstedOfName bool

var approveCmd = &cobra.Command{
	Use:     "approve your_id {uar_id | user repo}",
	Short:   "Approve a request",
	Aliases: []string{"a"},
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
		requestsUrl := "https://api.github.com/repos/praktykanci-IBM/user-access-records/contents/requests.json"
		res, err := http.Get(requestsUrl)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not fetch requests\n")
			os.Exit(1)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			fmt.Fprint(os.Stderr, "Could not fetch requests\n")
			os.Exit(1)
		}

		resBodyRequests, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not read response body\n")
			os.Exit(1)
		}

		var gitResponse GitResponse
		err = json.Unmarshal(resBodyRequests, &gitResponse)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not unmarshal response body\n")
			os.Exit(1)
		}

		requestsSha := gitResponse.Sha
		decoded, err := base64.StdEncoding.DecodeString(gitResponse.Content)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not decode content\n")
			os.Exit(1)
		}

		var requests Requests
		err = json.Unmarshal(decoded, &requests)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not unmarshal requests\n")
			os.Exit(1)
		}

		var approvedRequest Request
		if len(args) == 2 {
			for i, request := range requests.Requests {
				if request.ID == args[1] {
					approvedRequest = requests.Requests[i]

					requests.Requests[i] = requests.Requests[len(requests.Requests)-1]
					requests.Requests = requests.Requests[:len(requests.Requests)-1]

					break
				}
			}
		} else {
			for i, request := range requests.Requests {
				if request.Name == args[1] && request.Repo == args[2] {
					approvedRequest = requests.Requests[i]

					requests.Requests[i] = requests.Requests[len(requests.Requests)-1]
					requests.Requests = requests.Requests[:len(requests.Requests)-1]

					break
				}
			}
		}

		if approvedRequest.ID == "" {
			fmt.Fprint(os.Stderr, "Could not find request\n")
			os.Exit(1)
		}

		acceptedUrl := "https://api.github.com/repos/praktykanci-IBM/user-access-records/contents/granted.json"
		res, err = http.Get(acceptedUrl)
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

		err = json.Unmarshal(resBodyGranted, &gitResponse)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not unmarshal response body\n")
			os.Exit(1)
		}

		grantedSha := gitResponse.Sha
		decoded, err = base64.StdEncoding.DecodeString(gitResponse.Content)
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

		grantedRequests.Requests = append(grantedRequests.Requests, GrantedRequest{
			Name:          approvedRequest.Name,
			WhenRequested: approvedRequest.When,
			WhenAccepted:  time.Now().Unix(),
			Justification: approvedRequest.Justification,
			Repo:          approvedRequest.Repo,
			ID:            approvedRequest.ID,
			ApproverID:    args[0],
			AdminID:       "",
			WhenCompleted: 0,
		})

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

		req, err := http.NewRequest("PUT", acceptedUrl, bytes.NewBuffer(reqBody))
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

		if res.StatusCode != 200 {
			fmt.Fprint(os.Stderr, "Could not grant request\n")
			os.Exit(1)
		}

		requestsJson, err := json.Marshal(requests)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not marshal requests\n")
			os.Exit(1)
		}

		encodedRequests := base64.StdEncoding.EncodeToString(requestsJson)

		reqBody, err = json.Marshal(map[string]string{
			"message": "Remove request",
			"content": encodedRequests,
			"sha":     requestsSha,
		})
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not marshal request body\n")
			os.Exit(1)
		}

		req, err = http.NewRequest("PUT", requestsUrl, bytes.NewBuffer(reqBody))
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
			fmt.Fprint(os.Stderr, "Could not remove request\n")
			os.Exit(1)
		}

		fmt.Printf("Request granted successfully\n")
		fmt.Printf("ID of the request: %s\n", approvedRequest.ID)
	},
}

func init() {
	approveCmd.Flags().BoolVarP(&idInstedOfName, "id", "i", false, "Find user by ID instead of name")
	rootCmd.AddCommand(approveCmd)
}
