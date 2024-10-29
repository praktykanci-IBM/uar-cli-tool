package main

import (
	"flag"
	"fmt"
)


func main() {
	var helpFlag = flag.Bool("help",false,"display application flags") //help flag

	flag.BoolVar(helpFlag,"h",false,"alias for help flag") //alias h for help flag

	flag.Parse()

	if *helpFlag {
		fmt.Println("Commands:")
		flag.VisitAll(func(f *flag.Flag) {
			fmt.Printf("\t%s\t %s\n", f.Name, f.Usage)
		})
		return
	}
}