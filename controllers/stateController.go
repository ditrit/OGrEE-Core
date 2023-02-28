package controllers

import (
	l "cli/logger"
	"cli/models"
	"cli/readline"
	"container/list"
	"net/http"
	"strings"
	"time"
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

	paths := strings.Split(Path, "/")

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

	paths := strings.Split(Path, "/")

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
			if State.DebugLvl > NONE {
				println(e.Error())
			}
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

	pathSplit := strings.Split(Path, "/")

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
