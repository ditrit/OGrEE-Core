package main

import (
	"flag"
)

type Flags struct {
	verbose    string
	unityURL   string
	APIURL     string
	APIKEY     string
	listenPort int
	envPath    string
	histPath   string
	analyser   string
	script     string
}

// Assign value to flag with preference to 'x'
func NonDefault[T comparable](x, y, defaultValue T) T {
	if x != defaultValue {
		return x
	} else if y != defaultValue {
		return y
	} else {
		return defaultValue
	}
}

func main() {
	var listenPORT, l int
	var verboseLevel, v, unityURL, u, APIURL, a, APIKEY, k,
		envPath, e, histPath, h, analyse, s, file, f string

	flag.StringVar(&v, "v", "ERROR",
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")

	flag.StringVar(&verboseLevel, "verbose", "ERROR",
		"Indicates level of debugging messages."+
			"The levels are of in ascending order:"+
			"{NONE,ERROR,WARNING,INFO,DEBUG}.")

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

	var flags Flags
	flags.verbose = NonDefault(v, verboseLevel, "ERROR")
	flags.unityURL = NonDefault(u, unityURL, "")
	flags.APIURL = NonDefault(a, APIURL, "")
	flags.APIKEY = NonDefault(k, APIKEY, "")
	flags.listenPort = NonDefault(l, listenPORT, 0)
	flags.envPath = NonDefault(e, envPath, "./.env")
	flags.histPath = NonDefault(h, histPath, "./.history")
	flags.analyser = NonDefault(s, analyse, "true")
	flags.script = NonDefault(f, file, "")
	//Pass control to repl.go
	Start(&flags)
}
