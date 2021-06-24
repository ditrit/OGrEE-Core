package utils

//Builds json messages and
//returns json response

import (
	"bytes"
	"container/list"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	treeHierarchy list.List
}

type Node struct {
	ID     int
	Entity int
	Name   string
	Nodes  list.List
}

var State ShellState

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
		resp, e := Send("GET",
			"https://ogree.chibois.net/api/user/tenants", nil)
		if e != nil {
			println("Error while getting children!")
			Exit()
		}
		println("Turning the list into nodes...")
		return makeNodeArrFromResp(resp, TENANT)
	case TENANT:
		println("Getting list of sites now...")
		resp, e := Send("GET",
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
		resp, e := Send("GET",
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
		resp, e := Send("GET",
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
		resp, e := Send("GET",
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
		resp, e := Send("GET",
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
		resp, e := Send("GET",
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
		resp, e := Send("GET",
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

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func ErrLog(message, funcname, details string, r *http.Request) {
	f, err := os.OpenFile("resources/debug.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	ip := r.RemoteAddr

	log.SetOutput(f)
	log.Println(message + " FOR FUNCTION: " + funcname)
	log.Println("FROM IP: " + ip)
	log.Println(details)
}

func ParamsParse(link *url.URL) []byte {
	q, _ := url.ParseQuery(link.RawQuery)
	values := make(map[string]string)
	for key, _ := range q {
		values[key] = q.Get(key)
	}

	//If you marshal it then
	//Unmarshal it, you can parse
	//the URL into a struct of choice!
	//Note that you would have to
	//Unmarshal twice to catch the
	//inner struct
	js, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}

	return js

	/*
		mydata := &models.Tenant{}
		json.Unmarshal(query, mydata)
		json.Unmarshal(query, &(mydata.Attributes))
	*/
	//return values
}

func JoinQueryGen(entity string) string {
	return "JOIN " + entity +
		"_attributes ON " + entity + "_attributes.id = " + entity + ".id"
}

func InitState() {
	State.sessionBuffer = *State.sessionBuffer.Init()
	State.treeHierarchy = *State.treeHierarchy.Init()
	//State.CurrPath = "/"
	State.CurrPath = "/CED/BETA/A"
}

func UpdateSessionState(ln *string) {
	State.sessionBuffer.PushBack(*ln)
}

//Function is an abstraction of a normal exit
func Exit() {
	writeHistoryOnExit(&State.sessionBuffer)
	os.Exit(0)
}

//Function helps with API Requests
func Send(method, URL string, data map[string]interface{}) (*http.Response,
	error) {
	client := &http.Client{}
	key := os.Getenv("apikey")
	dataJSON, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
	req.Header.Set("Authorization", "Bearer "+key)
	return client.Do(req)

}
