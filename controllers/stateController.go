package controllers

import (
	l "cli/logger"
	"cli/models"
	"cli/readline"
	"container/list"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	TENANT = iota
	SITE
	BLDG
	ROOM
	RACK
	DEVICE
	AC
	PWRPNL
	CABINET
	CORIDOR
	SENSOR
	ROOMTMPL
	OBJTMPL
	GROUP
	STRAY_DEV
	STRAYSENSOR
)

var BuildTime string
var BuildHash string
var BuildTree string
var GitCommitDate string
var State ShellState

type ShellState struct {
	CurrPath         string
	PrevPath         string
	ClipBoard        *[]string
	TreeHierarchy    *Node
	EnvFilePath      string //Holds file path of '.env'
	HistoryFilePath  string //Holds file path of '.history'
	ScriptCalled     bool
	ScriptPath       string
	UnityClientURL   string
	APIURL           string
	APIKEY           string
	UnityClientAvail bool  //For deciding to message unity or not
	FilterDisplay    bool  //Set whether or not to send attributes to unity
	Analyser         bool  //Use static analysis before executing scripts
	ObjsForUnity     []int //Deciding what objects should be sent to unity
	DrawThreshold    int   //Number of objects to be sent at a time to unity
	DrawableObjs     []int //Indicate which objs drawable in unity
	DrawableJsons    map[string]map[string]interface{}
	DebugLvl         int
	LineNumber       int //Used exectuting scripts
	Terminal         **readline.Instance
	Timeout          time.Duration
}

type Node struct {
	ID     string
	PID    string
	Entity int
	Name   string
	Path   string
	Nodes  list.List
}

func IsInObjForUnity(x string) bool {
	entInt := EntityStrToInt(x)
	if entInt != -1 {

		for idx := range State.ObjsForUnity {
			if State.ObjsForUnity[idx] == entInt {
				return true
			}
		}
	}
	return false
}

func IsDrawableEntity(x string) bool {
	entInt := EntityStrToInt(x)

	for idx := range State.DrawableObjs {
		if State.DrawableObjs[idx] == entInt {
			return true
		}
	}
	return false
}

func GetLineNumber() int {
	return State.LineNumber
}

func GetScriptPath() string {
	return State.ScriptPath
}

func GetKey() string {
	return State.APIKEY
}

func SearchAndInsert(root **Node, node *Node, path string) {
	if root != nil {
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			if node.PID == (i.Value).(*Node).ID {
				//println("NODE ", node.Name, "WITH PID: ", node.PID)
				//println("Matched with PARENT ")
				//println()
				node.Path = path + "/" + (i.Value).(*Node).Name + "/" + node.Name
				(i.Value).(*Node).Nodes.PushBack(node)
				return
			}
			x := (i.Value).(*Node)
			SearchAndInsert(&x, node, path+"/"+x.Name)
		}
	}
}

// Function for debugging purposes
func View(root *Node, dt int) {
	if dt != 7 || root != nil {
		arr := (*root).Nodes
		for i := arr.Front(); i != nil; i = i.Next() {

			println("Now Printing children of: ",
				(*Node)((i.Value).(*Node)).Name)
			//println()
			View(((i.Value).(*Node)), dt+1)
		}
	}
}

func getNextInPath(name string, root *Node) *Node {
	for i := root.Nodes.Front(); i != nil; i = i.Next() {
		if (i.Value.(*Node)).Name == name {
			return (i.Value.(*Node))
		}
	}
	return nil
}

