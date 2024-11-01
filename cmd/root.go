package cmd

import (
	"fmt"
	"os"
	"praktykanci/uar/cmd/cbn"
	configdata "praktykanci/uar/configData"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "uar",
	Short: "uar is a tool for managing user access records",
}

// var GITHUB_PAT string

func Execute() {
	// GITHUB_PAT = viper.GetString("GITHUB_PAT")
	configdata.Init()

	// import the subcommands
	rootCmd.AddCommand(cbn.CbnCommand)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
