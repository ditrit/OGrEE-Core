package controllers

import (
	"cli/models"
	"cli/readline"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func PWD() string {
	println(State.CurrPath)
	return State.CurrPath
}

//Utility function that
//displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func ParseResponse(resp *http.Response, e error, purpose string) map[string]interface{} {
	ans := map[string]interface{}{}

	if e != nil {
		WarningLogger.Println("Error while sending "+purpose+" to server: ", e)
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		ErrorLogger.Println("Error while trying to read server response: ", err)
		if purpose == "POST" || purpose == "Search" {
		os.Exit(-1)
	}
		return nil
	}

	json.Unmarshal(bodyBytes, &ans)
	return ans
}

func PostObj(ent int, entity string, data map[string]interface{}) map[string]interface{} {
	var respMap map[string]interface{}
	resp, e := models.Send("POST",
		"http://localhost:3001/api/"+entity+"s", GetKey(), data)

	respMap = ParseResponse(resp, e, "POST")

	if resp.StatusCode == http.StatusCreated && respMap["status"].(bool) == true {
		//Print success message
		println(string(respMap["message"].(string)))

		//Insert object into tree
		node := &Node{}

		if ent == TENANT {
			node.ID, _ = respMap["data"].(map[string]interface{})["id"].(string)
			node.Name = respMap["data"].(map[string]interface{})["name"].(string)
			node.PID = ""

		} else if ent == OBJTMPL {
			node.PID = "1"
			node.ID = respMap["data"].(map[string]interface{})["slug"].(string)
			node.Name = node.ID

		} else if ent == ROOMTMPL {
			node.ID = respMap["data"].(map[string]interface{})["slug"].(string)
			node.Name = node.ID
			node.PID = "2"

		} else if ent == GROUP {
			node.Name = respMap["data"].(map[string]interface{})["name"].(string)
			node.ID = node.Name
			node.PID = "3"
		} else {
			node.ID, _ = respMap["data"].(map[string]interface{})["id"].(string)
			node.Name = respMap["data"].(map[string]interface{})["name"].(string)
			node.PID = respMap["data"].(map[string]interface{})["parentId"].(string)
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
	println("Error:", string(respMap["message"].(string)))
	return nil
}

func DeleteObj(path string) bool {
	URL := "https://ogree.chibois.net/api/"
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

	URL += EntityToString((*nd).Entity) + "s/" + (*nd).ID
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
	URL := "https://ogree.chibois.net/api/" + entity + "s?"

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
	jsonResp = ParseResponse(resp, e, "Search")

	if resp.StatusCode == http.StatusOK {
		obj := jsonResp["data"].(map[string]interface{})["objects"].([]interface{})
		objects := []map[string]interface{}{}
		for idx := range obj {
			println()
			println()
			println("OBJECT: ", idx)
			displayObject(obj[idx].(map[string]interface{}))
			objects = append(objects, obj[idx].(map[string]interface{}))
			println()
		}
		return objects

	}
	return nil
}

func GetObject(path string) map[string]interface{} {
	URL := "https://ogree.chibois.net/api/"
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

	URL += EntityToString((*nd).Entity) + "s/" + (*nd).ID
	resp, e := models.Send("GET", URL, GetKey(), nil)
	data = ParseResponse(resp, e, "GET")

	if resp.StatusCode == http.StatusOK {
		if data["data"] != nil {
			obj := data["data"].(map[string]interface{})

			if !silenced {
			displayObject(obj)
		}

		}
		return data["data"].(map[string]interface{})
	}
	return nil
}

//This is an auxillary function
//for writing proper JSONs
func GenUpdateJSON(m *map[string]interface{}, key string, value interface{}, del bool) (map[string]interface{}, bool) {

	//Base Cae
	if _, ok := (*m)[key]; ok {
		if del == true { //make a delete
			delete((*m), key)
		} else {
		(*m)[key] = value
		}

		return *m, true
	}

	for i := range *m {
		//We have a nested map
		if sub, ok := (*m)[i].(map[string]interface{}); ok {
			modifiedJ, ret := GenUpdateJSON(&sub, key, value, del)
			(*m)[i] = modifiedJ
			if ret == true {
				return *m, ret
			}
		}
	}

	return nil, false
}

func UpdateObj(path string, data map[string]interface{}, deleteAndPut bool) map[string]interface{} {
	println("OK. Attempting to update...")
	var resp *http.Response

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

		URL := "https://ogree.chibois.net/api/" +
			EntityToString((*nd).Entity) + "s/" + (*nd).ID

		//Make the proper Update JSON
		ogData := map[string]interface{}{}
		respGet, e := models.Send("GET", URL, GetKey(), nil)
		ogData = ParseResponse(respGet, e, "GET")

		ogData = ogData["data"].(map[string]interface{})
		attrs := map[string]interface{}{}

		for i := range data {
			nv, _ := GenUpdateJSON(&ogData, i, data[i], deleteAndPut)

			if nv == nil {
				//The key was not found so let's insert it
				//in attributes
				attrs[i] = data[i]
			} else {
				ogData = nv
			}

		}
		if len(attrs) > 0 {
			ogData["attributes"] = attrs
		}

		if deleteAndPut == true {
			resp, e = models.Send("PUT", URL, GetKey(), ogData)
		} else {
			resp, e = models.Send("PATCH", URL, GetKey(), ogData)
		}

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
			ans = append(ans, GetObject(path+"/"+res[i], true))
		}
		return ans
	} else if string(x[0]) == "/" {
		ans := []map[string]interface{}{}
		path := x
		res := DispAtLevel(&State.TreeHierarchy, *StrToStack(x))
		for i := range res {
			ans = append(ans, GetObject(path+"/"+res[i], true))
		}
		return ans
	} else {
		res := DispAtLevel(&State.TreeHierarchy, *StrToStack(State.CurrPath + "/" + x))
		ans := []map[string]interface{}{}
		path := State.CurrPath + "/" + x
		for i := range res {
			ans = append(ans, GetObject(path+"/"+res[i], true))
		}
		return ans
	}

}

