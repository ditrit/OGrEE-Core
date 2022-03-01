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

		//Insert tenant into tree
		node := &Node{}
		if ent == TENANT {
			node.ID, _ = respMap["data"].(map[string]interface{})["id"].(string)
			node.Name = respMap["data"].(map[string]interface{})["name"].(string)
			node.PID = "-2"
			SearchAndInsert(&State.TreeHierarchy, node, ent, "")
		}

		InformUnity("POST", "PostObj",
			map[string]interface{}{"type": "create", "data": respMap["data"]})

		return respMap["data"].(map[string]interface{})
	}
	println("Error:", string(respMap["message"].(string)))
	println()
	return nil
}

func DeleteObj(path string) bool {
	//We have to get object first since
	//there is a potential for multiple paths
	//we don't want to delete the wrong object
	objJSON, GETURL := GetObject(path, true)
	if objJSON == nil {
		println("Error while deleting Object!")
		WarningLogger.Println("Error while deleting Object!")
		return false
	}
	entities := filepath.Base(filepath.Dir(GETURL))
	URL := State.APIURL + "/api/" + entities + "/" + objJSON["id"].(string)

	resp, e := models.Send("DELETE", URL, GetKey(), nil)
	if e != nil {
		println("Error while deleting Object!")
		WarningLogger.Println("Error while deleting Object!", e)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		println("Success")
		//Delete Tenant nodes for now
		if entities[:len(entities)-2] == "tenant" {
			DeleteNodeInTree(&State.TreeHierarchy, objJSON["id"].(string), TENANT)
		}
	} else {
		println("Error while deleting Object!")
		WarningLogger.Println("Error while deleting Object!", e)
		return false
	}

	InformUnity("POST", "DeleteObj",
		map[string]interface{}{"type": "delete", "data": objJSON["id"].(string)})

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
func GetObject(path string, silenced bool) (map[string]interface{}, string) {
	var data map[string]interface{}

	pathSplit := PreProPath(path)
	paths := OnlinePathResolve(pathSplit)

	for i := range paths {
		resp, e := models.Send("GET", paths[i], GetKey(), nil)
		data = ParseResponse(resp, e, "GET")

		if resp.StatusCode == http.StatusOK {
			if data["data"] != nil {
				obj := data["data"].(map[string]interface{})

				if !silenced {
					displayObject(obj)
				}

			}

			return data["data"].(map[string]interface{}), paths[i]
		}
	}

	//Object wasn't found
	println("Error finding Object from given path!")
	WarningLogger.Println("Object to Get not found")
	println(path)

	return nil, ""
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
		var respJson map[string]interface{}
		//We have to get object first since
		//there is a potential for multiple paths
		//we don't want to update the wrong object
		objJSON, GETURL := GetObject(path, true)
		if objJSON == nil {
			println("Error while deleting Object!")
			WarningLogger.Println("Error while deleting Object!")
			return nil
		}
		entities := filepath.Base(filepath.Dir(GETURL))
		URL := State.APIURL + "/api/" + entities + "/" + objJSON["id"].(string)

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

		respJson = ParseResponse(resp, e, "UPDATE")
		if respJson != nil {
			println("Success")

			InformUnity("POST", "UpdateObj",
				map[string]interface{}{"type": "modify", "data": data["data"]})
		}

		//For now update tenants
		if entities == "tenants" && data["name"] != nil && data["name"] != "" {
			nd := FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
			(*nd).Name = data["name"].(string)
		}

		data = respJson

	} else {
		println("Error! Please enter desired parameters of Object to be updated")
	}
	return data
}

func LS(x string) []map[string]interface{} {
	ans := []map[string]interface{}{}
	var path string
	if x == "" || x == "." {
		path = State.CurrPath

	} else if string(x[0]) == "/" {
		path = x

	} else {
		path = State.CurrPath + "/" + x
	}

	res := FetchNodesAtLevel(path)

	for i := range res {
		println(res[i])
	}
	//Return an empty result for now
	//Getting a complete array is a single line
	//change in FetchNodes
	return ans

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
	obj, path := GetObject(x, true)
	if obj == nil {
		println("Error finding Object from given path!")
		WarningLogger.Println("Object to Get not found")
		return nil
	}

	entityDir, _ := filepath.Split(path)
	entities := filepath.Base(entityDir)
	objEnt := entities[:len(entities)-1]
	obi := EntityStrToInt(objEnt)
	if obi == -1 { //Something went wrong
		println("Error finding Object from given path!")
		WarningLogger.Println("Object to Get not found")
		return nil
	}

	//YouareAt -> obi
	//want 	   -> entity

	if (entity >= AC && entity <= CORIDOR) && !(obi <= BLDG) {
		return nil
	}

	if entity < AC && obi > entity {
		return nil
	}

	//println(entities)
	var idToSend string
	if obi == TENANT {
		idToSend = obj["name"].(string)
	} else {
		idToSend = obj["id"].(string)
	}
	//println(entities)
	//println(obi)
	//println("WANT:", EntityToString(entity))
	res := lsobjHelper(State.APIURL, idToSend, obi, entity)
	for i := range res {
		println("DEBUGSTMT")
		if res[i] != nil && res[i]["name"] != nil {
			println(res[i]["name"].(string))
		}

	}
	return res
	//return nil
}

