package main

import "cli/config"

func main() {
	conf := config.ReadConfig()
	//Pass control to repl.go
	Start(conf)
}
