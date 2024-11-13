package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCommand = &cobra.Command{
	Use:   "status",
	Short: "Check if user is logged in",
	Run: func(cmd *cobra.Command, args []string) {
		if viper.Get("GITHUB_PAT") == nil {
			fmt.Println("You are not logged in")
		} else {
			fmt.Println("You are logged in")
		}
	},
}

func init() {
	AuthCommand.AddCommand(statusCommand)
}