func lsobjHelper(api, objID string, curr, entity int) []map[string]interface{} {
	if entity-curr >= 2 {
		var ext, URL string
		println("DEBUG-should be here")
		ext = EntityToString(curr) + "s/" + objID + "/" + EntityToString(curr+2) + "s"
		URL = State.APIURL + "/api/" + ext
		println("DEBUG-URL:", URL)

		//EDGE CASE, if user is at a BLDG and requests object of room
		if curr == BLDG && (entity >= AC && entity <= CORIDOR) {
			ext = EntityToString(curr) + "s/" + objID + "/" + EntityToString(entity) + "s"
			r, e := models.Send("GET", State.APIURL+"/api/"+ext, GetKey(), nil)
			tmp := ParseResponse(r, e, "getting objects")
			if tmp == nil {
				return nil
			}

			tmpObjs := tmp["data"].(map[string]interface{})["objects"].([]interface{})
			res := infArrToMapStrinfArr(tmpObjs)
			return res
		}
		//END OF EDGE CASE BLOCK

		r, e := models.Send("GET", URL, GetKey(), nil)
		resp := ParseResponse(r, e, "getting objects")
		if resp == nil {
			println("return nil1")
			return nil
		}

		//objs -> resp["data"]["objects"]
		if data, ok := resp["data"].(map[string]interface{}); ok {
			if objs, ok1 := data["objects"].([]interface{}); ok1 {
				x := []map[string]interface{}{}

				if entity >= AC && entity <= CORIDOR {

					for q := range objs {
						id := objs[q].(map[string]interface{})["id"].(string)
						ext2 := "/api/" + EntityToString(curr+2) + "s/" + id + "/" + EntityToString(entity) + "s"

						tmp, e := models.Send("GET", State.APIURL+ext2, GetKey(), nil)
						tmp2 := ParseResponse(tmp, e, "get objects")
						if x != nil {
							tmpObjects := tmp2["data"].(map[string]interface{})["objects"].([]interface{})

							//convert []interface{} to []map[string]interface{}
							x = append(x, infArrToMapStrinfArr(tmpObjects)...)
						}
					}

				} else {
					for i := range objs {
						rest := lsobjHelper(api, objs[i].(map[string]interface{})["id"].(string), curr+2, entity)
						if rest != nil && len(rest) > 0 {
							x = append(x, rest...)
						}

					}
				}

				println("returning x")
				println(len(x))
				return x
			} else {
				println("return nil3")
				return nil
			}

		} else {
			println("return nil2")
			return nil
		}

	} else if entity-curr >= 1 {
		println("DEBUG-must be here")
		ext := EntityToString(curr) + "s/" + objID + "/" + EntityToString(curr+1) + "s"
		URL := State.APIURL + "/api/" + ext
		r, e := models.Send("GET", URL, GetKey(), nil)
		println("DEBUG-URL SENT:", URL)
		resp := ParseResponse(r, e, "getting objects")
		if resp == nil {
			println("return nil")
			return nil
		}
		//objs := resp["data"]["objects"]
		if data, ok := resp["data"].(map[string]interface{}); ok {
			if objs, ok1 := data["objects"].([]interface{}); ok1 {
				ans := []map[string]interface{}{}
				//For associated objects of room
				if entity >= AC && entity <= CORIDOR {
					for i := range objs {
						ext2 := "/api/" + EntityToString(curr) + "s/" +
							objs[i].(map[string]interface{})["id"].(string) +
							"/" + EntityToString(entity) + "s"

						tmp, e := models.Send("GET", State.APIURL+ext2, GetKey(), nil)
						x := ParseResponse(tmp, e, "get objects")
						if x != nil {
							ans = append(ans, x)
						}
					}
				} else {
					/*for idx := range objs {
						ans = append(ans, objs[idx].(map[string]interface{}))
					}*/
					ans = infArrToMapStrinfArr(objs)
				}

				return ans

			} else {
				return nil
			}

		} else {
			return nil
		}
	} else if entity-curr == 0 { //Base Case
		resp, e := models.Send("GET", State.APIURL+"/api/"+EntityToString(curr)+"s/"+objID, GetKey(), nil)
		x := ParseResponse(resp, e, "get object")
		return []map[string]interface{}{x["data"].(map[string]interface{})}
	}
	return nil

}

