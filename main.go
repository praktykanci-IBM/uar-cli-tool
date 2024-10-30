package main

import (
	"fmt"
	"os"
	"praktykanci/uar/cmd"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	
	cmd.Execute()
}
