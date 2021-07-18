package controllers

import (
	"bufio"
	"cli/models"
	"container/list"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"cli/readline"
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

func AddHistory(rl *readline.Instance) {
	readFile, err := os.Open(".resources/.history")

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileTextLines []string

	for fileScanner.Scan() {
		fileTextLines = append(fileTextLines, fileScanner.Text())
	}

	readFile.Close()

	for _, eachline := range fileTextLines {
		rl.SaveHistory(strings.TrimSuffix(eachline, "\n"))
	}

	return
}

func InitState() {
	//State.sessionBuffer = *State.sessionBuffer.Init()
	State.CurrPath = "/"
	BuildTree()
}

func InitStateDummy() {
	State.CurrPath = "/"
	(State.TreeHierarchy) = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	PopulateDummy(&State.TreeHierarchy, 0)
}

func UpdateSessionState(ln *string) {
	State.sessionBuffer.PushBack(*ln)
}

//Function is an abstraction of a normal exit
func Exit() {
	//writeHistoryOnExit(&State.sessionBuffer)
	//runtime.Goexit()
	os.Exit(0)
}

//Populate hierarchy into B Tree like
//structure
func Populate(root **Node, dt int) {
	if dt != SUBDEV1 || root != nil {
		arr := getChildren(*root)
		for i := range arr {
			Populate(&arr[i], dt+1)
			(*root).Nodes.PushBack(arr[i])
		}
	}
}

func PopulateDummy(root **Node, dt int) {
	if dt != SUBDEV1 || root != nil {
		arr := GetChildrenDummy(*root)
		for i := range arr {
			PopulateDummy(&arr[i], dt+1)
			(*root).Nodes.PushBack(arr[i])
		}
	}
}

func GetChildrenDummy(curr *Node) []*Node {
	switch curr.Entity {
	case -1:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/tenants", nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, TENANT)
	case TENANT:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/tenants/"+curr.Name+"/sites",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SITE)
	case SITE:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/sites/"+
				strconv.Itoa(curr.ID)+"/buildings",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, BLDG)
	case BLDG:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/buildings/"+
				strconv.Itoa(curr.ID)+"/rooms",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, ROOM)
	case ROOM:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/rooms/"+
				strconv.Itoa(curr.ID)+"/racks",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, RACK)
	case RACK:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/racks/"+
				strconv.Itoa(curr.ID)+"/devices",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, DEVICE)
	case DEVICE:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/devices/"+
				strconv.Itoa(curr.ID)+"/subdevices",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SUBDEV)
	}
	return nil
}

func getChildren(curr *Node) []*Node {
	switch curr.Entity {
	case -1:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/tenants", nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, TENANT)
	case TENANT:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/tenants/"+curr.Name+"/sites",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SITE)
	case SITE:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/sites/"+
				strconv.Itoa(curr.ID)+"/buildings",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, BLDG)
	case BLDG:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/buildings/"+
				strconv.Itoa(curr.ID)+"/rooms",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, ROOM)
	case ROOM:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/rooms/"+
				strconv.Itoa(curr.ID)+"/racks",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, RACK)
	case RACK:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/racks/"+
				strconv.Itoa(curr.ID)+"/devices",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, DEVICE)
	case DEVICE:
		resp, e := models.Send("GET",
			"https://ogree.chibois.net/api/user/devices/"+
				strconv.Itoa(curr.ID)+"/subdevices",
			nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		return makeNodeArrFromResp(resp, SUBDEV)
	case SUBDEV:
		resp, e := models.Send("GET",
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

func BuildTree() {
	(State.TreeHierarchy) = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	Populate(&State.TreeHierarchy, 0)
	/*DispAtLevel(&State.TreeHierarchy,
	*(strToStack(State.CurrPath)))*/
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

func StrToStackTAB(x string) *Stack {
	stk := Stack{}
	sarr := strings.Split(x, "/")
	for i := len(sarr) - 1; i >= 0; i-- {
		//println("PUSHING TO STACK: ", sarr[i])
		/*if sarr[i] == ".." {
			stk.Pop()
		}*/
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

func DispAtLevel(root **Node, x Stack) []string {
	if x.Len() > 0 {
		name := x.Peek()
		node := getNextInPath(name.(string), *root)
		if node == nil {
			println("Name doesn't exist! ", string(name.(string)))
			return nil
		}
		x.Pop()
		return DispAtLevel(&node, x)
	} else {
		var items = make([]string, 0)
		var nm string
		println("This is what we got:")
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

func CheckPath(root **Node, x, pstk *Stack) (bool, string) {
	if x.Len() == 0 {
		_, path := GetPathStrAtPtr(&State.TreeHierarchy, root, "")
		//println(path)
		return true, path
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
			return true, path
		}

		//Somewhere in tree
	} else if pstk.Len() > 0 && string(p.(string)) == ".." {
		prevNode := (pstk.Pop()).(*Node)
		return CheckPath(&prevNode, x, pstk)
	}

	nd := getNextInPath(string(p.(string)), *root)
	if nd == nil {
		return false, ""
	}

	pstk.Push(*root)
	return CheckPath(&nd, x, pstk)

}

/*func CheckPathExists(root **Node, x *Stack) bool {
	if x.Len() > 0 {
		nd := getNextInPath((x.Pop()).(string), *root)
		if nd == nil {
			return false
		}
		return CheckPathExists(&nd, x)
	}
	return true
}
*/
