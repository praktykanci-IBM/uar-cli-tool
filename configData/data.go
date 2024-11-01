package configData

import "github.com/spf13/viper"

var GITHUB_PAT string

func Init() {
	GITHUB_PAT = viper.GetString("GITHUB_PAT")
}
