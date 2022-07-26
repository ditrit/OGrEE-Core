package controllers

import (
	"bytes"
	l "cli/logger"
	"cli/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
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
		l.GetWarningLogger().Println("Error while sending "+purpose+" to server: ", e)
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		l.GetErrorLogger().Println("Error while trying to read server response: ", err)
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

		//If ent is in State.ObjsForUnity then notify Unity
		if IsInObjForUnity(entity) == true {
			entInt := EntityStrToInt(entity)
			InformUnity("POST", "PostObj", entInt,
				map[string]interface{}{"type": "create", "data": respMap["data"]})
		}

		return respMap["data"].(map[string]interface{})
	}
	println("Error:", string(respMap["message"].(string)))
	println()
	return nil
}

//Calls API's Validation
func ValidateObj(data map[string]interface{}, ent string, silence bool) bool {
	var respMap map[string]interface{}
	resp, e := models.Send("POST",
		State.APIURL+"/api/validate/"+ent+"s", GetKey(), data)

	respMap = ParseResponse(resp, e, "POST")

	if resp.StatusCode == http.StatusOK && respMap["status"].(bool) == true {
		//Print success message
		println(string(respMap["message"].(string)))

		return true
	}
	println("Error:", string(respMap["message"].(string)))
	println()
	return false
}

func DeleteObj(Path string) bool {
	//We have to get object first since
	//there is a potential for multiple paths
	//we don't want to delete the wrong object
	objJSON, GETURL := GetObject(Path, true)
	if objJSON == nil {
		println("Error while deleting Object!")
		l.GetWarningLogger().Println("Error while deleting Object!")
		return false
	}

	//Make sure we are deleting an object and not
	//an aggregate call result
	if objJSON["id"] == nil {
		println("Error: Cannot delete object")
		return false
	}
	entities := path.Base(path.Dir(GETURL))
	URL := State.APIURL + "/api/" + entities + "/" + objJSON["id"].(string)

	//Get curr object path to check if it is equivalent
	//to user received path
	_, currPathURL := GetObject(State.CurrPath, true)

	resp, e := models.Send("DELETE", URL, GetKey(), nil)
	if e != nil {
		println("Error while deleting Object!")
		l.GetWarningLogger().Println("Error while deleting Object!", e)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		println("Success")
	} else {
		println("Error while deleting Object!")
		l.GetWarningLogger().Println("Error while deleting Object!", e)
		return false
	}

	entity := entities[:len(entities)-1]
	if IsInObjForUnity(entity) == true {
		InformUnity("POST", "DeleteObj", -1,
			map[string]interface{}{"type": "delete", "data": objJSON["id"].(string)})
	}

	//Check if deleted object is current path
	if currPathURL == GETURL {
		CD("..")
	}

	return true
}

