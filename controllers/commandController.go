package controllers

import (
	"cli/logger"
	l "cli/logger"
	"cli/models"
	u "cli/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

func PWD() string {
	println(State.CurrPath)
	return State.CurrPath
}

func PostObj(ent int, entity string, data map[string]interface{}) (map[string]interface{}, error) {
	var respMap map[string]interface{}
	resp, e := models.Send("POST",
		State.APIURL+"/api/"+entity+"s", GetKey(), data)

	respMap = ParseResponse(resp, e, "POST")
	if respMap == nil {
		return nil, fmt.Errorf("Invalid Response received from API")
	}

	if resp.StatusCode == http.StatusCreated && respMap["status"].(bool) == true {
		//Print success message
		if State.DebugLvl > NONE {
			println(string(respMap["message"].(string)))
		}

		//If ent is in State.ObjsForUnity then notify Unity
		if IsInObjForUnity(entity) == true {
			entInt := EntityStrToInt(entity)
			InformUnity("PostObj", entInt,
				map[string]interface{}{"type": "create", "data": respMap["data"]})
		}

		return respMap["data"].(map[string]interface{}), nil
	}
	return nil, fmt.Errorf(APIErrorPrefix + respMap["message"].(string))
}

// Calls API's Validation
func ValidateObj(data map[string]interface{}, ent string, silence bool) bool {
	var respMap map[string]interface{}
	resp, e := models.Send("POST",
		State.APIURL+"/api/validate/"+ent+"s", GetKey(), data)

	respMap = ParseResponse(resp, e, "POST")
	if respMap == nil {
		if State.DebugLvl > 1 {
			println("Received invalid response from API")
		}
		return false
	}

	if resp.StatusCode == http.StatusOK && respMap["status"].(bool) == true {
		//Print success message
		if silence == false {
			println(string(respMap["message"].(string)))
		}

		return true
	}
	println("Error: ", string(APIErrorPrefix+respMap["message"].(string)))
	println()
	return false
}

func DeleteObj(Path string) bool {
	if Path == "" || Path == "." {
		Path = State.CurrPath

	} else if string(Path[0]) != "/" {
		Path = State.CurrPath + "/" + Path
	}

	//We have to get object first since
	//there is a potential for multiple paths
	//we don't want to delete the wrong object
	objJSON, GETURL := GetObject(Path, true)
	if objJSON == nil {
		if State.DebugLvl > 0 {
			println("Error while deleting Object!")
		}

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
		if State.DebugLvl > 0 {
			println("Error while deleting Object!")
		}

		l.GetWarningLogger().Println("Error while deleting Object!", e)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		println("Success")
	} else {
		if State.DebugLvl > 0 {
			println("Error while deleting Object!")
		}
		l.GetWarningLogger().Println("Error while deleting Object!", e)
		return false
	}

	entity := entities[:len(entities)-1]
	if IsInObjForUnity(entity) == true {
		InformUnity("DeleteObj", -1,
			map[string]interface{}{"type": "delete", "data": objJSON["id"].(string)})
	}

	//Check if deleted object is current path
	if currPathURL == GETURL {
		CD("..")
	}

	return true
}

func DeleteSelection() bool {
	res := false
	if State.ClipBoard != nil {
		for i := range *State.ClipBoard {
			println("Going to delete object: ", (*(State.ClipBoard))[i])
			if res = DeleteObj((*(State.ClipBoard))[i]); res != true {
				l.GetWarningLogger().Println("Couldn't delete obj in selection: ",
					(*(State.ClipBoard))[i])
				if State.DebugLvl > 0 {
					println("Couldn't delete obj in selection: ",
						(*(State.ClipBoard))[i])
				}

			}
			println()
		}

	}
	return res
}

// Search for objects
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
	if jsonResp == nil {
		if State.DebugLvl > 1 {
			println("Received invalid response from API")
		}
		return nil
	}

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
			InformUnity("Search", -1, resp)
		}

		return objects

	}
	return nil
}

// Check if the object exists in API
func CheckObject(path string, silenced bool) (string, bool) {
	pathSplit := PreProPath(path)
	paths := OnlinePathResolve(pathSplit)

	for i := range paths {
		resp, e := models.Send("OPTIONS", paths[i], GetKey(), nil)
		if e != nil {
			if !silenced {
				println(paths[i])
				println(e.Error())
			}

		}
		if resp.StatusCode == http.StatusOK {
			return paths[i], true
		}
	}
	return "", false
}

