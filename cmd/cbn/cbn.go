package cbn

import (
	"fmt"

	"github.com/spf13/cobra"
)

var CbnCommand = &cobra.Command{
	Use:     "cbn",
	Short:   "CBN commands",
	Aliases: []string{"c"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("no arguments are required")
		}
		return nil
	},
}
