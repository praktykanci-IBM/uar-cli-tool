package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var requestCmd = &cobra.Command{
    Use:     "request",
    Aliases: []string{"addition"},
    Short:   "Request access to repository",
    Long:    "Request access to selected repository with user ID, repository name and business justification",
    Args:    cobra.ExactArgs(3),
    Run: func(cmd *cobra.Command, args []string) {

		//check if user exists
		username:= args[0]
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
			repo:= args[1]

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
				//user and repo exist
				// fmt.Printf("Input data is %s %s %s.", args[0], args[1], args[2])
				// url := "https://api.github.com/repos/praktykanci-IBM/user-access-records/contents/awaiting.json"

				
			}
		}

    },
}

func init() {
    rootCmd.AddCommand(requestCmd)
}