func infArrToMapStrinfArr(x []interface{}) []map[string]interface{} {
	ans := []map[string]interface{}{}
	for i := range x {
		ans = append(ans, x[i].(map[string]interface{}))
	}
	return ans
}

func CD(x string) string {
	if x == ".." {
		State.PrevPath = State.CurrPath
		State.CurrPath = filepath.Dir(State.CurrPath)

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
			pth = filepath.Clean(State.CurrPath + "/" + x)
			exist, _ = CheckPathOnline(pth)
		} else {
			pth = filepath.Clean(x)
			exist, _ = CheckPathOnline(pth)
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

			if exist, _ := CheckPathOnline(State.CurrPath + "/" + x); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += "/" + x
			} else {
				println("OGREE: ", x, " : No such object")
				WarningLogger.Println("No such object: ", x)
			}

		} else {

			if exist, _ := CheckPathOnline(State.CurrPath + x); exist == true {
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
	names := FetchNodesAtLevel(base)

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
func GetOCLIAtrributes(path string, ent int, data map[string]interface{}) {
	attr := map[string]interface{}{}
	tmpl := map[string]interface{}{}
	var parent map[string]interface{}
	var domain string
	var parentURL string

	ogPath := path
	if ent > TENANT {
		path = "/Physical/" + filepath.Dir(path)
	}

	data["name"] = filepath.Base(ogPath)
	data["category"] = EntityToString(ent)

	//Retrieve Parent
	if ent != TENANT && ent != GROUP {
		parent, parentURL = GetObject(path, true)
		if parent == nil {
			println("Error! The parent was not found in path")
			return
		}

		//Retrieve parent name for domain
		tmp := strings.Split(parentURL, State.APIURL+"/api/tenants/")

		domIDX := strings.Index(tmp[1], "/")
		if domIDX == -1 {
			domain = tmp[1]
		} else {
			domain = tmp[1][:domIDX]
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
		data["domain"] = domain
		data["parentId"] = parent["id"]

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
		data["parentId"] = parent["id"]
		data["domain"] = domain

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

		data["parentId"] = parent["id"]
		data["domain"] = domain
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

		data["parentId"] = parent["id"]
		data["domain"] = domain
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

		data["domain"] = domain
		data["parentId"] = parent["id"]

		PostObj(ent, "device", data)

	case SEPARATOR, CORIDOR, GROUP:
		//name, category, domain, pid

		if ent != GROUP {
			data["domain"] = domain
			data["parentId"] = parent["id"]
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
	//Verify paths
	for _, val := range *x {
		//path := StrToStack(val)
		//nd := FindNodeInTree(&State.TreeHierarchy, path)
		obj, _ := GetObject(val, true)
		if obj != nil {
			data = map[string]interface{}{"type": "select", "data": obj["id"]}
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

//Utility functions

//displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

//Messages Unity Client
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

//Auxillary function that preprocesses
//strings to be used for Path Resolver funcs
func PreProPath(path string) []string {
	var pathSplit []string
	switch path {
	case "":
		pathSplit = strings.Split(State.CurrPath, "/")
		pathSplit = pathSplit[2:]
	default:
		if path[0] != '/' && len(State.CurrPath) > 1 {
			pathSplit = strings.Split(filepath.Clean(State.CurrPath+"/"+path), "/")
			pathSplit = pathSplit[2:]

		} else {
			pathSplit = strings.Split(filepath.Clean(path), "/")
			if strings.TrimSpace(pathSplit[0]) == "Physical" ||
				strings.TrimSpace(pathSplit[0]) == "Logical" ||
				strings.TrimSpace(pathSplit[0]) == "Enterprise" {
				pathSplit = pathSplit[1:]
			} else {
				pathSplit = pathSplit[2:]
			}
		}
	}
	return pathSplit
}

//Take 'user' abstraction path and
//convert to online URL for API
func OnlinePathResolve(path []string) []string {
	//We have to make an array since there can be
	//multiple possible paths for paths past room
	paths := []string{}
	basePath := State.APIURL + "/api"
	roomChildren := []string{"/acs", "/panels", "/cabinets",
		"/separators", "/rows", "/groups",
		"/corridors", "/tiles", "/sensors"}

	if len(path) == 0 {
		return nil
	}

	//Check if path is templates or groups type
	if path[0] == "ObjectTemplates" {
		basePath += "/obj-templates"
		if len(path) > 1 { // Check for name
			basePath += "/" + path[1]
		}

		return []string{basePath}
	}

	if path[0] == "RoomTemplates" {
		basePath += "/room-templates"
		if len(path) > 1 {
			basePath += "/" + path[1]
		}
		return []string{basePath}
	}

	if path[0] == "Groups" {
		basePath += "/groups"
		if len(path) > 1 {
			basePath += "/" + path[1]
		}
		return []string{basePath}
	} //END OF template group type check

	for i := range path {
		if i < 4 {
			basePath += "/" + EntityToString(i) + "s/" + path[i]
		}
	}

	if len(path) <= 4 {
		return []string{basePath}
	}

	if len(path) == 5 { //Possible paths for children of room
		for idx := range roomChildren {
			paths = append(paths, basePath+roomChildren[idx]+"/"+path[4])
		}
		paths = append(paths, basePath+"/racks/"+path[4])
		return paths
	}

	basePath += "/racks/" + path[4]

	if len(path) == 6 { //Possible paths for children of racks
		paths = append(paths, basePath+"/devices/"+path[5])
		paths = append(paths, basePath+"/sensors/"+path[5])
		paths = append(paths, basePath+"/groups/"+path[5])
		return paths
	}

	basePath += "/devices/" + path[5]
	if len(path) >= 7 { //Possible paths for recursive device family
		for i := 6; i < len(path)-1; i++ {
			basePath = basePath + "/devices/" + path[i]
		}
		paths = append(paths, basePath+"/devices/"+path[len(path)-1])
		paths = append(paths, basePath+"/sensors/"+path[len(path)-1])

	}

	return paths
}

//Auxillary function for FetchNodesAtLevel
//Take 'user' abstraction path and
//convert to online URL for API
func OnlineLevelResolver(path []string) []string {
	//We have to make an array since there can be
	//multiple possible paths for paths past room
	paths := []string{}
	basePath := State.APIURL + "/api"
	roomChildren := []string{"/acs", "/panels", "/cabinets",
		"/separators", "/rows", "/groups",
		"/corridors", "/tiles", "/sensors"}

	//Check if path is templates or groups type
	if path[0] == "ObjectTemplates" {
		basePath += "/obj-templates"
		if len(path) > 1 { // Check for name
			basePath += "/" + path[1]
		}

		return []string{basePath}
	}

	if path[0] == "RoomTemplates" {
		basePath += "/room-templates"
		if len(path) > 1 {
			basePath += "/" + path[1]
		}
		return []string{basePath}
	}

	if path[0] == "Groups" {
		basePath += "/groups"
		if len(path) > 1 {
			basePath += "/" + path[1]
		}
		return []string{basePath}
	} //END OF template group type check

	for i := range path {
		if i < 5 {
			basePath += "/" + EntityToString(i) + "s/" + path[i]
		}
	}

	if len(path) < 4 {
		x := strings.Split(basePath, "/")
		if len(x)%2 == 0 {
			//Grab 2nd last entity type and get its subentity
			tmp := x[len(x)-2]
			secondLastEnt := tmp[:len(tmp)-1]

			subEnt := EntityToString(EntityStrToInt(secondLastEnt) + 1)
			basePath = basePath + "/" + subEnt + "s"
		}
		return []string{basePath}
	}

	if len(path) == 4 { //Possible paths for children of room
		for idx := range roomChildren {
			paths = append(paths, basePath+roomChildren[idx])
		}
		paths = append(paths, basePath+"/racks")
		return paths
	}

	if len(path) == 5 {
		return []string{basePath + "/sensors",
			basePath + "/groups",
			basePath + "/devices"}
	}

	basePath += "/devices"

	if len(path) == 6 { //Possible paths for children of racks
		paths = append(paths, basePath+"/"+path[5]+"/sensors")
		paths = append(paths, basePath+"/"+path[5]+"/groups")
		paths = append(paths, basePath+"/"+path[5]+"/devices")
		return paths
	}

	basePath += "/" + path[5]
	if len(path) >= 7 { //Possible paths for recursive device family
		for i := 6; i < len(path); i++ {
			basePath = basePath + "/devices/" + path[i]
		}
		paths = append(paths, basePath+"/devices")
		paths = append(paths, basePath+"/sensors")

	}

	return paths
}
