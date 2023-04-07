package main

//This file inits the State and
//manages the interpreter and REPL
//(read eval print loop)

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

import (
	"cli/config"
	c "cli/controllers"
	l "cli/logger"
	"cli/readline"
	"fmt"
	"os"
	"strings"
)

func InterpretLine(str string) {
	root, parseErr := Parse(str)
	if parseErr != nil {
		fmt.Println(parseErr.Error())
		return
	}
	if root == nil {
		return
	}
	_, err := root.execute()
	if err != nil {
		l.GetErrorLogger().Println(err.Error())
		if c.State.DebugLvl > c.NONE {
			if traceErr, ok := err.(*stackTraceError); ok {
				fmt.Println(traceErr.Error())
			} else {
				fmt.Println("Error : " + err.Error())
			}
		}
	}
}

// Init the Shell
func Start(conf *config.Config) {
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

	Repl(rl, user)
}

// The loop of the program
func Repl(rl *readline.Instance, user string) {
	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		InterpretLine(line)
		//c.UpdateSessionState(&line)
		//Update Prompt
		rl.SetPrompt("\u001b[1m\u001b[32m" + user + "@" + "OGrEE3D:" +
			"\u001b[37;1m" + c.State.CurrPath + "\u001b[1m\u001b[32m$>\u001b[0m ")
	}
}
