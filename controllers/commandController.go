package controllers

import (
	"cli/models"
	"cli/readline"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func PWD() string {
	println(State.CurrPath)
	return State.CurrPath
}

func Disp(x map[string]interface{}) {
	/*for i, k := range x {
		println("We got: ", i, " and ", k)
	}*/

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func PostObj(ent int, entity string, data map[string]interface{}) map[string]interface{} {
	var respMap map[string]interface{}
	resp, e := models.Send("POST",
		"https://ogree.chibois.net/api/user/"+entity+"s", GetKey(), data)

	if e != nil {
		WarningLogger.Println("Error while sending POST to server: ", e)
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		ErrorLogger.Println("Error while trying to read server response: ", err)
		os.Exit(-1)
	}

	json.Unmarshal(bodyBytes, &respMap)
	println(string(respMap["message"].(string)) /*bodyBytes*/)
	if resp.StatusCode == http.StatusCreated && respMap["status"].(bool) == true {
		//Insert object into tree
		node := &Node{}
		node.ID, _ = strconv.Atoi(respMap["data"].(map[string]interface{})["id"].(string))
		node.Name = respMap["data"].(map[string]interface{})["name"].(string)
		_, ok := respMap["data"].(map[string]interface{})["parentId"].(float64)
		//node.PID = int(respMap["data"].(map[string]interface{})["parentId"].(float64))
		if ok {
			node.PID = int(respMap["data"].(map[string]interface{})["parentId"].(float64))
		} else {
			node.PID, _ = strconv.Atoi(respMap["data"].(map[string]interface{})["parentId"].(string))
		}
		node.Entity = ent
		switch ent {
		case TENANT:
			State.TreeHierarchy.Nodes.PushBack(node)
		default:
			UpdateTree(&State.TreeHierarchy, node)
		}
		return respMap["data"].(map[string]interface{})
	}
	return nil
}

func DeleteObj(path string) bool {
	URL := "https://ogree.chibois.net/api/user/"
	nd := new(*Node)

	switch path {
	case "":
		nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(State.CurrPath))
	default:
		if path[0] != '/' && len(State.CurrPath) > 1 {
			nd = FindNodeInTree(&State.TreeHierarchy,
				StrToStack(State.CurrPath+"/"+path))
		} else {
			nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
		}
	}

	if nd == nil {
		println("Error finding Object from given path!")
		WarningLogger.Println("Object to DELETE was not found")
		return false
	}

	URL += EntityToString((*nd).Entity) + "s/" + strconv.Itoa((*nd).ID)
	resp, e := models.Send("DELETE", URL, GetKey(), nil)
	if e != nil {
		println("Error while deleting Object!")
		WarningLogger.Println("Error while deleting Object!", e)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		DeleteNodeInTree(&State.TreeHierarchy, (*nd).ID, (*nd).Entity)
		println("Success")
	} else {
		println("Error while deleting object in cache!")
		WarningLogger.Println("Error while deleting object in tree")
		//json.Unmarshal()
	}

	return true
}

//Search for objects
func SearchObjects(entity string, data map[string]interface{}) []map[string]interface{} {
	var jsonResp map[string]interface{}
	URL := "https://ogree.chibois.net/api/user/" + entity + "s?"

	for i, k := range data {
		if i == "attributes" {
			for j, _ := range k.(map[string]string) {
				URL = URL + "&" + j + "=" + data[i].(map[string]string)[j]
			}
		} else {
			URL = URL + "&" + i + "=" + string(data[i].(string))
		}
	}

	//println("Here is URL: ", URL)
	InfoLogger.Println("Search query URL:", URL)

	resp, e := models.Send("GET", URL, GetKey(), nil)
	//println("Response Code: ", resp.Status)
	if e != nil {
		WarningLogger.Println("Error while sending GET to server", e)
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		ErrorLogger.Println("Error while trying to read server response: ", err)
		os.Exit(-1)
	}
	//println(string(bodyBytes))
	json.Unmarshal(bodyBytes, &jsonResp)
	if resp.StatusCode == http.StatusOK {
		obj := jsonResp["data"].(map[string]interface{})["objects"].([]interface{})
		for idx := range obj {
			println()
			println()
			println("OBJECT: ", idx)
			displayObject(obj[idx].(map[string]interface{}))
			println()
		}
		return jsonResp["data"].(map[string]interface{})["objects"].([]map[string]interface{})
	}
	return nil
}

func GetObject(path string) map[string]interface{} {
	URL := "https://ogree.chibois.net/api/user/"
	nd := new(*Node)
	var data map[string]interface{}

	switch path {
	case "":
		nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(State.CurrPath))
	default:
		if path[0] != '/' && len(State.CurrPath) > 1 {
			nd = FindNodeInTree(&State.TreeHierarchy,
				StrToStack(State.CurrPath+"/"+path))
		} else {
			nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
		}
	}

	if nd == nil {
		println("Error finding Object from given path!")
		WarningLogger.Println("Object to Get not found")
		return nil
	}

	URL += EntityToString((*nd).Entity) + "s/" + strconv.Itoa((*nd).ID)
	resp, e := models.Send("GET", URL, GetKey(), nil)
	if e != nil {
		println("Error while obtaining Object details!")
		WarningLogger.Println("Error while sending GET to server", e)
		return nil
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error while reading response!")
		ErrorLogger.Println("Error while trying to read server response: ", err)
		return nil
	}
	json.Unmarshal(bodyBytes, &data)
	if resp.StatusCode == http.StatusOK {
		if data["data"] != nil {
			obj := data["data"].(map[string]interface{})
			displayObject(obj)
		}
		return data["data"].(map[string]interface{})
	}
	return nil
}

