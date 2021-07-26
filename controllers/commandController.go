package controllers

import (
	"cli/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func Execute() {
	println("Congratulations, nobody cares")
	return
}

func PWD() {
	println(State.CurrPath)
}

func Disp(x map[string]interface{}) {
	/*for i, k := range x {
		println("We got: ", i, " and ", k)
	}*/

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func PostObj(entity, path string, data map[string]interface{}) {
	var respMap map[string]interface{}
	resp, e := models.Send("POST",
		"https://ogree.chibois.net/api/user/"+entity+"s", data)

	if e != nil {
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		os.Exit(-1)
	}

	json.Unmarshal(bodyBytes, &respMap)
	println(string(respMap["message"].(string)) /*bodyBytes*/)
	if resp.StatusCode == http.StatusCreated {
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
		switch entity {
		case "tenant":
			node.Entity = TENANT
			State.TreeHierarchy.Nodes.PushBack(node)
		case "site":
			node.Entity = SITE
			UpdateTree(&State.TreeHierarchy, node)

		case "building":
			node.Entity = BLDG
			val := UpdateTree(&State.TreeHierarchy, node)
			println("BLDG ADDED?", val)

		case "room":
			node.Entity = ROOM
			UpdateTree(&State.TreeHierarchy, node)

		case "rack":
			node.Entity = RACK
			UpdateTree(&State.TreeHierarchy, node)

		case "device":
			node.Entity = DEVICE
			UpdateTree(&State.TreeHierarchy, node)

		case "subdevice":
			node.Entity = SUBDEV
			UpdateTree(&State.TreeHierarchy, node)

		case "subdevice1":
			node.Entity = SUBDEV1
			UpdateTree(&State.TreeHierarchy, node)

		}

	}
	return
}

func DeleteObj(entity string, data map[string]interface{}) {
	resp, e := models.Send("DELETE",
		"https://ogree.chibois.net/api/user/"+entity+"s/"+
			string(data["id"].(string)), nil)
	if e != nil {
		println("There was an error!")
	}
	if resp.StatusCode != http.StatusNoContent {
		println("Unsuccessful!")
	} else {
		println("Success")
	}

	ID, _ := strconv.Atoi(data["id"].(string))
	var ent int
	switch entity {
	case "tenant":
		ent = TENANT
	case "site":
		ent = SITE
	case "building":
		ent = BLDG
	case "room":
		ent = ROOM
	case "rack":
		ent = RACK
	case "device":
		ent = DEVICE
	case "subdevice":
		ent = SUBDEV
	case "subdevice1":
		ent = SUBDEV1
	}

	DeleteNodeInTree(&State.TreeHierarchy, ID, ent)

	return
}

//Search for objects
func SearchObjects(entity string, data map[string]interface{}) {
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

	println("Here is URL: ", URL)

	resp, e := models.Send("GET", URL, nil)
	println("Response Code: ", resp.Status)
	if e != nil {
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		os.Exit(-1)
	}
	println(string(bodyBytes))
}

func GetObject(path string) {
	URL := "https://ogree.chibois.net/api/user/"
	nd := &Node{}
	var data map[string]interface{}

	switch path {
	case "":
		nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(State.CurrPath))
	default:
		nd = FindNodeInTree(&State.TreeHierarchy, StrToStack(path))
	}

	if nd == nil {
		println("Error finding Object from given path!")
		return
	}

	URL += EntityToString(nd.Entity) + "s/" + strconv.Itoa(nd.ID)
	resp, e := models.Send("GET", URL, nil)
	if e != nil {
		println("Error while obtaining Object details!")
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error while reading response!")
		return
	}
	json.Unmarshal(bodyBytes, &data)
	if resp.StatusCode == http.StatusOK {
		if data["data"] != nil {
			obj := data["data"].(map[string]interface{})
			for i := range obj {
				if i == "attributes" {
					for q := range obj[i].(map[string]interface{}) {
						println(q, ":", string(obj[i].(map[string]interface{})[q].(string)))
					}
				} else {
					println(i, ":", obj[i])
				}

			}
		}
	}

}

func UpdateObj(entity string, data map[string]interface{}) {
	println("OK. Attempting to update...")
	if data["id"] != nil {
		URL := "https://ogree.chibois.net/api/user/" + entity + "s/" +
			string(data["id"].(string))

		resp, e := models.Send("PUT", URL, data)
		println("Response Code: ", resp.Status)
		if e != nil {
			println("There was an error!")
		}
		defer resp.Body.Close()
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			println("Error: " + err.Error() + " Now Exiting")
			os.Exit(-1)
		}
		println(string(bodyBytes))
	} else {
		println("Error! Please enter ID of Object to be updated")
	}

}

func LS(x string) {
	if x == "" || x == "." {
		DispAtLevel(&State.TreeHierarchy, *StrToStack(State.CurrPath))
	} else if string(x[0]) == "/" {
		DispAtLevel(&State.TreeHierarchy, *StrToStack(x))
	} else {
		DispAtLevel(&State.TreeHierarchy, *StrToStack(State.CurrPath + "/" + x))
	}
}

func CD(x string) {
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
			exist, pth = CheckPath(&State.TreeHierarchy, StrToStack(State.CurrPath+"/"+x), New())
		} else {
			exist, pth = CheckPath(&State.TreeHierarchy, StrToStack(x), New())
		}
		if exist == true {
			println("THE PATH: ", pth)
			State.PrevPath = State.CurrPath
			State.CurrPath = pth
		} else {
			println("Path does not exist")
		}
	} else {
		if len(State.CurrPath) != 1 {
			if exist, _ := CheckPath(&State.TreeHierarchy,
				StrToStack(State.CurrPath+"/"+x), New()); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += "/" + x
			} else {
				println("OGREE: ", x, " : No such object")
			}

		} else {
			if exist, _ := CheckPath(&State.TreeHierarchy,
				StrToStack(State.CurrPath+x), New()); exist == true {
				State.PrevPath = State.CurrPath
				State.CurrPath += x
			} else {
				println("OGREE: ", x, " : No such object")
			}

		}

	}

}

func Help() {
	fmt.Printf(`A Shell interface to the API and your datacenter visualisation solution`)
}
