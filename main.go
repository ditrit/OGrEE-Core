package main

import (
	"flag"
)

func main() {
	var verboseLevel, listenPORT int
	var unityURL, APIURL, APIKEY, envPath, histPath string

	flags := map[string]interface{}{}

	flag.IntVar(&verboseLevel, "v", 0,
		"Indicates level of debugging messages. 0 being the least, 4 is max")

	flag.StringVar(&unityURL, "unity_url", "", "Unity URL")

	flag.StringVar(&APIURL, "api_url", "", "API URL")

	flag.IntVar(&listenPORT, "listen_port", 0,
		"Indicates which port to communicate to Unity")

	flag.StringVar(&APIKEY, "api_key", "", "Indicate the key of the API")

	flag.StringVar(&envPath, "env_path", "./.resources/.env",
		"Indicate the location of the Shell's env file")

	flag.StringVar(&histPath, "history_path", "./.history",
		"Indicate the location of the Shell's history file")

	flag.Parse()

	flags["v"] = verboseLevel

	flags["unity_url"] = unityURL

	flags["api_url"] = APIURL

	flags["api_key"] = APIKEY

	flags["listen_port"] = listenPORT
	flags["env_path"] = envPath
	flags["history_path"] = histPath

	//Pass control to repl.go
	Start(flags)
}
