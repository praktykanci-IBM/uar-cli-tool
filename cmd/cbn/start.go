package cbn

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	. "praktykanci/uar/configData"
	. "praktykanci/uar/types"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var negativeRevalidation bool
var startCmd = &cobra.Command{
	Use:     "start owner_name repo",
	Short:   "Start the CBN",
	Aliases: []string{"s"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("requires owner_name and repo")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		cbnUrl := "https://api.github.com/repos/praktykanci-IBM/continuous-business-need/contents/record.json"
		res, err := http.Get(cbnUrl)
		if err != nil {
			fmt.Print("Could not fetch CBN\n")
			os.Exit(1)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			fmt.Print("Could not fetch CBN\n")
			os.Exit(1)
		}

		resBodyCbn, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Print("Could not read response body\n")
			os.Exit(1)
		}

		var gitResponse GitResponseData
		err = json.Unmarshal(resBodyCbn, &gitResponse)
		if err != nil {
			fmt.Print("Could not unmarshal response body\n")
			os.Exit(1)
		}

		cbnSha := gitResponse.Sha
		decodedCbn, err := base64.StdEncoding.DecodeString(gitResponse.Content)
		if err != nil {
			fmt.Print("Could not decode response body\n")
			os.Exit(1)
		}

		var cbns CbnArray
		err = json.Unmarshal(decodedCbn, &cbns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not unmarshal response body, %v\n", err)
			os.Exit(1)
		}

		CbnID := uuid.NewString()
		cbns.Cbns = append(cbns.Cbns, Cbn{
			CbnID:	   CbnID,
			Owner:     args[0],
			Repo:      args[1],
			IsPositive: !negativeRevalidation,
			StartDate: time.Now().Unix(),
		})

		newCbns, err := json.Marshal(cbns)
		if err != nil {
			fmt.Print("Could not marshal response body\n")
			os.Exit(1)
		}

		updateCbnsBody, err := json.Marshal(map[string]string{
			"message": "Add new CBN",
			"content": base64.StdEncoding.EncodeToString(newCbns),
			"sha":     cbnSha,
		})
		if err != nil {
			fmt.Print("Could not marshal response body\n")
			os.Exit(1)
		}

		reqUpdateCbns, err := http.NewRequest("PUT", cbnUrl, bytes.NewBuffer(updateCbnsBody))
		if err != nil {
			fmt.Print("Could not make PUT request\n")
			os.Exit(1)
		}

		fmt.Printf("pat: %s\n", GITHUB_PAT)

		reqUpdateCbns.Header.Set("Authorization", fmt.Sprintf("Bearer %s", GITHUB_PAT))
		reqUpdateCbns.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		resUpdateCbns, err := http.DefaultClient.Do(reqUpdateCbns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not send request\n")
		}
		defer resUpdateCbns.Body.Close()

		fmt.Printf("res: %v\n", resUpdateCbns)

		if resUpdateCbns.StatusCode != 200 {
			fmt.Fprintf(os.Stderr, "Could not create CBN\n")
			os.Exit(1)
		}

		fmt.Print("Started new CBN\n")
		fmt.Printf("ID of the CBN: %s\n", CbnID)
	},
}

func init() {
	startCmd.Flags().BoolVarP(&negativeRevalidation, "negative-revalidation", "n", false, "Use negative revalidation insted of positive")
	CbnCommand.AddCommand(startCmd)
}
