package controllers

import (
	"cli/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func PWD() string {
	println(State.CurrPath)
	return State.CurrPath
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
		State.APIURL+"/api/"+entity+"s", GetKey(), data)

	respMap = ParseResponse(resp, e, "POST")

	if resp.StatusCode == http.StatusCreated && respMap["status"].(bool) == true {
		//Print success message
		println(string(respMap["message"].(string)))

		//Insert object into tree
		node := &Node{}

		if ent == TENANT {
			node.ID, _ = respMap["data"].(map[string]interface{})["id"].(string)
			node.Name = respMap["data"].(map[string]interface{})["name"].(string)
			node.PID = "-2"

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

		//switch ent {
		//case TENANT:
		//State.TreeHierarchy.Nodes.PushBack(node)
		//default:
		//UpdateTree(&State.TreeHierarchy, node)
		SearchAndInsert(&State.TreeHierarchy, node, ent, "")

		InformUnity("POST", "PostObj",
			map[string]interface{}{"type": "create", "data": respMap["data"]})

		return respMap["data"].(map[string]interface{})
	}
	println("Error:", string(respMap["message"].(string)))
	println()
	return nil
}

func DeleteObj(path string) bool {
	URL := State.APIURL + "/api/"
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

	InformUnity("POST", "DeleteObj",
		map[string]interface{}{"type": "delete", "data": (*nd).ID})

	return true
}

//Search for objects
func SearchObjects(entity string, data map[string]interface{}) []map[string]interface{} {
	var jsonResp map[string]interface{}
	URL := State.APIURL + "/api/" + entity + "s?"

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

//Silenced bool
//Useful for LS since
//otherwise the terminal would be polluted by debug statements
func GetObject(path string, silenced bool) map[string]interface{} {
	URL := State.APIURL + "/api/"
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

	//If the path uses dots instead of slashes
	if strings.Contains(path, ".") == true {
		path = strings.ReplaceAll(path, ".", "/")
	}

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

		URL := State.APIURL + "/api/" +
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

		InformUnity("POST", "UpdateObj",
			map[string]interface{}{"type": "modify", "data": data["data"]})

		//println(string(bodyBytes))
	} else {
		println("Error! Please enter desired parameters of Object to be updated")
	}
	return data
}

