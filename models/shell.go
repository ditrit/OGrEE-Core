package models

import (
	u "cli/utils"
	"container/list"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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
)

type ShellState struct {
	CurrPath      string
	sessionBuffer list.List
	TreeHierarchy *Node
}

type Node struct {
	ID     int
	Entity int
	Name   string
	Nodes  list.List
}

var State ShellState

func writeHistoryOnExit(ss *list.List) {
	f, err := os.OpenFile(".resources/.history",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	for i := ss.Back(); i != nil; i = ss.Back() {
		f.Write([]byte(string(ss.Remove(i).(string) + "\n")))
	}
	return
}

func InitState() {
	State.sessionBuffer = *State.sessionBuffer.Init()
	State.CurrPath = "/"
}

func UpdateSessionState(ln *string) {
	State.sessionBuffer.PushBack(*ln)
}

//Function is an abstraction of a normal exit
func Exit() {
	writeHistoryOnExit(&State.sessionBuffer)
	os.Exit(0)
}

//Populate hierarchy into B Tree like
//structure
func Populate(root **Node, dt int) {
	if dt != SUBDEV1 || root != nil {
		println("Now getting children...")
		arr := getChildren(*root)
		for i := range arr {
			println("Now populating: ", arr[i].Name)
			Populate(&arr[i], dt+1)
			(*root).Nodes.PushBack(arr[i])
		}
	}
}

func getChildren(curr *Node) []*Node {
	switch curr.Entity {
	case -1:
		println("Getting list of tenants now...")
		resp, e := u.Send("GET",
			"https://ogree.chibois.net/api/user/tenants", nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, TENANT)
	case TENANT:
		println("Getting list of sites now...")
		resp, e := u.Send("GET",
			"https://ogree.chibois.net/api/user/tenants/"+curr.Name+"/sites",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, SITE)
	case SITE:
		println("Getting list of bldgs now...")
		resp, e := u.Send("GET",
			"https://ogree.chibois.net/api/user/sites/"+
				strconv.Itoa(curr.ID)+"/buildings",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, BLDG)
	case BLDG:
		println("Getting list of rooms now...")
		resp, e := u.Send("GET",
			"https://ogree.chibois.net/api/user/buildings/"+
				strconv.Itoa(curr.ID)+"/rooms",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, ROOM)
	case ROOM:
		println("Getting list of racks now...")
		resp, e := u.Send("GET",
			"https://ogree.chibois.net/api/user/rooms/"+
				strconv.Itoa(curr.ID)+"/racks",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, RACK)
	case RACK:
		println("Getting list of devices now...")
		resp, e := u.Send("GET",
			"https://ogree.chibois.net/api/user/racks/"+
				strconv.Itoa(curr.ID)+"/devices",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, DEVICE)
	case DEVICE:
		println("Getting list of subdevices now...")
		resp, e := u.Send("GET",
			"https://ogree.chibois.net/api/user/devices/"+
				strconv.Itoa(curr.ID)+"/subdevices",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, SUBDEV)
	case SUBDEV:
		println("Getting list of subdevice1 now...")
		resp, e := u.Send("GET",
			"https://ogree.chibois.net/api/user/subdevices/"+
				strconv.Itoa(curr.ID)+"/all",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, SUBDEV1)
	}
	return nil
}

func makeNodeArrFromResp(resp *http.Response, entity int) []*Node {
	arr := []*Node{}
	var jsonResp map[string]interface{}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		Exit()
	}
	json.Unmarshal(bodyBytes, &jsonResp)

	objs, ok := ((jsonResp["data"]).(map[string]interface{})["objects"]).([]interface{})
	sd1obj, ok1 := ((jsonResp["data"]).(map[string]interface{})["subdevices1"]).([]interface{})
	if !ok && !ok1 {
		println("Nothing found!")
		return nil
	} else if ok1 && !ok {
		objs = sd1obj
	}
	println("Now creating the nodes..")
	for i, _ := range objs {
		node := &Node{}
		node.Entity = entity
		node.Name = (string((objs[i].(map[string]interface{}))["name"].(string)))
		node.ID, _ = strconv.Atoi((objs[i].(map[string]interface{}))["id"].(string))
		println("We got: ", node.Name)
		arr = append(arr, node)
	}
	println("Printing before returning the nodes...")
	for i := range arr {
		println(arr[i].Name)
	}
	return arr
}

func DispTree() {
	nd := &(Node{})
	nd.Entity = -1
	Populate(&nd, 0)
	println("Now viewing the tree...")
	View(nd, 0)
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

func DispTree1() {
	(State.TreeHierarchy) = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	Populate(&State.TreeHierarchy, 0)
	println("Now viewing the tree...")
	DispAtLevel(&State.TreeHierarchy,
		*(strToStack(State.CurrPath)))
}

func strToStack(x string) *Stack {
	stk := Stack{}
	sarr := strings.Split(x, "/")
	for i := len(sarr) - 1; i >= 0; i-- {
		println("PUSHING TO STACK: ", sarr[i])
		if sarr[i] != "" {
			stk.Push(sarr[i])
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

func DispAtLevel(root **Node, x Stack) {
	if x.Len() > 0 {
		name := x.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			println("Name doesn't exist! ", string(name.(string)))
			return
		}
		x.Pop()
		DispAtLevel(&node, x)
	} else {
		println("This is what we got:")
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			println(string(i.Value.(*Node).Name))
		}
	}
	return
}
