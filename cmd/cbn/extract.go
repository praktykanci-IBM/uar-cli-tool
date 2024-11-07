package cbn

import (
	"fmt"

	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:     "extract owner_name repo",
	Short:   "Extract data for the CBN",
	Aliases: []string{"e"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("requires owner_name and repo")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}
