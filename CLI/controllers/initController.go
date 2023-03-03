package controllers

//This file contains code associated with initialising the Shell

import (
	"bufio"
	l "cli/logger"
	"cli/models"
	"cli/readline"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func InitEnvFilePath(envPath string) {
	State.EnvFilePath = envPath
}

func InitHistoryFilePath(histPath string) {
	State.HistoryFilePath = histPath
}

func InitDebugLevel(verbose string) {
	var ok bool
	State.DebugLvl, ok = map[string]int{
		"NONE":    NONE,
		"ERROR":   ERROR,
		"WARNING": WARNING,
		"INFO":    INFO,
		"DEBUG":   DEBUG,
	}[verbose]
	if !ok {
		println("Invalid Logging Mode detected. Resorting to default: ERROR")
		State.DebugLvl = 1
	}
}

// Intialise the ShellState
func InitState(analyser string, env map[string]string) {

	State.ClipBoard = nil
	State.TreeHierarchy = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	State.TreeHierarchy.PID = ""
	State.CurrPath = "/Physical"
	State.PrevPath = "/Physical"
	State.LineNumber = 0

	State.UnityClientAvail = false

	//Set the filter attributes setting
	State.FilterDisplay = false

	//Set the Analyser setting to ON for now
	State.Analyser, _ = strconv.ParseBool(analyser)

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

	bTemplate := &Node{}
	bTemplate.ID = "3"
	bTemplate.PID = "0"
	bTemplate.Entity = -1
	bTemplate.Name = "BldgTemplates"
	bTemplate.Path = "/Logical"
	SearchAndInsert(&State.TreeHierarchy, bTemplate, "/Logical")

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
	State.ObjsForUnity = SetObjsForUnity("updates", env)
	State.DrawableObjs = SetObjsForUnity("drawable", env)
	State.DrawableJsons = make(map[string]map[string]interface{}, 16)

	for i := TENANT; i < GROUP+1; i++ {
		ent := EntityToString(i)
		State.DrawableJsons[ent] = SetDrawableTemplate(ent, env)
	}

	//Set Draw Threshold
	SetDrawThreshold(env)
}

// It is useful to have the state to hold
// a pointer to our readline terminal
func SetStateReadline(rl *readline.Instance) {
	State.Terminal = &rl
}

// Startup the go routine for listening
func InitUnityCom(rl *readline.Instance, addr string) {
	errConnect := models.ConnectToUnity(addr, State.Timeout)
	if errConnect != nil {
		if State.DebugLvl > ERROR {
			println(errConnect.Error())
		}
		return
	}
	State.UnityClientAvail = true

	data := map[string]interface{}{"api_url": State.APIURL, "api_token": GetKey()}
	req := map[string]interface{}{"type": "login", "data": data}
	errLogin := models.ContactUnity(req, State.DebugLvl)
	if errLogin != nil {
		println(errLogin.Error())
		return
	}
	fmt.Println("Unity Client is Reachable!")
	go models.ReceiveLoop(rl, addr, &State.UnityClientAvail)
}

func InitTimeout(env map[string]string) {
	if duration, ok := env["unityTimeout"]; ok && duration != "" {
		var timeLen int
		var durationType string
		fmt.Sscanf(duration, "%d%s", &timeLen, &durationType)
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
			if State.DebugLvl > 1 {
				println("Invalid duration unit found in env file. Resorting to default of ms")
			}

			State.Timeout = time.Millisecond * time.Duration(timeLen)
		}
		return
	}

	if State.DebugLvl > 1 {
		l.GetWarningLogger().Println("Unity deadline not found. Resorting to default time duration of 10 ms")
		println("Warning: Unity deadline not found in env file. Resorting to default of 10 ms")
	}

	State.Timeout = time.Millisecond * time.Duration(10)
	return
}

func InitKey(apiKey string, env map[string]string) string {
	if apiKey != "" {
		State.APIKEY = apiKey
		return State.APIKEY
	}
	envApiKey, ok := env["apiKey"]
	if ok {
		State.APIKEY = envApiKey
		return State.APIKEY
	}
	fmt.Println("Error: No API Key Found")
	if State.DebugLvl > 0 {
		l.GetErrorLogger().Println(
			"No API Key provided in env file nor as argument")
	}

	State.APIKEY = ""
	return ""
}

// Invoked on 'lsog' command
func GetEmail() string {
	file, err := os.Open("./.env")
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
		if State.DebugLvl > 0 {
			fmt.Println(err)
		}

		l.GetErrorLogger().Println(err.Error())
	}
	return ""
}