func UpdateObj(path string, data map[string]interface{}) map[string]interface{} {
	println("OK. Attempting to update...")
	if data != nil {
		var respJson map[string]string
		nd := new(*Node)
		switch path {
		case "":
			nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(State.CurrPath))
		default:
			if path[0] != '/' && len(State.CurrPath) > 1 {
				nd = FindNodeInTree(&State.TreeHierarchy,
					StrToStack(State.CurrPath+"/"+path))
			} else {
				nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
			}
		}

		if nd == nil {
			println("Error finding Object from given path!")
			WarningLogger.Println("Object to Update not found")
			return nil
		}

		URL := "https://ogree.chibois.net/api/user/" +
			EntityToString((*nd).Entity) + "s/" + strconv.Itoa((*nd).ID)

		resp, e := models.Send("PUT", URL, GetKey(), data)
		//println("Response Code: ", resp.Status)
		if e != nil {
			println("There was an error!")
			WarningLogger.Println("Error while sending UPDATE to server", e)
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			println("Error while reading response: " + err.Error())
			ErrorLogger.Println("Error while trying to read server response: ", err)
			return nil
		}
		json.Unmarshal(bodyBytes, &respJson)
		println(respJson["message"])
		if resp.StatusCode == http.StatusOK && data["name"] != nil {
			//Need to update name of Obj in tree
			(*nd).Name = string(data["name"].(string))
			(*nd).Path = (*nd).Path[:strings.LastIndex((*nd).Path, "/")+1] + (*nd).Name
		}
		//println(string(bodyBytes))
	} else {
		println("Error! Please enter desired parameters of Object to be updated")
	}
	return data
}

func LS(x string) []map[string]interface{} {
	if x == "" || x == "." {
		ans := []map[string]interface{}{}
		path := State.CurrPath
		res := DispAtLevel(&State.TreeHierarchy, *StrToStack(State.CurrPath))
		for i := range res {
			ans = append(ans, silencedGetObj(path+"/"+res[i]))
		}
		return ans
	} else if string(x[0]) == "/" {
		ans := []map[string]interface{}{}
		path := x
		res := DispAtLevel(&State.TreeHierarchy, *StrToStack(x))
		for i := range res {
			ans = append(ans, silencedGetObj(path+"/"+res[i]))
		}
		return ans
	} else {
		res := DispAtLevel(&State.TreeHierarchy, *StrToStack(State.CurrPath + "/" + x))
		ans := []map[string]interface{}{}
		path := State.CurrPath + "/" + x
		for i := range res {
			ans = append(ans, silencedGetObj(path+"/"+res[i]))
		}
		return ans
	}

}