func DeleteSelection() bool {
	res := true
	for i := range *State.ClipBoard {
		println("Going to delete object: ", (*(State.ClipBoard))[i])
		if DeleteObj((*(State.ClipBoard))[i]) != true {
			l.GetWarningLogger().Println("Couldn't delete obj in selection: ",
				(*(State.ClipBoard))[i])
			println("Couldn't delete obj in selection: ",
				(*(State.ClipBoard))[i])
			res = false
		}
		println()
	}
	return res
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
	l.GetInfoLogger().Println("Search query URL:", URL)

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

		if IsInObjForUnity(entity) {
			resp := map[string]interface{}{"type": "search", "data": objects}
			InformUnity("POST", "Search", -1, resp)
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
	l.GetWarningLogger().Println("Object to Get not found")
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

//This function recursively applies an update to an object and
//the rest of its subentities
//Thus being the only function thus far to call another exported
//function in this file
func RecursivePatch(Path, id, ent string, data map[string]interface{}) {
	var entities string
	var URL string
	println("OK. Attempting to update...")
	//var resp *http.Response

	if data != nil {
		if Path != "" {

			//We have to get object first since
			//there is a potential for multiple paths
			//we don't want to update the wrong object
			objJSON, GETURL := GetObject(Path, true)
			if objJSON == nil {
				println("Error while deleting Object!")
				l.GetWarningLogger().Println("Error while deleting Object!")
				return
			}
			entities = path.Base(path.Dir(GETURL))
			URL = State.APIURL + "/api/" + entities + "/" + objJSON["id"].(string) + "/all"
		} else {
			entities = ent + "s"
			URL = State.APIURL + "/api/" + entities + "/" + id + "/all"
		}
		//GET Object
		resp, e := models.Send("GET", URL, GetKey(), nil)
		r := ParseResponse(resp, e, "recursive update")
		if e != nil {
			return
		}
		recursivePatchAux(r["data"].(map[string]interface{}), data)
		println("Success")
		return

	}
	println("Error! Please enter desired parameters of Object to be updated")

}

func recursivePatchAux(res, data map[string]interface{}) {
	id := res["id"].(string)
	ent := res["category"].(string)
	UpdateObj("", id, ent, data, false)

	if childrenJson, ok := res["children"]; ok {
		if enfants, ok := childrenJson.([]interface{}); ok {
			for i := range enfants {
				if child, ok := enfants[i].(map[string]interface{}); ok {
					//id := child["id"].(string)
					//ent := child["entity"].(string)
					//UpdateObj("", id, ent,data, false)
					recursivePatchAux(child, data)
				}
			}
		}
	}

}

//You can either update obj by path or by ID and entity string type
//The deleteAndPut bool is for deleting an attribute
func UpdateObj(Path, id, ent string, data map[string]interface{}, deleteAndPut bool) map[string]interface{} {

	println("OK. Attempting to update...")
	var resp *http.Response

	if data != nil {
		var respJson map[string]interface{}
		var URL string
		var entities string

		if Path != "" || Path == "" && ent == "" {

			if Path == "" { //This means we should use curr path
				Path = State.CurrPath
			}

			//We have to get object first since
			//there is a potential for multiple paths
			//we don't want to update the wrong object
			objJSON, GETURL := GetObject(Path, true)
			if objJSON == nil {
				println("Error while getting Object!")
				l.GetWarningLogger().Println("Error while getting Object!")
				return nil
			}
			entities = path.Base(path.Dir(GETURL))
			URL = State.APIURL + "/api/" + entities + "/" + objJSON["id"].(string)
		} else {
			entities = ent + "s"
			URL = State.APIURL + "/api/" + entities + "/" + id
		}

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

			//Determine if Unity requires the message as
			//Interact or Modify
			message := map[string]interface{}{}
			interactData := map[string]interface{}{}
			var key string

			if entities == "rooms" && (data["tilesName"] != nil || data["tilesColor"] != nil) {
				println("Room modifier detected")
				Disp(data)
				message["type"] = "interact"

				//Get the interactive key
				key = determineStrKey(data, []string{"tilesName", "tilesColor"})

				interactData["id"] = ogData["id"]
				interactData["param"] = key
				interactData["value"] = data[key]
				message["data"] = interactData

			} else if entities == "racks" && data["U"] != nil {
				message["type"] = "interact"
				interactData["id"] = ogData["id"]
				interactData["param"] = "U"
				interactData["value"] = data["U"]
				message["data"] = interactData

			} else if (entities == "devices" || entities == "racks") &&
				(data["alpha"] != nil || data["slots"] != nil ||
					data["localCS"] != nil) {
				message["type"] = "interact"

				//Get interactive key
				key = determineStrKey(data, []string{"alpha", "U", "slots", "localCS"})

				interactData["id"] = ogData["id"]
				interactData["param"] = key
				interactData["value"] = data[key]

				message["data"] = interactData

			} else if entities == "groups" && data["content"] != nil {
				message["type"] = "interact"
				interactData["id"] = ogData["id"]
				interactData["param"] = "content"
				interactData["value"] = data["content"]

				message["data"] = interactData

			} else {
				message["type"] = "modify"
				message["data"] = respJson["data"]
			}

			entStr := entities[:len(entities)-1]
			if IsInObjForUnity(entStr) == true {
				entInt := EntityStrToInt(entStr)
				InformUnity("POST", "UpdateObj", entInt, message)
			}

		}

		data = respJson

	} else {
		println("Error! Please enter desired parameters of Object to be updated")
	}
	return data
}

func LS(x string) []map[string]interface{} {
	var path string
	if x == "" || x == "." {
		path = State.CurrPath

	} else if string(x[0]) == "/" {
		path = x

	} else {
		path = State.CurrPath + "/" + x
	}

	res := FetchJsonNodesAtLevel(path)

	//Display the objects by otherwise by name
	//or slug for templates
	for i := range res {
		if _, ok := res[i]["slug"].(string); ok {
			println(res[i]["slug"].(string))
		} else {
			println(res[i]["name"].(string))
		}
	}
	return res

}

func Clear() {

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Printf("\033[H")
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
	fmt.Println("DEBUG LEVEL: ", State.DebugLvl)
}

//Displays environment variable values
//to user
func Env() {
	fmt.Println("Unity: ", State.UnityClientAvail)
	fmt.Println("Filter: ", State.FilterDisplay)
	fmt.Println()
	fmt.Println("Objects Unity shall be informed of upon update:")
	for _, k := range State.ObjsForUnity {
		fmt.Println(EntityToString(k))
	}
	fmt.Println()
	fmt.Println("Objects Unity shall draw:")
	for _, k := range State.DrawableObjs {
		fmt.Println(EntityToString(k))
	}
}

func LSOBJECT(x string, entity int) []map[string]interface{} {
	obj, Path := GetObject(x, true)
	if obj == nil {
		println("Error finding Object from given path!")
		l.GetWarningLogger().Println("Object to Get not found")
		return nil
	}

	entityDir, _ := path.Split(Path)
	entities := path.Base(entityDir)
	objEnt := entities[:len(entities)-1]
	obi := EntityStrToInt(objEnt)
	if obi == -1 { //Something went wrong
		println("Error finding Object from given path!")
		l.GetWarningLogger().Println("Object to Get not found")
		return nil
	}

	//YouareAt -> obi
	//want 	   -> entity

	if (entity >= AC && entity <= CORIDOR) && obi > BLDG {
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
		if res[i] != nil && res[i]["name"] != nil {
			println(res[i]["name"].(string))
		}

	}
	return res
	//return nil
}

//NOTE: LSDEV is recursive while LSSENSOR is not
//Code could be more tidy
func lsobjHelper(api, objID string, curr, entity int) []map[string]interface{} {
	var ext, URL string
	if entity == SENSOR && (curr == BLDG || curr == ROOM || curr == RACK || curr == DEVICE) {
		ext = EntityToString(curr) + "s/" + objID + "/" + EntityToString(entity) + "s"
		URL = State.APIURL + "/api/" + ext
		r, e := models.Send("GET", URL, GetKey(), nil)
		tmp := ParseResponse(r, e, "getting objects")
		if tmp == nil {
			return nil
		}

		tmpObjs := LoadArrFromResp(tmp, "objects")
		if tmp == nil {
			return nil
		}
		res := infArrToMapStrinfArr(tmpObjs)
		return res

	} else if entity-curr >= 2 {

		//println("DEBUG-should be here")
		ext = EntityToString(curr) + "s/" + objID + "/" + EntityToString(curr+2) + "s"
		URL = State.APIURL + "/api/" + ext
		//println("DEBUG-URL:", URL)

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
		objs := LoadArrFromResp(resp, "objects")
		if objs != nil {
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
				if entity == DEVICE && curr == ROOM {
					x = append(x, infArrToMapStrinfArr(objs)...)
				}
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
		}

	} else if entity-curr >= 1 {
		//println("DEBUG-must be here")
		ext := EntityToString(curr) + "s/" + objID + "/" + EntityToString(curr+1) + "s"
		URL := State.APIURL + "/api/" + ext
		r, e := models.Send("GET", URL, GetKey(), nil)
		//println("DEBUG-URL SENT:", URL)
		resp := ParseResponse(r, e, "getting objects")
		if resp == nil {
			println("return nil")
			return nil
		}
		//objs := resp["data"]["objects"]
		objs := LoadArrFromResp(resp, "objects")
		if objs != nil {
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

				ans = infArrToMapStrinfArr(objs)
				if curr == RACK && entity == DEVICE {
					for idx := range ans {
						ext2 := "/api/" + EntityToString(entity) +
							"s/" + ans[idx]["id"].(string) + "/" + EntityToString(entity) + "s"
						subURL := State.APIURL + ext2
						r1, e1 := models.Send("GET", subURL, GetKey(), nil)
						tmp1 := ParseResponse(r1, e1, "getting objects")

						tmp2 := LoadArrFromResp(tmp1, "objects")
						if tmp2 != nil {
							//Swap ans and objs to keep order
							ans = append(ans, infArrToMapStrinfArr(tmp2)...)
						}

					}

				}
			}

			return ans
		}

	} else if entity-curr == 0 { //Base Case

		//For devices we have to make hierarchal call
		if entity == DEVICE {
			URL = State.APIURL + "/api/" + EntityToString(curr) + "s/" + objID + "/devices"
		} else {
			URL = State.APIURL + "/api/" + EntityToString(curr) + "s/" + objID
		}

		resp, e := models.Send("GET", URL, GetKey(), nil)
		x := ParseResponse(resp, e, "get object")
		if entity == DEVICE {
			tmp := x["data"].(map[string]interface{})["objects"].([]interface{})
			objArr := infArrToMapStrinfArr(tmp)
			return objArr
		}
		return []map[string]interface{}{x["data"].(map[string]interface{})}
	}
	return nil
}

//Convert []interface{} array to
//[]map[string]interface{} array
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
		State.CurrPath = path.Dir(State.CurrPath)

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
			pth = path.Clean(State.CurrPath + "/" + x)
			exist, _ = CheckPathOnline(pth)
		} else {
			pth = path.Clean(x)
			exist, _ = CheckPathOnline(pth)
		}
		if exist == true {
			if State.DebugLvl >= 1 {
				println("THE PATH: ", pth)
			}
			State.PrevPath = State.CurrPath
			State.CurrPath = pth
		} else {
			//Need to check that the path points to tree
			//before declaring it to be nonexistant
			if string(x[0]) != "/" {
				pth = State.CurrPath + "/" + x
			} else {
				pth = x
			}
			pth = path.Clean(pth)
			if FindNodeInTree(&State.TreeHierarchy, StrToStack(pth), true) != nil {
				State.PrevPath = State.CurrPath
				State.CurrPath = pth
				//println(("DEBUG not in tree either"))
				//println("DEBUG ", x)
				//println()
			} else {
				println("Path does not exist")
				l.GetWarningLogger().Println("Path: ", x, " does not exist")
			}

		}
	} else {
		if len(State.CurrPath) != 1 {
			if exist, _ := CheckPathOnline(State.CurrPath + "/" + x); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += "/" + x
			} else {
				println("OGREE: ", x, " : No such object")
				l.GetWarningLogger().Println("No such object: ", x)
			}

		} else {

			if exist, _ := CheckPathOnline(State.CurrPath + x); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += x
			} else {
				println("OGREE: ", x, " : No such object")
				l.GetWarningLogger().Println("No such object: ", x)
			}

		}

	}
	return State.CurrPath
}

