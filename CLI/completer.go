package main

import (
	c "cli/controllers"
	"os"
	pathutil "path"
	"strings"
)

//Functions for autocompleter

func ListEntities(line string) []string {
	var path string
	var trimmed string
	//Instead let's trim to the first instance of whitespace
	idx := strings.Index(line, " ")
	if idx == -1 {
		return nil
	}
	idx += 1
	if line[idx:] == "" {
		path = c.State.CurrPath
	} else {
		path = TrimToSlash(line[idx:])
		trimmed = line[idx:]
		if len(line) > idx+1 {
			if len(trimmed) > 2 && trimmed[2:] == ".." || len(trimmed) > 0 && trimmed != "/" {
				path = c.State.CurrPath + "/" + path
			}
		}

		if path == "" {
			path = c.State.CurrPath
		}

		//Helps to make autocompletion at the root
		if trimmed[0] == '/' {
			if strings.Count(trimmed, "/") > 1 {
				path = TrimToSlash(trimmed)

			} else {
				path = "/"
			}
		}
	}
	items, _ := c.Ls(pathutil.Clean(path))
	return items
}

func ListLocal(line string) []string {
	var path string
	//Algorithm to strip the string from both ends
	//to extract the file path
	q := strings.Index(line, ":") + 1
	if q < 0 {
		path = "./"
	} else {
		path = strings.TrimSpace(TrimToSlash(line[q:]))
		if len(line) > 4 {
			check := strings.TrimSpace(line[q+1:])
			if len(check) > 0 {
				if string(check[0]) != "/" {
					path = "./" + path
				}
			}

		}

		if path == "" {
			path = "./"
		}

	}
	//End of Algorithm

	names := make([]string, 0)
	files, e := os.ReadDir(path)
	if e != nil {
		if c.State.DebugLvl > c.NONE {
			Println("\n", e.Error())
		}
		return []string{}
	}
	for _, f := range files {
		names = append(names, f.Name())
	}
	return names

}

func UnLinkObjCompleter(line string) []string {
	splitted := strings.SplitAfter(line, "link")
	length := len(splitted)
	if length < 1 {
		return nil
	}

	partTwo := false

	if strings.Contains(splitted[1], "@") {
		partTwo = true
	}

	entities := ListEntities(splitted[1])
	if !partTwo {
		entities = append(entities, " @ ")
	}

	return entities
}

func ListForUI(line string) []string {
	var path string
	var trimmed string
	//Instead let's trim to the first instance of '='
	idx := strings.Index(line, "= ")
	if idx == -1 {
		return nil
	}
	idx += 1
	if line[idx:] == "" {
		path = c.State.CurrPath
	} else {
		path = TrimToSlash(line[idx:])
		trimmed = line[idx:]
		if len(line) > idx+1 {
			if len(trimmed) > 2 && trimmed[2:] == ".." || len(trimmed) > 0 && trimmed != "/" {
				path = c.State.CurrPath + "/" + strings.TrimSpace(path)
			}
		}

		if path == "" {
			path = c.State.CurrPath
		}

		//Helps to make autocompletion at the root
		if trimmed[0] == '/' {
			if strings.Count(trimmed, "/") > 1 {
				path = TrimToSlash(trimmed)

			} else {
				path = "/"
			}
		}
	}
	items, _ := c.Ls(pathutil.Clean(path))
	return items

}