func LSOG() {
	fmt.Println("USER EMAIL:", GetEmail())
	fmt.Println("API URL:", "https://ogree.chibois.net/api/user/")
	fmt.Println("BUILD DATE:", BuildTime)
	fmt.Println("BUILD TREE:", BuildTree)
	fmt.Println("BUILD HASH:", BuildHash)
	fmt.Println("COMMIT DATE: ", GitCommitDate)
	fmt.Println("LOG PATH:", "./log.txt")
	fmt.Println("HISTORY FILE PATH:", ".resources/.history")
}

func LSOBJECT(x string, entity int) []map[string]interface{} {
	objs := []*Node{}
	if x == "" || x == "." {
		ok, _, r := CheckPath(&State.TreeHierarchy,
			StrToStack(State.CurrPath), New())
		if !ok {
			println("Path not valid!")
			WarningLogger.Println("Path not found: ", x)
			return nil
		}
		objs = GetNodes(r, entity)
	} else if string(x[0]) == "/" {
		ok, _, r := CheckPath(&State.TreeHierarchy, StrToStack(x), New())
		if !ok {
			println("Path not valid!")
			WarningLogger.Println("Path not found: ", x)
			return nil
		}
		objs = GetNodes(r, entity)
	} else {
		ok, _, r := CheckPath(&State.TreeHierarchy,
			StrToStack(State.CurrPath+"/"+x), New())
		if !ok {
			println("Path not valid!")
			WarningLogger.Println("Path not found: ", x)
			return nil
		}
		objs = GetNodes(r, entity)
	}

	for i := range objs {
		println(i, ":", objs[i].Name)
	}

	//Slow but necessary part
	ans := []map[string]interface{}{}
	for i := range objs {
		ans = append(ans, silencedGetObj(objs[i].Path))
	}

	return ans
}

func CD(x string) string {
	if x == ".." {
		lastIdx := strings.LastIndexByte(State.CurrPath, '/')
		if lastIdx != -1 {
			if lastIdx == 0 {
				lastIdx += 1
			}
			State.PrevPath = State.CurrPath
			State.CurrPath =
				State.CurrPath[0:lastIdx]
		}

	} else if x == "" {
		State.PrevPath = State.CurrPath
		State.CurrPath = "/"
	} else if x == "." {
		//Do nothing
	} else if x == "-" {
		//Change to previous path
		tmp := State.CurrPath
		State.CurrPath = State.PrevPath
		State.PrevPath = tmp
	} else if strings.Count(x, "/") >= 1 {
		exist := false
		var pth string

		if string(x[0]) != "/" {
			exist, pth, _ = CheckPath(&State.TreeHierarchy, StrToStack(State.CurrPath+"/"+x), New())
		} else {
			exist, pth, _ = CheckPath(&State.TreeHierarchy, StrToStack(x), New())
		}
		if exist == true {
			if State.DebugLvl >= 1 {
				println("THE PATH: ", pth)
			}
			State.PrevPath = State.CurrPath
			State.CurrPath = pth
		} else {
			println("Path does not exist")
			WarningLogger.Println("Path: ", x, " does not exist")
		}
	} else {
		if len(State.CurrPath) != 1 {
			if exist, _, _ := CheckPath(&State.TreeHierarchy,
				StrToStack(State.CurrPath+"/"+x), New()); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += "/" + x
			} else {
				println("OGREE: ", x, " : No such object")
				WarningLogger.Println("No such object: ", x)
			}

		} else {
			if exist, _, _ := CheckPath(&State.TreeHierarchy,
				StrToStack(State.CurrPath+x), New()); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += x
			} else {
				println("OGREE: ", x, " : No such object")
				WarningLogger.Println("No such object: ", x)
			}

		}

	}
	return State.CurrPath
}

