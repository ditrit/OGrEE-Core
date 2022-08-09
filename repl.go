package main

import (
	"bufio"
	c "cli/controllers"
	l "cli/logger"
	p "cli/preprocessor"
	"cli/readline"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//This file inits the State and
//manages the interpreter and REPL
//(read eval print loop)

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

func InterpretLine(str *string) bool {
	lex := NewLexer(strings.NewReader(*str))
	result := yyParse(lex)
	if root != nil {
		root.execute()
		root = nil
	}

	if result != 0 {
		return false
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

//Functions for autocompleter

func listEntities(path string) func(string) []string {
	return func(line string) []string {

		if strings.TrimSpace(line[2:]) == "" {
			path = c.State.CurrPath
		} else {
			path = TrimToSlash(line[3:])
			if len(line) > 4 {
				trimmed := strings.TrimSpace(line[2:])
				if len(trimmed) > 2 && trimmed[2:] == ".." || len(trimmed) > 0 && trimmed != "/" {
					path = c.State.CurrPath + "/" + path
				}
			}

			if path == "" {
				path = c.State.CurrPath
			} /*else if path == "/" {
				path = "/"
			}*/
		}

		items := c.FetchNodesAtLevel(path)
		return items
	}
}

func listLocal(path string) func(string) []string {
	return func(line string) []string {

		//Algorithm to strip the string from both ends
		//to extract the file path
		q := strings.Index(line, ":") + 1
		if q < 0 {
			path = "./"
		} else {
			path = strings.TrimSpace(TrimToSlash(line[q:]))
			if len(line) > 4 {
				if strings.TrimSpace(line[q:]) != "/" {
					path = "./" + path
				}
			}

			if path == "" {
				path = "./"
			}

		}
		//End of Algorithm

		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

func TrimToSlash(x string) string {
	idx := strings.LastIndex(x, "/")
	return x[:idx+1]
}

//End of Functions for autocompleter

//Init the Shell
func Start(flags map[string]interface{}) {

	env := map[string]interface{}{}

	l.InitLogs()
	c.InitDebugLevel(flags) //Set the Debug level
	c.LoadEnvFile(env, flags["env_path"].(string))
	c.InitTimeout(env)    //Set the Unity Timeout
	c.GetURLs(flags, env) //Set the URLs
	c.InitKey(flags, env) //Set the API Key
	user, _ := c.Login(env)

	c.InitState(flags, env)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\u001b[32m\u001b[1m" + user + "@" + "OGrEE3D:" + "\u001b[37;1m" + c.State.CurrPath + "$> " + "\u001b[0m",
		HistoryFile:     flags["history_path"].(string),
		AutoComplete:    getPrefixCompleter(),
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

	//If Unity Client is present we have to Trigger
	//the listen routine
	if c.State.UnityClientAvail == true {
		c.SetListener(flags, env)
		c.TriggerListen(rl, c.State.ListenAddr)
	}

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

//Helper function that returns the prefix completer
//It is placed in a helper function to maintain readability
//and organisation in the Start() func
func getPrefixCompleter() *readline.PrefixCompleter {
	return readline.NewPrefixCompleter(false,
		readline.PcItem("cd", true,
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("pwd", false),
		readline.PcItem("clear", false),
		readline.PcItem("exit", false),
		readline.PcItem("env", false),
		readline.PcItem("grep", false),
		readline.PcItem("drawable(", false),
		readline.PcItem("draw", false),
		readline.PcItem("ls", true,
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("man", false,
			readline.PcItem("pwd", false),
			readline.PcItem("print", false),
			readline.PcItem("clear", false),
			readline.PcItem("grep", false),
			readline.PcItem("ls", true),
			readline.PcItem("cd", false),
			readline.PcItem("tree", false),
			readline.PcItem("selection", false),
			readline.PcItem("if", false),
			readline.PcItem("for", false),
			readline.PcItem("while", false),
			readline.PcItem(".cmds", false),
			readline.PcItem("lsog", false),
			readline.PcItem("env", false),
			readline.PcItem("link", false),
			readline.PcItem("unlink", false),
			readline.PcItem("lsten", false),
			readline.PcItem("lssite", false),
			readline.PcItem("lsbldg", false),
			readline.PcItem("lsroom", false),
			readline.PcItem("lsrack", false),
			readline.PcItem("lsdev", false),
			readline.PcItem("lscabinet", false),
			readline.PcItem("lscorridor", false),
			readline.PcItem("lsac", false),
			readline.PcItem("lspanel", false),
			readline.PcItem("lssensor", false),
			readline.PcItem("lsu", false),
			readline.PcItem("create", false),
			readline.PcItem("gt", false),
			readline.PcItem("update", false),
			readline.PcItem("hc", false),
			readline.PcItem("drawable", false),
			readline.PcItem("draw", false),
			readline.PcItem("camera", false),
			readline.PcItem("ui", false),
			readline.PcItem(".template", false),
			readline.PcItem(".var", false),
			readline.PcItem("unset", false),
			readline.PcItem("=", false),
			readline.PcItem("-", false),
			readline.PcItem("+", false),
			readline.PcItem(">", false),
			readline.PcItem("delete", false)),
		readline.PcItem("+", false,
			readline.PcItem("tn:", false),
			readline.PcItem("si:", false),
			readline.PcItem("bd:", false),
			readline.PcItem("ro:", false),
			readline.PcItem("rk:", false),
			readline.PcItem("dv:", false),
			readline.PcItem("gp:", false),
			readline.PcItem("co:", false),
			readline.PcItem("orphan:device:", false)),

		readline.PcItem("create", false,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItem("corridor", false),
			readline.PcItem("group", false),
			readline.PcItem("panel", false),
			readline.PcItem("cabinet", false),
			readline.PcItem("sensor", false),
			readline.PcItem("obj_template", false),
			readline.PcItem("room_template", false),
		),

		readline.PcItem("gt", true,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("update", false),
		readline.PcItem("delete", false),
		readline.PcItem("selection", false),
		readline.PcItem(".cmds:", true,
			readline.PcItemDynamic(listLocal(""), false)),

		readline.PcItem(".template:", true,
			readline.PcItemDynamic(listLocal(""), false)),
		readline.PcItem(".var:", false),
		readline.PcItem("tree", false),
		readline.PcItem("lsten", false),
		readline.PcItem("lssite", false),
		readline.PcItem("lsbldg", false),
		readline.PcItem("lsroom", false),
		readline.PcItem("lsrack", false),
		readline.PcItem("lsdev", false),
		readline.PcItem("lscabinet", false),
		readline.PcItem("lscorridor", false),
		readline.PcItem("lsac", false),
		readline.PcItem("lspanel", false),
		readline.PcItem("lssensor", false),
		readline.PcItem("lsog", false),
		readline.PcItem("lsu", true,
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("print", false),
		readline.PcItem("unset", false,
			readline.PcItem("-v", false),
			readline.PcItem("-f", false)),
		readline.PcItem("while", false,
			readline.PcItem("done", false),
		),
		readline.PcItem("for", false,
			readline.PcItem("in", false),
			readline.PcItem("done", false),
		),
		readline.PcItem("if", false,
			readline.PcItem("then", false),
			readline.PcItem("elif", false),
			readline.PcItem("else", false),
			readline.PcItem("fi", false),
		),
		readline.PcItem("camera", false,
			readline.PcItem(".", false,
				readline.PcItem("move", false),
				readline.PcItem("translate", false),
				readline.PcItem("wait", false)),
		),
		readline.PcItem("ui", false,
			readline.PcItem(".", false,
				readline.PcItem("highlight", false),
				readline.PcItem("hl", false),
				readline.PcItem("debug", false),
				readline.PcItem("infos", false),
				readline.PcItem("wireframe", false),
				readline.PcItem("delay", false)),
		),
		readline.PcItem(">", false),
		readline.PcItem("hc", true,
			readline.PcItemDynamic(listEntities(""), false)),
		/*readline.PcItem("gt", false,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItem("subdevice", false),
			readline.PcItem("subdevice1", false),
		),*/
		readline.PcItem("link:", false),
		readline.PcItem("unlink:", false),
		readline.PcItem("-", false,
			readline.PcItem("selection", false)),
		readline.PcItem("=", true,
			readline.PcItemDynamic(listEntities(""), false)),
	)
}
