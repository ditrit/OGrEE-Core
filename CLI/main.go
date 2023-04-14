package main

import (
	"cli/config"
	c "cli/controllers"
	l "cli/logger"
	"cli/readline"
	"os"
	"strings"
)

func main() {
	conf := config.ReadConfig()

	l.InitLogs()
	c.InitConfigFilePath(conf.ConfigPath)
	c.InitHistoryFilePath(conf.HistPath)
	c.InitDebugLevel(conf.Verbose)         //Set the Debug level
	c.InitTimeout(conf.UnityTimeout)       //Set the Unity Timeout
	c.InitURLs(conf.APIURL, conf.UnityURL) //Set the URLs

	conf.User, conf.APIKEY = c.Login(conf.User, conf.APIKEY)
	c.InitEmail(conf.User) //Set the User email
	c.InitKey(conf.APIKEY) //Set the API Key
	c.InitState(conf)
	err := InitVars(conf.Variables)
	if err != nil {
		println("Error while initializing variables :", err.Error())
		return
	}

	user := strings.Split(conf.User, "@")[0]

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "\u001b[1m\u001b[32m" + user + "@" + "OGrEE3D:" +
			"\u001b[37;1m" + c.State.CurrPath + "\u001b[1m\u001b[32m$>\u001b[0m ",
		HistoryFile:     c.State.HistoryFilePath,
		AutoComplete:    GetPrefixCompleter(),
		InterruptPrompt: "^C",
		//EOFPrompt:       "exit",

		HistorySearchFold: true,
		//FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	//Allow the ShellState to hold a ptr to readline
	c.SetStateReadline(rl)

	//Execute Script if provided as arg and exit
	if conf.Script != "" {
		if strings.Contains(conf.Script, ".ocli") {
			LoadFile(conf.Script)
			os.Exit(0)
		}
	}
	c.InitUnityCom(rl, c.State.UnityClientURL)
	//Pass control to repl.go
	Start(rl, user)
}