// Replaces DispAtLevel since we are no longer
// storing objects in a tree and returns string arr
func FetchNodesAtLevel(Path string) []string {
	names := []string{}
	urls := []string{}

	paths := strings.Split(path.Clean(Path), "/")

	/*if len(paths) == 1 || len(paths) == 0 {
		println("DEBUG only / encountered")
	}*/

	if len(paths) == 2 && paths[1] == "Physical" {
		urls = []string{State.APIURL + "/api/tenants"}
		names = NodesAtLevel(&State.TreeHierarchy, *StrToStack(Path))
	} else {
		if len(paths) == 3 && paths[2] == "Stray" {
			names = NodesAtLevel(&State.TreeHierarchy, *StrToStack(Path))
		}

		if len(paths) < 3 { // /Physical or / or /Logical
			//println("Should be here")
			//println("LEN:", len(paths))
			//println("YO DEBUG", path)
			return NodesAtLevel(&State.TreeHierarchy, *StrToStack(Path))
		}

		// 2: since first idx is useless
		// and 2nd is just /Physical or /Logical etc
		urls = OnlineLevelResolver(paths[2:])
	}

	for i := range urls {
		//println("DEBUG URL to send:", urls[i])
		r, e := models.Send("GET", urls[i], GetKey(), nil)
		if e != nil {
			println(e.Error())
			return nil
		}

		if r.StatusCode == http.StatusOK { //Retrieved nodes
			parsedResp := ParseResponse(r, e, "get request")
			if parsedResp == nil {
				return nil
			}

			if parsedResp["data"] != nil {

				if objs, ok := parsedResp["data"].(map[string]interface{})["objects"]; ok {
					data := objs.([]interface{})

					for i := range data {
						//If we have templates, check for slug
						if _, ok := data[i].(map[string]interface{})["slug"]; ok {
							names = append(names, data[i].(map[string]interface{})["slug"].(string))
						} else {
							names = append(names, data[i].(map[string]interface{})["name"].(string))
						}

						//println(data[i].(map[string]interface{})["name"].(string))
					}

				}

			}
		}
	}
	return names
}

// Same as FetchNodesAtLevel but returns the JSONs
// in map[string]inf{} format
func FetchJsonNodesAtLevel(Path string) []map[string]interface{} {
	objects := []map[string]interface{}{}
	urls := []string{}

	paths := strings.Split(path.Clean(Path), "/")

	if len(paths) == 2 && paths[1] == "Physical" {
		x := NodesAtLevel(&State.TreeHierarchy, *StrToStack(Path))
		objects = append(objects, strArrToMapStrInfArr(x)...)
		urls = []string{State.APIURL + "/api/tenants"}

	} else {
		if len(paths) == 3 && paths[2] == "Stray" || len(paths) < 3 {
			x := NodesAtLevel(&State.TreeHierarchy, *StrToStack(Path))
			return strArrToMapStrInfArr(x)
		}

		if len(paths) == 3 && paths[2] == "Domain" {
			//println("DEBUG this section for the new nodes")
			//println("DEBUG path2: ", paths[3])
			urls = []string{State.APIURL + "/api/domains"}

		}

		if len(paths) == 4 && paths[2] == "Stray" {
			//println("DEBUG this section for the new nodes")
			//println("DEBUG path2: ", paths[3])
			if paths[3] == "Device" {
				urls = []string{State.APIURL + "/api/stray-devices"}
			}
			if paths[3] == "Sensor" {
				urls = []string{State.APIURL + "/api/stray-sensors"}
			}

		} else {
			//if len(paths) < 3 { // /Physical or / or /Logical
			//println("DEBUG Should be here")
			//println("DEBUG LEN:", len(paths))
			//println("DEBUG: ", path)
			//	x := NodesAtLevel(&State.TreeHierarchy, *StrToStack(path))
			//	return strArrToMapStrInfArr(x)
			//}

			// 2: since first idx is useless
			// and 2nd is just /Physical or /Logical etc
			urls = OnlineLevelResolver(paths[2:])
		}

	}

	for i := range urls {
		//println("URL to send:", urls[i])
		r, e := models.Send("GET", urls[i], GetKey(), nil)
		if e != nil {
			println(e.Error())
			return nil
		}

		if r.StatusCode == http.StatusOK { //Retrieved nodes
			parsedResp := ParseResponse(r, e, "get request")
			if parsedResp == nil {
				return nil
			}

			if parsedResp["data"] != nil {

				if objs, ok := parsedResp["data"].(map[string]interface{})["objects"]; ok {
					data := objs.([]interface{})

					for i := range data {
						//If we have templates, check for slug
						if object, ok := data[i].(map[string]interface{}); ok {
							objects = append(objects, object)
						}
					}

				}

			}
		}
	}
	return objects
}

