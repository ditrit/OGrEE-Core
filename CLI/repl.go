package main

//This file inits the State and
//manages the interpreter and REPL
//(read eval print loop)

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

import (
	c "cli/controllers"
	l "cli/logger"
	"cli/readline"
	"fmt"
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
		manageError(err, true)
	}
}

func manageError(err error, addErrorPrefix bool) {
	l.GetErrorLogger().Println(err.Error())
	if c.State.DebugLvl > c.NONE {
		if traceErr, ok := err.(*stackTraceError); ok {
			fmt.Println(traceErr.Error())
		} else if errWithInternalErr, ok := err.(c.ErrorWithInternalError); ok {
			printError(errWithInternalErr.UserError, addErrorPrefix)
			if c.State.DebugLvl > c.ERROR {
				println(errWithInternalErr.InternalError.Error())
			}
		} else {
			printError(err, addErrorPrefix)
		}
	}
}

func printError(err error, addErrorPrefix bool) {
	errMsg := err.Error()
	if !addErrorPrefix || strings.Contains(strings.ToLower(errMsg), "error") {
		fmt.Println(errMsg)
	} else {
		fmt.Println("Error:", errMsg)
	}
}

// The loop of the program
func Start(rl *readline.Instance, user string) {
	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		InterpretLine(line)

		//Update Prompt
		rl.SetPrompt(SetPrompt(user))
	}
}
