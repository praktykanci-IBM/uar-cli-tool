package main

import (
	"fmt"
	"praktykanci/uar/cmd"

	"github.com/spf13/viper"
)

func main() {
	// err := godotenv.Load()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	// 	os.Exit(1)
	// }

	viper.SetConfigFile(".env")
	viper.ReadInConfig()

	test := viper.Get("GITHUB_PAT")
	fmt.Println(test)
	
	cmd.Execute()
}