// If the path refers to local tree the
// func will still verify it with local tree
func CheckPathOnline(Path string) (bool, string) {

	pathSplit := strings.Split(path.Clean(Path), "/")

	//Check if path refers to object in local State Tree
	//There is an edge case for Stray object paths ending
	//with Device or Sensor
	pathLen := len(pathSplit)
	if pathLen <= 3 || pathSplit[pathLen-1] == "Device" || pathSplit[pathLen-1] == "Sensor" {
		nd := FindNodeInTree(&State.TreeHierarchy, StrToStack(Path), true)
		if nd != nil {
			return true, Path
		}
	}

	paths := OnlinePathResolve(pathSplit[2:])

	for i := range paths {
		r, e := models.Send("GET", paths[i], GetKey(), nil)
		if e != nil {
			return false, ""
		}
		if r.StatusCode == http.StatusOK {
			return true, paths[i]
		}
	}
	return false, ""
}

// Return extra bool so that the Parent can delete
// leaf and keep track without stack
func DeleteNodeInTree(root **Node, ID string, ent int) (bool, bool) {
	if root == nil {
		return false, false
	}

	//Delete only when the PID matches Parent's ID
	if (*root).ID == ID && ent == (*root).Entity {
		return true, false
	}

	for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
		nxt := (i.Value).(*Node)
		first, deleted := DeleteNodeInTree(&nxt, ID, ent)
		if first == true && deleted == false {
			(*root).Nodes.Remove(i)
			return true, true
		}
	}
	return false, false
}

func FindNodeInTree(root **Node, path *Stack, silenced bool) **Node {
	if root == nil {
		return nil
	}

	if path.Len() > 0 {
		name := path.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			if !silenced {
				if State.DebugLvl > 0 {
					println("Name doesn't exist! ", string(name.(string)))
				}

			}

			l.GetWarningLogger().Println("Name doesn't exist! ", string(name.(string)))
			return nil
		}
		path.Pop()
		return FindNodeInTree(&node, path, silenced)
	} else {
		return root
	}
}

// Same thing as FindNodeInTree but instead we return the root if the
// desired node was not found
// NOTE: This func still returns nil and so a small 'hack' was
// implemented in the caller to avoid this
func FindNearestNodeInTree(root **Node, path *Stack, silenced bool) **Node {
	if root == nil {
		return nil
	}

	if path.Len() > 0 {
		name := path.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			return root
		}
		path.Pop()

		if node := FindNodeInTree(&node, path, silenced); node != nil {
			return node
		}
		return root
	} else {
		return root
	}
}

func EntityToString(entity int) string {
	switch entity {
	case TENANT:
		return "tenant"
	case SITE:
		return "site"
	case BLDG:
		return "building"
	case ROOM:
		return "room"
	case RACK:
		return "rack"
	case DEVICE:
		return "device"
	case AC:
		return "ac"
	case PWRPNL:
		return "panel"
	case STRAY_DEV:
		return "stray_device"
	case ROOMTMPL:
		return "room_template"
	case OBJTMPL:
		return "obj_template"
	case CABINET:
		return "cabinet"
	case GROUP:
		return "group"
	case CORIDOR:
		return "corridor"
	case SENSOR:
		return "sensor"
	default:
		return "INVALID"
	}
}

func EntityStrToInt(entity string) int {
	switch entity {
	case "tenant", "tn":
		return TENANT
	case "site", "si":
		return SITE
	case "building", "bldg", "bd":
		return BLDG
	case "room", "ro":
		return ROOM
	case "rack", "rk":
		return RACK
	case "device", "dv":
		return DEVICE
	case "ac":
		return AC
	case "panel", "pn":
		return PWRPNL
	case "stray_device":
		return STRAY_DEV
	case "room_template":
		return ROOMTMPL
	case "obj_template":
		return OBJTMPL
	case "cabinet", "cb":
		return CABINET
	case "group", "gr":
		return GROUP
	case "corridor", "co":
		return CORIDOR
	case "sensor", "sr":
		return SENSOR
	default:
		return -1
	}
}

func GetParentOfEntity(ent int) int {
	switch ent {
	case TENANT:
		return -1
	case SITE:
		return ent - 1
	case BLDG:
		return ent - 1
	case ROOM:
		return ent - 1
	case RACK:
		return ent - 1
	case DEVICE:
		return ent - 1
	case AC:
		return ROOM
	case PWRPNL:
		return ROOM
	case ROOMTMPL:
		return -1
	case OBJTMPL:
		return -1
	case CABINET:
		return ROOM
	case GROUP:
		return -1
	case CORIDOR:
		return ROOM
	case SENSOR:
		return -2
	default:
		return -3
	}
}