// Automatically assign Unity and API URLs
func GetURLs(apiURL string, unityURL string, env map[string]string) {
	if apiURL != "" {
		State.APIURL = apiURL
	}

	if unityURL != "" {
		State.UnityClientURL = unityURL
	}

	if State.UnityClientURL == "" {
		if envUnityURL, ok := env["unityURL"]; ok {
			State.UnityClientURL = envUnityURL
		}
	}

	if State.APIURL == "" {
		if envApiURL, ok := env["apiURL"]; ok {
			// if present, remove the last / to avoid path issues in ls command
			envApiURL = strings.TrimRight(envApiURL, "/")

			// check if URL is valid
			_, err := url.ParseRequestURI(envApiURL)
			if err != nil {
				println("apiURL is not valid! ")
			} else {
				State.APIURL = envApiURL
			}
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

// Helper for InitState will insert objs
// and init DrawThreshold
func SetObjsForUnity(x string, env map[string]string) []int {
	res := []int{}
	allDetected := false

	if objStr, ok := env[x]; ok && objStr != "" {
		//ObjStr is equal to everything after 'updates='
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

	//Use default values
	//Set the array to all and exit
	//GROUP is the greatest value int enum type
	//So we use that for the cond guard
	if allDetected || len(res) == 0 {
		if len(res) == 0 && !allDetected {
			l.GetWarningLogger().Println(x + " key not found, going to use defaults")
			if State.DebugLvl > 1 {
				println(x + " key not found, going to use defaults")
			}

		}
		for idx := 0; idx < GROUP+1; idx++ {
			res = append(res, idx)
		}
	}
	return res
}

func SetDrawThreshold(env map[string]string) {
	//Set Draw Threshold
	limit, e := strconv.Atoi(env["drawLimit"])
	if e != nil || limit < 0 {
		if State.DebugLvl > 0 {
			println("Setting Draw Limit to default")
		}
		State.DrawThreshold = 50 //50 is default value
	} else {
		State.DrawThreshold = limit
	}
}

func SetDrawableTemplate(entity string, env map[string]string) map[string]interface{} {
	templateKey := entity + "DrawableJson"
	if objStr, ok := env[templateKey]; ok && objStr != "" {
		objStr = strings.Trim(objStr, "'\"")
		//Now retrieve file
		ans := map[string]interface{}{}
		f, e := os.ReadFile(objStr)
		if e == nil {
			json.Unmarshal(f, &ans)
			return ans
		}
	}
	l.GetWarningLogger().Println("Specified template for " + entity + " not found")
	if State.DebugLvl > 1 {
		println("Specified template for " + entity +
			" not found, resorting to defaults")
	}
	return nil
}

func CreateCredentials() (string, string) {
	var tp map[string]interface{}

	user, _ := readline.Line("Please Enter desired user email: ")
	pass, _ := readline.Password("Please Enter desired password: ")
	data := map[string]interface{}{"email": user, "password": string(pass)}

	resp, e := models.Send("POST", State.APIURL+"/api", "", data)
	tp = ParseResponse(resp, e, "Create credentials")
	if tp == nil {
		println(e.Error())
		os.Exit(-1)
	}

	if !tp["status"].(bool) {
		errMessage := "Error while creating credentials : " + tp["message"].(string)
		if State.DebugLvl > 0 {
			println(errMessage)
		}
		l.GetErrorLogger().Println(errMessage)
		os.Exit(-1)
	}

	token := (tp["account"].(map[string]interface{}))["token"].(string)

	envMap, err := godotenv.Read(State.EnvFilePath)
	if err != nil {
		panic(err)
	}
	envMap["user"] = user
	envMap["apiKey"] = token
	godotenv.Write(envMap, State.EnvFilePath)

	l.GetInfoLogger().Println("Credentials created")
	return user, token
}

func CheckKeyIsValid(key string) bool {
	resp, err := models.Send("GET", State.APIURL+"/api/token/valid", key, nil)
	if err != nil {
		if State.DebugLvl > 0 {
			l.GetErrorLogger().Println("Unable to connect to API: ", State.APIURL)
			l.GetErrorLogger().Println(err.Error())
			println(err.Error())
		}
		return false
	}

	if resp.StatusCode != 200 {
		readline.Line("HTTP Response Status code: " +
			strconv.Itoa(resp.StatusCode))
		if State.DebugLvl > NONE {
			x := ParseResponse(resp, err, " Read API Response message")
			if x != nil {
				println("[API] " + x["message"].(string))
			} else {
				println("Was not able to read API Response message")
			}
		}

		return false
	}

	return true
}

func Login(env map[string]string) (string, string) {
	user, userOk := env["user"]
	key, keyOk := env["apiKey"]

	if !userOk || !keyOk || (userOk && user == "") || (keyOk && key == "") {
		l.GetInfoLogger().Println("Key not found, going to generate..")
		user, key = CreateCredentials()
	}

	if !CheckKeyIsValid(key) {
		if State.DebugLvl > 0 {
			println("Error while checking key. Now exiting")
		}

		l.GetErrorLogger().Println("Error while checking key. Now exiting")
		os.Exit(-1)
	}

	//println("Checking credentials...")
	//println(CheckKeyIsValid(key))

	user = (strings.Split(user, "@"))[0]
	l.GetInfoLogger().Println("Successfully Logged In")
	return user, key
}
