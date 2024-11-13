package configData

import (
	"os"
	"runtime"

	"github.com/spf13/viper"
)

var GITHUB_PAT string
var ORG_NAME string
var UAR_DB_NAME string
var CBN_DB_NAME string

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	if runtime.GOOS == "windows" {
		configDir, _ := os.UserConfigDir()
		configDir = configDir + "\\uar"

		viper.AddConfigPath(configDir)
	} else {
		viper.AddConfigPath("~/.config/uar")
	}

	viper.ReadInConfig()
	GITHUB_PAT = viper.GetString("GITHUB_PAT")

	// GITHUB_PAT = viper.GetString("GITHUB_PAT")
	// GITHUB_PAT = viper.GetString("GITHUB_PAT")
	ORG_NAME = "praktykanci-IBM"
	UAR_DB_NAME = "user-access-records"
	CBN_DB_NAME = "continuous-business-need"
}