func NodesAtLevel(root **Node, x Stack) []string {
	if x.Len() > 0 {
		name := x.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			if State.DebugLvl > 0 {
				println("Name doesn't exist! ", string(name.(string)))
			}

			l.GetWarningLogger().Println("Node name: ", string(name.(string)), "doesn't exist!")
			return nil
		}
		x.Pop()
		return NodesAtLevel(&node, x)
	} else {
		var items = make([]string, 0)
		var nm string
		//println("This is what we got:")
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			nm = string(i.Value.(*Node).Name)
			//println(nm)
			items = append(items, nm)
		}
		return items
	}
	return nil
}

// Utility function used by FetchJsonNodes
func strArrToMapStrInfArr(x []string) []map[string]interface{} {
	ans := []map[string]interface{}{}
	for i := range x {
		ans = append(ans, map[string]interface{}{"name": x[i]})
	}
	return ans
}

// Provides a mapping for stray
// and normal objects
func MapStrayString(x string) string {
	if x == "device" {
		return "stray-device"
	}
	if x == "sensor" {
		return "stray-sensor"
	}

	if x == "stray-device" {
		return "device"
	}
	if x == "stray-sensor" {
		return "sensor"
	}
	return "INVALID-MAP"
}

func MapStrayInt(x int) int {
	if x == DEVICE {
		return STRAY_DEV
	}
	if x == SENSOR {
		return STRAYSENSOR
	}

	if x == STRAY_DEV {
		return DEVICE
	}
	if x == STRAYSENSOR {
		return SENSOR
	}
	return -1
}

// New Tree funcs here
func StrayWalk(root **Node, prefix string, depth int) {

	if depth > 0 {
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			node := i.Value.(*Node)

			if i.Next() == nil {
				fmt.Println(prefix+"└──", node.Name)
				StrayWalk(&node, prefix+"    ", depth-1)
			} else {
				fmt.Println(prefix+("├──"), node.Name)
				StrayWalk(&node, prefix+"│   ", depth-1)
			}
		}

		if (*root).Nodes.Len() == 0 && depth > 0 {
			switch (*root).Name {
			case "Device":
				//Get Stray Devices and print them
				StrayAndDomain("stray-devices", prefix, depth)
			case "Sensor":
				//Get Stray Sensors and print them
				r, e := models.Send("GET",
					State.APIURL+"/api/stray-sensors", GetKey(), nil)
				resp := ParseResponse(r, e, "fetch objects")
				if resp != nil {
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				}
			default: //Error, execution should not reach here

			}
			return
		}
	}

}

func RootWalk(root **Node, path string, depth int) {
	org := FindNodeInTree(root, StrToStack("/Organisation"), true)
	fmt.Println("├──" + "Organisation")
	OrganisationWalk(org, "│   ", depth-1)

	logical := FindNodeInTree(root, StrToStack("/Logical"), true)
	fmt.Println("├──" + "Logical")
	LogicalWalk(logical, "│   ", depth-1)

	phys := FindNodeInTree(root, StrToStack("/Physical"), true)
	fmt.Println("└──" + "Physical")
	PhysicalWalk(phys, "    ", path, depth-1)
}

func LogicalWalk(root **Node, prefix string, depth int) {

	if root != nil {
		if depth >= 0 {
			if (*root).Nodes.Len() == 0 {
				switch (*root).Name {
				case "ObjectTemplates":
					//Get All Obj Templates and print them
					r, e := models.Send("GET",
						State.APIURL+"/api/obj-templates", GetKey(), nil)
					resp := ParseResponse(r, e, "fetching objects")
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				case "RoomTemplates":
					//Get All Room Templates and print them
					r, e := models.Send("GET",
						State.APIURL+"/api/room-templates", GetKey(), nil)
					resp := ParseResponse(r, e, "fetching objects")
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				case "Groups":
					//Get All Groups and print them
					r, e := models.Send("GET",
						State.APIURL+"/api/groups", GetKey(), nil)
					resp := ParseResponse(r, e, "fetching objects")
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				default: //Error case, execution should not reach here

				}
				return
			}

			for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
				if i.Next() == nil {
					fmt.Println(prefix+"└──", (i.Value.(*Node).Name))
					value := i.Value.(*Node)
					LogicalWalk(&(value), prefix+"    ", depth-1)

				} else {
					fmt.Println(prefix+("├──"), (i.Value.(*Node).Name))
					value := i.Value.(*Node)
					LogicalWalk(&(value), prefix+"│   ", depth-1)

				}
			}
		}
	}

}

