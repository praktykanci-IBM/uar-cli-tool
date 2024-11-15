package cbn

import (
	"context"
	"fmt"
	"os"
	"praktykanci/uar/configData"
	"praktykanci/uar/types"
	"strings"

	"github.com/google/go-github/v66/github"
	"gopkg.in/yaml.v3"
)

func getCbnID(repoName string, githubClient *github.Client) string {
	_, currentCbns, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, "CBN", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var cbnID string
	for _, cbn := range currentCbns {
		cbnFile, _, _, err := githubClient.Repositories.GetContents(context.Background(), configData.ORG_NAME, configData.DB_NAME, fmt.Sprintf("CBN/%s", *cbn.Name), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cbnContentMarshaled, err := cbnFile.GetContent()
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

		if cbnContent.Repo == repoName && cbnContent.ExecutedBy == "" {
			cbnID = strings.Split(*cbn.Name, ".")[0]
			break
		}

	}

	if cbnID == "" {
		fmt.Fprintf(os.Stderr, "No active CBN for this repository\n")
		os.Exit(1)
	}

	return cbnID
}
