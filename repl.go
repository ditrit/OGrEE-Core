package main

//This file inits the State and
//manages the interpreter and REPL
//(read eval print loop)

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

import (
	"bufio"
	c "cli/controllers"
	l "cli/logger"
	p "cli/preprocessor"
	"cli/readline"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func InterpretLine(str *string) bool {
	lex := NewLexer(strings.NewReader(*str))
	result := yyParse(lex)
	if result != 0 {
		return false
	}
	if root != nil {
		_, err := root.execute()
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				l.GetWarningLogger().Println(err.Error())
				fmt.Println(err.Error())
				return true
			} else {
				l.GetErrorLogger().Println(err.Error())
				fmt.Println("Error : " + err.Error())
				return false
			}
		}
	}
	return true
}

func validateFile(comBuf *[]map[string]int, file string) bool {
	invalidCommands := []string{}
	for i := range *comBuf {
		for k := range (*comBuf)[i] {
			lex := NewLexer(strings.NewReader(k))
			if yyAnalyse(lex) != 0 {
				invalidCommands = append(invalidCommands,
					" LINE#: "+k)
			}
		}
	}

	if len(invalidCommands) > 0 {
		println("Syntax errors were found in the file: ", file)
		println("The following commands were invalid")
		for i := range invalidCommands {
			println(invalidCommands[i])
		}
		return false
	}
	return true
}

func executeFile(comBuf *[]map[string]int, file string) {
	for i := range *comBuf {
		for st := range (*comBuf)[i] {
			c.State.LineNumber = (*comBuf)[i][st]
			if InterpretLine(&st) == false {
				//println("Command: ", st)
				return
			}
		}
	}
}

func loadFile(path string) {
	originalPath := path
	newBackup := p.ProcessFile(path, c.State.DebugLvl)
	file, err := os.Open(newBackup)
	if err != nil {
		if c.State.DebugLvl > 0 {
			println("Error:", err.Error())
		}

		l.GetWarningLogger().Println("Error:", err)
	}
	defer file.Close()
	fullcom := ""
	keepScanning := false
	scanner := bufio.NewScanner(file)
	c.State.LineNumber = 1 //Indicate Line Number
	commandBuffer := []map[string]int{}

	for scanner.Scan() {
		x := scanner.Text()
		if len(x) > 0 {
			if commentIdx := strings.Index(x, "//"); commentIdx != -1 { //Comment found
				fullcom += x[:commentIdx]
			} else if string(x[len(x)-1]) == "\\" {
				fullcom += x
				keepScanning = true
			} else if keepScanning == true {
				fullcom += x
				//InterpretLine(&fullcom)
				commandBuffer = append(commandBuffer,
					map[string]int{fullcom: c.State.LineNumber})
				keepScanning = false
				fullcom = ""
			} else {
				//InterpretLine(&x)
				commandBuffer = append(commandBuffer,
					map[string]int{x: c.State.LineNumber})
			}
		}

		if originalPath != c.State.ScriptPath { //Nested Execution
			loadFile(c.State.ScriptPath)
			c.State.ScriptPath = originalPath
		}

		c.State.LineNumber++ //Increment
	}

	//Validate the commandbuffer
	fName := filepath.Base(path)
	if c.State.Analyser == true {
		if validateFile(&commandBuffer, fName) == true {
			executeFile(&commandBuffer, fName)
		}
	} else {
		executeFile(&commandBuffer, fName)
	}

	ResetStateScriptData()
}

func ResetStateScriptData() {
	//Reset
	c.State.LineNumber = 0
	c.State.ScriptCalled = false
}

//Init the Shell
func Start(flags map[string]interface{}) {

	env := map[string]interface{}{}

	l.InitLogs()
	c.InitEnvFilePath(flags)
	c.InitHistoryFilePath(flags)
	c.InitDebugLevel(flags) //Set the Debug level
	c.LoadEnvFile(env, flags["env_path"].(string))
	c.InitTimeout(env)    //Set the Unity Timeout
	c.GetURLs(flags, env) //Set the URLs
	c.InitKey(flags, env) //Set the API Key
	user, _ := c.Login(env)

	c.InitState(flags, env)

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
	if flags["script"] != "" {
		if strings.Contains(flags["script"].(string), ".ocli") {
			c.State.ScriptCalled = true
			c.State.ScriptPath = flags["script"].(string)
			loadFile(flags["script"].(string))
			os.Exit(0)
		}
	}
	c.InitUnityCom(rl, c.State.UnityClientURL)

	Repl(rl, user)
}

//The loop of the program
func Repl(rl *readline.Instance, user string) {
	for {
		if c.State.ScriptCalled == true {
			//Load the path and
			//call interpret line
			loadFile(c.State.ScriptPath)
			c.State.ScriptCalled = false
		} else {
			line, err := rl.Readline()
			if err != nil { // io.EOF
				break
			}
			InterpretLine(&line)
		}

		//c.UpdateSessionState(&line)
		//Update Prompt
		rl.SetPrompt("\u001b[1m\u001b[32m" + user + "@" + "OGrEE3D:" +
			"\u001b[37;1m" + c.State.CurrPath + "\u001b[1m\u001b[32m$>\u001b[0m ")
	}
}