func OrganisationWalk(root **Node, prefix string, depth int) {

	if root != nil {
		if depth >= 0 {
			if (*root).Nodes.Len() == 0 {
				switch (*root).Name {
				case "Domain":
					StrayAndDomain("domains", prefix, depth)
				case "Enterprise":
					//Most likely same as Domain case
				}
			}

			for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
				if i.Next() == nil {
					fmt.Println(prefix+"└──", (i.Value.(*Node).Name))
					value := i.Value.(*Node)
					OrganisationWalk(&(value), prefix+"    ", depth-1)

				} else {
					fmt.Println(prefix+("├──"), (i.Value.(*Node).Name))
					value := i.Value.(*Node)
					OrganisationWalk(&(value), prefix+"│   ", depth-1)

				}
			}
		}
	}

}

func PhysicalWalk(root **Node, prefix, path string, depth int) {
	arr := strings.Split(path, "/")
	if len(arr) == 3 {
		//println("DEBUG ENTERED")
		if arr[2] == "Stray" {
			fmt.Println(prefix + "├──Device")
			if depth >= 1 {
				//Get and Print Stray Devices
				StrayAndDomain("stray-devices", prefix+"│   ", depth)

				//Get and Print Stray Sensors
				fmt.Println(prefix + "└──Sensor")
				r1, e1 := models.Send("GET",
					State.APIURL+"/api/stray-sensors", GetKey(), nil)
				resp1 := ParseResponse(r1, e1, "fetch objects")

				if resp1 != nil {
					RemoteGetAllWalk(resp1["data"].(map[string]interface{}),
						prefix+"    ")
				}
			} else { //Extra else block for correct printing
				fmt.Println(prefix + "└──Sensor")
			}

		} else { //Interacting with Tenants
			//Check Depth
			//if depth > 1 { //Get Obj Hierarchy ???
			depthStr := strconv.Itoa(depth + 1)

			//Need to convert path to URL then append /all?limit=depthStr
			_, urls := CheckPathOnline(path)
			r, e := models.Send("GET", urls, GetKey(), nil)

			parsed := ParseResponse(r, e, "get object")
			if parsed != nil {

				obj := parsed["data"].(map[string]interface{})
				cat := obj["category"].(string)
				ID := obj["id"].(string)
				URL := State.APIURL + "/api/" +
					cat + "s/" + ID + "/all?limit=" + depthStr
				r1, e1 := models.Send("GET", URL, GetKey(), nil)
				parsedRoot := ParseResponse(r1, e1, "get object hierarchy")
				if parsedRoot != nil {
					if _, ok := parsedRoot["data"]; ok {
						RemoteHierarchyWalk(
							parsedRoot["data"].(map[string]interface{}),
							prefix, depth+1)
					}

				}
			}

		}
	}
	if len(arr) == 2 { //Means path== "/Physical"

		var resp map[string]interface{}
		if arr[1] == "Physical" { //Means path== "/Physical"

			//Need to check num tenants before passing the prefix
			//Get and Print Tenants Block

			r, e := models.Send("GET",
				State.APIURL+"/api/tenants", GetKey(), nil)
			resp = ParseResponse(r, e, "fetch objects")
			strayNode := FindNodeInTree(&State.TreeHierarchy,
				StrToStack("/Physical/Stray"), true)

			if length, _ := GetRawObjectsLength(resp); length > 0 {
				fmt.Println(prefix + "├──" + " Stray")
				StrayWalk(strayNode, prefix+"│   ", depth)
			} else {
				fmt.Println(prefix + "└──" + " Stray")
				StrayWalk(strayNode, prefix+"   ", depth)
			}

			if resp != nil {
				if depth == 0 {
					if _, ok := resp["data"]; ok {
						RemoteGetAllWalk(resp["data"].(map[string]interface{}),
							prefix)
					}
					return
				}

			}

			if depth > 0 {
				if _, ok := resp["data"]; ok {
					tenants := resp["data"].(map[string]interface{})["objects"].([]interface{})

					size := len(tenants)
					for idx, tInf := range tenants {
						tenant := tInf.(map[string]interface{})
						ID := tenant["id"].(string)
						depthStr := strconv.Itoa(depth)

						var subPrefix string
						var currPrefix string
						if idx == size-1 {
							subPrefix = prefix + "    "
							currPrefix = prefix + "└──"
						} else {
							subPrefix = prefix + "│   "
							currPrefix = prefix + "├──"
						}

						fmt.Println(currPrefix + tenant["name"].(string))

						//Get Hierarchy for each tenant and walk
						r, e := models.Send("GET",
							State.APIURL+"/api/tenants/"+ID+"/all?limit="+depthStr, GetKey(), nil)
						resp := ParseResponse(r, e, "fetch objects")
						if resp != nil {
							RemoteHierarchyWalk(resp["data"].(map[string]interface{}),
								subPrefix, depth)
						}

					}
				}

			}
		} else { //Means path == "/"

			if depth >= 0 {

				strayNode := FindNodeInTree(&State.TreeHierarchy,
					StrToStack("/Physical/Stray"), true)

				//Get and Print Tenants Block
				r, e := models.Send("GET",
					State.APIURL+"/api/tenants", GetKey(), nil)
				resp = ParseResponse(r, e, "fetch objects")

				//Need to check num tenants before passing the prefix
				if length, _ := GetRawObjectsLength(resp); length > 0 {
					fmt.Println(prefix + "├──" + " Stray")
					StrayWalk(strayNode, prefix+"│   ", depth)
				} else {
					fmt.Println(prefix + "└──" + " Stray")
					StrayWalk(strayNode, prefix+"   ", depth)
				}
				if resp != nil {
					if depth == 0 {
						if _, ok := resp["data"]; ok {
							RemoteGetAllWalk(resp["data"].(map[string]interface{}),
								prefix)
						}
						return
					}

				}

				//If hierarchy happens to be greater than 1
				if depth > 0 && resp != nil {
					if tenants := GetRawObjects(resp); tenants != nil {
						size := len(tenants)
						for idx, tInf := range tenants {
							tenant := tInf.(map[string]interface{})
							ID := tenant["id"].(string)
							depthStr := strconv.Itoa(depth)

							//Get Hierarchy for each tenant and walk
							r, e := models.Send("GET",
								State.APIURL+"/api/tenants/"+ID+"/all?limit="+depthStr, GetKey(), nil)
							resp := ParseResponse(r, e, "fetch objects")

							var subPrefix string
							var currPrefix string
							if idx == size-1 {
								subPrefix = prefix + "    "
								currPrefix = prefix + "└──"
							} else {
								subPrefix = prefix + "│   "
								currPrefix = prefix + "├──"
							}

							fmt.Println(currPrefix + tenant["name"].(string))
							if resp != nil {
								RemoteHierarchyWalk(resp["data"].(map[string]interface{}),
									subPrefix, depth)
							}
						}
					}
				}

			}
		}

	}

	if len(arr) > 3 { //Could still be Stray not sure yet
		if arr[2] == "Stray" && len(arr) <= 4 {
			//println("DEBUG IS THIS EDGE CASE?")
			//strayNode := FindNodeInTree(&State.TreeHierarchy,
			//	StrToStack("/Physical/Stray"), true)
			//StrayWalk(strayNode, prefix, depth+1)
			//println("DEBUG LEN ARR:", len(arr))
			StrayAndDomain("stray-devices", prefix, depth)
		} else {
			//Get Object hierarchy and walk
			depthStr := strconv.Itoa(depth + 1)

			//Need to convert path to URL then append /all?limit=depthStr
			_, urls := CheckPathOnline(path)
			r, e := models.Send("GET", urls, GetKey(), nil)
			//WE need to get the Object in order for us to create
			//the correct GET /all?limit=depthStr URL
			//we get the object category and ID in the JSON response

			parsed := ParseResponse(r, e, "get object")
			if parsed != nil {

				obj := parsed["data"].(map[string]interface{})
				cat := obj["category"].(string)
				ID := obj["id"].(string)
				URL := State.APIURL + "/api/" +
					cat + "s/" + ID + "/all?limit=" + depthStr
				r1, e1 := models.Send("GET", URL, GetKey(), nil)
				parsedRoot := ParseResponse(r1, e1, "get object hierarchy")
				if parsedRoot != nil {
					if _, ok := parsedRoot["data"]; ok {
						RemoteHierarchyWalk(
							parsedRoot["data"].(map[string]interface{}),
							prefix, depth+1)
					}

				}
			}
		}
	}
}

