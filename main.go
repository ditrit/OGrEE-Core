package main

import (
	"flag"
)

func main() {
	var verboseLevel int

	flag.IntVar(&verboseLevel, "v", 0,
		"Indicates level of debugging messages. 0 being the least, 4 is max")

	flag.Parse()

	//Pass control to repl.go
	Start(verboseLevel)
}