func Help(entry string) {
	switch entry {
	case "ls":
		fmt.Println(`Usage: `, entry, "[PATH] (optional)")
		fmt.Println(`Displays objects in a given directory`)
	case "cd":
		fmt.Println(`Usage: `, entry, "[PATH] (optional)")
		fmt.Println(`Changes current directory`)
	case "tree":
		fmt.Println(`Usage: `, entry, "[PATH] (optional) DEPTH (optional)")
		fmt.Println(`Recursively display hierarchy with depth indentation`)
		fmt.Println(`If no options specified then tree executes with`)
		fmt.Println(`current path and depth of 0`)
	case "create":
		fmt.Println(`Usage: `, entry, "ENTITY [PATH](optional)  [ATTRIBUTES]")
		fmt.Println(`Creates an object in a given directory`)
		printAttributeOptions()
	case "gt":
		fmt.Println(`Usage: `, entry, "ENTITY (optional) [PATH](optional)  [ATTRIBUTES](optional)")
		fmt.Println(`Obtains object(s) details. 
				If ENTITY is specified then it will enter a 'search mode' 
				and at least 1 ATTRIBUTE must be specified. Otherwise an 
				object's details will be retrieved`)
		printAttributeOptions()
	case "update":
		fmt.Println(`Usage: `, entry, "[PATH](optional)  [ATTRIBUTES]")
		fmt.Println(`Modify an object by specifying new attribute values`)
		printAttributeOptions()
	case "delete":
		fmt.Println(`Usage: `, entry, "[PATH]")
		fmt.Println(`Delete an object`)
	case "lsog":
		fmt.Println(`Usage: `, entry)
		fmt.Println(`Displays system information`)
	case "grep":
		fmt.Println("NOT YET IMPLEMENTED")

	case "for":
		fmt.Println("Usage: for ((init; condition; increment)); commands; done")
		fmt.Println("OR: for var in {INTEGER..INTEGER}; commands; done")
		fmt.Println("OR: for var in expression; commands; done")
		fmt.Println("\nFor  is  a  looping command, the interpreter will execute init once.")
		fmt.Println("Then it evaluates condition as a bool expression. It will execute the commands and increment")
		fmt.Println("Then it loops again until the condition evaluates to false")
		fmt.Println("In the 2nd for loop type. A variable is created")
		fmt.Println("which will iterate between the provided range.")
		fmt.Println("This will repeatedly execute commands until the iterator reaches end of range")
		fmt.Println("\nIn the 3rd type, an internal iterator is created and will")
		fmt.Println("iterate the range of the variable or expression given.")
		fmt.Println("For now the return value of a for loop is NULL.")
		fmt.Println("\nEXAMPLE\n")
		fmt.Println("\n\tIterate 5 times and print current directory")
		fmt.Println("\tfor ((x=0; x < 5; $x=$x+1)); pwd done")
		fmt.Println("\n\tIterate 8 times execute tree and print current directory")
		fmt.Println("\tfor x in {2..10}; pwd; tree done")
		fmt.Println("\n\tIterate 5 times and print current directory")
		fmt.Println("\tfor x in 5; pwd; done")
		fmt.Println("\n\tIterate 3 times and print current directory")
		fmt.Println("\t.var:p=3; for x in $p; pwd; done")
		fmt.Println("\n\tIterate in array and print current directory")
		fmt.Println("\tfor x in len(array); pwd; done")
		fmt.Println("\n\tIterate in LS and print current directory")
		fmt.Println("\tfor x in $ls; pwd; done")

	case "while":
		fmt.Println("Usage: while (test) body done")
		fmt.Println("Execute command(s) repeatedly as long as a condition is met")
		fmt.Println("\nThe  while command evaluates test as a boolean.")
		fmt.Println("If it is true  value  then body  is  executed.")
		fmt.Println("Once body has been executed then test is evaluated again,")
		fmt.Println("and the process repeats until eventually test evaluates to a false boolean value.")
		fmt.Println("The while command for now returns NULL.")
		fmt.Println("\nEXAMPLE\n")
		fmt.Println("\t.var:x=0")
		fmt.Println("\twhile ($x < 5) pwd; $x=$x+1; done ")
	case "if":
		fmt.Println(`Usage: if [expr1] then body1 elif [expr2] then body2 elif ... else bodyN fi`)
		fmt.Println()
		fmt.Println()
		fmt.Println("The  if command evaluates expr1 as an expression.")
		fmt.Println("The value of the expression must be a boolean.")
		fmt.Println("If it is true then body1 is executed. Otherwise expr2 is")
		fmt.Println("evaluated as an expression and if it is true then body2 is executed, and so on.")
		fmt.Println("If no expression evaluates to true then bodyN is executed.")
		fmt.Println("The Else argument is optional.")
		fmt.Println("There may be any number of elif clauses,")
		fmt.Println("including zero.  BodyN may also be omitted as long as else is omitted too.")
		fmt.Println("The  return  value for now is NULL.")
		fmt.Println("\nEXAMPLE\n")
		fmt.Println("\tif [5 < 6] then ls fi")
		fmt.Println("\tif [5 < 6] then ls else tree fi")
		fmt.Println("\tif [5 == 6] then ls elif [5 == 4 ] then tree else pwd fi")
	case "cmds":
		fmt.Println("Usage: .cmds: [PATH]")
		fmt.Println("Loads script file and executes OGREE commands in file")
		fmt.Println("NOTE: Complete path must be provided")
	case "var":
		fmt.Println("Usage: .var:myVarName=[x]")
		fmt.Println("Declares and assigns a variable named myVarName")
		fmt.Println("to have value x")
		fmt.Println("NOTE: This is the only way to declare variables")

	case "unset":
		fmt.Println("Usage: ", entry, "[VAR/FUNC NAME]", "[OPTIONS]")
		fmt.Println("Deletes function or variable")
		fmt.Println()
		fmt.Println("OPTIONS")
		fmt.Println()
		fmt.Println("-v\t\tDeletes variable")
		fmt.Println("-f\t\tDeletes function")
	case "select":
		fmt.Println("Usage: selection")
		fmt.Println("Displays selected objects in clipboard")
		fmt.Println("Select objects by: ={obj1 obj2 ...}")
	case "lsobj", "lsten", "lssite", "lsbldg", "lsroom", "lsrack",
		"lsdev", "lssubdev", "lssubdev1":
		fmt.Println("Usage: ", entry, "[PATH]")
		fmt.Println("Recursively displays specified object type from given path")
	default:
		fmt.Printf(`A Shell interface to the API and your datacenter visualisation solution`)
		fmt.Println()
		fmt.Printf(`Meta+B means press Esc and n separately.  
		Users can change that in terminal simulator(i.e. iTerm2) to Alt+B  
		Notice: Meta+B is equals with Alt+B in windows.
		
		* Shortcut in normal mode
		
		| Shortcut           | Comment                           |
		| ------------------ | --------------------------------- |
		| Ctrl+A         | Beginning of line                 |
		| Ctrl+B / ←   	 | Backward one character            |
		| Meta+B         | Backward one word                 |
		| Ctrl+C         | Send io.EOF                       |
		| Ctrl+D         | Delete one character              |
		| Meta+D         | Delete one word                   |
		| Ctrl+E         | End of line                       |
		| Ctrl+F / →   	 | Forward one character             |
		| Meta+F         | Forward one word                  |
		| Ctrl+G         | Cancel                            |
		| Ctrl+H         | Delete previous character         |
		| Ctrl+I / Tab 	 | Command line completion           |
		| Ctrl+J         | Line feed                         |
		| Ctrl+K         | Cut text to the end of line       |
		| Ctrl+L         | Clear screen                      |
		| Ctrl+M         | Same as Enter key                 |
		| Ctrl+N / ↓   	 | Next line (in history)            |
		| Ctrl+P / ↑   	 | Prev line (in history)            |
		| Ctrl+R         | Search backwards in history       |
		| Ctrl+S         | Search forwards in history        |
		| Ctrl+T         | Transpose characters              |
		| Meta+T         | Transpose words (TODO)            |
		| Ctrl+U         | Cut text to the beginning of line |
		| Ctrl+W         | Cut previous word                 |
		| Backspace      | Delete previous character         |
		| Meta+Backspace | Cut previous word                 |
		| Enter          | Line feed                         |
		
		
		* Shortcut in Search Mode (Ctrl+S or Ctrl+r to enter this mode)
		
		| Shortcut                | Comment                                 |
		| ----------------------- | --------------------------------------- |
		| Ctrl+S              | Search forwards in history              |
		| Ctrl+R              | Search backwards in history             |
		| Ctrl+C / Ctrl+G 	  | Exit Search Mode and revert the history |
		| Backspace           | Delete previous character               |
		| Other               | Exit Search Mode                        |
		
		* Shortcut in Complete Select Mode (double Tab to enter this mode)
		
		| Shortcut                | Comment                                  |
		| ----------------------- | ---------------------------------------- |
		| Ctrl+F              | Move Forward                             |
		| Ctrl+B              | Move Backward                            |
		| Ctrl+N              | Move to next line                        |
		| Ctrl+P              | Move to previous line                    |
		| Ctrl+A              | Move to the first candicate in current line |
		| Ctrl+E              | Move to the last candicate in current line |
		| Tab / Enter         | Use the word on cursor to complete       |
		| Ctrl+C / Ctrl+G 	  | Exit Complete Select Mode                |
		| Other               | Exit Complete Select Mode                |`)
	}

}

