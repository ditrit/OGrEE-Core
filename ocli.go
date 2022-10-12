package main

//This file loads and executes OCLI script files

import (
	"bufio"
	c "cli/controllers"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ValidateFile(comBuf *[]map[string]int, file string) bool {
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

func ExecuteFile(comBuf *[]map[string]int, file string) {
	for i := range *comBuf {
		for st := range (*comBuf)[i] {
			c.State.LineNumber = (*comBuf)[i][st]
			fmt.Println(st)
			if InterpretLine(&st) == false {
				//println("Command: ", st)
				return
			}
		}
	}
}

func LoadFile(path string) {
	originalPath := path
	file, err := os.Open(originalPath)
	if err != nil {
		if c.State.DebugLvl > 0 {
			println("Error:", err.Error())
		}
	}

	scanner := bufio.NewScanner(file)
	c.State.LineNumber = 1 //Indicate Line Number
	commandBuffer := []map[string]int{}

	for scanner.Scan() {
		x := scanner.Text()
		if len(x) > 0 {
			commandBuffer = append(commandBuffer,
				map[string]int{x: c.State.LineNumber})
		}

		//if originalPath != c.State.ScriptPath { //Nested Execution
		//	LoadFile(c.State.ScriptPath)
		//	c.State.ScriptPath = originalPath
		//}

		c.State.LineNumber++ //Increment
	}

	//Validate the commandbuffer
	fName := filepath.Base(path)
	if c.State.Analyser == true {
		if ValidateFile(&commandBuffer, fName) == true {
			ExecuteFile(&commandBuffer, fName)
		}
	} else {
		ExecuteFile(&commandBuffer, fName)
	}

	ResetStateScriptData()
}

func ResetStateScriptData() {
	//Reset
	c.State.LineNumber = 0
	c.State.ScriptCalled = false
}
