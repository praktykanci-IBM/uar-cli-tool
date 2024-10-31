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

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:     "start owner_name repo kind_of_revalidation",
	Short:   "Start the CBN",
	Aliases: []string{"s"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("requires owner_name, repo and kind_of_revalidation")
		}

		if args[2] != "positive" && args[2] != "negative" {
			return fmt.Errorf("kind_of_revalidation must be either positive or negative")

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

		var cbns []Cbn
		err = json.Unmarshal(decodedCbn, &cbns)
		if err != nil {
			fmt.Print("Could not unmarshal response body\n")
			os.Exit(1)
		}

		cbns = append(cbns, Cbn{
			Owner:     args[0],
			Repo:      args[1],
			Kind:      args[2],
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

		req, err := http.NewRequest("PUT", cbnUrl, bytes.NewBuffer(updateCbnsBody))
		if err != nil {
			fmt.Print("Could not make PUT request\n")
			os.Exit(1)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", GITHUB_PAT))

	},
}

func init() {
	CbnCommand.AddCommand(startCmd)
}
