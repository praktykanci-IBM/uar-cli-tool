package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"praktykanci/uar/configData"
	. "praktykanci/uar/configData"
	. "praktykanci/uar/types"

	"github.com/google/go-github/v66/github"
	"github.com/spf13/cobra"
)

var testCommand = &cobra.Command{
	Use:   "test",
	Short: "Test command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Test command")
		githubClient := github.NewClient(nil).WithAuthToken(GITHUB_PAT)
		fContent, dContent, res, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.UAR_DB_NAME, "granted.json", nil)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not fetch granted requests\n")
			os.Exit(1)
		}

		fmt.Printf("File content: %s\n\n", fContent)
		fmt.Printf("Directory content: %s\n\n", dContent)
		fmt.Printf("Response: %v\n\n", res)

		decodedContent, err := base64.StdEncoding.DecodeString(*fContent.Content)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not decode content\n")
			os.Exit(1)
		}

		fmt.Printf("Decoded content: %s\n\n", decodedContent)

		var unmarshaledContent GrantedRequests
		err = json.Unmarshal(decodedContent, &unmarshaledContent)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not unmarshal response body\n")
			os.Exit(1)
		}

		fmt.Printf("Unmarshaled content: %v\n\n", unmarshaledContent)

	},
}

func init() {
	rootCmd.AddCommand(testCommand)
}
