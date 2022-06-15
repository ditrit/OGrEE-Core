package controllers

import (
	"bufio"
	l "cli/logger"
	"cli/models"
	"cli/readline"
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
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
	SEPARATOR
	CABINET
	ROW
	TILE
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
	ScriptCalled     bool
	ScriptPath       string
	UnityClientURL   string
	APIURL           string
	UnityClientAvail bool  //For deciding to message unity or not
	FilterDisplay    bool  //Set whether or not to send attributes to unity
	ObjsForUnity     []int //Deciding what objects should be sent to unity
	DrawableObjs     []int //Indicate which objs drawable in unity
	DrawableJsons    map[string]map[string]interface{}
	DebugLvl         int
	LineNumber       int //Used exectuting scripts
	Terminal         **readline.Instance
}

type Node struct {
	ID     string
	PID    string
	Entity int
	Name   string
	Path   string
	Nodes  list.List
}

//Intialise the ShellState
func InitState(debugLvl int) {
	State.DebugLvl = debugLvl
	State.ClipBoard = nil
	State.TreeHierarchy = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	State.TreeHierarchy.PID = ""
	State.CurrPath = "/Physical"
	State.PrevPath = "/Physical"
	State.LineNumber = 0

	//Send login notification
	data := map[string]interface{}{"api_url": State.APIURL, "api_token": GetKey()}
	req := map[string]interface{}{"type": "login", "data": data}
	e := models.ContactUnity("POST", State.UnityClientURL, req)
	if e != nil {
		l.WarningLogger.Println("Note: Unity Client Unreachable")
		fmt.Println("Note: Unity Client Unreachable ")
		State.UnityClientAvail = false
	} else {
		fmt.Println("Unity Client is Reachable!")
		State.UnityClientAvail = true
	}
	//Set the filter attributes setting
	State.FilterDisplay = false

	phys := &Node{}
	phys.Name = "Physical"
	phys.PID = ""
	phys.ID = "-2"
	State.TreeHierarchy.Nodes.PushBack(phys)

	stray := &Node{}
	stray.Name = "Stray"
	stray.PID = "-2"
	stray.ID = "-3"
	stray.Path = "/Physical/"
	SearchAndInsert(&State.TreeHierarchy, stray, "/Physical")

	strayDev := &Node{}
	strayDev.Name = "Device"
	strayDev.PID = "-3"
	strayDev.ID = "-4"
	strayDev.Path = "/Physical/Stray"
	SearchAndInsert(&State.TreeHierarchy, strayDev, "/Physical/Stray")

	straySens := &Node{}
	straySens.Name = "Sensor"
	straySens.PID = "-3"
	straySens.ID = "-5"
	straySens.Path = "/Physical/Stray"
	SearchAndInsert(&State.TreeHierarchy, straySens, "/Physical/Stray")

	// SETUP LOGICAL HIERARCHY START
	// TODO: PUT THIS SECTION IN A LOOP
	logique := &Node{}
	logique.ID = "0"
	logique.Name = "Logical"
	logique.Path = "/"
	State.TreeHierarchy.Nodes.PushBack(logique)

	oTemplate := &Node{}
	oTemplate.ID = "1"
	oTemplate.PID = "0"
	oTemplate.Entity = -1
	oTemplate.Name = "ObjectTemplates"
	oTemplate.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, oTemplate, "/Logical")

	rTemplate := &Node{}
	rTemplate.ID = "2"
	rTemplate.PID = "0"
	rTemplate.Entity = -1
	rTemplate.Name = "RoomTemplates"
	rTemplate.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, rTemplate, "/Logical")

	group := &Node{}
	group.ID = "3"
	group.PID = "0"
	group.Entity = -1
	group.Name = "Groups"
	group.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, group, "/Logical")

	//SETUP LOGICAL HIERARCHY END

	//SETUP DOMAIN/ENTERPRISE
	organisation := &Node{}
	organisation.ID = "5"
	organisation.Name = "Organisation"
	organisation.Path = "/"
	State.TreeHierarchy.Nodes.PushBack(organisation)

	domain := &Node{}
	domain.Name = "Domain"
	domain.PID = "5"
	domain.ID = "-6"
	domain.Path = "/Organisation"
	SearchAndInsert(&State.TreeHierarchy, domain, "/Organisation")

	enterprise := &Node{}
	enterprise.ID = "0"
	enterprise.PID = "5"
	enterprise.Name = "Enterprise"
	enterprise.Path = "/Organisation"
	SearchAndInsert(&State.TreeHierarchy, enterprise, "/Organisation")

	//Set which objects Unity will be notified about
	State.ObjsForUnity = SetObjsForUnity("updates")
	State.DrawableObjs = SetObjsForUnity("drawable")
	State.DrawableJsons = make(map[string]map[string]interface{}, 16)

	for i := TENANT; i < GROUP+1; i++ {
		ent := EntityToString(i)
		State.DrawableJsons[ent] = SetDrawableTemplate(ent)
	}
}

