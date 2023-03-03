package main

//This file loads and executes OCLI script files

import (
	"bufio"
	c "cli/controllers"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func ValidateFile(comBuf *[]map[string]int, file string) bool {
	invalidCommands := []string{}
	for i := range *comBuf {
		for k := range (*comBuf)[i] {
			_, err := Parse(k)
			if err != nil {
				invalidCommands = append(invalidCommands,
					" LINE#: "+strconv.Itoa((*comBuf)[i][k])+"\t"+"COMMAND:"+k)
			}
		}
	}

	if len(invalidCommands) > 0 {
		if c.State.DebugLvl > c.NONE {
			println("Syntax errors were found in the file: ", file)
			println("The following commands were invalid")
			for i := range invalidCommands {
				println(invalidCommands[i])
			}
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
			if !InterpretLine(st) {
				//println("Command: ", st)
				return
			}

			if file != c.State.ScriptPath { //Nested Execution
				LoadFile(c.State.ScriptPath)
				c.State.ScriptPath = file
			}
		}
	}

}

func LoadFile(path string) {
	originalPath := path
	file, err := os.Open(originalPath)
	if err != nil {
		if c.State.DebugLvl > c.NONE {
			println("Error:", err.Error())
		}
		return
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

		c.State.LineNumber++ //Increment
	}

	//Validate the commandbuffer
	fName := filepath.Base(path)
	if c.State.Analyser {
		if ValidateFile(&commandBuffer, fName) {
			c.State.ScriptCalled = false
			ExecuteFile(&commandBuffer, path)
		}
	} else {
		ExecuteFile(&commandBuffer, path)
	}

	ResetStateScriptData()
}

func ResetStateScriptData() {
	//Reset
	c.State.LineNumber = 0
	c.State.ScriptCalled = false
}