func ListUserVars(path string, appendDeref bool) func(string) []string {
	return func(line string) []string {
		ans := []string{}
		varMap := c.State.DynamicSymbolTable
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

func ListUserFuncs(line string) []string {
	ans := []string{}
	funcMap := c.State.FuncTable
	for i := range funcMap {
		ans = append(ans, i)
	}
	return ans
}

func SiteOCLICompleter(line string) []string {
	//Trim everything up to and including the ':'
	idx := strings.Index(line, ":")
	if idx == -1 {
		return nil
	}

	ans := ListEntities(line[idx:])

	return ans
}

func BldgOCLICompleter(line string) []string {
	//Trim everything up to and including the ':'
	ans := SiteOCLICompleter(line)
	if strings.Count(line, "@") == 1 {
		ans = append(ans, " @ ")
	}
	return ans
}

func TrimToSlash(x string) string {
	idx := strings.LastIndex(x, "/")
	return x[:idx+1]
}

func DrawCompleter(path string) func(string) []string {
	return func(line string) []string {
		//Trim everything until the "("

		ans := ListEntities(line)
		if !strings.Contains(line, ")") {
			ans = append(ans, ")")
		}
		return ans
	}
}

var completer *PrefixCompleter = getPrefixCompleter()

func Complete(currentLine string) []string {
	lineRunes := []rune(currentLine)
	var completions []string
	results, _ := completer.Do(lineRunes, len(lineRunes))

	for _, result := range results {
		completions = append(completions, string(result))
	}

	return completions
}

func getPrefixCompleter() *PrefixCompleter {
	return NewPrefixCompleter(
		PcItem("cd",
			PcItemDynamic(ListEntities)),
		PcItem("pwd"),
		PcItem("clear"),
		PcItem("exit"),
		PcItem("env",
			PcItem("Unity",
				PcItem("=",
					PcItem("true"),
					PcItem("false"))),
			PcItem("Filter",
				PcItem("=",
					PcItem("true"),
					PcItem("false")))),
		PcItem("grep"),
		PcItem("drawable",
			PcItemDynamic(ListEntities)),
		PcItem("draw",
			PcItemDynamic(ListEntities)),
		PcItem("ls",
			PcItemDynamic(ListEntities)),
		PcItem("man",
			PcItem("pwd"),
			PcItem("print"),
			PcItem("clear"),
			PcItem("grep"),
			PcItem("ls"),
			PcItem("cd"),
			PcItem("tree"),
			PcItem("selection"),
			PcItem("if"),
			PcItem("for"),
			PcItem("while"),
			PcItem(".cmds"),
			PcItem("lsog"),
			PcItem("env"),
			PcItem("link"),
			PcItem("unlink"),
			PcItem("lssite"),
			PcItem("lsbldg"),
			PcItem("lsroom"),
			PcItem("lsrack"),
			PcItem("lsdev"),
			PcItem("lscabinet"),
			PcItem("lscorridor"),
			PcItem("lsac"),
			PcItem("lspanel"),
			PcItem("lssensor"),
			PcItem("lsenterprise"),
			PcItem("get"),
			PcItem("getu"),
			PcItem("getslot"),
			PcItem("hc"),
			PcItem("drawable"),
			PcItem("draw"),
			PcItem("camera"),
			PcItem("ui"),
			PcItem(".template"),
			PcItem(".var"),
			PcItem("undraw"),
			PcItem("unset"),
			PcItem("="),
			PcItem("-"),
			PcItem("+"),
			PcItem(">")),
		PcItem("+",
			PcItem("domain:",
				PcItemDynamic(SiteOCLICompleter)),
			PcItem("si:",
				PcItemDynamic(SiteOCLICompleter)),
			PcItem("bd:",
				PcItemDynamic(BldgOCLICompleter)),
			PcItem("ro:",
				PcItemDynamic(SiteOCLICompleter)),
			PcItem("rk:",
				PcItemDynamic(SiteOCLICompleter)),
			PcItem("dv:",
				PcItemDynamic(SiteOCLICompleter)),
			PcItem("gr:",
				PcItemDynamic(SiteOCLICompleter)),
			PcItem("co:",
				PcItemDynamic(SiteOCLICompleter)),
			PcItem("orphan sensor:",
				PcItemDynamic(SiteOCLICompleter)),
			PcItem("orphan device:",
				PcItemDynamic(SiteOCLICompleter))),
		PcItem("get",
			PcItemDynamic(ListEntities)),
		PcItem("getu",
			PcItemDynamic(ListEntities)),

		PcItem("getslot",
			PcItemDynamic(ListEntities)),
		PcItem("selection"),
		PcItem(".cmds:",
			PcItemDynamic(ListLocal)),

		PcItem(".template:",
			PcItemDynamic(ListLocal)),
		PcItem(".var:"),
		PcItem("tree",
			PcItemDynamic(ListEntities)),
		PcItem("lssite",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lsbldg",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lsroom",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lsrack",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lsdev",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lscabinet",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lscorridor",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lsac",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lspanel",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lssensor",
			PcItem("-r"),
			PcItemDynamic(ListEntities)),
		PcItem("lsog"),
		PcItem("print",
			PcItemDynamic(ListUserVars("", true))),
		PcItem("lsenterprise"),
		PcItem("undraw",
			PcItemDynamic(ListEntities)),
		PcItem("unset",
			PcItem("-v",
				PcItemDynamic(ListUserVars("", false))),
			PcItem("-f",
				PcItemDynamic(ListUserFuncs))),
		PcItem("while"),
		PcItem("for"),
		PcItem("if"),
		PcItem("camera.",
			PcItem("move",
				PcItem("=")),
			PcItem("translate",
				PcItem("=")),
			PcItem("wait",
				PcItem("="))),
		PcItem("ui.highlight",
			PcItem("=",
				PcItemDynamic(ListForUI))),
		PcItem("ui.hl",
			PcItem("=",
				PcItemDynamic(ListForUI))),
		PcItem("ui.debug",
			PcItem("=",
				PcItem("true"),
				PcItem("false"))),
		PcItem("ui.infos",
			PcItem("=",
				PcItem("true"),
				PcItem("false"))),
		PcItem("ui.wireframe",
			PcItem("=",
				PcItem("true"),
				PcItem("false"))),
		PcItem("ui.delay",
			PcItem(" = ")),
		PcItem("ui.clearcache"),
		PcItem(">",
			PcItemDynamic(ListEntities)),
		PcItem("hc",
			PcItemDynamic(ListEntities)),
		/*PcItem("gt",
			PcItem("site"),
			PcItem("building"),
			PcItem("room"),
			PcItem("rack"),
			PcItem("device"),
			PcItem("subdevice"),
			PcItem("subdevice1"),
		),*/
		PcItem("link",
			PcItemDynamic(UnLinkObjCompleter)),
		PcItem("unlink",
			PcItemDynamic(UnLinkObjCompleter)),
		PcItem("-",
			PcItem("selection"),
			PcItemDynamic(ListEntities),
		),
		PcItem("=", PcItemDynamic(ListEntities)),
	)
}
