package controllers

import (
	"cli/models"
	"container/list"
	"encoding/json"
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
	SUBDEV
	SUBDEV1
	WALL
	CORRIDOR
	GROUP
)

var BuildTime string
var BuildHash string
var BuildTree string
var GitCommitDate string
var State ShellState

type ShellState struct {
	CurrPath      string
	PrevPath      string
	ClipBoard     *[]string
	TreeHierarchy *Node
	ScriptCalled  bool
	ScriptPath    string
	DebugLvl      int
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
	State.TreeHierarchy.PID = ""
	State.CurrPath = "/"
	x := GetChildren(0)
	for i := range x {
		x[i].Path = "/" + x[i].Name
		State.TreeHierarchy.Nodes.PushBack(x[i])
	}

	for i := 1; i < DEVICE; i++ {
		//time.Sleep(2 * time.Second)
		x := GetChildren(i)
		for k := range x {
			SearchAndInsert(&State.TreeHierarchy, x[k], i, "")
		}
	}
}

func GetChildren(curr int) []*Node {

	//Loop because sometimes a
	//Stream Error occurs
	for {
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/"+EntityToString(curr)+"s",
			GetKey(), nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		//println("REQ:", "https://ogree.chibois.net/api/"+EntityToString(curr)+"s")

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
		node.Name = (string((objs[i].(map[string]interface{}))["name"].(string)))
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
	if (*root).ID == curr.PID && curr.Entity == (*root).Entity+1 {
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
	case SUBDEV:
		return "subdevice"
	default:
		return "subdevice1"
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
	case "subdevice":
		return SUBDEV
	default:
		return SUBDEV1
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
