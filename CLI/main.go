package main

import (
	"cli/config"
	c "cli/controllers"
	l "cli/logger"
	"cli/readline"
	"os"
	"strings"
)

func SetPrompt(user string) string {
	c.State.Prompt = "\u001b[1m\u001b[32m" + user + "@" + c.State.Customer + ":" +
		"\u001b[37;1m" + c.State.CurrPath + "\u001b[1m\u001b[32m>\u001b[0m "
	c.State.BlankPrompt = user + "@" + c.State.Customer + c.State.CurrPath + "> "
	return c.State.Prompt
}

func main() {
	conf := config.ReadConfig()

	l.InitLogs()
	c.InitConfigFilePath(conf.ConfigPath)
	c.InitHistoryFilePath(conf.HistPath)
	c.InitDebugLevel(conf.Verbose)
	c.InitTimeout(conf.UnityTimeout)
	c.InitURLs(conf.APIURL, conf.UnityURL)

	if !c.PingAPI() {
		println("Cannot reach API at", c.State.APIURL)
		return
	}

	var err error
	var apiKey string
	user, apiKey, err := c.Login(conf.User)
	if err != nil {
		println(err.Error())
		return
	} else {
		println("Successfully connected")
	}
	c.State.User = *user
	c.InitKey(apiKey)
	c.InitState(conf)
	err = InitVars(conf.Variables)
	if err != nil {
		println("Error while initializing variables :", err.Error())
		return
	}

	userShort := strings.Split(c.State.User.Email, "@")[0]

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          SetPrompt(userShort),
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
	Start(rl, userShort)
}