func Filter(root map[string]interface{}, depth int, ent string) {
	var arr []interface{}
	var replacement []interface{}
	if root == nil {
		return
	}

	if _, ok := root["objects"]; !ok {
		return
	}

	if _, ok := root["objects"].([]interface{}); !ok {
		return
	}
	arr = root["objects"].([]interface{})
	//length = len(arr)

	for _, m := range arr {
		if object, ok := m.(map[string]interface{}); ok {
			if object["parentId"] == nil {
				//Change m -> result of hierarchal API call
				ext := object["id"].(string) + "/all?limit=" + strconv.Itoa(depth)
				URL := State.APIURL + "/api/" + ent + "/" + ext
				r, _ := models.Send("GET", URL, GetKey(), nil)
				parsed := ParseResponse(r, nil, "Fetch "+ent)
				m = parsed["data"].(map[string]interface{})
				replacement = append(replacement, m)
				//Disp(m.(map[string]interface{}))
			}
		}
	}

	root["objects"] = replacement
}

func StrayAndDomain(ent, prefix string, depth int) {
	//Do the call, filter and perform remote
	//hierarchy walk
	//Get All Domains OR Stray Devices and print them
	r, e := models.Send("GET",
		State.APIURL+"/api/"+ent, GetKey(), nil)
	resp := ParseResponse(r, e, "fetching objects")
	if resp != nil {
		if _, ok := resp["data"]; ok {
			data := resp["data"].(map[string]interface{})
			Filter(data, depth, ent)

			if objects, ok := data["objects"]; ok {
				length := len(objects.([]interface{}))
				for i, obj := range objects.([]interface{}) {
					if m, ok := obj.(map[string]interface{}); ok {
						subname := m["name"].(string)

						if i == length-1 {
							fmt.Println(prefix+"└──", subname)
							RemoteHierarchyWalk(m, prefix+"    ", depth-1)
						} else {
							fmt.Println(prefix+("├──"), subname)
							RemoteHierarchyWalk(m, prefix+"│   ", depth-1)
						}
					}

				}
			}
		}

	}

}