//It is useful to have the state to hold
//a pointer to our readline terminal
func SetStateReadline(rl *readline.Instance) {
	State.Terminal = &rl
}

//Startup the go routine for listening
func TriggerListen(rl *readline.Instance) {
	go models.ListenForUnity(rl)
}

//Helper for InitState will
//insert objs
func SetObjsForUnity(x string) []int {
	res := []int{}
	key := x + "="
	allDetected := false
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), key) {
			//ObjStr is equal to everything after 'updates='
			objStr := strings.SplitAfter(scanner.Text(), key)[1]
			arr := strings.Split(objStr, ",")

			for i := range arr {
				arr[i] = strings.ToLower(arr[i])

				if val := EntityStrToInt(arr[i]); val != -1 {
					res = append(res, val)

				} else if arr[i] == "all" {
					//Exit the loop and use default code @ end of function
					allDetected = true
					i = len(arr)
				}
			}

		}
	}

	if err := scanner.Err(); err != nil {
		l.ErrorLogger.Println(err)
		fmt.Println(err)
	}

	//Use default values
	//Set the array to all and exit
	//GROUP is the greatest value int enum type
	//So we use that for the cond guard
	if allDetected || len(res) == 0 {
		if len(res) == 0 && !allDetected {
			l.WarningLogger.Println(x + " key not found, going to use defaults")
			println(x + " key not found, going to use defaults")
		}
		for idx := 0; idx < GROUP; idx++ {
			res = append(res, idx)
		}
	}
	return res
}

func SetDrawableTemplate(entity string) map[string]interface{} {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords

	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), entity) {
			objStr := strings.Split(scanner.Text(), entity+"DrawableJson=")[1]
			objStr = strings.Trim(objStr, "'\"")

			//Now retrieve file
			ans := map[string]interface{}{}
			f, e := ioutil.ReadFile(objStr)
			if e != nil {
				l.WarningLogger.Println("Specified template for" + entity + "not found")
				if State.DebugLvl > 2 {
					println("Specified template for " + entity +
						" not found, resorting to defaults")
				}
				return nil
			}
			json.Unmarshal(f, &ans)
			return ans
		}
	}

	return nil
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

func GetChildren(curr int) []*Node {

	resp, e := models.Send("GET",
		State.APIURL+"/api/"+EntityToString(curr)+"s",
		GetKey(), nil)
	if e != nil {
		println("Error while getting children!")
		Exit()
	}
	//println("REQ:", "http://localhost:3001/api/"+EntityToString(curr)+"s")

	x := makeNodeArrFromResp(resp, curr)
	return x

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

//Automatically assign Unity and API URLs
func GetURLs() {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Falling back to default URLs")
		l.InfoLogger.Println("Falling back to default URLs")
		State.UnityClientURL = "http://localhost:5500"
		State.APIURL = "http://localhost:3001"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "unityURL=") {
			State.UnityClientURL = scanner.Text()[9:]
		}

		if strings.HasPrefix(scanner.Text(), "apiURL=") {
			State.APIURL = scanner.Text()[7:]
		}
	}

	if State.APIURL == "" {
		fmt.Println("Falling back to default API URL:" +
			"http://localhost:3001")
		l.InfoLogger.Println("Falling back to default API URL:" +
			"http://localhost:3001")
		State.APIURL = "http://localhost:3001"
	}

	if State.UnityClientURL == "" {
		fmt.Println("Falling back to default Unity URL:" +
			"http://localhost:5500")
		l.InfoLogger.Println("Falling back to default Unity URL:" +
			"http://localhost:5500")
		State.APIURL = "http://localhost:5500"
	}

}

