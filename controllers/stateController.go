package controllers

import (
	"bufio"
	"cli/models"
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	UnityClientAvail bool //For deciding to message unity or not
	DebugLvl         int
	LineNumber       int //Used exectuting scripts
	TemplateTable    map[string]map[string]interface{}
}

type Node struct {
	ID     string
	PID    string
	Entity int
	Name   string
	Path   string
	Nodes  list.List
}

//Populate hierarchy into B Tree like
//structure
func InitState(debugLvl int) {
	State.DebugLvl = debugLvl
	State.ClipBoard = nil
	State.TreeHierarchy = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	State.TemplateTable = map[string]map[string]interface{}{}
	State.TreeHierarchy.PID = ""
	State.CurrPath = "/Physical"
	State.LineNumber = 0
	e := models.ContactUnity("GET", State.UnityClientURL, nil)
	if e != nil {
		WarningLogger.Println("Note: Unity Client Unreachable")
		fmt.Println("Note: Unity Client Unreachable")
		State.UnityClientAvail = false
	} else {
		fmt.Println("Unity Client is Reachable!")
		State.UnityClientAvail = true
	}

	phys := &Node{}
	phys.Name = "Physical"
	phys.PID = ""
	phys.ID = "-2"
	x := GetChildren(0)
	for i := range x {
		x[i].Path = "/Physical/" + x[i].Name
		phys.Nodes.PushBack(x[i])
		//State.TreeHierarchy.Nodes.PushBack(x[i])
	}
	State.TreeHierarchy.Nodes.PushBack(phys)

	for i := 1; i < SENSOR+1; i++ {
		x := GetChildren(i)
		for k := range x {
			SearchAndInsert(&State.TreeHierarchy, x[k], i, "")
		}
	}

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
	SearchAndInsert(&State.TreeHierarchy, oTemplate, 0, "/Logical")
	q := GetChildren(OBJTMPL)
	for k := range q {
		q[k].PID = "1"
		SearchAndInsert(&State.TreeHierarchy, q[k], 1, "")
	}

	rTemplate := &Node{}
	rTemplate.ID = "2"
	rTemplate.PID = "0"
	rTemplate.Entity = -1
	rTemplate.Name = "RoomTemplates"
	rTemplate.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, rTemplate, 0, "/Logical")
	q = GetChildren(ROOMTMPL)
	for k := range q {
		q[k].PID = "2"
		SearchAndInsert(&State.TreeHierarchy, q[k], 1, "")
	}

	group := &Node{}
	group.ID = "3"
	group.PID = "0"
	group.Entity = -1
	group.Name = "Groups"
	group.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, group, 0, "/Logical")
	q = GetChildren(GROUP)
	for k := range q {
		q[k].PID = "3"
		SearchAndInsert(&State.TreeHierarchy, q[k], 1, "")
	}
	//SETUP LOGICAL HIERARCHY END

	//SETUP DOMAIN/ENTERPRISE
	enterprise := &Node{}
	enterprise.ID = "0"
	enterprise.Name = "Enterprise"
	enterprise.Path = "/"
	State.TreeHierarchy.Nodes.PushBack(enterprise)
}

func GetLineNumber() int {
	return State.LineNumber
}

func GetScriptPath() string {
	return State.ScriptPath
}

func GetChildren(curr int) []*Node {

	//Loop because sometimes a
	//Stream Error occurs
	for {
		resp, e := models.Send("GET",
			State.APIURL+"/api/"+EntityToString(curr)+"s",
			GetKey(), nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		//println("REQ:", "http://localhost:3001/api/"+EntityToString(curr)+"s")

		x := makeNodeArrFromResp(resp, curr)
		if x != nil {
			return x
		}
	}
}

func SearchAndInsert(root **Node, node *Node, dt int, path string) {
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
			SearchAndInsert(&x, node, dt+1, path+"/"+x.Name)
		}
	}
	return
}

//Automatically assign Unity and API URLs
func GetURLs() {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Falling back to default URLs")
		InfoLogger.Println("Falling back to default URLs")
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
		InfoLogger.Println("Falling back to default API URL:" +
			"http://localhost:3001")
		State.APIURL = "http://localhost:3001"
	}

	if State.UnityClientURL == "" {
		fmt.Println("Falling back to default Unity URL:" +
			"http://localhost:5500")
		InfoLogger.Println("Falling back to default Unity URL:" +
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
			ErrorLogger.Println("Object obtained does not have name or slug!" +
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
			WarningLogger.Println("Node name: ", string(name.(string)), "doesn't exist!")
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
	return nil
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

func UpdateTree(root **Node, curr *Node) bool {
	if root == nil {
		return false
	}

	//Add only when the PID matches Parent's ID
	//And (possibly if) the parent is indeed the correct Entity
	/*&& GetParentOfEntity(curr.Entity) == (*root).Entity*/
	if (*root).ID == curr.PID {
		(*root).Nodes.PushBack(curr)
		return true
	}

	for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
		nxt := (i.Value).(*Node)
		x := UpdateTree(&nxt, curr)
		if x != false {
			return true
		}
	}
	return false
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

func FindNodeInTree(root **Node, path *Stack) **Node {
	if root == nil {
		return nil
	}

	if path.Len() > 0 {
		name := path.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			println("Name doesn't exist! ", string(name.(string)))
			WarningLogger.Println("Name doesn't exist! ", string(name.(string)))
			return nil
		}
		path.Pop()
		return FindNodeInTree(&node, path)
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
	case "tenant":
		return TENANT
	case "site":
		return SITE
	case "building", "bldg":
		return BLDG
	case "room":
		return ROOM
	case "rack":
		return RACK
	case "device":
		return DEVICE
	case "ac":
		return AC
	case "panel":
		return PWRPNL
	case "separator":
		return SEPARATOR
	case "room_template":
		return ROOMTMPL
	case "obj_template":
		return OBJTMPL
	case "cabinet":
		return CABINET
	case "row":
		return ROW
	case "tile":
		return TILE
	case "group":
		return GROUP
	case "corridor":
		return CORIDOR
	case "sensor":
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
			WarningLogger.Println("Node name: ", string(name.(string)), "doesn't exist!")
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