// Silenced bool
// Useful for LS since
// otherwise the terminal would be polluted by debug statements
func GetObject(path string, silenced bool) (map[string]interface{}, string) {
	var data map[string]interface{}

	pathSplit := PreProPath(path)
	paths := OnlinePathResolve(pathSplit)

	for i := range paths {
		resp, e := models.Send("GET", paths[i], GetKey(), nil)
		if e != nil {
			println(paths[i])
			println(e.Error())
		}
		data = ParseResponse(resp, e, "GET")

		if resp == nil {
			return nil, ""
		}

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
	if State.DebugLvl > 0 && !silenced {
		println("Error finding Object from given path!")
		println(path)
	}

	l.GetWarningLogger().Println("Object to Get not found :", path)

	return nil, ""
}

// This is an auxillary function
// for writing proper JSONs
func GenUpdateJSON(m map[string]interface{}, key string, value interface{}, del bool) bool {
	//Base Cae
	if _, ok := m[key]; ok {
		if del == true { //make a delete
			delete(m, key)
		} else {
			m[key] = value
		}
		return true
	}
	for i := range m {
		//We have a nested map
		if sub, ok := m[i].(map[string]interface{}); ok {
			ret := GenUpdateJSON(sub, key, value, del)
			if ret {
				return true
			}
		}
	}
	return false
}

// This function recursively applies an update to an object and
// the rest of its subentities
func RecursivePatch(Path, id, ent string, data map[string]interface{}) error {
	var entities string
	var URL string
	println("OK. Attempting to update...")

	if data != nil {
		if Path != "" {
			//We have to get object first since
			//there is a potential for multiple paths
			//we don't want to update the wrong object
			objJSON, GETURL := GetObject(Path, true)
			if objJSON == nil {
				l.GetWarningLogger().Println("Error while deleting Object!")
				return fmt.Errorf("error while deleting Object")
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
		if r == nil {
			return fmt.Errorf("Failure while getting root object")
		}
		if e != nil {
			return nil
		}
		recursivePatchAux(r["data"].(map[string]interface{}), data)
		println("Success")
		return nil
	}
	return fmt.Errorf("error! Please enter desired parameters of Object to be updated")
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

// You can either update obj by path or by ID and entity string type
// The deleteAndPut bool is for deleting an attribute
func UpdateObj(Path, id, ent string, data map[string]interface{}, deleteAndPut bool) (map[string]interface{}, error) {
	println("OK. Attempting to update...")
	var resp *http.Response
	var objJSON map[string]interface{}
	var GETURL string

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
			objJSON, GETURL = GetObject(Path, true)
			if objJSON == nil {
				l.GetWarningLogger().Println("Error while getting Object!")
				return nil, fmt.Errorf("error while getting Object")
			}
			entities = path.Base(path.Dir(GETURL))
			URL = State.APIURL + "/api/" + entities + "/" + objJSON["id"].(string)

			//Check if the description keyword was specified
			//if it is we need to do extra processing

			//Local anonfunc for parsing descriptionX
			//where X is a number
			fn := func(description string) (int, error) {
				//Split description and number off of 'i'
				//key := i[:10]
				numStr := description[11:]

				num, e := strconv.Atoi(numStr)
				if e != nil {
					return -1, e
				}

				if num < 0 {
					msg := "Index for description" +
						" cannot be negative"
					return -1, fmt.Errorf(msg)
				}
				return num, nil
			}

			for i := range data {

				if strings.Contains(i, "description") {
					//Modify the JSON itself and send the object
					//JSON instead

					//Get description as an array from JSON
					descInf := objJSON["description"]
					if desc, ok := descInf.([]interface{}); ok {

						if i == "description" {
							if len(desc) == 0 {
								desc = []interface{}{data[i]}
							} else {
								desc[0] = data[i]
							}

							data = map[string]interface{}{"description": desc}
						} else {

							num, e := fn(i)
							if e != nil {
								return nil, e
							}

							if num >= len(desc) {
								//desc[len(desc)-1] = data[i]
								desc = append(desc, data[i])
							} else {
								desc[num] = data[i]
							}

							//We must send a PUT since this will modify the JSON
							i = "description"
							data = map[string]interface{}{"description": desc}

						}

					} else if _, ok := descInf.(string); ok {
						var num int
						var e error

						if i == "description" {
							num = 0
						} else {
							num, e = fn(i)
							if e != nil {
								return nil, e
							}
						}

						//Assume the string takes idx 0
						if num > 0 {
							objJSON["description"] = []interface{}{descInf, data[i]}

						} else {
							objJSON["description"] = []interface{}{data[i]}
						}

					} else { //Description is some invalid value
						objJSON["description"] = []interface{}{data[i]}
					}

				}

				if strings.Contains(i, "temperature_") {
					category := EntityStrToInt(objJSON["category"].(string))
					if category == RACK || category == DEVICE ||
						category == ROOM {
						switch data[i].(type) {
						case float64, float32: //GOOD
						default:
							msg := "The temperature for sensors should be a float"
							return nil, fmt.Errorf(msg)
						}
					}

				}

				if i == "usableColor" || i == "reservedColor" ||
					i == "technicalColor" {
					category := EntityStrToInt(objJSON["category"].(string))
					if category == SITE {
						//Same function as AssertColor in semantic.go
						var colorStr string
						switch data[i].(type) {
						case string, int, float64, float32:
							if _, ok := data[i].(string); ok {
								colorStr = data[i].(string)
							}

							if _, ok := data[i].(int); ok {
								colorStr = strconv.Itoa(data[i].(int))
							}

							if _, ok := data[i].(float32); ok {
								colorStr = strconv.FormatFloat(data[i].(float64), 'f', -1, 64)
							}

							if _, ok := data[i].(float64); ok {
								colorStr = strconv.FormatFloat(data[i].(float64), 'f', -1, 64)
							}

							for len(colorStr) < 6 {
								colorStr = "0" + colorStr
							}

							if len(colorStr) != 6 {
								msg := "Please provide a valid 6 length hex value for the color"
								return nil, fmt.Errorf(msg)
							}

							//Eliminate 'odd length' errors
							if len(colorStr)%2 != 0 {
								colorStr = "0" + colorStr
							}

							_, err := hex.DecodeString(colorStr)
							if err != nil {
								msg := "Please provide a valid 6 length hex value for the color"
								return nil, fmt.Errorf(msg)
							}

						default:
							msg := "Please provide a valid 6 length hex value for the color"
							return nil, fmt.Errorf(msg)
						}
						data[i] = colorStr
					}
				}
			}

		} else {
			entities = ent + "s"
			URL = State.APIURL + "/api/" + entities + "/" + id
		}

		//Make the proper Update JSON
		var e error
		var ogData map[string]interface{}
		if objJSON == nil {
			r, e1 := models.Send("GET", URL, GetKey(), nil)
			objJSON = ParseResponse(r, e1, "GET")
			if objJSON == nil {
				return nil, fmt.Errorf("Couldn't get object for update")
			}
			ogData = objJSON["data"].(map[string]interface{})
		} else {
			ogData = objJSON
		}
		//respGet, e := models.Send("GET", URL, GetKey(), nil)
		//ogData := ParseResponse(respGet, e, "GET")

		attrs := map[string]interface{}{}

		for i := range data {
			// Since all data of obj attributes must be string
			// stringify the data before sending
			if u.IsNestedAttr(i, ent) {
				data[i] = Stringify(data[i])
			}

			found := GenUpdateJSON(ogData, i, data[i], deleteAndPut)
			if !found {
				//The key was not found so let's insert it
				//in attributes
				attrs[i] = data[i]
			}
		}
		if len(attrs) > 0 {
			ogData["attributes"] = attrs
		}

		if deleteAndPut {
			resp, e = models.Send("PUT", URL, GetKey(), ogData)
		} else {
			resp, e = models.Send("PATCH", URL, GetKey(), ogData)
		}

		respJson = ParseResponse(resp, e, "UPDATE")
		if respJson != nil {
			if resp.StatusCode == 200 {
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
					InformUnity("UpdateObj", entInt, message)
				}

			} else {
				if mInf, ok := respJson["message"]; ok {
					if m, ok := mInf.(string); ok {
						return nil, fmt.Errorf(APIErrorPrefix + m)
					}
				}
				msg := "Cannot update. Please ensure that your attributes " +
					"are modifiable and try again. For more details see the " +
					"OGREE wiki: https://github.com/ditrit/OGrEE-3D/wiki"
				return nil, fmt.Errorf(msg)
			}

		}

		data = respJson

	} else {
		println("Error! Please enter desired parameters of Object to be updated")
	}
	return data, nil
}

// Specific update for deleting elements in an array of an obj
func UnsetInObj(Path, attr string, idx int) (map[string]interface{}, error) {
	var arr []interface{}

	//Check for valid idx
	if idx < 0 {
		return nil,
			fmt.Errorf("Index out of bounds. Please provide an index greater than 0")
	}

	//Get the object
	objJSON, _ := GetObject(Path, true)
	if objJSON == nil {
		l.GetWarningLogger().Println("Error while getting Object!")
		return nil, fmt.Errorf("error while getting Object")
	}

	//Check if attribute exists in object
	existing, nested := AttrIsInObj(objJSON, attr)
	if !existing {
		if State.DebugLvl > ERROR {
			logger.GetErrorLogger().Println("Attribute :" + attr + " was not found")
		}
		return nil, fmt.Errorf("Attribute :" + attr + " was not found")
	}

	//Check if attribute is an array
	if nested {
		objAttributes := objJSON["attributes"].(map[string]interface{})
		if _, ok := objAttributes[attr].([]interface{}); !ok {
			if State.DebugLvl > ERROR {
				println("Attribute is not an array")
			}
			return nil, fmt.Errorf("Attribute is not an array")

		}
		arr = objAttributes[attr].([]interface{})

	} else {
		if _, ok := objJSON[attr].([]interface{}); !ok {
			if State.DebugLvl > ERROR {
				logger.GetErrorLogger().Println("Attribute :" + attr + " was not found")
			}
			return nil, fmt.Errorf("Attribute :" + attr + " was not found")
		}
		arr = objJSON[attr].([]interface{})
	}

	//Ensure that we can delete elt in array
	if len(arr) == 0 {
		if State.DebugLvl > ERROR {
			println("Cannot delete anymore elements")
		}
		return nil, fmt.Errorf("Cannot delete anymore elements")
	}

	//Perform delete
	if idx >= len(arr) {
		idx = len(arr) - 1
	}
	arr = slices.Delete(arr, idx, idx+1)

	//Save back into obj
	if nested {
		objJSON["attributes"].(map[string]interface{})[attr] = arr
	} else {
		objJSON[attr] = arr
	}

	//Send to API and update Unity
	entity := objJSON["category"].(string)
	id := objJSON["id"].(string)
	URL := State.APIURL + "/api/" + entity + "s/" + id

	resp, e := models.Send("PUT", URL, GetKey(), objJSON)
	respJson := ParseResponse(resp, e, "UPDATE")
	if respJson != nil {
		if resp.StatusCode == 200 {
			println("Success")

			message := map[string]interface{}{
				"type": "modify", "data": respJson["data"]}

			//Update and inform unity
			if IsInObjForUnity(entity) == true {
				entInt := EntityStrToInt(entity)
				InformUnity("UpdateObj", entInt, message)
			}
		}
	}

	return nil, nil
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
		fmt.Printf("\033[2J\033[H")
	}
}

