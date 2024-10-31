package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "uar",
	Short: "uar is a tool for managing user access records",
}

var GITHUB_PAT string

func Execute() {
	GITHUB_PAT = viper.GetString("GITHUB_PAT")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