func displayObject(obj map[string]interface{}) {
	for i := range obj {
		if i == "attributes" {
			for q := range obj[i].(map[string]interface{}) {
				val := string(obj[i].(map[string]interface{})[q].(string))
				if val == "" {
					println(q, ":", "NONE")
				} else {
					println(q, ":", val)
				}
			}
		} else {
			if i == "description" {
				print(i)
				inf := obj[i].([]interface{})
				for idx := range inf {
					println(inf[idx].(string))
				}
			} else if val, ok := obj[i].(string); ok == true {
				if val == "" {
					println(i, ":", "NONE")
				} else {
					println(i, ":", val)
				}
			} else {
				println(obj[i].(float64))
			}
		}

	}
}

func printAttributeOptions() {
	attrArr := []string{"address", "category", "city", "color",
		"country", "description", "domain", "gps", "height",
		"heightUnit", "id", "mainContact", "mainEmail", "mainPhone",
		"model", "name", "nbFloors", "orientation", "parentId", "posU",
		"posXY", "posXYUnit", "posZ", "posZUnit", "reserved", "reservedColor",
		"serial", "size", "sizeU", "sizeUnit", "slot", "technical",
		"technicalColor", "template", "token", "type", "usableColor",
		"vendor", "zipcode"}
	fmt.Println("Attributes: ")
	//fmt.Println("")
	for i := range attrArr {
		fmt.Println("", attrArr[i])
	}
}