func EasyUpdate(path, op string, data map[string]interface{}) map[string]interface{} {
	println("OK. Attempting to update...")
	var resp *http.Response
	var respJson map[string]interface{}
	var e error
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

	URL := State.APIURL + "/api/" +
		EntityToString((*nd).Entity) + "s/" + (*nd).ID

	if data != nil {
		resp, e = models.Send(op, URL, GetKey(), data)

		if e != nil {
			println("There was an error!")
			WarningLogger.Println("Error while sending UPDATE (via easy syntax) to server", e)
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

		InformUnity("POST", "EasyUpdate",
			map[string]interface{}{"type": "modify", "data": data["data"]})

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
	fmt.Println("API URL:", State.APIURL+"/api/")
	fmt.Println("UNITY URL:", State.UnityClientURL)
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
		"cmds", "var", "unset", "select", "camera", "ui":
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
		path = "./other/man/cmds.md"

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
		println("Manual Page not found!")
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
	if depth < 0 {
		WarningLogger.Println("Tree command cannot accept negative value")
		println("Error: Tree command cannot accept negative value")
		return
	}
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

//Helps to create the Object (thru OCLI syntax)
func GetOCLIAtrributes(path *Stack, ent int, data map[string]interface{}) {
	attr := map[string]interface{}{}
	tmpl := map[string]interface{}{}
	var nd **Node

	if ent > TENANT {
		path.Push("Physical")
		path.ReversePop()
	}

	data["name"] = string(path.PeekLast().(string))
	println("NAME:", string(data["name"].(string)))
	data["category"] = EntityToString(ent)

	//Retrieve Parent
	if ent != TENANT && ent != GROUP {
		nd = FindNodeInTree(&State.TreeHierarchy, path)
		if nd == nil {
			if nd == nil {
				println("Error! The parent was not found in path")
				return
			}
		}
	}

	switch ent {
	case TENANT:

		data["domain"] = data["name"]
		data["parentId"] = nil
		PostObj(ent, "tenant", data)

	case SITE:

		//Default values
		data["attributes"].(map[string]interface{})["usableColor"] = "DBEDF2"
		data["attributes"].(map[string]interface{})["reservedColor"] = "F2F2F2"
		data["attributes"].(map[string]interface{})["technicalColor"] = "EBF2DE"
		data["domain"] = (*nd).Name
		data["parentId"] = (*nd).ID

		//println("Top:", path.Peek().(string))
		//println("Last:", path.Peek().(string))
		//return
		PostObj(ent, "site", data)
	case BLDG:

		attr = data["attributes"].(map[string]interface{})

		//User provided x,y,z coordinates which must be
		//formatted into a string (as per client specifications)
		if arr, ok := attr["size"].(map[string]interface{}); ok {
			res := ""
			for k, v := range arr {
				switch v.(type) {
				case int:
					res += k + ":" + strconv.Itoa(v.(int)) + ","
				case float64:
					res += k + strconv.FormatFloat(v.(float64), 'E', -1, 64) + ","
				}
			}
			attr["size"] = res[:len(res)-1]
		}

		attr["posXYUnit"] = "m"
		attr["sizeUnit"] = "m"
		attr["heightUnit"] = "m"
		attr["height"] = 0 //Should be set from parser by default
		data["parentId"] = (*nd).ID
		data["domain"] = strings.Split(((*nd).Path), "/")[2]

		PostObj(ent, "building", data)
	case ROOM:

		attr = data["attributes"].(map[string]interface{})

		baseAttrs := map[string]interface{}{
			"orientation": "+N+E", "floorUnit": "t",
			"posXYUnit": "m", "sizeUnit": "m",
			"height":     0,
			"heightUnit": "m"}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into attributes
		if q, ok := attr["template"]; ok {
			tmpl = State.TemplateTable[q.(string)]
			MergeMaps(attr, tmpl, true)
		}

		data["parentId"] = (*nd).ID
		data["domain"] = strings.Split(((*nd).Path), "/")[2]
		data["attributes"] = attr

		PostObj(ent, "room", data)
	case RACK:

		attr = data["attributes"].(map[string]interface{})

		baseAttrs := map[string]interface{}{
			"sizeUnit":    "m",
			"height":      0,
			"heightUnit":  "m",
			"posXYUnit":   "t",
			"orientation": "front",
		}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		if q, ok := attr["template"]; ok {
			tmpl = State.TemplateTable[q.(string)]
			MergeMaps(attr, tmpl, true)
		}

		data["parentId"] = (*nd).ID
		data["domain"] = strings.Split(((*nd).Path), "/")[2]
		data["attributes"] = attr

		PostObj(ent, "rack", data)
	case DEVICE:

		attr = data["attributes"].(map[string]interface{})

		baseAttrs := map[string]interface{}{
			"orientation": "front",
			"size":        0,
			"height":      0,
			"heightUnit":  "mm",
		}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		if q, ok := attr["template"]; ok {
			tmpl = State.TemplateTable[q.(string)]
			MergeMaps(attr, tmpl, true)
		}

		data["domain"] = strings.Split(((*nd).Path), "/")[2]
		data["parentId"] = (*nd).ID

		PostObj(ent, "device", data)

	case SEPARATOR, CORIDOR, GROUP:
		//name, category, domain, pid

		if ent != GROUP {
			data["domain"] = strings.Split(((*nd).Path), "/")[2]
			data["parentId"] = (*nd).ID
		}
		data["attributes"] = map[string]interface{}{}

		if ent == SEPARATOR {
			PostObj(ent, "separator", data)
		} else if ent == CORIDOR {
			PostObj(ent, "corridor", data)
		} else {
			PostObj(ent, "group", data)
		}
	}
}

//Used for Unity Client commands
func HandleUI(data map[string]interface{}) {
	Disp(data)
	InformUnity("POST", "HandleUI", data)
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

func LoadTemplate(data map[string]interface{}, filePath string) {
	if cat, ok := data["category"]; !ok {
		ErrorLogger.Println("Received Invalid Template!")
		fmt.Println("Error! Invalid Template")
	} else {
		if category, ok := cat.(string); !ok {
			ErrorLogger.Println("Category not a string Template!")
			fmt.Println("Error! Category must be string in Template." +
				"Please indicate object type as per OGrEE docs")

		} else if EntityStrToInt(category) < 0 { //Category is not an entity
			ErrorLogger.Println("Invalid Category in Template!")
			fmt.Println("Error! Invalid Category in Template." +
				"Please indicate object type as per OGrEE docs")

		} else { //We have a valid category, so let's add it
			//Retrieve Name from path and store it
			//The code assumes that the file does not
			//begin with a '.'
			fileName := filepath.Base(filePath)

			if fileName == "/" {
				WarningLogger.Println("Template not found")
				fmt.Println("Error! Template not found")
				return
			}

			if idx := strings.Index(fileName, "."); idx > 0 {
				fileName = fileName[:idx]
			}

			State.TemplateTable[fileName] = data

			//Inform Unity Client
			InformUnity("POST", "load template",
				map[string]interface{}{"type": "load template", "data": data})

		}

	}

}

func SetClipBoard(x *[]string) []string {
	State.ClipBoard = x
	data := map[string]interface{}{}
	//Verify nodes
	for _, val := range *x {
		path := StrToStack(val)
		nd := FindNodeInTree(&State.TreeHierarchy, path)
		if nd != nil {
			data = map[string]interface{}{"type": "select", "data": (*nd).ID}
			InformUnity("POST", "SetClipBoard", data)
		}
	}
	return *State.ClipBoard
}

func Print(a []interface{}) string {
	ans := ""

	for i := range a {
		ans += fmt.Sprintf("%v", a[i])
	}
	fmt.Println(ans)

	return ans
}

func SetEnv(arg string, val interface{}) {
	switch arg {
	case "Unity":
		if _, ok := val.(bool); !ok {
			msg := "Can only assign bool values for Unity Env Var"
			WarningLogger.Println(msg)
		} else {
			State.UnityClientAvail = val.(bool)
		}

	default:

	}
}

//Utility function that
//displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

//Utility Function to message Unity Client
func InformUnity(method, caller string, data map[string]interface{}) {
	//If unity is available message it
	//otherwise do nothing
	if State.UnityClientAvail == true {
		e := models.ContactUnity(method, State.UnityClientURL, data)
		if e != nil {
			WarningLogger.Println("Unable to contact Unity Client @" + caller)
			fmt.Println("Error while updating Unity: ", e.Error())
		} else {
			fmt.Println("Successfully updated Unity")
		}
		println()
		println()
	}
}

func MergeMaps(x, y map[string]interface{}, overwrite bool) {
	for i := range y {
		//Conflict case
		if _, ok := x[i]; ok {
			if overwrite == true {
				WarningLogger.Println("Conflict while merging maps")
				println("Conflict while merging data, resorting to overwriting!")
				x[i] = y[i]
			}
		} else {
			x[i] = y[i]
		}

	}
}
