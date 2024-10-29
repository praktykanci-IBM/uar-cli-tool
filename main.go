package main

import (
	"flag"
	"fmt"
)


func main() {
	var helpFlag = flag.Bool("help",false,"display application's flags")
	flag.Parse()

	if(*helpFlag){
		fmt.Printf("Commands: \n\thelp: \tdisplay application's flags")
		return 
	}
}