func tree(base string, prefix string, depth int) {
	names := NodesAtLevel(&State.TreeHierarchy, *StrToStack(base))

	for index, name := range names {
		/*if name[0] == '.' {
			continue
		}*/
		//subpath := path.Join(base, name)
		subpath := base + "/" + name
		//counter.index(subpath)

		if index == len(names)-1 {
			fmt.Println(prefix+"└──", (name))
			if depth != 0 {
				tree(subpath, prefix+"    ", depth-1)
			}

		} else {
			fmt.Println(prefix+("├──"), (name))
			if depth != 0 {
				tree(subpath, prefix+("│   "), depth-1)
			}
		}
	}
}

func Tree(x string, depth int) {
	if x == "" || x == "." {
		println(State.CurrPath)
		tree(State.CurrPath, "", depth)
	} else if string(x[0]) == "/" {
		println(x)
		tree(x, "", depth)
	} else {
		println(State.CurrPath + "/" + x)
		tree(State.CurrPath+"/"+x, "", depth)
	}
}

func getAttrAndVal(x string) (string, string) {
	i := 0
	end := 0
	iter := 0
	for ; iter < len(x); iter++ {
		if string(x[iter]) == "." {
			i = iter
		}

		if string(x[iter]) == "=" {
			end = iter
			iter = len(x)
		}
	}

	a := x[i+1 : end]
	v := x[end+1:]
	return a, v
}