func Help(entry string) {
	var path string
	switch entry {
	case "ls", "pwd", "print", "cd", "tree", "create", "gt", "clear",
		"update", "delete", "lsog", "grep", "for", "while", "if", "env",
		"cmds", "var", "unset", "select", "camera", "ui", "hc", "drawable",
		"link", "unlink", "draw":
		path = "./other/man/" + entry + ".md"

	case ">":
		path = "./other/man/focus.md"

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

//Function is an abstraction of a normal exit
func Exit() {
	//writeHistoryOnExit(&State.sessionBuffer)
	//runtime.Goexit()
	os.Exit(0)
}

func Tree(x string, depth int) {
	if depth < 0 {
		l.GetWarningLogger().Println("Tree command cannot accept negative value")
		println("Error: Tree command cannot accept negative value")
		return
	}
	objJSON, _ := GetObject(x, true)
	if objJSON == nil {
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

func GetHierarchy(x string, depth int, silence bool) []map[string]interface{} {
	//Variable declarations
	var URL, ext, depthStr string
	var ans []map[string]interface{}

	//Get object first
	obj, e := GetObject(x, true)
	if obj == nil {
		println("Error: ", e)
		return nil
	}

	//Then obtain hierarchy
	id := obj["id"].(string)
	if ent, ok := obj["category"]; ok {
		if entity, ok := ent.(string); ok {
			//Make URL
			depthStr = strconv.Itoa(depth)
			ext = "/api/" + entity + "s/" + id + "/all?limit=" + depthStr
			URL = State.APIURL + ext

			r, e := models.Send("GET", URL, GetKey(), nil)
			if e != nil {
				println("Error: " + e.Error())
				l.GetErrorLogger().Println("Error: " + e.Error())
				return nil
			}

			data := ParseResponse(r, nil, "get hierarchy")
			if data == nil {
				l.GetWarningLogger().Println("Hierarchy call response was nil")
				println("No data")
				return nil
			}

			objs := LoadArrFromResp(data, "children")
			if objs == nil {
				l.GetWarningLogger().Println("No objects found in hierarchy call")
				println("No objects found in hierarchy call")
				return nil
			}

			ans = infArrToMapStrinfArr(objs)

		}

	}
	if silence == false {
		DispMapArr(ans)
	}

	return ans
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
func GetOCLIAtrributes(Path string, ent int, data map[string]interface{}) {
	attr := map[string]interface{}{}
	var parent map[string]interface{}
	var domain string
	var parentURL string

	ogPath := Path
	//The only place where the STRAY_DEV path is
	//properly fixed
	if ent == STRAY_DEV {
		Path = "/Physical/Stray/Devices/" + path.Dir(Path)
	} else if ent > TENANT {
		Path = "/Physical/" + path.Dir(Path)
	}

	name := path.Base(ogPath)
	if name == "." || name == "" {
		l.GetWarningLogger().Println("Invalid path name provided for OCLI object creation")
		println("Error please provide a valid name for this object")
		return
	}

	data["name"] = name
	data["category"] = EntityToString(ent)
	data["description"] = []interface{}{}

	//Retrieve Parent
	if ent != TENANT && ent != GROUP && ent != STRAY_DEV {
		parent, parentURL = GetObject(Path, true)
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
		data["domain"] = domain
		data["parentId"] = parent["id"]

		PostObj(ent, "site", data)
	case BLDG:
		attr = data["attributes"].(map[string]interface{})

		//Serialise size and posXY if given
		if _, ok := attr["size"].(string); ok {
			attr["size"] = serialiseAttr(attr, "size")
		} else {
			attr["size"] = serialiseAttr2(attr, "size")
		}

		if _, ok := attr["posXY"].(string); ok {
			attr["posXY"] = serialiseAttr(attr, "posXY")
		} else {
			attr["posXY"] = serialiseAttr2(attr, "posXY")
		}

		attr["posXYUnit"] = "m"
		attr["sizeUnit"] = "m"
		attr["heightUnit"] = "m"
		//attr["height"] = 0 //Should be set from parser by default
		data["parentId"] = parent["id"]
		data["domain"] = domain

		PostObj(ent, "building", data)
	case ROOM:

		attr = data["attributes"].(map[string]interface{})

		baseAttrs := map[string]interface{}{
			"orientation": "+N+E", "floorUnit": "t",
			"posXYUnit": "m", "sizeUnit": "m",
			"height":     "5",
			"heightUnit": "m"}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		GetOCLIAtrributesTemplateHelper(attr, data, ent)

		if _, ok := attr["posXY"].(string); ok {
			attr["posXY"] = serialiseAttr(attr, "posXY")
		} else {
			attr["posXY"] = serialiseAttr2(attr, "posXY")
		}

		data["parentId"] = parent["id"]
		data["domain"] = domain
		data["attributes"] = attr
		PostObj(ent, "room", data)
	case RACK:

		attr = data["attributes"].(map[string]interface{})

		baseAttrs := map[string]interface{}{
			"sizeUnit":    "cm",
			"height":      "5",
			"heightUnit":  "U",
			"posXYUnit":   "t",
			"orientation": "front",
		}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		GetOCLIAtrributesTemplateHelper(attr, data, ent)

		//Serialise posXY if given
		if _, ok := attr["posXY"].(string); ok {
			attr["posXY"] = serialiseAttr(attr, "posXY")
		} else {
			attr["posXY"] = serialiseAttr2(attr, "posXY")
		}

		data["parentId"] = parent["id"]
		data["domain"] = domain
		data["attributes"] = attr

		PostObj(ent, "rack", data)
	case DEVICE:

		attr = data["attributes"].(map[string]interface{})

		//Special routine to perform on device
		//based on if the parent has a "slot" attribute

		//First check if attr has only posU & sizeU
		//reject if true while also converting sizeU to string if numeric
		if len(attr) == 2 {
			if sizeU, ok := attr["sizeU"]; ok {
				sizeUValid := checkNumeric(attr["sizeU"])

				if _, ok := attr["template"]; !ok && sizeUValid == false {
					l.GetWarningLogger().Println("Invalid parameter provided for device ")
					println("Please provide an existing template or valid SizeU value")
					return
				}

				//Convert block
				//And Set height
				if _, ok := sizeU.(int); ok {
					attr["sizeU"] = strconv.Itoa(sizeU.(int))
					attr["height"] = strconv.FormatFloat(
						(float64(sizeU.(int)) * 44.5), 'G', -1, 64)
				} else if _, ok := sizeU.(float64); ok {
					attr["sizeU"] = strconv.FormatFloat(sizeU.(float64), 'G', -1, 64)
					attr["height"] = strconv.FormatFloat(sizeU.(float64)*44.5, 'G', -1, 64)
				}
				//End of convert block
				if _, ok := attr["slot"]; ok {
					l.GetWarningLogger().Println("Invalid device syntax encountered")
					println("Invalid syntax: If you have provided",
						" a template, it was not found")
					return
				}
			}
		}

		//If slot not found
		if x, ok := attr["posU/slot"]; ok {
			delete(attr, "posU/slot")
			//Convert posU to string if numeric
			if _, ok := x.(float64); ok {
				x = strconv.FormatFloat(x.(float64), 'G', -1, 64)
				attr["posU"] = x
				attr["slot"] = ""
			} else if _, ok := x.(int); ok {
				x = strconv.Itoa(x.(int))
				attr["posU"] = x
				attr["slot"] = ""
			} else {
				attr["slot"] = x
			}
		}

		//If user provided templates, get the JSON
		//and parse into templates
		if _, ok := attr["template"]; ok {
			GetOCLIAtrributesTemplateHelper(attr, data, DEVICE)
		} else {
			attr["template"] = ""
			if parAttr, ok := parent["attributes"].(map[string]interface{}); ok {
				if rackSizeInf, ok := parAttr["size"]; ok {
					values := map[string]interface{}{}

					if rackSizeComplex, ok := rackSizeInf.(string); ok {
						q := json.NewDecoder(strings.NewReader(rackSizeComplex))
						q.Decode(&values)
						if determineStrKey(values, []string{"x"}) == "x" &&
							determineStrKey(values, []string{"y"}) == "y" {
							if _, ok := values["x"].(int); ok {
								values["x"] = values["x"].(int) / 10

							} else if _, ok := values["x"].(float64); ok {
								values["x"] = values["x"].(float64) / 10.0
							}

							if _, ok := values["y"].(int); ok {
								values["y"] = values["y"].(int) / 10

							} else if _, ok := values["y"].(float64); ok {
								values["y"] = values["y"].(float64) / 10.0
							}
							newValues, _ := json.Marshal(values)
							attr["size"] = string(newValues)

						}
					}

				}
			}
		}

		//End of device special routine

		baseAttrs := map[string]interface{}{
			"orientation": "front",
			"size":        "{\"x\":5,\"y\":5}",
			"sizeUnit":    "mm",
			"height":      "5",
			"heightUnit":  "mm",
		}

		MergeMaps(attr, baseAttrs, false)

		data["domain"] = domain
		data["parentId"] = parent["id"]
		data["attributes"] = attr
		PostObj(ent, "device", data)

	case SEPARATOR, CORIDOR, GROUP:
		//name, category, domain, pid

		if ent != GROUP {
			data["domain"] = domain
			data["parentId"] = parent["id"]
		}
		attr := map[string]interface{}{}

		if ent == SEPARATOR {
			//serialiseAttr(attr)
			data["attributes"] = attr
			PostObj(ent, "separator", data)
		} else if ent == CORIDOR {
			PostObj(ent, "corridor", data)
		} else {
			PostObj(ent, "group", data)
		}

	case STRAYSENSOR:
		attr = data["attributes"].(map[string]interface{})
		if _, ok := attr["template"]; ok {
			//GetOCLIAtrributesTemplateHelper(attr, data, DEVICE)
			tmpl := fetchTemplate(attr["template"].(string), STRAYSENSOR)
			MergeMaps(attr, tmpl, true)
		} else {
			attr["template"] = ""
		}
		PostObj(ent, "stray-sensor", data)

	case STRAY_DEV:
		attr = data["attributes"].(map[string]interface{})
		if _, ok := attr["template"]; ok {
			GetOCLIAtrributesTemplateHelper(attr, data, DEVICE)
		} else {
			attr["template"] = ""
		}
		PostObj(ent, "stray-device", data)
	}
}

//If user provided templates, get the JSON
//and parse into templates
func GetOCLIAtrributesTemplateHelper(attr, data map[string]interface{}, ent int) {
	//Inner func declaration used for importing
	//data from templates
	attrSerialiser := func(someVal interface{}, idx string, ent int) string {
		if x, ok := someVal.(int); ok {
			if ent == DEVICE || ent == ROOM {
				return strconv.Itoa(x)
			}
			return strconv.Itoa(x / 10)
		} else if x, ok := someVal.(float64); ok {
			if ent == DEVICE || ent == ROOM {
				return strconv.FormatFloat(x, 'G', -1, 64)
			}
			return strconv.FormatFloat(x/10.0, 'G', -1, 64)
		} else {
			msg := "Warning: Invalid " + idx +
				" value detected in size." +
				" Resorting to default"
			println(msg)
			return "5"
		}
	}

	if q, ok := attr["template"]; ok {
		if qS, ok := q.(string); ok {
			//Determine the type of template
			tInt := 0
			if ent == ROOM {
				tInt = ROOMTMPL
			} else {
				tInt = OBJTMPL
			} //End of determine block

			if tmpl := fetchTemplate(qS, tInt); tmpl != nil {
				//MergeMaps(attr, tmpl, true)
				key := determineStrKey(tmpl, []string{"sizeWDHmm", "sizeWDHm"})

				if sizeInf, ok := tmpl[key].([]interface{}); ok && len(sizeInf) == 3 {
					var xS, yS, zS string
					xS = attrSerialiser(sizeInf[0], "x", ent)
					yS = attrSerialiser(sizeInf[1], "y", ent)
					zS = attrSerialiser(sizeInf[2], "height", ent)

					attr["size"] = "{\"x\":" + xS + ", \"y\":" + yS + "}"
					attr["height"] = zS

					if ent == DEVICE {
						attr["sizeUnit"] = "mm"
						attr["heightUnit"] = "mm"
						if tmpx, ok := tmpl["attributes"]; ok {
							if x, ok := tmpx.(map[string]interface{}); ok {
								if tmp, ok := x["type"]; ok {
									if t, ok := tmp.(string); ok {
										if t == "chassis" || t == "server" {
											res := 0
											if val, ok := sizeInf[2].(float64); ok {
												res = int((val / 1000) / 0.04445)
											} else if val, ok := sizeInf[2].(int); ok {
												res = int((float64(val) / 1000) / 0.04445)
											} else {
												//Resort to default value
												msg := "Warning, invalid value provided for" +
													" sizeU. Defaulting to 5"
												println(msg)
												res = int((5 / 1000) / 0.04445)
											}
											attr["sizeU"] = strconv.Itoa(res)

										}
									}
								}
							}
						}

					} else if ent == ROOM {
						attr["sizeUnit"] = "m"
						attr["heightUnit"] = "m"

						//Copy additional Room specific attributes
						var tmp []byte
						CopyAttr(attr, tmpl, "technicalArea")
						if _, ok := attr["technicalArea"]; ok {
							//tmp, _ := json.Marshal(attr["technicalArea"])
							attr["technical"] = attr["technicalArea"]
							delete(attr, "technicalArea")
						}

						CopyAttr(attr, tmpl, "reservedArea")
						if _, ok := attr["reservedArea"]; ok {
							//tmp, _ = json.Marshal(attr["reservedArea"])
							attr["reserved"] = attr["reservedArea"]
							delete(attr, "reservedArea")
						}

						parseReservedTech(attr)

						CopyAttr(attr, tmpl, "separators")
						if _, ok := attr["separators"]; ok {
							tmp, _ = json.Marshal(attr["separators"])
							attr["separators"] = string(tmp)
						}

						CopyAttr(attr, tmpl, "tiles")
						if _, ok := attr["tiles"]; ok {
							tmp, _ = json.Marshal(attr["tiles"])
							attr["tiles"] = string(tmp)
						}

						CopyAttr(attr, tmpl, "rows")
						if _, ok := attr["rows"]; ok {
							tmp, _ = json.Marshal(attr["rows"])
							attr["rows"] = string(tmp)
						}

						CopyAttr(attr, tmpl, "aisles")
						if _, ok := attr["aisles"]; ok {
							tmp, _ = json.Marshal(attr["aisles"])
							attr["aisles"] = string(tmp)
						}

						CopyAttr(attr, tmpl, "colors")
						if _, ok := attr["colors"]; ok {
							tmp, _ = json.Marshal(attr["colors"])
							attr["colors"] = string(tmp)
						}

					} else {
						attr["sizeUnit"] = "cm"
						attr["heightUnit"] = "cm"
					}

					//Copy Description
					if _, ok := tmpl["description"]; ok {
						if descTable, ok := tmpl["description"].([]interface{}); ok {
							data["description"] = descTable
						} else {
							data["description"] = []interface{}{tmpl["description"]}
						}
					} else {
						data["description"] = []string{}
					}

					//fbxModel section
					if check := CopyAttr(attr, tmpl, "fbxModel"); !check {
						attr["fbxModel"] = ""
					}

					//Copy orientation if available
					CopyAttr(attr, tmpl, "orientation")

					//Merge attributes if available
					if tmplAttrsInf, ok := tmpl["attributes"]; ok {
						if tmplAttrs, ok := tmplAttrsInf.(map[string]interface{}); ok {
							MergeMaps(attr, tmplAttrs, false)
						}
					}
				} else {
					println("Warning, invalid size value in template.",
						"Default values will be assigned")

				}
			} else {
				attr["template"] = ""
				println("Warning: template was not found.",
					"it will not be used")
				l.GetWarningLogger().Println("Invalid data type or incorrect name used to invoke template")
			}

		} else {
			attr["template"] = ""
			println("Warning: template must be a string that",
				" refers to an existing imported template.",
				q, " will not be used")
			l.GetWarningLogger().Println("Invalid data type used to invoke template")
		}

	} else {
		attr["template"] = ""
		//Serialise size and posXY if given
		if _, ok := attr["size"].(string); ok {
			attr["size"] = serialiseAttr(attr, "size")
		} else {
			attr["size"] = serialiseAttr2(attr, "size")
		}
	}
}

//Used for Unity Client commands
func HandleUI(data map[string]interface{}) {
	//Extra code for the highlight command
	//since client wants an object ID instead of the name/path
	if data["type"].(string) == "ui" &&
		(data["data"].(map[string]interface{})["command"] == "highlight" ||
			data["data"].(map[string]interface{})["command"] == "hl") {

		//check if the object to highlight was provided as a string
		if objArg, ok := data["data"].(map[string]interface{})["data"].(string); ok {

			if objArg == "" || objArg == "." {
				objArg = State.CurrPath
			} else if string(objArg[0]) == "/" {
				//do nothing
			} else {
				objArg = State.CurrPath + "/" + objArg
			}

			obj, _ := GetObject(objArg, true)
			if obj != nil {
				data["data"].(map[string]interface{})["data"] = obj["id"]
			} else {
				println("Please provide a valid path")
				return
			}
		} else if data["data"].(map[string]interface{})["data"] == nil {
			l.GetWarningLogger().Println("Invalid parameter provided for highlighting")
			println("OGREE: Error Invalid parameter provided for highlighting")
			return
		}
	}
	Disp(data)
	InformUnity("POST", "HandleUI", -1, data)
}

func FocusUI(path string) {
	var id string
	if path != "" {
		obj, e := GetObject(path, true)
		if e != "" {
			println(e)
		}
		id = obj["id"].(string)
	} else {
		id = ""
	}

	data := map[string]interface{}{"type": "focus", "data": id}
	InformUnity("POST", "FocusUI", -1, data)
}

func LinkObject(paths []interface{}) {

	var h []map[string]interface{}

	//Stray-device retrieval and validation
	sdev, _ := GetObject(paths[0].(string), true)
	//println("DEBUG OUR PATH 1st:", spath)
	if sdev == nil {
		println("Object doesn't exist")
		return
	}
	if _, ok := sdev["category"]; !ok {
		l.GetWarningLogger().Println("Attempted to link non stray-device ")
		println("Error: Invalid object. Only stray-devices can be linked")
		return
	}
	if cat, _ := sdev["category"]; cat != "stray-device" {
		l.GetWarningLogger().Println("Attempted to link non stray-device ")
		println("Error: Invalid object. Only stray-devices can be linked")
		return
	}

	//Retrieve the stray-device hierarchy
	h = GetHierarchy(paths[0].(string), 50, true)

	//Parent retrieval and validation block
	parent, _ := GetObject(paths[1].(string), true)
	if parent == nil {
		println("Destination is not valid")
		return
	}
	if _, ok := parent["category"]; !ok {
		l.GetWarningLogger().Println("Attempted to link with invalid target")
		println("Error: Invalid destination object")
		println("Please use a rack or a device as a link target")
		return
	}
	if cat, _ := parent["category"].(string); cat != "device" && cat != "rack" {
		l.GetWarningLogger().Println("Attempted to link with invalid target")
		println("Error: Invalid destination object")
		println("Please use a rack or a device as a link target")
		return
	}

	//Need to make sure that origin and destination are
	//not the same!
	if parent["id"] == sdev["id"] && parent["name"] == sdev["name"] {
		l.GetWarningLogger().Println("Attempted to object to itself")
		println("Error you must provide a unique stray-device" +
			" and a unique destination for it")
	}

	//Ensure that the stray device can be imported as device
	//First set the parentId of stray device to point to parent ID
	//Then dive, set the parentID (Which PID is not exactly important
	//we just need to point to a valid PID.)
	//and invoke API validation endpoint
	sdev["parentId"] = parent["id"]
	if len(paths) == 3 {
		//sdev[]
		if attrInf, ok := sdev["attributes"]; ok {
			//attr["slot"] = paths[2]
			if attr, ok := attrInf.(map[string]interface{}); ok {
				attr["slot"] = paths[2]
			} else {
				sdev["attributes"] = map[string]interface{}{"slot": paths[2]}
			}
		} else {
			sdev["attributes"] = map[string]interface{}{"slot": paths[2]}
		}
	}

	sdev = PostObj(DEVICE, "device", sdev)
	if sdev == nil {
		println("Aborting link operation")
		return
	}

	var localValid func(x []map[string]interface{}, entity string, pid interface{}) (bool, map[string]interface{})
	localValid = func(x []map[string]interface{}, entity string, pid interface{}) (bool, map[string]interface{}) {
		if x != nil {
			for i := range x {
				x[i]["parentId"] = pid

				var ent string
				catInf := x[i]["category"]
				if catInf == "device" {
					ent = "stray-device"
				} else if catInf == "sensor" {
					ent = "stray-sensor"
				} else if catInf == "stray-device" {
					ent = "device"
				} else if catInf == "stray-sensor" {
					ent = "sensor"
					//x[i]["cate"]
				} else {
					ent = entity
				}

				res := ValidateObj(x[i], ent, true)
				if res == false {
					return false, x[i]
				}

				childrenInfArr := x[i]["children"]
				if childrenInfArr != nil {
					children := infArrToMapStrinfArr(childrenInfArr.([]interface{}))
					localValid(children, entity, pid)
				}
			}
		}
		return true, nil
	}

	//valid, x := validFn(h, "device", parent["id"])
	valid, x := localValid(h, "device", sdev["id"])
	if !valid {
		desiredObj := MapStrayString(x["category"].(string))
		println("In the target's hierarchy the following "+
			x["category"].(string)+" does not satisfy "+
			desiredObj+" validation requirements: ", x["name"].(string))
		println("Aborting link operation")
		DeleteObj(paths[1].(string) + "/" + sdev["name"].(string))
		l.GetWarningLogger().Println("Link failure")
		return
	}

	var localfn func(x []map[string]interface{}, pid interface{})
	localfn = func(x []map[string]interface{}, pid interface{}) {
		if x != nil {
			for i := range x {
				var entInt int
				var ent string
				x[i]["parentId"] = pid
				childrenInfArr := x[i]["children"]
				delete(x[i], "children")

				if x[i]["category"].(string) == "stray-sensor" {
					ent = "sensor"
					entInt = SENSOR
				} else {
					entInt = DEVICE
					ent = "device"
				}

				newObj := PostObj(entInt, ent, x[i])

				if childrenInfArr != nil {
					var newpid interface{}
					if entInt == DEVICE {
						newpid = newObj["id"]
					} else {
						newpid = pid
					}

					children := infArrToMapStrinfArr(childrenInfArr.([]interface{}))
					localfn(children, newpid)
				}
			}
		}
	}

	//Create the device and Reconstruct it's hierarchy
	localfn(h, sdev["id"])

	//Delete the stray-device
	DeleteObj(paths[0].(string))
}

//This function validates a hierarchy to be imported into another category
func validFn(x []map[string]interface{}, entity string, pid interface{}) (bool, map[string]interface{}) {
	if x != nil {
		for i := range x {
			x[i]["parentId"] = pid

			var ent string
			/*if catInf, _ := x[i]["category"]; catInf != entity {
				ent = catInf.(string)
			} else {
				ent = entity
			}*/
			catInf := x[i]["category"]
			if catInf == "device" {
				ent = "stray-device"
			} else if catInf == "sensor" {
				ent = "stray-sensor"
			} else if catInf == "stray-device" {
				ent = "device"
			} else if catInf == "stray-sensor" {
				ent = "sensor"
			} else {
				ent = entity
			}

			res := ValidateObj(x[i], ent, true)
			if res == false {
				return false, x[i]
			}

			childrenInfArr := x[i]["children"]
			if childrenInfArr != nil {
				children := infArrToMapStrinfArr(childrenInfArr.([]interface{}))
				validFn(children, entity, pid)
			}
		}
	}
	return true, nil
}

func fn(x []map[string]interface{}, pid interface{}, entity string, ent int) {
	if x != nil {
		for i := range x {
			x[i]["parentId"] = pid
			childrenInfArr := x[i]["children"]
			delete(x[i], "children")

			var entStr string
			catInf := x[i]["category"]
			if catInf == "device" {
				entStr = "stray-device"
			} else if catInf == "sensor" {
				entStr = "stray-sensor"
			} else {
				entStr = entity
			}

			newObj := PostObj(ent, entStr, x[i])

			if childrenInfArr != nil {
				newpid := newObj["id"]
				children := infArrToMapStrinfArr(childrenInfArr.([]interface{}))
				fn(children, newpid, entity, ent)
			}
		}
	}
}

//paths should only have a length of 1 or 2
func UnlinkObject(paths []interface{}) {
	//paths[0] ===> device to unlink
	//paths[1] ===> new location in stray-dev (can be optionally empty)
	dev := map[string]interface{}{}
	h := []map[string]interface{}{}

	//first we need to check that the path corresponds to a device
	//we also need to ignore groups
	//arbitrarily set depth to 50 since it doesn't make sense
	//for a device to have a deeper hierarchy paths[0].(string)
	dev, _ = GetObject(paths[0].(string), true)
	if dev == nil {
		l.GetErrorLogger().Println("User attempted to unlink non-existing object")
		println("Error: This object does not exist ")
		return
	}

	//Exit if device not found
	if _, ok := dev["category"]; !ok {
		l.GetErrorLogger().Println("User attempted to unlink non-device object")
		println("Error: This object is not a device. You can only unlink devices ")
		return
	}

	if catInf, _ := dev["category"].(string); catInf != "device" {
		l.GetErrorLogger().Println("User attempted to unlink non-device object")
		println("Error: This object is not a device. You can only unlink devices ")
		return
	}

	h = GetHierarchy(paths[0].(string), 50, true)

	//Dive POST
	var parent map[string]interface{}

	if len(paths) > 1 {
		if parentObjPath, _ := paths[1].(string); parentObjPath != "" {
			parent, _ = GetObject(parentObjPath, true)
			if parent != nil {
				dev["parentId"] = parent["id"]
			}
		}
	}

	if parent == nil {
		DeleteAttr(dev, "parentId")
	}

	newDev := PostObj(STRAY_DEV, "stray-device", dev)
	if newDev == nil {
		l.GetWarningLogger().Println("Unable to unlink target: ", paths[0].(string))
		println("Error: Unable to unlink target: ", paths[0].(string))
		return
	}
	var newPID interface{}
	newPID = newDev["id"]

	if ok, obj := validFn(h, "stray-device", nil); !ok {
		println("Object with name: ", obj["name"].(string), " could not be added")
		println("Unable to unlink")

		//Would also have to delete the parent object in this case
		DeleteObj("/Physical/Stray/Device/" + dev["name"].(string))
		return
	}

	fn(h, newPID, "stray-device", STRAY_DEV)

	//Delete device and we are done
	DeleteObj(paths[0].(string))
}

//Unity UI will draw already existing objects
//by retrieving the hierarchy
func Draw(x string, depth int) {
	if depth == 0 {
		res, _ := GetObject(x, true)
		data := map[string]interface{}{"type": "create", "data": res}
		InformUnity("POST", "Draw", 0, data)
		return
	}

	res := GetHierarchy(x, depth, true)
	if res == nil {
		l.GetErrorLogger().Println("User attempted to draw non drawable object")
		return
	}
	data := map[string]interface{}{"type": "create", "data": res[0]}

	//0 to include the JSON filtration
	InformUnity("POST", "Draw", 0, data)
}

func IsEntityDrawable(path string) bool {
	ans := false
	//Fix path
	if path == "" || path == "." {
		path = (State.CurrPath)
	} else if string(path[0]) == "/" {
		//Do nothing
	} else {
		path = (State.CurrPath + "/" + path)
	}

	//Get Object first
	obj, _ := GetObject(path, true)
	if obj == nil {
		l.GetWarningLogger().Println("Error: object was not found")
		return false
	}

	//Check entity by looking @ category
	//Return if it is drawable in Unity
	if catInf, ok := obj["category"]; ok {
		if category, ok := catInf.(string); ok {
			ans = IsDrawableEntity(category) //State Controller call
			println(ans)
			return ans
		}
	}
	println("false")
	return false
}

func IsAttrDrawable(path, attr string, object map[string]interface{}, silence bool) bool {
	ans := false
	var category string
	var templateJson map[string]interface{}
	if object == nil {
		//Fix path
		if path == "" || path == "." {
			path = (State.CurrPath)
		} else if string(path[0]) == "/" {
			//Do nothing
		} else {
			path = (State.CurrPath + "/" + path)
		}

		//Get Object first
		obj, err := GetObject(path, true)
		if obj == nil {
			l.GetWarningLogger().Println(err)
			if silence != true {
				println(err)
			}

			return false
		}

		//Ensure that we can get the category
		//from object
		if catInf, ok := obj["category"]; ok {
			if cat, ok := catInf.(string); !ok {
				l.GetErrorLogger().Println("Object does not have category")
				if silence != true {
					println("Error: Object does not have category")
				}

				return false
			} else if EntityStrToInt(cat) == -1 {
				l.GetErrorLogger().Println("Object has invalid category")
				if silence != true {
					println("Error: Object does has invalid category")
				}

				return false
			}
		} else {
			l.GetErrorLogger().Println("Object does not have category")
			if silence != true {
				println("Error: Object does not have category")
			}

			return false
		}
		//Check is Done
		category = obj["category"].(string)
	} else {
		if catInf, ok := object["category"]; ok {
			if cat, ok := catInf.(string); !ok {
				l.GetErrorLogger().Println("Object does not have category")
				if silence != true {
					println("Error: Object does not have category")
				}

				return false
			} else if EntityStrToInt(cat) == -1 {
				l.GetErrorLogger().Println("Object has invalid category")
				if silence != true {
					println("Error: Object does has invalid category")
				}

				return false
			}
		} else {
			l.GetErrorLogger().Println("Object does not have category")
			if silence != true {
				println("Error: Object does not have category")
			}

			return false
		}
		category = object["category"].(string)
	}

	templateJson = State.DrawableJsons[category]

	//Return true here by default
	if templateJson == nil {
		if silence != true {
			println(true)
		}

		return true
	}
	switch attr {
	case "id", "name", "category", "parentID",
		"description", "domain", "parentid", "parentId":
		if val, ok := templateJson[attr]; ok {
			if _, ok := val.(bool); ok {
				ans = val.(bool)
				if silence != true {
					println(ans)
				}

				return ans
			}
		}
		ans = false

	default:
		//ans = templateJson["attributes"].(map[string]interface{})[attr].(bool)
		if tmp, ok := templateJson["attributes"]; ok {
			if attributes, ok := tmp.(map[string]interface{}); ok {
				if val, ok := attributes[attr]; ok {
					if _, ok := val.(bool); ok {
						ans = val.(bool)
						if silence != true {
							println(ans)
						}
						return ans
					}
				}
			}
		}
		ans = false
	}

	if silence != true {
		println(ans)
	}
	return ans
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
			UpdateObj(k, "", "", data, false)
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
	var URL string

	if cat, _ := data["category"]; cat == "room" {
		//Room template
		URL = State.APIURL + "/api/room_templates"
	} else {
		//Obj template
		URL = State.APIURL + "/api/obj_templates"
	}

	r, e := models.Send("POST", URL, GetKey(), data)
	if e != nil {
		l.GetErrorLogger().Println(e.Error())
		println("Error: ", e.Error())
	}

	if r.StatusCode == http.StatusCreated {
		println("Template Loaded")
	} else {
		l.GetWarningLogger().Println("Couldn't load template, Status Code: ", r.StatusCode)
		println("Error template wasn't loaded")
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
			InformUnity("POST", "SetClipBoard", -1, data)
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
			l.GetWarningLogger().Println(msg)
			println(msg)
		} else {
			State.UnityClientAvail = val.(bool)
			if val.(bool) == true {
				TriggerListen(*State.Terminal, State.ListenAddr)
			}
			println("Unity Environment variable set")
		}

	case "Filter":
		if _, ok := val.(bool); !ok {
			msg := "Can only assign bool values for Filter Env Var"
			l.GetWarningLogger().Println(msg)
			println(msg)
		} else {
			State.FilterDisplay = val.(bool)
			println("Filter Display Environment variable set")
		}

	case "Analyser":
		if _, ok := val.(bool); !ok {
			msg := "Can only assign bool values for SAnalyser Env Var"
			l.GetWarningLogger().Println(msg)
			println(msg)
		} else {
			State.Analyser = val.(bool)
			println("Static Analyser Environment variable set")
		}

	default:
		println(arg + " is not an environment variable")
	}
}

//Utility functions
func determineStrKey(x map[string]interface{}, possible []string) string {
	for idx := range possible {
		if _, ok := x[possible[idx]]; ok {
			return possible[idx]
		}
	}
	return "" //The code should not reach this point!
}

//Serialising size & posXY is inefficient but
//the team wants it for now
//"size":"[25,29.4,0]" -> "size": "{\"x\":25,\"y\":29.4,\"z\":0}"
func serialiseAttr(attr map[string]interface{}, want string) string {
	var newSize string
	if size, ok := attr[want].(string); ok {
		left := strings.Index(size, "[")
		right := strings.Index(size, "]")
		coords := []string{"x", "y", "z"}

		if left != -1 && right != -1 {
			var length int
			subStr := size[left+1 : right]
			nums := strings.Split(subStr, ",")

			if len(nums) == 3 && want == "size" {
				length = 2
			} else {
				length = len(nums)
			}

			for idx := 0; idx < length; idx++ {
				newSize += "\"" + coords[idx] + "\":" + nums[idx]

				if idx < length-1 {
					newSize += ","
				}
			}
			newSize = "{" + newSize + "}"

			if len(nums) == 3 && want == "size" {
				attr["height"] = nums[2]
			}
		}
	}
	return newSize
}

//Same utility func as above but we have an arbitrary array
//and want to cast it to -> "size": "{\"x\":25,\"y\":29.4,\"z\":0}"
func serialiseAttr2(attr map[string]interface{}, want string) string {
	var newSize string
	if items, ok := attr[want].([]interface{}); ok {
		coords := []string{"x", "y", "z"}
		var length int

		if len(items) == 3 && want == "size" {
			length = 2
		} else {
			length = len(items)
		}

		for idx := 0; idx < length; idx++ {
			r := bytes.NewBufferString("")
			fmt.Fprintf(r, "%v ", items[idx])
			//itemStr :=
			newSize += "\"" + coords[idx] + "\":" + r.String()

			if idx < length-1 {
				newSize += ","
			}
		}
		newSize = "{" + newSize + "}"

		if len(items) == 3 && want == "size" {
			if _, ok := items[2].(int); ok {
				items[2] = strconv.Itoa(items[2].(int))
			} else if _, ok := items[2].(float64); ok {
				items[2] = strconv.FormatFloat(items[2].(float64), 'G', -1, 64)
			}
			attr["height"] = items[2]
		}
	}
	return newSize
}

//This func is used for when the user wants to filter certain
//attributes from being sent/displayed to Unity viewer client
func GenerateFilteredJson(x map[string]interface{}) map[string]interface{} {
	ans := map[string]interface{}{}
	attrs := map[string]interface{}{}
	if catInf, ok := x["category"]; ok {
		if cat, ok := catInf.(string); ok {
			if EntityStrToInt(cat) != -1 {

				//Start the filtration
				for i := range x {
					if i == "attributes" {
						for idx := range x[i].(map[string]interface{}) {
							if IsAttrDrawable("", idx, x, true) == true {
								attrs[idx] = x[i].(map[string]interface{})[idx]
							}
						}
					} else {
						if IsAttrDrawable("", i, x, true) == true {
							ans[i] = x[i]
						}
					}
				}
				if len(attrs) > 0 {
					ans["attributes"] = attrs
				}
				return ans
			}
		}
	}
	return x //Nothing will be filtered
}

//Display contents of []map[string]inf array
func DispMapArr(x []map[string]interface{}) {
	for idx := range x {
		println()
		println()
		println("OBJECT: ", idx)
		displayObject(x[idx])
		println()
	}
}

//displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func LoadArrFromResp(resp map[string]interface{}, idx string) []interface{} {
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if objs, ok1 := data[idx].([]interface{}); ok1 {
			return objs
		}
	}
	return nil
}

//Messages Unity Client
func InformUnity(method, caller string, entity int, data map[string]interface{}) {
	//If unity is available message it
	//otherwise do nothing
	if State.UnityClientAvail == true {
		if entity > -1 && entity < SENSOR+1 {
			data = GenerateFilteredJson(data)
		}
		e := models.ContactUnity(method, State.UnityClientURL, data, State.Timeout, State.DebugLvl)
		if e != nil {
			l.GetWarningLogger().Println("Unable to contact Unity Client @" + caller)
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
				l.GetWarningLogger().Println("Conflict while merging maps")
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
func PreProPath(Path string) []string {
	var pathSplit []string
	switch Path {
	case "":
		pathSplit = strings.Split(State.CurrPath, "/")
		pathSplit = pathSplit[2:]
	default:
		if Path[0] != '/' && len(State.CurrPath) > 1 {
			pathSplit = strings.Split(path.Clean(State.CurrPath+"/"+Path), "/")
			pathSplit = pathSplit[2:]

		} else {
			pathSplit = strings.Split(path.Clean(Path), "/")
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
	if path[0] == "Stray" {
		var objs string
		if len(path) > 1 && path[1] == "Device" {
			basePath += "/stray-devices"
			objs = "/devices/"
		} else if len(path) > 1 && path[1] == "Sensor" {
			basePath += "/stray-sensors"
			objs = "/stray-sensors/"
		}
		//sensorPath := basePath

		if len(path) > 2 { // Check for name

			basePath += "/" + path[2]
			//sensorPath += "/" + path[2]
			for i := 3; i < len(path); i++ {
				basePath += objs + path[i]
				//sensorPath += "/stray-sensors/" + path[i]
			}
		}

		//if basePath == sensorPath {
		return []string{basePath}
		//}
		//return []string{basePath, sensorPath}
	}

	if path[0] == "Domain" {
		var objs string
		if len(path) > 1 {
			basePath += "/domains/" + path[1]
			objs = "/domains/"
		}

		if len(path) > 2 { // Check for name

			basePath += objs + path[2]
			for i := 3; i < len(path); i++ {
				basePath += objs + path[i]
			}
		}
		return []string{basePath}
	}

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

	//Check if path is templates or groups type or stray device
	if path[0] == "Stray" {
		var objs string
		if len(path) > 1 && path[1] == "Device" {
			basePath += "/stray-devices"
			objs = "/devices"
		} else if len(path) > 1 && path[1] == "Sensor" {
			basePath += "/stray-sensors"
			objs = "/stray-sensors"
		}

		if len(path) > 2 { // Check for name
			basePath += "/" + path[2] + objs
			for i := 3; i < len(path); i++ {
				basePath += "/" + path[i] + objs
			}

		}
		return []string{basePath}
	}

	if path[0] == "Domain" {
		basePath += "/domains"
		objs := "/domains"
		if len(path) > 1 { // Check for name
			basePath += "/" + path[1] + objs

		}

		if len(path) > 2 {
			//basePath += "/" + path[2] + objs
			for i := 2; i < len(path); i++ {
				basePath += "/" + path[i] + objs
			}
		}

		return []string{basePath}
	}

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

//Helper func that safely deletes a string key in a map
func DeleteAttr(x map[string]interface{}, key string) {
	if _, ok := x[key]; ok {
		delete(x, key)
	}
}

//Helper func that safely copies a value in a map
func CopyAttr(dest, source map[string]interface{}, key string) bool {
	if _, ok := source[key]; ok {
		dest[key] = source[key]
		return true
	}
	return false
}

//Helper function for GetOCLIAttr which retrieves
//template from server if available, this func mainly helps
//to keep code organised
func fetchTemplate(name string, objType int) map[string]interface{} {
	var URL string
	if objType == ROOMTMPL {
		URL = State.APIURL + "/api/room_templates/" + name
	} else {
		URL = State.APIURL + "/api/obj_templates/" + name
	}
	r, e := models.Send("GET", URL, GetKey(), nil)
	res := ParseResponse(r, e, "fetch template")
	if res != nil {
		if tmplInf, ok := res["data"]; ok {
			if tmpl, ok := tmplInf.(map[string]interface{}); ok {
				return tmpl
			}
		}
	}

	return nil
}

//Helper func is used to check if sizeU is numeric
//this is necessary since the OCLI command for creating a device
//needs to distinguish if the parameter is a valid sizeU or template
func checkNumeric(x interface{}) bool {
	switch x.(type) {
	case int, float64, float32:
		return true
	default:
		return false
	}
}

//Hack function for the reserved and technical areas
//which copies that room areas function in ast.go
//[room]:areas=[r1,r2,r3,r4]@[t1,t2,t3,t4]
func parseReservedTech(x map[string]interface{}) map[string]interface{} {
	var reservedStr string
	var techStr string
	if reserved, ok := x["reserved"].([]interface{}); ok {
		if tech, ok := x["technical"].([]interface{}); ok {
			if len(reserved) == 4 && len(tech) == 4 {
				r4 := bytes.NewBufferString("")
				fmt.Fprintf(r4, "%v", reserved[3].(float64))
				r3 := bytes.NewBufferString("")
				fmt.Fprintf(r3, "%v", reserved[2].(float64))
				r2 := bytes.NewBufferString("")
				fmt.Fprintf(r2, "%v", reserved[1].(float64))
				r1 := bytes.NewBufferString("")
				fmt.Fprintf(r1, "%v", reserved[0].(float64))

				t4 := bytes.NewBufferString("")
				fmt.Fprintf(t4, "%v", tech[3].(float64))
				t3 := bytes.NewBufferString("")
				fmt.Fprintf(t3, "%v", tech[2].(float64))
				t2 := bytes.NewBufferString("")
				fmt.Fprintf(t2, "%v", tech[1].(float64))
				t1 := bytes.NewBufferString("")
				fmt.Fprintf(t1, "%v", tech[0].(float64))

				reservedStr = "{\"left\":" + r4.String() + ",\"right\":" + r3.String() + ",\"top\":" + r1.String() + ",\"bottom\":" + r2.String() + "}"
				techStr = "{\"left\":" + t4.String() + ",\"right\":" + t3.String() + ",\"top\":" + t1.String() + ",\"bottom\":" + t2.String() + "}"
				x["reserved"] = reservedStr
				x["technical"] = techStr
			}
		}
	}
	return x
}