//Function is an abstraction of a normal exit
func Exit() {
	//writeHistoryOnExit(&State.sessionBuffer)
	//runtime.Goexit()
	os.Exit(0)
}

func makeNodeArrFromResp(resp *http.Response, entity int) []*Node {
	arr := []*Node{}
	var jsonResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&jsonResp)
	defer resp.Body.Close()
	if err != nil {
		println("Error: " + err.Error())
		return nil
	}

	//println("NOW@,", entity)
	//println("MSG: ", jsonResp["message"].(string))
	//for i := range jsonResp {
	//	println("KEY:", i)
	//}
	//println("STATUS:", jsonResp["status"].(bool))

	objs, ok := ((jsonResp["data"]).(map[string]interface{})["objects"]).([]interface{})
	sd1obj, ok1 := ((jsonResp["data"]).(map[string]interface{})["subdevices1"]).([]interface{})
	if !ok && !ok1 {
		println("Nothing found!")
		return nil
	} else if ok1 && !ok {
		objs = sd1obj
	}
	//println("LEN-OBJS:", len(objs))
	for i, _ := range objs {
		node := &Node{}
		node.Path = ""
		node.Entity = entity
		if v, ok := (objs[i].(map[string]interface{}))["name"]; ok {
			node.Name = v.(string)
		} else if v, ok := (objs[i].(map[string]interface{}))["slug"]; ok {
			node.Name = v.(string)
		} else {
			l.ErrorLogger.Println("Object obtained does not have name or slug!" +
				"Now Exiting")
			println("Object obtained does not have name or slug!" +
				"Now Exiting")
		}
		//node.Name = (string((objs[i].(map[string]interface{}))["name"].(string)))
		node.ID, _ = (objs[i].(map[string]interface{}))["id"].(string)
		num, ok := objs[i].(map[string]interface{})["parentId"].(string)
		if !ok {
			if entity == 0 { //We have TENANT
				node.PID = ""
			} else {
				//ERROR Case
				node.PID = "ERR"
			}
		} else {
			node.PID = num
		}
		arr = append(arr, node)
	}
	return arr
}

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

func StrToStack(x string) *Stack {
	stk := Stack{}
	numPrev := 0
	sarr := strings.Split(x, "/")
	for i := len(sarr) - 1; i >= 0; i-- {
		if sarr[i] == ".." {
			numPrev += 1
		} else if sarr[i] != "" {
			if numPrev == 0 {
				stk.Push(sarr[i])
			} else {
				numPrev--
			}
		}

	}
	return &stk
}

func getNextInPath(name string, root *Node) *Node {
	for i := root.Nodes.Front(); i != nil; i = i.Next() {
		if (i.Value.(*Node)).Name == name {
			return (i.Value.(*Node))
		}
	}
	return nil
}

func DispAtLevel(root **Node, x Stack) []string {
	if x.Len() > 0 {
		name := x.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			println("Name doesn't exist! ", string(name.(string)))
			l.WarningLogger.Println("Node name: ", string(name.(string)), "doesn't exist!")
			return nil
		}
		x.Pop()
		return DispAtLevel(&node, x)
	} else {
		var items = make([]string, 0)
		var nm string
		if State.DebugLvl >= 2 {
			println("This is what we got:")
		}
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			nm = string(i.Value.(*Node).Name)
			println(nm)
			items = append(items, nm)
		}
		return items
	}
}