func LSOG() {
	fmt.Println("USER EMAIL:", GetEmail())
	fmt.Println("API URL:", "https://ogree.chibois.net/api/")
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
		ans = append(ans, GetObject(objs[i].Path, true))
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
	var path string
	switch entry {
	case "ls", "pwd", "print", "cd", "tree", "create", "gt",
		"update", "delete", "lsog", "grep", "for", "while", "if",
		"cmds", "var", "unset", "select":
		path = "./other/man/" + entry + ".md"

	case "+":
		path = "./other/man/plus.md"

	case "=":
		path = "./other/man/equal.md"

	case "-":
		path = "./other/man/minus.md"

	case ".template":
		path = "./other/man/template.md"

	case ".cmds":
		path = "./other/man/cmd.md"

	case ".var":
		path = "./other/man/var.md"

	case "lsobj", "lsten", "lssite", "lsbldg", "lsroom", "lsrack",
		"lsdev":
		path = "./other/man/lsobj.md"

	default:
		path = "./other/man/default.md"
	}
		
	text, e := ioutil.ReadFile(path)
	if e != nil {
		println("Error yo")
	} else {
		println(string(text))
	}

}

func displayObject(obj map[string]interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	if err := enc.Encode(obj); err != nil {
		log.Fatal(err)
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

//When creating via OCLI syntax
//{entity}.attribute=someVal
//Gets stripped and returns
//attribute, someVal
func getAttrAndVal(x string) (string, string) {
	arr := strings.Split(x, "=")

	a := strings.TrimSpace(arr[0])
	v := strings.TrimSpace(arr[1])
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
			//println("Checking for domain:", data["domain"].(string))
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
			println("Enter attribute yo")
			x, e := term.Readline()
			if e != nil {
				println("Error reading attribute: ", e)
				ErrorLogger.Println("Error reading attribute: ", e)
				return
			}
			x = strings.TrimSpace(x)
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
			data["attributes"].(map[string]interface{})["floorUnit"] == nil ||
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

	case SEPARATOR, CORIDOR, GROUP:
		//name, category, domain, pid
		data["attributes"] = map[string]interface{}{}
		for data["name"] == nil || data["domain"] == nil || data["category"] == nil ||
			data["parentId"] == nil {
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

		if ent == SEPARATOR {
			PostObj(ent, "separator", data)
		} else if ent == CORIDOR {
			PostObj(ent, "corridor", data)
		} else {
			PostObj(ent, "group", data)
		}
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
			UpdateObj(k, data, false)
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

func Print(a ...interface{}) string {
	ans := ""
	for i := range a {
		ans += fmt.Sprintf("%v ", a[i])
	}
	//need to strip the [ ] surrounding
	//the string
	println(ans[1 : len(ans)-2])
	return ans
}