func GetOCLIAtrributes(path *Stack, ent int, data map[string]interface{}, term *readline.Instance) {
	//data["name"] = string(path.Peek().(string))
	//data["attributes"] = map[string]interface{}{}
	data["name"] = string(path.PeekLast().(string))
	println("NAME:", string(data["name"].(string)))
	switch ent {
	case TENANT:
		for data["domain"] == nil || data["category"] == nil {
			println("Enter attribute")
			x, e := term.Readline()
			if e != nil {
				println("Error reading attribute: ", e)
				ErrorLogger.Println("Error reading attribute: ", e)
				return
			}
			a, v := getAttrAndVal(x)
			switch a {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
				data[a] = v

			default:
				//data["attributes"].(map[string]interface{})[a] = v
				if _, ok := data["attributes"].(map[string]interface{}); ok {
					data["attributes"].(map[string]interface{})[a] = v
				} else {
					data["attributes"].(map[string]string)[a] = v
				}
			}
		}
		//println("Color:", data["attributes"].(map[string]string)["color"])
		PostObj(ent, "tenant", data)

	case SITE:
		//loop until user gives all neccessary attributes
		for data["domain"] == nil || data["category"] == nil ||
			data["parentId"] == nil ||
			data["attributes"].(map[string]interface{})["orientation"] == nil ||
			data["attributes"].(map[string]interface{})["usableColor"] == nil ||
			data["attributes"].(map[string]interface{})["reservedColor"] == nil ||
			data["attributes"].(map[string]interface{})["technicalColor"] == nil {
			println("Enter attribute")
			x, e := term.Readline()
			if e != nil {
				println("Error reading attribute: ", e)
				ErrorLogger.Println("Error reading attribute: ", e)
				return
			}
			a, v := getAttrAndVal(x)
			switch a {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
				data[a] = v

			default:
				if _, ok := data["attributes"].(map[string]interface{}); ok {
					data["attributes"].(map[string]interface{})[a] = v
				} else {
					data["attributes"].(map[string]string)[a] = v
				}
			}
		}
		PostObj(ent, "site", data)
	case BLDG:
		for data["domain"] == nil || data["category"] == nil ||
			data["parentId"] == nil ||
			data["attributes"].(map[string]interface{})["posXY"] == nil ||
			data["attributes"].(map[string]interface{})["posXYUnit"] == nil ||
			data["attributes"].(map[string]interface{})["posZ"] == nil ||
			data["attributes"].(map[string]interface{})["posZUnit"] == nil ||
			data["attributes"].(map[string]interface{})["size"] == nil ||
			data["attributes"].(map[string]interface{})["sizeUnit"] == nil ||
			data["attributes"].(map[string]interface{})["height"] == nil ||
			data["attributes"].(map[string]interface{})["heightUnit"] == nil {
			println("Enter attribute")
			x, e := term.Readline()
			if e != nil {
				println("Error reading attribute: ", e)
				ErrorLogger.Println("Error reading attribute: ", e)
				return
			}
			a, v := getAttrAndVal(x)
			switch a {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
				data[a] = v

			default:
				if _, ok := data["attributes"].(map[string]interface{}); ok {
					data["attributes"].(map[string]interface{})[a] = v
				} else {
					data["attributes"].(map[string]string)[a] = v
				}
			}
		}
		PostObj(ent, "building", data)
	case ROOM:
		for data["domain"] == nil || data["category"] == nil ||
			data["parentId"] == nil ||
			data["attributes"].(map[string]interface{})["orientation"] == nil ||
			data["attributes"].(map[string]interface{})["posXYUnit"] == nil ||
			data["attributes"].(map[string]interface{})["posZ"] == nil ||
			data["attributes"].(map[string]interface{})["posZUnit"] == nil ||
			data["attributes"].(map[string]interface{})["sizeUnit"] == nil ||
			data["attributes"].(map[string]interface{})["height"] == nil ||
			data["attributes"].(map[string]interface{})["heightUnit"] == nil {
			println("Enter attribute")
			x, e := term.Readline()
			if e != nil {
				println("Error reading attribute: ", e)
				ErrorLogger.Println("Error reading attribute: ", e)
				return
			}
			a, v := getAttrAndVal(x)
			switch a {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
				data[a] = v

			default:
				if _, ok := data["attributes"].(map[string]interface{}); ok {
					data["attributes"].(map[string]interface{})[a] = v
				} else {
					data["attributes"].(map[string]string)[a] = v
				}
			}
		}
		PostObj(ent, "room", data)
	case RACK:
		for data["domain"] == nil || data["category"] == nil ||
			data["parentId"] == nil ||
			data["attributes"].(map[string]interface{})["orientation"] == nil ||
			data["attributes"].(map[string]interface{})["posXYUnit"] == nil ||
			data["attributes"].(map[string]interface{})["posZ"] == nil ||
			data["attributes"].(map[string]interface{})["posZUnit"] == nil ||
			data["attributes"].(map[string]interface{})["sizeUnit"] == nil ||
			data["attributes"].(map[string]interface{})["height"] == nil ||
			data["attributes"].(map[string]interface{})["heightUnit"] == nil {
			println("Enter attribute")
			x, e := term.Readline()
			if e != nil {
				println("Error reading attribute: ", e)
				ErrorLogger.Println("Error reading attribute: ", e)
				return
			}
			a, v := getAttrAndVal(x)
			switch a {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
				data[a] = v

			default:
				if _, ok := data["attributes"].(map[string]interface{}); ok {
					data["attributes"].(map[string]interface{})[a] = v
				} else {
					data["attributes"].(map[string]string)[a] = v
				}
			}
		}
		PostObj(ent, "rack", data)
	case DEVICE:
		for data["domain"] == nil || data["category"] == nil ||
			data["parentId"] == nil ||
			data["attributes"].(map[string]interface{})["orientation"] == nil ||
			data["attributes"].(map[string]interface{})["size"] == nil ||
			data["attributes"].(map[string]interface{})["sizeUnit"] == nil ||
			data["attributes"].(map[string]interface{})["height"] == nil ||
			data["attributes"].(map[string]interface{})["heightUnit"] == nil {
			println("Enter attribute")
			x, e := term.Readline()
			if e != nil {
				println("Error reading attribute: ", e)
				ErrorLogger.Println("Error reading attribute: ", e)
				return
			}
			a, v := getAttrAndVal(x)
			switch a {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
				data[a] = v

			default:
				if _, ok := data["attributes"].(map[string]interface{}); ok {
					data["attributes"].(map[string]interface{})[a] = v
				} else {
					data["attributes"].(map[string]string)[a] = v
				}
			}
		}
		PostObj(ent, "device", data)
	}
}

