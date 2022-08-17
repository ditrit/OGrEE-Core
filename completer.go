package main

import (
	c "cli/controllers"
	"cli/readline"
	"io/ioutil"
	"strings"
)

//Functions for autocompleter

func ListEntities(path string) func(string) []string {
	return func(line string) []string {

		//Instead let's trim to the first instance of whitespace
		idx := strings.Index(line, " ")
		if idx == -1 {
			return nil
		}
		idx += 1
		if line[idx:] == "" {
			path = c.State.CurrPath
			//println("DEBUG path is current")
		} else {
			path = TrimToSlash(line[idx:])
			if len(line) > idx+1 {
				trimmed := line[idx:]
				if len(trimmed) > 2 && trimmed[2:] == ".." || len(trimmed) > 0 && trimmed != "/" {
					path = c.State.CurrPath + "/" + path
				}
			}

			if path == "" {
				path = c.State.CurrPath
			}
		}

		items := c.FetchNodesAtLevel(path)
		return items
	}
}

func ListLocal(path string) func(string) []string {
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

func UnLinkObjCompleter(path string) func(string) []string {
	return func(line string) []string {
		splitted := strings.SplitAfter(line, ":")
		length := len(splitted)
		if length < 1 {
			return nil
		}

		partTwo := false

		if strings.Contains(splitted[1], "@") {
			partTwo = true
		}

		fn := ListEntities("")
		//if !partTwo {

		entities := fn(splitted[1])
		if !partTwo {
			entities = append(entities, " @ ")
		}

		return entities
		/*} else {
			splitPartTwo := strings.SplitAfter(splitted[1], "@")
			if len(splitPartTwo) < 1 {
				println("DEBUG RETURNING NIL")
				return nil
			}
			//println("DEBUG PART2")
			//println("DEBUG: Path to check", splitPartTwo[1])
			entities := fn(splitPartTwo[1])
			return entities
		}*/

	}
}

func ListUserVars(path string, appendDeref bool) func(string) []string {
	return func(line string) []string {
		ans := []string{}
		varMap := GetDynamicMap()
		for i := range varMap {
			if appendDeref {
				ans = append(ans, "$"+i)
			} else {
				ans = append(ans, i)
			}

		}
		return ans
	}
}

func ListUserFuncs(path string) func(string) []string {
	return func(line string) []string {
		ans := []string{}
		funcMap := GetFuncTable()
		for i := range funcMap {
			ans = append(ans, i)
		}
		return ans
	}
}

func ListForCreate(path string) func(string) []string {
	return func(line string) []string {
		//We want to trim until first whitespace
		split := strings.Index(line, " ")
		fn := ListEntities("")
		ans := fn(line[split+1:])
		return ans

	}
}

func AttrCompleter(path string) func(string) []string {
	return func(line string) []string {

		return []string{" :name=", " :desc="}
	}
}

func TenantSiteOCLICompleter(path string) func(string) []string {
	return func(line string) []string {

		//Trim everything up to and including the ':'
		idx := strings.Index(line, ":")
		if idx == -1 {
			return nil
		}

		fn := ListEntities("")
		ans := fn(line[idx:])
		if !strings.Contains(line, "@") {
			ans = append(ans, " @ ")
		}

		return ans
	}
}

func BldgOCLICompleter(path string) func(string) []string {
	return func(line string) []string {

		//Trim everything up to and including the ':'
		fn := TenantSiteOCLICompleter("")
		ans := fn(line)
		if strings.Count(line, "@") == 1 {
			ans = append(ans, " @ ")
		}
		return ans
	}
}

func TrimToSlash(x string) string {
	idx := strings.LastIndex(x, "/")
	return x[:idx+1]
}

//End of Functions for autocompleter

//Helper function that returns the prefix completer
//It is placed in an additional GO file as a function to maintain readability
//and organisation
func GetPrefixCompleter() *readline.PrefixCompleter {
	return readline.NewPrefixCompleter(false,
		readline.PcItem("cd", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("pwd", false),
		readline.PcItem("clear", false),
		readline.PcItem("exit", false),
		readline.PcItem("env", false),
		readline.PcItem("grep", false),
		readline.PcItem("drawable(", true,
			readline.PcItemDynamic(ListEntities(""), false),
			readline.PcItem(")", false)),
		readline.PcItem("draw", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("ls", true,
			readline.PcItemDynamic(ListEntities(""), false)),
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
			readline.PcItem("lsslot", false),
			readline.PcItem("lsu", false),
			readline.PcItem("create", false),
			readline.PcItem("gt", false),
			readline.PcItem("getu", false),
			readline.PcItem("getslot", false),
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
			readline.PcItem("tn:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true)),
			readline.PcItem("si:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true)),
			readline.PcItem("bd:", true,
				readline.PcItemDynamic(BldgOCLICompleter(""), true)),
			readline.PcItem("ro:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true)),
			readline.PcItem("rk:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true)),
			readline.PcItem("dv:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true)),
			readline.PcItem("gp:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true)),
			readline.PcItem("co:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true)),
			readline.PcItem("orphan:sensor:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true)),
			readline.PcItem("orphan:device:", true,
				readline.PcItemDynamic(TenantSiteOCLICompleter(""), true))),

		readline.PcItem("create", false,
			readline.PcItem("tenant", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("site", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("building", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("room", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("rack", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("device", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("corridor", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("group", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("panel", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("cabinet", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
			readline.PcItem("sensor", true,
				readline.PcItemDynamic(ListForCreate(""), true),
				readline.PcItemDynamic(AttrCompleter(""), false)),
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
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("getu", true,
			readline.PcItemDynamic(ListEntities(""), false)),

		readline.PcItem("getslot", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("update", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("delete", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("selection", false),
		readline.PcItem(".cmds:", true,
			readline.PcItemDynamic(ListLocal(""), false)),

		readline.PcItem(".template:", true,
			readline.PcItemDynamic(ListLocal(""), false)),
		readline.PcItem(".var:", false),
		readline.PcItem("tree", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lsten", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lssite", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lsbldg", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lsroom", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lsrack", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lsdev", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lscabinet", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lscorridor", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lsac", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lspanel", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lssensor", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lsog", false),
		readline.PcItem("lsslot", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("lsu", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("print", false,
			readline.PcItemDynamic(ListUserVars("", true), false)),
		readline.PcItem("unset", false,
			readline.PcItem("-v", false,
				readline.PcItemDynamic(ListUserVars("", false), false)),
			readline.PcItem("-f", false,
				readline.PcItemDynamic(ListUserFuncs(""), false))),
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
		readline.PcItem(">", true,
			readline.PcItemDynamic(ListEntities(""), false)),
		readline.PcItem("hc", true,
			readline.PcItemDynamic(ListEntities(""), false)),
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

		readline.PcItem("link:", true,
			readline.PcItemDynamic(UnLinkObjCompleter(""), false)),
		readline.PcItem("unlink:", true,
			readline.PcItemDynamic(UnLinkObjCompleter(""), false)),
		readline.PcItem("-", true,
			readline.PcItem("selection", false),
			readline.PcItemDynamic(ListEntities(""), false),
		),
		readline.PcItem("=", true, readline.PcItemDynamic(ListEntities(""), false)),
	)
}
