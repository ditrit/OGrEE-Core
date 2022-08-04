package main

import (
	"flag"
)

//Assign value to flags[key] with preference to 'x'
func SetArgFlags(x, y, defaultValue interface{}, key string, flags map[string]interface{}) {
	if x != defaultValue {
		flags[key] = x
	} else if y != defaultValue {
		flags[key] = y
	} else {
		flags[key] = defaultValue
	}
}

func main() {
	var listenPORT, l int
	var verboseLevel, v, unityURL, u, APIURL, a, APIKEY, k,
		envPath, e, histPath, h, analyse, s, file, f string

	flags := map[string]interface{}{}

	flag.StringVar(&v, "v", "ERROR",
		"Indicates level of debugging messages. 0 being the least, 4 is max")

	flag.StringVar(&verboseLevel, "verbose", "ERROR",
		"Indicates level of debugging messages. 0 being the least, 4 is max")

	flag.StringVar(&unityURL, "unity_url", "", "Unity URL")
	flag.StringVar(&u, "u", "", "Unity URL")

	flag.StringVar(&APIURL, "api_url", "", "API URL")
	flag.StringVar(&a, "a", "", "API URL")

	flag.IntVar(&listenPORT, "listen_port", 0,
		"Indicates which port to communicate to Unity")
	flag.IntVar(&l, "l", 0,
		"Indicates which port to communicate to Unity")

	flag.StringVar(&APIKEY, "api_key", "", "Indicate the key of the API")
	flag.StringVar(&k, "k", "", "Indicate the key of the API")

	flag.StringVar(&envPath, "env_path", "./.env",
		"Indicate the location of the Shell's env file")
	flag.StringVar(&e, "e", "./.env",
		"Indicate the location of the Shell's env file")

	flag.StringVar(&histPath, "history_path", "./.history",
		"Indicate the location of the Shell's history file")
	flag.StringVar(&h, "h", "./.history",
		"Indicate the location of the Shell's history file")

	flag.StringVar(&analyse, "analyser", "true", "Dictate if the Shell shall"+
		" use the Static Analyser before script execution")
	flag.StringVar(&s, "s", "true", "Dictate if the Shell shall"+
		" use the Static Analyser before script execution")

	flag.StringVar(&file, "file", "", "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")
	flag.StringVar(&f, "f", "", "Launch the shell as an interpreter "+
		" by only executing an OCLI script file")

	flag.Parse()

	if v == "ERROR" {
		flags["v"] = 1
	} else {
		switch v {
		case "NONE":
			flags["v"] = 0
		case "WARNING":
			flags["v"] = 2
		case "INFO":
			flags["v"] = 3
		case "DEBUG":
			flags["v"] = 4
		default:
			switch verboseLevel {
			case "NONE":
				flags["v"] = 0
			case "WARNING":
				flags["v"] = 2
			case "INFO":
				flags["v"] = 3
			case "DEBUG":
				flags["v"] = 4
			default:
				flags["v"] = 1
			}
		}
	}

	SetArgFlags(u, unityURL, "", "unity_url", flags)
	SetArgFlags(a, APIURL, "", "api_url", flags)
	SetArgFlags(k, APIKEY, "", "api_key", flags)
	SetArgFlags(l, listenPORT, 0, "listen_port", flags)
	SetArgFlags(e, envPath, "./.env", "env_path", flags)
	SetArgFlags(h, histPath, "./.history", "history_path", flags)
	SetArgFlags(s, analyse, "true", "analyser", flags)
	SetArgFlags(f, file, "", "script", flags)

	//Pass control to repl.go
	Start(flags)
}
