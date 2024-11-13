package auth

import (
	"fmt"

	"github.com/spf13/cobra"
)

var AuthCommand = &cobra.Command{
	Use:   "auth",
	Short: "Auth commands",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("no arguments are required")
		}
		return nil
	},
}
