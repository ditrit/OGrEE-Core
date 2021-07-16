package main

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

// Adding TAB completion support
//https://thoughtbot.com/blog/tab-completion-in-gnu-readline
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

/*
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}
*/

func listEntities(path string) func(string) []string {
	return func(line string) []string {
		items := c.DispAtLevelTAB(&c.State.TreeHierarchy,
			*c.StrToStackTAB(path))
		return items
	}
}

func main() {

	user, _ := c.Login()

	var completer = readline.NewPrefixCompleter(
		readline.PcItem("cd",
			readline.PcItemDynamic(listEntities(""))),
		readline.PcItem("pwd"),
		readline.PcItem("clear"),
		readline.PcItem("exit"),
		readline.PcItem("grep"),
		readline.PcItem("ls",
			readline.PcItemDynamic(listEntities(""))),
		readline.PcItem("man"),
		readline.PcItem("create",
			readline.PcItem("tenant"),
			readline.PcItem("site"),
			readline.PcItem("building"),
			readline.PcItem("room"),
			readline.PcItem("rack"),
			readline.PcItem("device"),
			readline.PcItem("subdevice"),
			readline.PcItem("subdevice1"),
		),

		readline.PcItem("get",
			readline.PcItem("tenant"),
			readline.PcItem("site"),
			readline.PcItem("building"),
			readline.PcItem("room"),
			readline.PcItem("rack"),
			readline.PcItem("device"),
			readline.PcItem("subdevice"),
			readline.PcItem("subdevice1"),
		),
		readline.PcItem("update",
			readline.PcItem("tenant"),
			readline.PcItem("site"),
			readline.PcItem("building"),
			readline.PcItem("room"),
			readline.PcItem("rack"),
			readline.PcItem("device"),
			readline.PcItem("subdevice"),
			readline.PcItem("subdevice1"),
		),
		readline.PcItem("delete",
			readline.PcItem("tenant"),
			readline.PcItem("site"),
			readline.PcItem("building"),
			readline.PcItem("room"),
			readline.PcItem("rack"),
			readline.PcItem("device"),
			readline.PcItem("subdevice"),
			readline.PcItem("subdevice1"),
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
	c.InitState()
	//c.AddHistory(rl)
	//readline.NewPrefixCompleter
	//readline.SegmentCompleter
	//readline.SetAutoComplete
	//readline.AutoCompleter
	//readline.FuncListener
	//readline.AutoCompleter
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
