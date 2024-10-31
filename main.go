package main

import (
	"praktykanci/uar/cmd"

	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	cmd.Execute()
}
