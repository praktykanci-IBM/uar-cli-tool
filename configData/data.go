package configData

import "github.com/spf13/viper"

var GITHUB_PAT string
var ORG_NAME string
var UAR_DB_NAME string
var CBN_DB_NAME string

func Init() {
	GITHUB_PAT = viper.GetString("GITHUB_PAT")
	ORG_NAME = "praktykanci-IBM"
	UAR_DB_NAME = "user-access-records"
	CBN_DB_NAME = "continuous-business-need"
}