func ShowClipBoard() []string {
	if State.ClipBoard != nil {
		for _, k := range *State.ClipBoard {
			println(k)
		}
		return *State.ClipBoard
	}
	return nil
}

func UpdateSelection(data map[string]interface{}) {
	if State.ClipBoard != nil {
		for _, k := range *State.ClipBoard {
			UpdateObj(k, data)
		}
	}

}

func LoadFile(path string) string {
	State.ScriptCalled = true
	State.ScriptPath = path
	return path
	//scanner := bufio.NewScanner(file)
}

func SetClipBoard(x *[]string) []string {
	State.ClipBoard = x
	return *State.ClipBoard
}

func Print(x string) string {
	fmt.Println(x)
	return x
}

//Silenced functions
//These are useful for command output assignments etc
//otherwise the terminal would be polluted by debug statements
func silencedGetObj(path string) map[string]interface{} {
	URL := "https://ogree.chibois.net/api/user/"
	nd := new(*Node)
	var data map[string]interface{}

	switch path {
	case "":
		nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(State.CurrPath))
	default:
		if path[0] != '/' && len(State.CurrPath) > 1 {
			nd = FindNodeInTree(&State.TreeHierarchy,
				StrToStack(State.CurrPath+"/"+path))
		} else {
			nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
		}
	}

	if nd == nil {
		println("Error finding Object from given path!")
		WarningLogger.Println("Object to Get not found")
		return nil
	}

	URL += EntityToString((*nd).Entity) + "s/" + strconv.Itoa((*nd).ID)
	resp, e := models.Send("GET", URL, GetKey(), nil)
	if e != nil {
		println("Error while obtaining Object details!")
		WarningLogger.Println("Error while sending GET to server", e)
		return nil
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error while reading response!")
		ErrorLogger.Println("Error while trying to read server response: ", err)
		return nil
	}
	json.Unmarshal(bodyBytes, &data)
	if resp.StatusCode == http.StatusOK {
		return data["data"].(map[string]interface{})
	}
	return nil
}
