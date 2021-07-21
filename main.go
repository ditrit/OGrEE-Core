package main

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

import (
	c "cli/controllers"
	"strings"

	"cli/readline"
)

func InterpretLine(str *string) {
	lex := NewLexer(strings.NewReader(*str))
	e := yyParse(lex)
	println("\nReturn Code: ", e)
	return
}

func listEntities(path string) func(string) []string {
	return func(line string) []string {

		if strings.TrimSpace(line[2:]) == "" {
			path = c.State.CurrPath
		} else {
			path = TrimToSlash(line[3:])
			if len(line) > 4 {
				if strings.TrimSpace(line[2:])[:2] == ".." || strings.TrimSpace(line[2:])[:1] != "/" {
					path = c.State.CurrPath + "/" + path
				}
			}

			if path == "" {
				path = c.State.CurrPath
			} /*else if path == "/" {
				path = "/"
			}*/
		}

		items := c.DispAtLevelTAB(&c.State.TreeHierarchy,
			*c.StrToStack(path))
		return items
	}
}

func TrimToSlash(x string) string {
	idx := strings.LastIndex(x, "/")
	return x[:idx+1]
}

func main() {

	user, _ := c.Login()

	var completer = readline.NewPrefixCompleter(false,
		readline.PcItem("cd", true,
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("pwd", false),
		readline.PcItem("clear", false),
		readline.PcItem("exit", false),
		readline.PcItem("grep", false),
		readline.PcItem("ls", true,
			readline.PcItemDynamic(listEntities(""), false)),
		readline.PcItem("man", false),
		readline.PcItem("create", false,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItem("subdevice", false),
			readline.PcItem("subdevice1", false),
		),

		readline.PcItem("get", false,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItem("subdevice", false),
			readline.PcItem("subdevice1", false),
		),
		readline.PcItem("update", false,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItem("subdevice", false),
			readline.PcItem("subdevice1", false),
		),
		readline.PcItem("delete", false,
			readline.PcItem("tenant", false),
			readline.PcItem("site", false),
			readline.PcItem("building", false),
			readline.PcItem("room", false),
			readline.PcItem("rack", false),
			readline.PcItem("device", false),
			readline.PcItem("subdevice", false),
			readline.PcItem("subdevice1", false),
		),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          user + "@" + "OGrEE3D:$> ",
		HistoryFile:     ".resources/.history",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		//EOFPrompt:       "exit",

		HistorySearchFold: true,
		//FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()
	//c.InitState()
	//c.InitStateDummy()
	c.NewBuildTree()
	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		InterpretLine(&line)
		//c.UpdateSessionState(&line)
		//Update Prompt
		rl.SetPrompt(user + "@" + "OGrEE3D:$" + c.State.CurrPath + "> ")
	}
}
