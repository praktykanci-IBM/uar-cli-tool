package configdata

import "github.com/spf13/viper"

var GITHUB_PAT = viper.GetString("GITHUB_PAT")