func DispAtLevelTAB(root **Node, x Stack) []string {
	if x.Len() > 0 {
		name := x.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			//println("Name doesn't exist! ", string(name.(string)))
			return nil
		}
		x.Pop()
		return DispAtLevelTAB(&node, x)
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

//Replaces DispAtLevel since we are no longer
//storing objects in a tree and returns string arr
func FetchNodesAtLevel(Path string) []string {
	names := []string{}
	urls := []string{}

	paths := strings.Split(path.Clean(Path), "/")

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

//Same as FetchNodesAtLevel but returns the JSONs
//in map[string]inf{} format
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

func DispStk(x Stack) {
	for i := x.Pop(); i != nil; i = x.Pop() {
		println((i.(*Node)).Name)
	}
}

func GetPathStrAtPtr(root, curr **Node, path string) (bool, string) {
	if root == nil || *root == nil {
		return false, ""
	}

	if *root == *curr {
		if path == "" {
			path = "/"
		}
		return true, path
	}

	for i := (**root).Nodes.Front(); i != nil; i = i.Next() {
		nd := (*Node)((i.Value.(*Node)))
		exist, path := GetPathStrAtPtr(&nd,
			curr, path+"/"+i.Value.(*Node).Name)
		if exist == true {
			return exist, path
		}
	}
	return false, path
}

func CheckPath(root **Node, x, pstk *Stack) (bool, string, **Node) {
	if x.Len() == 0 {
		_, path := GetPathStrAtPtr(&State.TreeHierarchy, root, "")
		//println(path)
		return true, path, root
	}

	p := x.Pop()

	//At Root
	if pstk.Len() == 0 && string(p.(string)) == ".." {
		//Pop until p != ".."
		for ; p != nil && string(p.(string)) == ".."; p = x.Pop() {
		}
		if p == nil {
			_, path := GetPathStrAtPtr(&State.TreeHierarchy, root, "/")
			//println(path)
			return true, path, root
		}

		//Somewhere in tree
	} else if pstk.Len() > 0 && string(p.(string)) == ".." {
		prevNode := (pstk.Pop()).(*Node)
		return CheckPath(&prevNode, x, pstk)
	}

	nd := getNextInPath(string(p.(string)), *root)
	if nd == nil {
		return false, "", nil
	}

	pstk.Push(*root)
	return CheckPath(&nd, x, pstk)

}

//If the path refers to local tree the
//func will still verify it with local tree
func CheckPathOnline(Path string) (bool, string) {

	pathSplit := strings.Split(path.Clean(Path), "/")

	//Check if path refers to object in local State Tree
	if len(pathSplit) <= 3 {
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

//Return extra bool so that the Parent can delete
//leaf and keep track without stack
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
				println("Name doesn't exist! ", string(name.(string)))
			}

			l.WarningLogger.Println("Name doesn't exist! ", string(name.(string)))
			return nil
		}
		path.Pop()
		return FindNodeInTree(&node, path, silenced)
	} else {
		return root
	}
}

func GetNodes(root **Node, entity int) []*Node {
	if root == nil {
		return nil
	}

	if (*root).Entity == entity {
		return []*Node{(*root)}
	}

	ans := []*Node{}
	for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
		nd := i.Value.(*Node)
		ans = append(ans, GetNodes(&nd, entity)...)
	}
	return ans
}

func FindNodeByIDP(root **Node, ID, PID string) *Node {
	if root != nil {

		if (*root).PID == PID && (*root).ID == ID {
			return (*root)
		}

		for i := (**root).Nodes.Front(); i != nil; i = i.Next() {
			nd := (*Node)((i.Value.(*Node)))
			if ans := FindNodeByIDP(&nd, ID, PID); ans != nil {
				return ans
			}
		}
	}

	return nil
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
	case SEPARATOR:
		return "separator"
	case STRAY_DEV:
		return "stray_device"
	case ROOMTMPL:
		return "room_template"
	case OBJTMPL:
		return "obj_template"
	case CABINET:
		return "cabinet"
	case ROW:
		return "row"
	case TILE:
		return "tile"
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
	case "separator", "sp":
		return SEPARATOR
	case "stray_device":
		return STRAY_DEV
	case "room_template":
		return ROOMTMPL
	case "obj_template":
		return OBJTMPL
	case "cabinet", "cb":
		return CABINET
	case "row":
		return ROW
	case "tile", "tl":
		return TILE
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
	case SEPARATOR:
		return ROOM
	case ROOMTMPL:
		return -1
	case OBJTMPL:
		return -1
	case CABINET:
		return ROOM
	case ROW:
		return ROOM
	case TILE:
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
			println("Name doesn't exist! ", string(name.(string)))
			l.WarningLogger.Println("Node name: ", string(name.(string)), "doesn't exist!")
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

//Utility function used by FetchJsonNodes
func strArrToMapStrInfArr(x []string) []map[string]interface{} {
	ans := []map[string]interface{}{}
	for i := range x {
		ans = append(ans, map[string]interface{}{"name": x[i]})
	}
	return ans
}

//Provides a mapping for stray
//and normal objects
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
