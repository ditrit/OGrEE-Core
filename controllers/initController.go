package controllers

//This file contains code associated with initialising the Shell

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	l "cli/logger"
	"cli/models"
	"cli/readline"
)

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
	e := models.ContactUnity("POST", State.UnityClientURL, req, State.Timeout)
	if e != nil {
		l.GetWarningLogger().Println("Note: Unity Client (" + State.UnityClientURL + ") Unreachable")
		fmt.Println("Note: Unity Client (" + State.UnityClientURL + ") Unreachable ")
		State.UnityClientAvail = false
	} else {
		fmt.Println("Unity Client is Reachable!")
		State.UnityClientAvail = true
	}
	//Set the filter attributes setting
	State.FilterDisplay = false
	//Set the Analyser setting to ON for now
	State.Analyser = true

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

func InitTimeout() {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "unityDeadline=") {
			if len(scanner.Text()[14:]) > 14 {
				timeArr := strings.Split(scanner.Text()[14:], " ")
				if len(timeArr) > 1 {
					timeDurationStr := timeArr[0]
					durationType := timeArr[1]

					timeLen, err := strconv.Atoi(timeDurationStr)
					if err != nil {
						l.GetWarningLogger().Println("Invalid value given for time duration. Resorting to default of 10")
						println("Invalid value given for time duration in env file. Resorting to default of 10")
						timeLen = 10
					}
					switch durationType {
					case "ns":
						State.Timeout = time.Nanosecond * time.Duration(timeLen)
					case "us":
						State.Timeout = time.Microsecond * time.Duration(timeLen)
					case "ms":
						State.Timeout = time.Millisecond * time.Duration(timeLen)
					case "s":
						State.Timeout = time.Second * time.Duration(timeLen)
					default:
						l.GetWarningLogger().Println("Invalid duration unit found. Resorting to default of ms")
						println("Invalid duration unit found in env file. Resorting to default of ms")
						State.Timeout = time.Millisecond * time.Duration(timeLen)
					}
				} else { // Error Case
					l.GetWarningLogger().Println("Invalid format given for unity deadline. Resorting to default time duration of 10 ms")
					println("Warning: Invalid duration unit found in env file. Resorting to default of ms")
					State.Timeout = time.Millisecond * time.Duration(10)
				}

			} else { //Error Case
				l.GetWarningLogger().Println("Unity deadline not found. Resorting to default time duration of 10 ms")
				println("Warning: Unity deadline not found in env file. Resorting to default of 10 ms")
				State.Timeout = time.Millisecond * time.Duration(10)
			}
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		l.GetErrorLogger().Println(err.Error())
	}
	l.GetWarningLogger().Println("Unity deadline not found. Resorting to default time duration of 10 ms")
	println("Warning: Unity deadline not found in env file. Resorting to default of 10 ms")
	State.Timeout = time.Millisecond * time.Duration(10)
	return
}

func InitKey() string {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "apikey=") {
			State.APIKEY = scanner.Text()[7:]
			return scanner.Text()[7:]
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		l.GetErrorLogger().Println(err.Error())
	}
	State.APIKEY = ""
	return ""
}

func GetEmail() string {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "user=") {
			return scanner.Text()[5:]
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		l.GetErrorLogger().Println(err.Error())
	}
	return ""
}

//Automatically assign Unity and API URLs
func GetURLs() {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Falling back to default URLs")
		l.GetInfoLogger().Println("Falling back to default URLs")
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
		l.GetInfoLogger().Println("Falling back to default API URL:" +
			"http://localhost:3001")
		State.APIURL = "http://localhost:3001"
	}

	if State.UnityClientURL == "" {
		fmt.Println("Falling back to default Unity URL:" +
			"http://localhost:5500")
		l.GetInfoLogger().Println("Falling back to default Unity URL:" +
			"http://localhost:5500")
		State.APIURL = "http://localhost:5500"
	}

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
		l.GetErrorLogger().Println(err)
		fmt.Println(err)
	}

	//Use default values
	//Set the array to all and exit
	//GROUP is the greatest value int enum type
	//So we use that for the cond guard
	if allDetected || len(res) == 0 {
		if len(res) == 0 && !allDetected {
			l.GetWarningLogger().Println(x + " key not found, going to use defaults")
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
				l.GetWarningLogger().Println("Specified template for " + entity + " not found")
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

func CreateCredentials() (string, string) {
	var tp map[string]interface{}
	var key string
	client := &http.Client{}

	user, _ := readline.Line("Please Enter desired user email: ")
	pass, _ := readline.Password("Please Enter desired password: ")

	buf, _ := json.Marshal(map[string]interface{}{"email": user,
		"password": pass})

	req, _ := http.NewRequest("POST",
		State.APIURL+"/api/user",
		bytes.NewBuffer(buf))

	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != http.StatusCreated {
		println("Error while creating credentials on server! Now exiting")
		l.GetErrorLogger().Println("Error while creating credentials on server! Now exiting")
		os.Exit(-1)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		readline.Line("Error: " + err.Error() + " Now Exiting")
		l.GetErrorLogger().Println("Error while trying to read server response: ", err)
		os.Exit(-1)
	}
	json.Unmarshal(bodyBytes, &tp)
	key = (tp["account"].(map[string]interface{}))["token"].(string)

	os.Mkdir(".resources", 0755)
	os.WriteFile("./.resources/.env",
		[]byte("user="+user+"\n"+"apikey="+key),
		0666)

	l.GetInfoLogger().Println("Credentials created")
	return user, key
}

func CheckKeyIsValid(key string) bool {
	client := &http.Client{}

	req, _ := http.NewRequest("GET",
		State.APIURL+"/api/token/valid", nil)

	req.Header.Set("Authorization", "Bearer "+key)

	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != 200 {
		//readline.Line(e.Error())
		if resp != nil {
			readline.Line("Status code" + strconv.Itoa(resp.StatusCode))
		} else {
			l.GetErrorLogger().Println("Unable to connect to API: ", State.APIURL)
			println("Unable to connect to API: ", State.APIURL)
		}

		return false
	}
	return true
}

func Login() (string, string) {
	var user, key string
	file, err := os.Open("./.resources/.env")
	if err != nil {
		l.GetInfoLogger().Println("Key not found, going to generate..")
		user, key = CreateCredentials()
	} else {
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords) // use scanwords
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "apikey=") {
				key = scanner.Text()[7:]
			}

			if strings.HasPrefix(scanner.Text(), "user=") {
				user = scanner.Text()[5:]
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println(err)
			l.GetErrorLogger().Println(err)
		}

		if !CheckKeyIsValid(key) {
			println("Error while checking key. Now exiting")
			l.GetErrorLogger().Println("Error while checking key. Now exiting")
			os.Exit(-1)
		}
	}
	defer file.Close()

	//println("Checking credentials...")
	//println(CheckKeyIsValid(key))

	user = (strings.Split(user, "@"))[0]
	l.GetInfoLogger().Println("Successfully Logged In")
	return user, key
}
