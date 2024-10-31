package cmd

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "praktykanci/uar/types"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var requestCmd = &cobra.Command{
	Use:     "request",
	Aliases: []string{"r"},
	Short:   "Request access to repository",
	Long:    "Request access to selected repository with user ID, repository name and business justification",
	Args:    cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {

		//check if user exists
		username := args[0]
		url := fmt.Sprintf("https://api.github.com/users/%s", username)
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		if message, ok := result["message"]; ok && message == "Not Found" {
			fmt.Println("User does not exist.")
		} else {
			//check if repo exists
			repo := args[1]

			url := fmt.Sprintf("https://api.github.com/repos/%s", repo)
			resp, err := http.Get(url)

			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			defer resp.Body.Close()

			var result map[string]interface{}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				fmt.Println("Error decoding JSON:", err)
				return
			}

			if message, ok := result["message"]; ok && message == "Not Found" {
				fmt.Println("Repo does not exist.")
			} else {
				url := "https://api.github.com/repos/praktykanci-IBM/user-access-records/contents/requests.json"

				resp, err := http.Get(url)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				defer resp.Body.Close()

				var result map[string]interface{}
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					fmt.Println("Error decoding JSON:", err)
					return
				}

				if encodedContent, ok := result["content"].(string); ok {
					decodedContent, err := base64.StdEncoding.DecodeString(encodedContent)
					if err != nil {
						fmt.Println("Error decoding Base64 content:", err)
						return
					}
					sha := result["sha"].(string)
					var data Requests
					errr := json.Unmarshal(decodedContent, &data)
					if errr != nil {
						fmt.Println("Error", errr)
						return
					}
					id := uuid.New().String()
					newRequest := Request{
						Name:          username,
						When:          time.Now().Unix(),
						Justification: args[2],
						Repo:          repo,
						ID:            id,
					}

					data.Requests = append(data.Requests, newRequest)

					jsonData, err := json.Marshal(data)
					if err != nil {
						fmt.Println("Error: ", err)
						return
					}
					encodedData := base64.StdEncoding.EncodeToString(jsonData)

					body := map[string]string{
						"message": "Add new request",
						"content": encodedData,
						"sha":     sha,
					}

					bodyData, _ := json.Marshal(body)

					req, err := http.NewRequest("PUT", url, bytes.NewBuffer(bodyData))
					if err != nil {
						fmt.Println("Error: ", err)
						return
					}
					req.Header.Set("Authorization", "Bearer "+GITHUB_PAT)
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}

					resp, err := client.Do(req)
					if err != nil {
						fmt.Println("Error making PUT request:", err)
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
						fmt.Println("Request added successfully.")
						fmt.Println("ID of your request:", newRequest.ID)
					} else {
						fmt.Printf("Failed to update file. Status: %s\n", resp.Status)
					}

				} else {
					fmt.Println("Content field not found or is not a string.")
				}

			}
		}

	},
}

func init() {
	rootCmd.AddCommand(requestCmd)
}