func LSOG() {

	//Need to add GET /api/version call data here
	r, e := models.Send("GET", State.APIURL+"/api/version", GetKey(), nil)
	parsedResp := ParseResponse(r, e, "get API information request")

	fmt.Println("********************************************")
	fmt.Println("OGREE Shell Information")
	fmt.Println("********************************************")

	fmt.Println("USER EMAIL:", GetEmail())
	fmt.Println("API URL:", State.APIURL+"/api/")
	fmt.Println("UNITY URL:", State.UnityClientURL)
	fmt.Println("BUILD DATE:", BuildTime)
	fmt.Println("BUILD TREE:", BuildTree)
	fmt.Println("BUILD HASH:", BuildHash)
	fmt.Println("COMMIT DATE: ", GitCommitDate)
	fmt.Println("ENV FILE PATH: ", State.EnvFilePath)
	fmt.Println("LOG PATH:", "./log.txt")
	fmt.Println("HISTORY FILE PATH:", State.HistoryFilePath)
	fmt.Println("DEBUG LEVEL: ", State.DebugLvl)

	if parsedResp != nil {
		if apiInfo, ok := parsedResp["data"].(map[string]interface{}); ok {
			fmt.Println("********************************************")
			fmt.Println("API Information")
			fmt.Println("********************************************")
			fmt.Println("BUILD DATE:", apiInfo["BuildDate"].(string))
			fmt.Println("BUILD TREE:", apiInfo["BuildTree"].(string))
			fmt.Println("BUILD HASH:", apiInfo["BuildHash"].(string))
			fmt.Println("COMMIT DATE: ", apiInfo["CommitDate"].(string))

		} else if State.DebugLvl > 1 {
			msg := "Received invalid response from API on GET /api/version"
			l.GetWarningLogger().Println(msg)
			fmt.Println("NOTE: " + msg)
		}

	} else {
		if State.DebugLvl > 1 {
			msg := "Received nil response from API on GET /api/version"
			l.GetWarningLogger().Println(msg)
			fmt.Println("NOTE: " + msg)
		}
	}
}

func LSEnterprise() {
	r, e := models.Send("GET", State.APIURL+"/api/stats",
		GetKey(), nil)
	resp := ParseResponse(r, e, "lsenterprise")
	if resp != nil {
		displayObject(resp)
	}
}

// Displays environment variable values
// and user defined variables and funcs
func Env(userVars, userFuncs map[string]interface{}) {
	fmt.Println("Unity: ", State.UnityClientAvail)
	fmt.Println("Filter: ", State.FilterDisplay)
	fmt.Println("Analyser: ", State.Analyser)
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

	fmt.Println()
	fmt.Println("Currently defined user variables:")
	for name, k := range userVars {
		if k != nil {
			fmt.Println("Name:", name, "  Value: ", k)
		}

	}

	fmt.Println()
	fmt.Println("Currently defined user functions:")
	for name, _ := range userFuncs {
		fmt.Println("Name:", name)
	}
}

func LSOBJECT(x string, entity int) []interface{} {
	var obj map[string]interface{}
	var Path string

	if entity == TENANT { //Special for tenants case
		if x == "/Physical" {
			Path = State.APIURL + "/api"
		} else {
			//Return nothing
			return nil
		}
	} else {
		obj, Path = GetObject(x, true)
		if obj == nil {
			if State.DebugLvl > 0 {
				println("Error finding Object from given path!")
			}

			l.GetWarningLogger().Println("Object to Get not found")
			return nil
		}
	}

	//Retrieve the desired objects under the working path
	entStr := EntityToString(entity) + "s"
	r, e := models.Send("GET", Path+"/"+entStr, GetKey(), nil)
	parsed := ParseResponse(r, e, "list objects")
	if parsed == nil {
		return nil
	}
	return GetRawObjects(parsed)
}

func GetByAttr(x string, u interface{}) {
	var path string
	if x == "" || x == "." {
		path = State.CurrPath

	} else if string(x[0]) == "/" {
		path = x

	} else {
		path = State.CurrPath + "/" + x
	}

	//Let's do a quick GET and check for rack because otherwise we
	//may have to get (costly) many devices and then
	//test if the result is a device array
	obj, url := GetObject(path, true)
	if obj == nil {
		return
	}

	if cat, ok := obj["category"]; !ok || cat != "rack" {
		if State.DebugLvl > 0 {
			println("Error command may only be performed on rack objects!")
		}

		l.GetWarningLogger().Println("Object to Get not found")
		return
	}

	//GET the devices and process the response
	req, code := models.Send("GET", url+"/devices", GetKey(), nil)
	reqParsed := ParseResponse(req, code, "get devices request")
	if reqParsed == nil {
		return
	}
	devInf := reqParsed["data"].(map[string]interface{})["objects"].([]interface{})
	devices := infArrToMapStrinfArr(devInf)

	switch u.(type) {
	case int:
		for i := range devices {
			if attr, ok := devices[i]["attributes"].(map[string]interface{}); ok {
				uStr := strconv.Itoa(u.(int))
				if attr["height"] == uStr {
					displayObject(devices[i])
					return //What if the user placed multiple devices at same height?
				}
			}
		}
	default: //String
		for i := range devices {
			if attr, ok := devices[i]["attributes"].(map[string]interface{}); ok {
				if attr["slot"] == u.(string) {
					displayObject(devices[i])
					return //What if the user placed multiple devices at same slot?
				}
			}
		}
	}
}

// This function display devices in a sorted order according
// to the attribute specified
func LSATTR(x, attr string) {
	var path string
	if x == "" || x == "." {
		path = State.CurrPath

	} else if string(x[0]) == "/" {
		path = x

	} else {
		path = State.CurrPath + "/" + x
	}

	//Let's do a quick GET and check for rack because otherwise we
	//may have to get (costly) many devices and then
	//test if the result is a device array
	obj, url := GetObject(path, true)
	if obj == nil {
		return
	}

	if cat, ok := obj["category"]; !ok || cat != "rack" {
		if State.DebugLvl > 0 {
			println("Error command may only be performed on rack objects!")
		}

		l.GetWarningLogger().Println("Object to Get not found")
		return
	}

	//GET the devices and process the response
	req, code := models.Send("GET", url+"/devices", GetKey(), nil)
	reqParsed := ParseResponse(req, code, "get devices request")
	if reqParsed == nil {
		return
	}
	devInf := reqParsed["data"].(map[string]interface{})["objects"].([]interface{})
	//devices := infArrToMapStrinfArr(devInf)

	sortedDevices := SortObjects(&devInf, attr)

	//Print the objects received
	if len(sortedDevices.GetData()) > 0 {
		println("Devices")
		println()
		sortedDevices.Print()
	}

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
		State.CurrPath, State.PrevPath = State.PrevPath, State.CurrPath

	} else if strings.Count(x, "/") >= 1 {
		exist := false
		var pth string

		if string(x[0]) != "/" {
			pth = State.CurrPath + "/" + x
			exist, _ = CheckPathOnline(pth)
		} else {
			pth = x
			exist, _ = CheckPathOnline(pth)
		}
		if exist == true {
			if State.DebugLvl >= 3 {
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
			if FindNodeInTree(&State.TreeHierarchy, StrToStack(pth), true) != nil {
				State.PrevPath = State.CurrPath
				State.CurrPath = pth
				//println(("DEBUG not in tree either"))
				//println("DEBUG ", x)
				//println()
			} else {
				if State.DebugLvl > 0 {
					println("Path does not exist")
				}

				l.GetWarningLogger().Println("Path: ", x, " does not exist")
			}

		}
	} else {
		if len(State.CurrPath) != 1 {
			if exist, _ := CheckPathOnline(State.CurrPath + "/" + x); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += "/" + x
			} else {
				if State.DebugLvl > 0 {
					println("OGREE: ", x, " : No such object")
				}

				l.GetWarningLogger().Println("No such object: ", x)
			}
		} else {

			if exist, _ := CheckPathOnline(State.CurrPath + x); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += x
			} else {
				if State.DebugLvl > 0 {
					println("OGREE: ", x, " : No such object")
				}

				l.GetWarningLogger().Println("No such object: ", x)
			}
		}
	}
	return State.CurrPath
}

