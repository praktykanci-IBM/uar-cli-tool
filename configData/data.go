package configData

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var GITHUB_PAT string
var ORG_NAME string
var UAR_DB_NAME string
var CBN_DB_NAME string

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	viper.AddConfigPath(filepath.Join(configDir, "uar"))
	viper.ReadInConfig()

	GITHUB_PAT = viper.GetString("GITHUB_PAT")

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.ReadInConfig()

	envPAT := viper.GetString("GITHUB_PAT")
	if envPAT != "" {
		GITHUB_PAT = envPAT
	}

	ORG_NAME = "praktykanci-IBM"
	UAR_DB_NAME = "user-access-records"
	CBN_DB_NAME = "continuous-business-need"
}