func RemoteGetAllWalk(root map[string]interface{}, prefix string) {
	var arr []interface{}
	var length int
	if root == nil {
		return
	}

	if _, ok := root["objects"]; !ok {
		return
	}

	if _, ok := root["objects"].([]interface{}); !ok {
		return
	}
	arr = root["objects"].([]interface{})
	length = len(arr)

	for i, m := range arr {
		var subname string
		if n, ok := m.(map[string]interface{})["name"].(string); ok {
			subname = n
		} else {
			subname = m.(map[string]interface{})["slug"].(string)
		}

		if i == length-1 {
			fmt.Println(prefix+"└──", subname)

		} else {
			fmt.Println(prefix+("├──"), subname)
		}
	}
}

func RemoteHierarchyWalk(root map[string]interface{}, prefix string, depth int) {

	if depth == 0 || root == nil {
		return
	}
	if infants, ok := root["children"]; !ok || infants == nil {
		return
	}

	//name := root["name"].(string)
	//println(prefix + name)

	//or cast to []interface{}
	arr := root["children"].([]interface{})

	//or cast to []interface{}
	length := len(arr)

	//or cast to []interface{}
	for i, mInf := range arr {
		m := mInf.(map[string]interface{})
		subname := m["name"].(string)

		if i == length-1 {
			fmt.Println(prefix+"└──", subname)
			RemoteHierarchyWalk(m, prefix+"    ", depth-1)
		} else {
			fmt.Println(prefix+("├──"), subname)
			RemoteHierarchyWalk(m, prefix+"│   ", depth-1)
		}
	}
}