func Help(entry string) {
	var path string
	entry = strings.TrimSpace(entry)
	switch entry {
	case "ls", "pwd", "print", "cd", "tree", "create", "get", "clear",
		"update", "delete", "lsog", "grep", "for", "while", "if", "env",
		"cmds", "var", "unset", "select", "camera", "ui", "hc", "drawable",
		"link", "unlink", "draw", "getu", "getslot", "undraw",
		"lsenterprise":
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

	text, e := os.ReadFile(path)
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

// Function is an abstraction of a normal exit
func Exit() {
	//writeHistoryOnExit(&State.sessionBuffer)
	//runtime.Goexit()
	os.Exit(0)
}

func Tree(x string, depth int) {
	if depth < 0 {
		l.GetWarningLogger().Println("Tree command cannot accept negative value")
		if State.DebugLvl > 0 {
			println("Error: Tree command cannot accept negative value")
		}
		return
	}

	var Path string

	if x == "" || x == "." {
		println(State.CurrPath)
		Path = State.CurrPath
	} else if string(x[0]) == "/" {
		println(x)
		Path = x
	} else {
		println(State.CurrPath + "/" + x)
		Path = State.CurrPath + "/" + x
	}
	tree(Path, depth)
}

func tree(path string, depth int) {
	arr := strings.Split(path, "/")

	if path == "/" {
		//RootWalk
		//if checking "/" doesn't work as intended then
		//test for arr[0] == "" && arr[1] == "" instead
		RootWalk(&State.TreeHierarchy, path, depth)
		return
	}

	switch arr[1] {
	case "Physical":
		//Get the Physical Node!
		physical := FindNodeInTree(&State.TreeHierarchy,
			StrToStack("/Physical"), true)
		PhysicalWalk(physical, "", path, depth)
	case "Logical":

		if len(arr) >= 4 { //This is the threshold
			return
		}

		//Get the Logical Node!
		logi := FindNearestNodeInTree(&State.TreeHierarchy,
			StrToStack(path), true)
		LogicalWalk(logi, "", depth)

	case "Organisation":

		if len(arr) >= 4 { //This refers to a remote path
			//Fetch the object and objects at level then walk
			//Get Object hierarchy and walk
			ObjectAndHierarchWalk(path, "", depth)
			return
		}

		//Get the Organisation Node!
		org := FindNearestNodeInTree(&State.TreeHierarchy,
			StrToStack(path), true)

		OrganisationWalk(org, "", depth)
	default: //Error! This should never occur
		println("DEBUG ERROR!")
		println("DEBUG LEN:", len(arr))
	}

}

func GetHierarchy(x string, depth int, silence bool) []map[string]interface{} {
	//Variable declarations
	var URL, ext, depthStr string
	var ans []map[string]interface{}

	if FindNodeInTree(&State.TreeHierarchy, StrToStack(x), true) != nil {
		if State.DebugLvl > 0 {
			println("This function can only be invoked on an object")
		}
		return nil
	}

	//Get object first
	obj, e := GetObject(x, true)
	if obj == nil {
		if e != "" {
			println("Error: ", e)
		}
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
				if State.DebugLvl > 0 {
					println("Error: " + e.Error())
				}

				l.GetErrorLogger().Println("Error: " + e.Error())
				return nil
			}

			data := ParseResponse(r, nil, "get hierarchy")
			if data == nil {
				l.GetWarningLogger().Println("Hierarchy call response was nil")
				if State.DebugLvl > 0 {
					println("No data")
				}

				return nil
			}

			objs := LoadArrFromResp(data, "children")
			if objs == nil {
				l.GetWarningLogger().Println("No objects found in hierarchy call")
				if State.DebugLvl > 0 {
					println("No objects found in hierarchy call")
				}

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

// Helps to create the Object (thru OCLI syntax)
func GetOCLIAtrributes(Path string, ent int, data map[string]interface{}) error {
	var attr map[string]interface{}
	var parent map[string]interface{}
	var domain string
	var parentURL string

	ogPath := Path
	Path = path.Dir(Path)
	name := path.Base(ogPath)
	if name == "." || name == "" {
		l.GetWarningLogger().Println("Invalid path name provided for OCLI object creation")
		return fmt.Errorf("Invalid path name provided for OCLI object creation")
	}

	data["name"] = name
	data["category"] = EntityToString(ent)
	data["description"] = []interface{}{}

	//Retrieve Parent
	if ent != TENANT && ent != STRAY_DEV && ent != STRAYSENSOR {
		parent, parentURL = GetObject(Path, true)
		if parent == nil {
			return fmt.Errorf("The parent was not found in path")
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
	var err error
	switch ent {
	case TENANT:
		data["domain"] = data["name"]

	case SITE:
		//Default values
		data["domain"] = domain
		data["parentId"] = parent["id"]
		data["attributes"] = map[string]interface{}{}

	case BLDG:
		attr = data["attributes"].(map[string]interface{})

		//Check for template
		if _, ok := attr["template"]; ok {
			GetOCLIAtrributesTemplateHelper(attr, data, BLDG)

		} else {
			//Serialise size and posXY manually instead
			if _, ok := attr["size"].(string); ok {
				attr["size"] = serialiseAttr(attr, "size")
			} else {
				attr["size"] = serialiseAttr2(attr, "size")
			}

			//Since template was not provided, set it empty
			attr["template"] = ""
		}

		if attr["size"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid size value for creating building")
				return fmt.Errorf("Invalid size attribute provided." +
					" \nIt must be an array/list/vector with 3 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		if _, ok := attr["posXY"].(string); ok {
			attr["posXY"] = serialiseAttr(attr, "posXY")
		} else {
			attr["posXY"] = serialiseAttr2(attr, "posXY")
		}

		if attr["posXY"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXY value for creating building")
				return fmt.Errorf("Invalid posXY attribute provided." +
					" \nIt must be an array/list/vector with 2 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		//Check rotation
		if _, ok := attr["rotation"].(float64); ok {
			attr["rotation"] =
				strconv.FormatFloat(attr["rotation"].(float64), 'f', -1, 64)
		}

		attr["posXYUnit"] = "m"
		attr["sizeUnit"] = "m"
		attr["heightUnit"] = "m"
		//attr["height"] = 0 //Should be set from parser by default
		data["parentId"] = parent["id"]
		data["domain"] = domain

	case ROOM:
		attr = data["attributes"].(map[string]interface{})

		baseAttrs := map[string]interface{}{
			"floorUnit": "t",
			"posXYUnit": "m", "sizeUnit": "m",
			"height":     "5",
			"heightUnit": "m"}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		//NOTE this function also assigns value for "size" attribute
		GetOCLIAtrributesTemplateHelper(attr, data, ent)

		if _, ok := attr["posXY"].(string); ok {
			attr["posXY"] = serialiseAttr(attr, "posXY")
		} else {
			attr["posXY"] = serialiseAttr2(attr, "posXY")
		}

		if attr["posXY"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXY value for creating room")
				return fmt.Errorf("Invalid posXY attribute provided." +
					" \nIt must be an array/list/vector with 2 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		//Check rotation
		if _, ok := attr["rotation"].(float64); ok {
			attr["rotation"] =
				strconv.FormatFloat(attr["rotation"].(float64), 'f', -1, 64)
		}

		if attr["size"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid size value for creating room")
				return fmt.Errorf("Invalid size attribute provided." +
					" \nIt must be an array/list/vector with 3 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		data["parentId"] = parent["id"]
		data["domain"] = domain
		data["attributes"] = attr
		if State.DebugLvl >= 3 {
			println("DEBUG VIEW THE JSON")
			Disp(data)
		}

	case RACK:
		attr = data["attributes"].(map[string]interface{})
		parentAttr := parent["attributes"].(map[string]interface{})
		//Save orientation because it gets overwritten by
		//GetOCLIAtrributesTemplateHelper()
		orientation := attr["orientation"]

		baseAttrs := map[string]interface{}{
			"sizeUnit":   "cm",
			"heightUnit": "U",
			"posXYUnit":  parentAttr["floorUnit"],
		}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		GetOCLIAtrributesTemplateHelper(attr, data, ent)

		//Restore the orientation overwritten
		//by the helper func
		attr["orientation"] = orientation

		if attr["size"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid size value for creating rack")
				return fmt.Errorf("Invalid size attribute/template provided." +
					" \nThe size must be an array/list/vector with " +
					"3 elements." + "\n\nIf you have provided a" +
					" template, please check that you are referring to " +
					"an existing template" +
					"\n\nFor more information " +
					"please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		//Serialise posXY if given
		if _, ok := attr["posXYZ"].(string); ok {
			attr["posXYZ"] = serialiseAttr(attr, "posXYZ")
		} else {
			attr["posXYZ"] = serialiseAttr2(attr, "posXYZ")
		}

		if attr["posXYZ"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXYZ value for creating rack")
				return fmt.Errorf("Invalid posXYZ attribute provided." +
					" \nIt must be an array/list/vector with 2 or 3 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		data["parentId"] = parent["id"]
		data["domain"] = domain
		data["attributes"] = attr

	case DEVICE:
		attr = data["attributes"].(map[string]interface{})

		//Special routine to perform on device
		//based on if the parent has a "slot" attribute

		//First check if attr has only posU & sizeU
		//reject if true while also converting sizeU to string if numeric
		//if len(attr) == 2 {
		if sizeU, ok := attr["sizeU"]; ok {
			sizeUValid := checkNumeric(attr["sizeU"])

			if _, ok := attr["template"]; !ok && sizeUValid == false {
				l.GetWarningLogger().Println("Invalid template / sizeU parameter provided for device ")
				return fmt.Errorf("Please provide a valid device template or sizeU")
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
				return fmt.Errorf("Invalid device syntax: If you have provided a template, it was not found")
			}
		}
		//}

		//Process the posU/slot attribute
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

		//Ensure slot is a string
		if _, ok := attr["slot"]; ok {
			if _, ok := attr["slot"].(string); !ok {
				return fmt.Errorf("The slot name must be a string")
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

	case GROUP:
		//name, category, domain, pid
		data["domain"] = domain
		data["parentId"] = parent["id"]
		attr := data["attributes"].(map[string]interface{})

		groups := strings.Join(attr["content"].([]string), ",")
		attr["content"] = groups

	case CORIDOR:
		//name, category, domain, pid
		attr = data["attributes"].(map[string]interface{})

		//Client demands that the group color be
		//the same as Tenant/Domain thus we have
		//to retrieve it
		arr := strings.Split(Path, "/")
		if len(arr) >= 2 {
			for i := range arr {
				if arr[i] == "Physical" {
					tenantName := arr[i+1]

					//GET Tenant/Domain
					r, e := models.Send("GET",
						State.APIURL+"/api/tenants/"+tenantName, GetKey(), nil)
					parsed := ParseResponse(r, e, "get color")
					if parsed == nil {
						msg := "Unable to retrieve color from server"
						return fmt.Errorf(msg)
					}

					if tenantData, ok := parsed["data"]; ok {
						if tenant, ok := tenantData.(map[string]interface{}); ok {
							if attrInf, ok := tenant["attributes"]; ok {
								if a, ok := attrInf.(map[string]interface{}); ok {
									if colorInf, ok := a["color"]; ok {
										if color, ok := colorInf.(string); ok {

											attr["color"] = color
										}
									}
								}
							}
						}
					}

				}
			}
		}

		if attr["color"] == nil {
			return fmt.Errorf("Couldn't get respective color from server")
		}

		data["domain"] = domain
		data["parentId"] = parent["id"]

	case STRAYSENSOR:
		attr = data["attributes"].(map[string]interface{})
		if _, ok := attr["template"]; ok {
			//GetOCLIAtrributesTemplateHelper(attr, data, DEVICE)
			tmpl := fetchTemplate(attr["template"].(string), STRAYSENSOR)
			MergeMaps(attr, tmpl, true)
		} else {
			attr["template"] = ""
		}

	case STRAY_DEV:
		attr = data["attributes"].(map[string]interface{})
		if _, ok := attr["template"]; ok {
			GetOCLIAtrributesTemplateHelper(attr, data, DEVICE)
		} else {
			attr["template"] = ""
		}

	default:
		//Execution should not reach here!
		return fmt.Errorf("Invalid Object Specified!")
	}

	//Stringify the attributes if not already
	if _, ok := data["attributes"]; ok {
		if attributes, ok := data["attributes"].(map[string]interface{}); ok {
			for i := range attributes {
				attributes[i] = Stringify(attributes[i])
			}
		}
	}

	//Because we already stored the string conversion in category
	//we can do the conversion for templates here
	data["category"] = strings.Replace(data["category"].(string), "_", "-", 1)

	_, err = PostObj(ent, data["category"].(string), data)
	if err != nil {
		return err
	}
	return nil
}

// If user provided templates, get the JSON
// and parse into templates
func GetOCLIAtrributesTemplateHelper(attr, data map[string]interface{}, ent int) {
	//Inner func declaration used for importing
	//data from templates
	attrSerialiser := func(someVal interface{}, idx string, ent int) string {
		if x, ok := someVal.(int); ok {
			if ent == DEVICE || ent == ROOM || ent == BLDG {
				return strconv.Itoa(x)
			}
			return strconv.Itoa(x / 10)
		} else if x, ok := someVal.(float64); ok {
			if ent == DEVICE || ent == ROOM || ent == BLDG {
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
			} else if ent == BLDG {
				tInt = BLDGTMPL
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
												res = int((val / 1000) / RACKUNIT)
											} else if val, ok := sizeInf[2].(int); ok {
												res = int((float64(val) / 1000) / RACKUNIT)
											} else {
												//Resort to default value
												msg := "Warning, invalid value provided for" +
													" sizeU. Defaulting to 5"
												println(msg)
												res = int((5 / 1000) / RACKUNIT)
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

						CopyAttr(attr, tmpl, "axisOrientation")

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

						CopyAttr(attr, tmpl, "pillars")
						if _, ok := attr["pillars"]; ok {
							tmp, _ = json.Marshal(attr["pillars"])
							attr["pillars"] = string(tmp)
						}

						CopyAttr(attr, tmpl, "floorUnit")
						if _, ok := attr["floorUnit"]; ok {
							if floorUnit, ok := attr["floorUnit"].(string); ok {
								attr["floorUnit"] = floorUnit
							}
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

						CopyAttr(attr, tmpl, "vertices")
						if _, ok := attr["vertices"]; ok {
							tmp, _ = json.Marshal(attr["vertices"])
							attr["vertices"] = string(tmp)
						}

						CopyAttr(attr, tmpl, "colors")
						if _, ok := attr["colors"]; ok {
							tmp, _ = json.Marshal(attr["colors"])
							attr["colors"] = string(tmp)
						}

						CopyAttr(attr, tmpl, "tileAngle")
						if _, ok := attr["tileAngle"]; ok {
							if tileAngle, ok := attr["tileAngle"].(int); ok {
								attr["tileAngle"] = strconv.Itoa(tileAngle)
							}

							if tileAngleF, ok := attr["tileAngle"].(float64); ok {
								tileAngleStr := strconv.FormatFloat(tileAngleF, 'f', -1, 64)
								attr["tileAngle"] = tileAngleStr
							}
						}

					} else if ent == BLDG {
						attr["sizeUnit"] = "m"
						attr["heightUnit"] = "m"

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
						if ent != BLDG {
							attr["fbxModel"] = ""
						}

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
					if State.DebugLvl > 1 {
						println("Warning, invalid size value in template.",
							"Default values will be assigned")
					}

				}
			} else {
				attr["template"] = ""
				if State.DebugLvl > 1 {
					println("Warning: template was not found.",
						"it will not be used")
				}

				l.GetWarningLogger().Println("Invalid data type or incorrect name used to invoke template")
			}

		} else {
			attr["template"] = ""
			if State.DebugLvl > 1 {
				println("Warning: template must be a string that",
					" refers to an existing imported template.",
					q, " will not be used")
			}

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

func UIDelay(time float64) {
	subdata := map[string]interface{}{"command": "delay", "data": time}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}
	InformUnity("HandleUI", -1, data)
}

func UIToggle(feature string, enable bool) {
	subdata := map[string]interface{}{"command": feature, "data": enable}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}
	InformUnity("HandleUI", -1, data)
}

func UIHighlight(objArg string) error {
	obj, _ := GetObject(objArg, true)
	if obj == nil {
		return fmt.Errorf("please provide a valid path")
	}
	subdata := map[string]interface{}{"command": "highlight", "data": obj["id"]}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}
	InformUnity("HandleUI", -1, data)
	return nil
}

func CameraMove(command string, position []float64, rotation []float64) {
	subdata := map[string]interface{}{"command": command}
	subdata["position"] = map[string]interface{}{"x": position[0], "y": position[1], "z": position[2]}
	subdata["rotation"] = map[string]interface{}{"x": rotation[0], "y": rotation[1]}
	data := map[string]interface{}{"type": "camera", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}
	InformUnity("HandleUI", -1, data)
}

func CameraWait(time float64) {
	subdata := map[string]interface{}{"command": "wait"}
	subdata["position"] = map[string]interface{}{"x": 0, "y": 0, "z": 0}
	subdata["rotation"] = map[string]interface{}{"x": 999, "y": time}
	data := map[string]interface{}{"type": "camera", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}
	InformUnity("HandleUI", -1, data)
}

func FocusUI(path string) {
	var id string
	if path != "" {
		obj, _ := GetObject(path, true)
		if obj == nil {
			if State.DebugLvl > 0 {
				msg := "Unable to focus on this object. Please" +
					" ensure that the object exists and" +
					" is not a directory and try again"
				println(msg)
				return
			}
		}
		category := EntityStrToInt(obj["category"].(string))
		if category == TENANT || category == SITE ||
			category == BLDG || category == ROOM {
			if State.DebugLvl > 0 {
				msg := "You cannot focus on this object. Note you cannot" +
					" focus on Tenants, Sites, Buildings and Rooms. " +
					"For more information please refer to the help doc  (man >)"
				println(msg)
				return
			}
		}

		id = obj["id"].(string)
	} else {
		id = ""
	}

	data := map[string]interface{}{"type": "focus", "data": id}
	InformUnity("FocusUI", -1, data)
	CD(path)
}

func LinkObject(source, destination string, destinationSlot interface{}) {

	var h []map[string]interface{}

	//Stray-device retrieval and validation
	sdev, _ := GetObject(source, true)
	if sdev == nil {
		if State.DebugLvl > 0 {
			println("Object doesn't exist")
		}

		return
	}
	if _, ok := sdev["category"]; !ok {
		l.GetWarningLogger().Println("Attempted to link non stray-device ")
		if State.DebugLvl > 0 {
			println("Error: Invalid object. Only stray-devices can be linked")
		}

		return
	}
	if cat, _ := sdev["category"]; cat != "stray-device" {
		l.GetWarningLogger().Println("Attempted to link non stray-device ")
		if State.DebugLvl > 0 {
			println("Error: Invalid object. Only stray-devices can be linked")
		}

		return
	}

	//Retrieve the stray-device hierarchy
	h = GetHierarchy(source, 50, true)

	//Parent retrieval and validation block
	parent, _ := GetObject(destination, true)
	if parent == nil {
		if State.DebugLvl > 0 {
			println("Destination is not valid")
		}

		return
	}
	if _, ok := parent["category"]; !ok {
		l.GetWarningLogger().Println("Attempted to link with invalid target")
		if State.DebugLvl > 0 {
			println("Error: Invalid destination object")
			println("Please use a rack or a device as a link target")

		}
		return
	}
	if cat, _ := parent["category"].(string); cat != "device" && cat != "rack" {
		l.GetWarningLogger().Println("Attempted to link with invalid target")
		if State.DebugLvl > 0 {
			println("Error: Invalid destination object")
			println("Please use a rack or a device as a link target")

		}
		return
	}

	//Need to make sure that origin and destination are
	//not the same!
	if parent["id"] == sdev["id"] && parent["name"] == sdev["name"] {
		l.GetWarningLogger().Println("Attempted to object to itself")
		if State.DebugLvl > 0 {
			println("Error you must provide a unique stray-device" +
				" and a unique destination for it")
		}

	}

	//Ensure that the stray device can be imported as device
	//First set the parentId of stray device to point to parent ID
	//Then dive, set the parentID (Which PID is not exactly important
	//we just need to point to a valid PID.)
	//and invoke API validation endpoint
	sdev["parentId"] = parent["id"]
	if destinationSlot != nil && destinationSlot != "" {
		if attrInf, ok := sdev["attributes"]; ok {
			//attr["slot"] = destinationSlot
			if attr, ok := attrInf.(map[string]interface{}); ok {
				attr["slot"] = destinationSlot
			} else {
				sdev["attributes"] = map[string]interface{}{"slot": destinationSlot}
			}
		} else {
			sdev["attributes"] = map[string]interface{}{"slot": destinationSlot}
		}
	}

	sdev, _ = PostObj(DEVICE, "device", sdev)
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
		if State.DebugLvl > 0 {
			println("In the target's hierarchy the following "+
				x["category"].(string)+" does not satisfy "+
				desiredObj+" validation requirements: ", x["name"].(string))
			println("Aborting link operation")
		}

		DeleteObj(destination + "/" + sdev["name"].(string))
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

				newObj, _ := PostObj(entInt, ent, x[i])

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
	DeleteObj(source)
}

// This function validates a hierarchy to be imported into another category
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

			newObj, _ := PostObj(ent, entStr, x[i])

			if childrenInfArr != nil {
				newpid := newObj["id"]
				children := infArrToMapStrinfArr(childrenInfArr.([]interface{}))
				fn(children, newpid, entity, ent)
			}
		}
	}
}

// paths should only have a length of 1 or 2
func UnlinkObject(source, destination string) {
	//source ===> device to unlink
	//destination ===> new location in stray-dev (can be optionally empty)
	dev := map[string]interface{}{}
	h := []map[string]interface{}{}

	//first we need to check that the path corresponds to a device
	//we also need to ignore groups
	//arbitrarily set depth to 50 since it doesn't make sense
	//for a device to have a deeper hierarchy
	dev, _ = GetObject(source, true)
	if dev == nil {
		if State.DebugLvl > 0 {
			println("Error: This object does not exist ")
		}
		l.GetErrorLogger().Println("User attempted to unlink non-existing object")

		return
	}

	//Exit if device not found
	if _, ok := dev["category"]; !ok {
		if State.DebugLvl > 0 {
			println("Error: This object is not a device. You can only unlink devices ")
		}
		l.GetErrorLogger().Println("User attempted to unlink non-device object")

		return
	}

	if catInf, _ := dev["category"].(string); catInf != "device" {
		if State.DebugLvl > 0 {
			println("Error: This object is not a device. You can only unlink devices ")
		}
		l.GetErrorLogger().Println("User attempted to unlink non-device object")

		return
	}

	h = GetHierarchy(source, 50, true)

	//Dive POST
	var parent map[string]interface{}

	if destination != "" {
		parent, _ = GetObject(destination, true)
		if parent != nil {
			dev["parentId"] = parent["id"]
		}
	}

	if parent == nil {
		DeleteAttr(dev, "parentId")
	}

	newDev, _ := PostObj(STRAY_DEV, "stray-device", dev)
	if newDev == nil {
		l.GetWarningLogger().Println("Unable to unlink target: ", source)
		if State.DebugLvl > 0 {
			println("Error: Unable to unlink target: ", source)
		}

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
	DeleteObj(source)
}

// TODO
// Move object counting to API side
func objectCounter(parent *map[string]interface{}) int {
	count := 0
	if (*parent) != nil {
		count += 1
		if _, ok := (*parent)["children"]; ok {
			if arr, ok := (*parent)["children"].([]interface{}); ok {

				for _, childInf := range arr {
					if child, ok := childInf.(map[string]interface{}); ok {
						count += objectCounter(&(child))
					}
				}
			}

			if arr, ok := (*parent)["children"].([]map[string]interface{}); ok {
				for _, child := range arr {
					count += objectCounter(&(child))

				}

			}
		}
	}

	return count
}

// Unity UI will draw already existing objects
// by retrieving the hierarchy. 'force' bool is useful
// for scripting where the user can 'force' input if
// the num objects to draw surpasses threshold
func Draw(x string, depth int, force bool) error {
	obj, _ := GetObject(x, true)
	if obj == nil {
		return fmt.Errorf("object not found")
	}
	if depth < 0 {
		return fmt.Errorf("draw command cannot accept negative value")
	} else {
		if depth != 0 {
			children := GetHierarchy(x, depth, true)
			if children != nil {
				obj["children"] = children
			}
		}

		count := objectCounter(&obj)
		if State.UnityClientAvail {
			okToGo := true
			if count > State.DrawThreshold && !force {
				msg := "You are about to send " + strconv.Itoa(count) +
					" objects to the Unity 3D client. " +
					"Do you want to continue ? (y/n)\n"
				(*State.Terminal).Write([]byte(msg))
				(*State.Terminal).SetPrompt(">")
				ans, _ := (*State.Terminal).Readline()
				if ans != "y" && ans != "Y" {
					okToGo = false
				}
			} else if force {
				okToGo = true
			} else if !force && count > State.DrawThreshold {
				okToGo = false
			}
			if okToGo {
				data := map[string]interface{}{"type": "create", "data": obj}
				//0 to include the JSON filtration
				unityErr := InformUnity("Draw", 0, data)
				if unityErr != nil {
					return unityErr
				}
			}
		}

	}
	return nil
}

// Unity UI will draw already existing objects
// by retrieving the hierarchy
func Undraw(x string) error {
	var id string
	if x == "" {
		id = ""
	} else {
		obj, _ := GetObject(x, true)
		if obj == nil {
			return fmt.Errorf("object not found")
		}
		id = obj["id"].(string)
	}

	data := map[string]interface{}{"type": "delete", "data": id}
	unityErr := InformUnity("Undraw", 0, data)
	if unityErr != nil {
		return unityErr
	}

	return nil
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
				if State.DebugLvl > 0 {
					println(err)
				}

			}

			return false
		}

		//Ensure that we can get the category
		//from object
		if catInf, ok := obj["category"]; ok {
			if cat, ok := catInf.(string); !ok {
				l.GetErrorLogger().Println("Object does not have category")
				if silence != true {
					if State.DebugLvl > 0 {
						println("Error: Object does not have category")
					}

				}

				return false
			} else if EntityStrToInt(cat) == -1 {
				l.GetErrorLogger().Println("Object has invalid category")
				if silence != true {
					if State.DebugLvl > 0 {
						println("Error: Object has invalid category")
					}

				}

				return false
			}
		} else {
			l.GetErrorLogger().Println("Object does not have category")
			if silence != true {
				if State.DebugLvl > 0 {
					println("Error: Object does not have category")
				}

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
					if State.DebugLvl > 0 {
						println("Error: Object does not have category")
					}

				}

				return false
			} else if EntityStrToInt(cat) == -1 {
				l.GetErrorLogger().Println("Object has invalid category")
				if silence != true {
					if State.DebugLvl > 0 {
						println("Error: Object has invalid category")
					}

				}

				return false
			}
		} else {
			l.GetErrorLogger().Println("Object does not have category")
			if silence != true {
				if State.DebugLvl > 0 {
					println("Error: Object does not have category")
				}

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

func UpdateSelection(data map[string]interface{}) error {
	if State.ClipBoard != nil {
		for _, k := range *State.ClipBoard {
			_, err := UpdateObj(k, "", "", data, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func LoadFile(path string) {
	//By setting the 'ScriptCalled' variable
	//the REPL will recognise this and invoke the
	//LoadFile() Function in ocli.go

	//Alternative to this would be to pass the LoadFile()
	//function as an argument here
	State.ScriptCalled = true
	State.ScriptPath = path
}

func LoadTemplate(data map[string]interface{}, filePath string) {
	var URL string

	if cat, _ := data["category"]; cat == "room" {
		//Room template
		URL = State.APIURL + "/api/room-templates"
	} else if cat == "bldg" || cat == "building" {
		//Bldg template
		URL = State.APIURL + "/api/bldg-templates"
	} else if cat == "rack" || cat == "device" {
		// Obj template
		URL = State.APIURL + "/api/obj-templates"
	} else {
		println("This template does not have a valid category. Please add a category attribute with a value of building or room or rack or device")
		return
	}

	r, e := models.Send("POST", URL, GetKey(), data)
	if e != nil {
		l.GetErrorLogger().Println(e.Error())
		if State.DebugLvl > NONE {
			println("Error: ", e.Error())
		}

	}

	//Crashes here if API timeout
	if r == nil {
		if State.DebugLvl > NONE {
			println("Unable to recieve response from API")
		}
		return
	}

	if r.StatusCode == http.StatusCreated {
		println("Template Loaded")
	} else {
		l.GetWarningLogger().Println("Couldn't load template, Status Code :", r.StatusCode, " filePath :", filePath)
		parsedResp := ParseResponse(r, e, "sending template")
		if State.DebugLvl > 0 {
			println("Error template wasn't loaded")
			if mInf, ok := parsedResp["message"]; ok {
				if msg, ok := mInf.(string); ok {
					println(APIErrorPrefix + msg)
				}
			}

		}

	}
}

func SetClipBoard(x []string) ([]string, error) {
	State.ClipBoard = &x
	var data map[string]interface{}

	if len(x) == 0 { //This means deselect
		data = map[string]interface{}{"type": "select", "data": "[]"}
		err := InformUnity("SetClipBoard", -1, data)
		if err != nil {
			return nil, fmt.Errorf("cannot reset clipboard : %s", err.Error())
		}
	} else {
		//Verify paths
		arr := make([]string, len(x))
		for idx, val := range x {
			obj, _ := GetObject(val, true)
			if obj != nil {
				arr[idx] = obj["id"].(string)
			} else {
				return nil, fmt.Errorf("cannot set clipboard")
			}
		}
		serialArr := "[\"" + strings.Join(arr, "\",\"") + "\"]"
		data = map[string]interface{}{"type": "select", "data": serialArr}
		err := InformUnity("SetClipBoard", -1, data)
		if err != nil {
			return nil, fmt.Errorf("cannot set clipboard : %s", err.Error())
		}
	}

	return *State.ClipBoard, nil
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
	case "Filter", "Unity":
		if _, ok := val.(bool); !ok {
			msg := "Can only assign bool values for " + arg + " Env Var"
			l.GetWarningLogger().Println(msg)
			if State.DebugLvl > 0 {
				println(msg)
			}

		} else {
			if arg == "Unity" {
				State.UnityClientAvail = val.(bool)
			} else if arg == "Filter" {
				State.FilterDisplay = val.(bool)
			}

			println(arg + " Display Environment variable set")
		}

	case "Analyser":
		if _, ok := val.(bool); !ok {
			msg := "Can only assign bool values for SAnalyser Env Var"
			l.GetWarningLogger().Println(msg)
			if State.DebugLvl > 0 {
				println(msg)
			}

		} else {
			State.Analyser = val.(bool)
			println("Static Analyser Environment variable set")
		}

	default:
		println(arg + " is not an environment variable")
	}
}

// Utility functions
func determineStrKey(x map[string]interface{}, possible []string) string {
	for idx := range possible {
		if _, ok := x[possible[idx]]; ok {
			return possible[idx]
		}
	}
	return "" //The code should not reach this point!
}

// Function called by update node for interact commands (ie label, labelFont)
func InteractObject(path string, keyword string, val interface{}, fromAttr bool) error {
	//First retrieve the object
	obj, e := GetObject(path, true)
	if e == "" {
		msg := "Object not found please check the path" +
			" you provided and try again"
		return fmt.Errorf(msg)
	}

	//Verify labelFont has valid values
	if fromAttr == true {
		//Check if the val refers to an attribute field in the object
		//this means to retrieve value from object
		if value, ok := val.(string); ok {

			innerMap := obj["attributes"].(map[string]interface{})

			if _, ok := obj[value]; ok {
				if value == "description" {

					desc := obj["description"].([]interface{})
					val = ""
					//Combine entire the description array into a string
					for i := 0; i < len(desc); i++ {
						if i == 0 {
							val = desc[i].(string)
						} else {
							val = val.(string) + "-" + desc[i].(string)
						}

					}
				} else {
					val = obj[value]
				}

			} else if _, ok := innerMap[value]; ok {
				val = innerMap[value]
			} else {
				if strings.Contains(value, "description") == true {
					if desc, ok := obj["description"].([]interface{}); ok {
						if len(value) > 11 { //descriptionX format
							//split the number and description
							numStr := strings.Split(value, "description")[1]
							num, e := strconv.Atoi(numStr)
							if e != nil {
								return e
							}

							if num < 0 {
								return fmt.Errorf("Description index must be positive")
							}

							if num >= len(desc) {
								msg := "Description index is out of" +
									" range. The length for this object is: " +
									strconv.Itoa(len(desc))
								return fmt.Errorf(msg)
							}
							val = desc[num]

						} else {
							val = innerMap[value]
						}
					} //Otherwise the description is a string

				} else {
					msg := "The specified attribute does not exist" +
						" in the object. \nPlease view the object" +
						" (ie. $> get) and try again"
					return fmt.Errorf(msg)
				}

			}

		} else {
			return fmt.Errorf("The label value must be a string")
		}
	}

	data := map[string]interface{}{"id": obj["id"],
		"param": keyword, "value": val}
	ans := map[string]interface{}{"type": "interact", "data": data}

	//-1 since its not neccessary to check for filtering
	return InformUnity("Interact", -1, ans)
}

// Messages Unity Client
func InformUnity(caller string, entity int, data map[string]interface{}) error {
	//If unity is available message it
	//otherwise do nothing
	if State.UnityClientAvail {
		if entity > -1 && entity < SENSOR+1 {
			data = GenerateFilteredJson(data)
		}
		if State.DebugLvl > INFO {
			println("DEBUG VIEW THE JSON")
			Disp(data)
		}

		e := models.ContactUnity(data, State.DebugLvl)
		if e != nil {
			l.GetWarningLogger().Println("Unable to contact Unity Client @" + caller)
			if State.DebugLvl > 1 {
				fmt.Println("Error while updating Unity: ", e.Error())
			}
			return fmt.Errorf("error while contacting unity : %s", e.Error())
		}
	}
	return nil
}

// x is path
func LSOBJECTRecursive(x string, entity int) []interface{} {
	var obj map[string]interface{}
	var Path string

	if entity == TENANT { //Edge case
		if x == "/Physical" {
			r, e := models.Send("GET",
				State.APIURL+"/api/tenants", GetKey(), nil)
			obj = ParseResponse(r, e, "Get Tenants")
			return LoadArrFromResp(obj, "objects")
		} else {
			//Return nothing
			return nil
		}
	} else {
		obj, Path = GetObject(x, true)
		if obj == nil {
			if State.DebugLvl > 0 {
				println("Error finding Object from given path!")
			}

			l.GetWarningLogger().Println("Object to Get not found")
			return nil
		}
	}

	entityDir, _ := path.Split(Path)
	entities := path.Base(entityDir)
	objEnt := entities[:len(entities)-1]
	obi := EntityStrToInt(objEnt)
	if obi == -1 { //Something went wrong
		if State.DebugLvl > 0 {
			println("Error finding Object from given path!")
		}

		l.GetWarningLogger().Println("Object to Get not found")
		return nil
	}

	//YouareAt -> obi
	//want 	   -> entity

	if (entity >= AC && entity <= CORIDOR) && obi > ROOM {
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
	return lsobjHelperRecursive(State.APIURL, idToSend, obi, entity)
	//return nil
}

// NOTE: LSDEV is recursive while LSSENSOR is not
// Code could be more tidy
func lsobjHelperRecursive(api, objID string, curr, entity int) []interface{} {
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
		//res := infArrToMapStrinfArr(tmpObjs)
		return tmpObjs

	} else if entity-curr >= 2 {

		//println("DEBUG-should be here")
		ext = EntityToString(curr) + "s/" + objID + "/" + EntityToString(curr+2) + "s"
		URL = State.APIURL + "/api/" + ext
		//println("DEBUG-URL:", URL)

		//EDGE CASE, if user is at a BLDG and requests object of room
		if (curr == BLDG || curr == ROOM) && (entity >= AC && entity <= CORIDOR) {
			ext = EntityToString(curr) + "s/" + objID + "/" + EntityToString(entity) + "s"
			r, e := models.Send("GET", State.APIURL+"/api/"+ext, GetKey(), nil)
			tmp := ParseResponse(r, e, "getting objects")
			if tmp == nil {
				return nil
			}

			return GetRawObjects(tmp)

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
			x := []interface{}{}

			if entity >= AC && entity <= CORIDOR {

				for q := range objs {
					id := objs[q].(map[string]interface{})["id"].(string)
					ext2 := "/api/" + EntityToString(curr+2) + "s/" + id + "/" + EntityToString(entity) + "s"

					tmp, e := models.Send("GET", State.APIURL+ext2, GetKey(), nil)
					tmp2 := ParseResponse(tmp, e, "get objects")
					if tmp2 != nil {
						x = GetRawObjects(tmp2)
					}
				}
			} else {
				if entity == DEVICE && curr == ROOM {
					x = append(x, objs...)
				}
				for i := range objs {
					rest := lsobjHelperRecursive(api, objs[i].(map[string]interface{})["id"].(string), curr+2, entity)
					if rest != nil && len(rest) > 0 {
						x = append(x, rest...)
					}

				}
			}

			if State.DebugLvl > 3 {
				println(len(x))
			}

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
			ans := []interface{}{}
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

				ans = objs
				if curr == RACK && entity == DEVICE {
					for idx := range ans {
						ext2 := "/api/" + EntityToString(entity) +
							"s/" +
							ans[idx].(map[string]interface{})["id"].(string) +
							"/" + EntityToString(entity) + "s"

						subURL := State.APIURL + ext2
						r1, e1 := models.Send("GET", subURL, GetKey(), nil)
						tmp1 := ParseResponse(r1, e1, "getting objects")

						tmp2 := LoadArrFromResp(tmp1, "objects")
						if tmp2 != nil {
							//Swap ans and objs to keep order
							ans = append(ans, tmp2...)
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
			return GetRawObjects(x)

		}
		return []interface{}{x["data"]}
	}
	return nil
}

// Auxillary function that preprocesses
// strings to be used for Path Resolver funcs
func PreProPath(Path string) []string {
	var pathSplit []string

	switch Path {
	case "":
		pathSplit = strings.Split(State.CurrPath, "/")
		pathSplit = pathSplit[2:]
	default:
		if Path[0] != '/' && len(State.CurrPath) > 1 {
			pathSplit = strings.Split(State.CurrPath+"/"+Path, "/")
			pathSplit = pathSplit[2:]
		} else {
			pathSplit = strings.Split(Path, "/")
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

// Take 'user' abstraction path and
// convert to online URL for API
func OnlinePathResolve(path []string) []string {
	//We have to make an array since there can be
	//multiple possible paths for paths past room
	paths := []string{}
	basePath := State.APIURL + "/api"
	roomChildren := []string{"/acs", "/panels", "/cabinets",
		"/groups", "/corridors", "/sensors"}

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

	if path[0] == "BldgTemplates" {
		basePath += "/bldg-templates"
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

// Auxillary function for FetchNodesAtLevel
// Take 'user' abstraction path and
// convert to online URL for API
func OnlineLevelResolver(path []string) []string {
	//We have to make an array since there can be
	//multiple possible paths for paths past room
	paths := []string{}
	basePath := State.APIURL + "/api"
	roomChildren := []string{"/acs", "/panels", "/cabinets",
		"/groups", "/corridors", "/sensors"}

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

	if path[0] == "BldgTemplates" {
		basePath += "/bldg-templates"
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

// Helper function for GetOCLIAttr which retrieves
// template from server if available, this func mainly helps
// to keep code organised
func fetchTemplate(name string, objType int) map[string]interface{} {
	var URL string
	if objType == ROOMTMPL {
		URL = State.APIURL + "/api/room_templates/" + name
	} else if objType == BLDGTMPL {
		URL = State.APIURL + "/api/bldg_templates/" + name
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
