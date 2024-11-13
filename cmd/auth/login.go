package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var loginCommand = &cobra.Command{
	Use:   "login new_token",
	Short: "Login to the system",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("github pat is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		newToken := args[0]

		viper.SetConfigName("config")
		viper.SetConfigType("toml")

		configDir, err := os.UserConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		configPath := filepath.Join(configDir, "uar")
		viper.AddConfigPath(configPath)

		_, err = os.Open(filepath.Join(configPath, "config.toml"))
		if err != nil {
			if os.IsNotExist(err) {
				err = os.WriteFile(filepath.Join(configPath, "config.toml"), []byte(""), 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		viper.Set("GITHUB_PAT", newToken)

		err = viper.WriteConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	AuthCommand.AddCommand(loginCommand)
}
