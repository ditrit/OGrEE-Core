package main

import (
	"flag"
)

func main() {
	var verboseLevel int
	var unityURL string
	var APIURL string
	var listenPORT int
	var APIKEY string
	flags := map[string]interface{}{}

	flag.IntVar(&verboseLevel, "v", 0,
		"Indicates level of debugging messages. 0 being the least, 4 is max")

	flag.StringVar(&unityURL, "unity_url", "", "Unity URL")

	flag.StringVar(&APIURL, "api_url", "", "API URL")

	flag.IntVar(&listenPORT, "listen_port", 0,
		"Indicates which port to communicate to Unity")

	flag.StringVar(&APIKEY, "api_key", "", "Indicate the key of the API")

	flag.Parse()

	flags["v"] = verboseLevel

	flags["unity_url"] = unityURL

	flags["api_url"] = APIURL

	flags["api_key"] = APIKEY

	flags["listen_port"] = listenPORT

	//Pass control to repl.go
	Start(flags)
}
