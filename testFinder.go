package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"io/ioutil"
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
	PrevPath      string
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

type (
	Stack struct {
		top    *node
		length int
	}
	node struct {
		value interface{}
		prev  *node
	}
)

// Create a new stack
func New() *Stack {
	return &Stack{nil, 0}
}

// Return the number of items in the stack
func (this *Stack) Len() int {
	return this.length
}

// View the top item on the stack
func (this *Stack) Peek() interface{} {
	if this.length == 0 {
		return nil
	}
	return this.top.value
}

// Pop the top item of the stack and return it
func (this *Stack) Pop() interface{} {
	if this.length == 0 {
		return nil
	}

	n := this.top
	this.top = n.prev
	this.length--
	return n.value
}

// Push a value onto the top of the stack
func (this *Stack) Push(value interface{}) {
	n := &node{value, this.top}
	this.top = n
	this.length++
}

func StrToStack(x string) *Stack {
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

func dispStk(x *Stack) {
	for i := x.Pop(); i != nil; i = x.Pop() {
		println(string(i.(string)))
	}
}

func printJSON(x *map[string]interface{}) {
	for i, k := range *x {
		println(i)
		if i == "message" {
			println(string(k.(string)))
		}
	}
}

func Exit() {
	os.Exit(-1)
}

//Function helps with API Requests
func Send(method, URL string, data map[string]interface{}) (*http.Response,
	error) {
	client := &http.Client{}
	key := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjY2MjMxODA4MzY4ODYyODIyNX0.0kMvRpCU8bwZS61liCjz7yWtcArybwQby1WyPiaMvPg"
	dataJSON, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
	req.Header.Set("Authorization", "Bearer "+key)
	return client.Do(req)

}

func getNextInPath(name string, root *Node) *Node {
	for i := root.Nodes.Front(); i != nil; i = i.Next() {
		if (i.Value.(*Node)).Name == name {
			return (i.Value.(*Node))
		}
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
	for i, _ := range objs {
		node := &Node{}
		node.Entity = entity
		node.Name = (string((objs[i].(map[string]interface{}))["name"].(string)))
		node.ID, _ = strconv.Atoi((objs[i].(map[string]interface{}))["id"].(string))
		arr = append(arr, node)
	}
	return arr
}

func getChildren(curr *Node) []*Node {
	switch curr.Entity {
	case -1:
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/tenants", nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, TENANT)
	case TENANT:
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/tenants/"+curr.Name+"/sites",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SITE)
	case SITE:
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/sites/"+
				strconv.Itoa(curr.ID)+"/buildings",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, BLDG)
	case BLDG:
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/buildings/"+
				strconv.Itoa(curr.ID)+"/rooms",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, ROOM)
	case ROOM:
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/rooms/"+
				strconv.Itoa(curr.ID)+"/racks",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, RACK)
	case RACK:
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/racks/"+
				strconv.Itoa(curr.ID)+"/devices",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, DEVICE)
	case DEVICE:
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/devices/"+
				strconv.Itoa(curr.ID)+"/subdevices",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SUBDEV)
	case SUBDEV:
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/subdevices/"+
				strconv.Itoa(curr.ID)+"/all",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SUBDEV1)
	}
	return nil
}

func Populate(root **Node, dt int) {
	if dt != SUBDEV1 || root != nil {
		arr := getChildren(*root)
		for i := range arr {
			Populate(&arr[i], dt+1)
			(*root).Nodes.PushBack(arr[i])
		}
	}
}

func BuildTree() {
	(State.TreeHierarchy) = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	Populate(&State.TreeHierarchy, 0)
	/*DispAtLevel(&State.TreeHierarchy,
	*(strToStack(State.CurrPath)))*/
}

func CheckPath(root **Node, x, pstk *Stack) bool {
	if x.Len() == 0 {
		_, path := GetPathStrAtPtr(&State.TreeHierarchy, root, "")
		println(path)
		return true
	}

	p := x.Pop()

	//At Root
	if pstk.Len() == 0 && string(p.(string)) == ".." {
		//Pop until p != ".."
		for ; p != nil && string(p.(string)) == ".."; p = x.Pop() {
		}
		if p == nil {
			_, path := GetPathStrAtPtr(&State.TreeHierarchy, root, "/")
			println(path)
			return true
		}

		//Somewhere in tree
	} else if pstk.Len() > 0 && string(p.(string)) == ".." {
		prevNode := (pstk.Pop()).(*Node)
		return CheckPath(&prevNode, x, pstk)
	}

	nd := getNextInPath(string(p.(string)), *root)
	if nd == nil {
		return false
	}

	pstk.Push(*root)
	return CheckPath(&nd, x, pstk)

}

func GetPathStrAtPtr(root, curr **Node, path string) (bool, string) {
	if root == nil || *root == nil {
		return false, ""
	}

	if *root == *curr {
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

func InitState() {
	//State.sessionBuffer = *State.sessionBuffer.Init()
	State.CurrPath = "/"
	BuildTree()
}

func main() {

	InitState()
	//TEST CASE 1
	dispStk(StrToStack("/CED/BETA"))

	println("Now going to test the path finder")
	State.CurrPath = "/CED/BETA/A/R1/A09"
	query := "../../R2"
	x := StrToStack(State.CurrPath + "/" + query)
	val := CheckPath(&State.TreeHierarchy, x, New())

	if val == true {
		println("TEST CASE 1 Passed!")
	}

	//Test Case 2
	if CheckPath(&State.TreeHierarchy, StrToStack(query), New()) == false {
		println("TEST CASE 2 Passed!")
	}

	//Test Case 3
	if CheckPath(&State.TreeHierarchy, StrToStack("../../CED"), New()) == true {
		println("TEST CASE 3 Passed!")
	}

	//Test Case 4
	if CheckPath(&State.TreeHierarchy, StrToStack("../../CED/BETA"), New()) == true {
		println("TEST CASE 4 Passed!")
	}

	if CheckPath(&State.TreeHierarchy, StrToStack("../.."), New()) == true {
		println("TEST CASE 5 Passed!")
	}

	query = "../../../../../PERF"
	State.CurrPath = "/CED/BETA/A/R1/A08"
	if CheckPath(&State.TreeHierarchy, StrToStack(State.CurrPath+"/"+query), New()) == true {
		println("TEST CASE 6 Passed!")
	}

	//TEST CASE 7
	State.CurrPath = "/CED/BETA/A/R1/A09"
	query = "../../../../..//CED/BETA/A/R1/A09/chassis01"
	if CheckPath(&State.TreeHierarchy,
		StrToStack(State.CurrPath+"/"+query), New()) == true {
		println("TEST CASE 7 Passed!")
	}
}
