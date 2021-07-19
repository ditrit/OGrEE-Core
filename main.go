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
		//var path string
		//println("TAB COMPLETER INVOKED")
		x := rlPtr.RetrieveDynamicQuery()
		//println("WE GOT", x)
		if strings.TrimSpace(x[2:]) == "" {
			path = c.State.CurrPath
		} else {
			path = TrimToSlash(x[3:])
			if len(x) > 4 {
				if strings.TrimSpace(x[2:])[:2] == ".." {
					path = c.State.CurrPath + "/" + path
				}
			}
		}
		//_, path = c.CheckPath(&c.State.TreeHierarchy, c.StrToStackTAB(path), c.New())
		items := c.DispAtLevelTAB(&c.State.TreeHierarchy,
			*c.StrToStackTAB(path))
		return items
	}
}

func TrimToSlash(x string) string {
	idx := strings.LastIndex(x, "/")
	return x[:idx+1]
}

/*func listMy(x string) func(string) []string {
	return func(line string) []string {
		var tmp string
		x := rlPtr.RetrieveDynamicQuery()
		ans := ls(TrimToSlash(x))
		for i := range ans {
			tmp += " " + ans[i]
		}
		//println("NAMES: ", tmp)
		return ans
	}
}*/

var rlPtr *readline.Instance

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
	rlPtr = rl
	rl.Operation.SetDynamicFileSystemCompletion((true))
	//rl.Operation.SetDynamicFileSystemCompletion(true)
	//rl.Operation.GetConfig().AutoComplete.Do()
	//rl.
	defer rl.Close()
	//c.InitState()
	c.InitStateDummy()
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
