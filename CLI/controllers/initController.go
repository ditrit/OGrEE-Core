package controllers

//This file contains code associated with initialising the Shell

import (
	"cli/config"
	l "cli/logger"
	"cli/models"
	"cli/readline"
	"cli/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func InitConfigFilePath(configPath string) {
	State.ConfigPath = configPath
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
func InitState(conf *config.Config) {

	State.ClipBoard = nil
	State.TreeHierarchy = &(Node{})
	(*(State.TreeHierarchy)).Entity = -1
	State.TreeHierarchy.PID = ""
	State.CurrPath = "/Physical"
	State.PrevPath = "/Physical"

	State.UnityClientAvail = false

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
	State.ObjsForUnity = SetObjsForUnity(conf.Updates)
	State.DrawableObjs = SetObjsForUnity(conf.Drawable)
	State.DrawableJsons = make(map[string]map[string]interface{}, 16)

	for i := SITE; i < GROUP+1; i++ {
		ent := EntityToString(i)
		State.DrawableJsons[ent] = SetDrawableTemplate(ent, conf.DrawableJson)
	}

	//Set Draw Threshold
	SetDrawThreshold(conf.DrawLimit)

	//Set customer / tenant name
	resp, e := models.Send("GET", State.APIURL+"/api/version", GetKey(), nil)
	parsed := ParseResponse(resp, e, "Get API Information request")
	if parsed != nil {
		if info, ok := LoadObjectFromInf(parsed["data"]); ok {
			if cInf, ok := info["Customer"]; ok {
				if customer, ok := cInf.(string); ok {
					State.Customer = customer
				}
			}

		}
	}

	if State.Customer == "" {
		if State.DebugLvl > NONE {
			println("Tenant Information not found!")
		}
		State.Customer = "UNKNOWN"
	}
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

func InitTimeout(duration string) {
	var timeLen int
	var durationType string
	_, err := fmt.Sscanf(duration, "%d%s", &timeLen, &durationType)
	if err != nil {
		msg := "Invalid unity timeout format. Resorting to default of 10ms"
		l.GetWarningLogger().Println(msg)
		if State.DebugLvl > 1 {
			println(msg)
		}
		State.Timeout = time.Millisecond * time.Duration(10)
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
		msg := "Invalid duration unit found. Resorting to default of ms"
		l.GetWarningLogger().Println(msg)
		if State.DebugLvl > 1 {
			println(msg)
		}
		State.Timeout = time.Millisecond * time.Duration(timeLen)
	}
}

func InitUser(user User) {

}

func InitKey(apiKey string) {
	if apiKey != "" {
		State.APIKEY = apiKey
	} else {
		fmt.Println("Error: No API Key Found")
		if State.DebugLvl > 0 {
			l.GetErrorLogger().Println(
				"No API Key provided in env file nor as argument")
		}
	}
}

func InitURLs(apiURL string, unityURL string) {
	apiURL = strings.TrimRight(apiURL, "/")
	_, err := url.ParseRequestURI(apiURL)
	if err != nil {
		msg := "apiURL is not valid!\n"
		msg += "Falling back to default API URL: http://localhost:3001"
		fmt.Println(msg)
		l.GetInfoLogger().Println(msg)
		State.APIURL = "http://localhost:3001"
	} else {
		State.APIURL = apiURL
	}
	State.UnityClientURL = unityURL
	if State.UnityClientURL == "" {
		msg := "Falling back to defaul Unity URL: localhost:5500"
		fmt.Println(msg)
		l.GetInfoLogger().Println(msg)
		State.UnityClientURL = "localhost:5500"
	}
}

// Helper for InitState will insert objs
// and init DrawThreshold
func SetObjsForUnity(objs []string) []int {
	res := []int{}
	allDetected := false
	for _, obj := range objs {
		obj = strings.ToLower(obj)
		if val := EntityStrToInt(obj); val != -1 {
			res = append(res, val)
		} else if obj == "all" {
			//Exit the loop and use default code @ end of function
			allDetected = true
			break
		}
	}
	//Set the array to all and exit
	//GROUP is the greatest value int enum type
	//So we use that for the cond guard
	if allDetected {
		for idx := 0; idx < GROUP+1; idx++ {
			res = append(res, idx)
		}
		res = append(res, DOMAIN)
	}
	return res
}

func SetDrawThreshold(limit int) {
	//Set Draw Threshold
	if limit < 0 {
		if State.DebugLvl > 0 {
			println("Setting Draw Limit to default")
		}
		State.DrawThreshold = 50 //50 is default value
	} else {
		State.DrawThreshold = limit
	}
}

func SetDrawableTemplate(entity string, DrawableJson map[string]string) map[string]interface{} {
	if objStr, ok := DrawableJson[entity]; ok && objStr != "" {
		objStr = strings.Trim(objStr, "'\"")
		//Now retrieve file
		ans := map[string]interface{}{}
		if !filepath.IsAbs(objStr) {
			objStr = utils.ExeDir() + "/" + objStr
		}
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

func Login(user string) (*User, string, error) {
	var err error
	if user == "" {
		user, err = readline.Line("User: ")
		if err != nil {
			return nil, "", fmt.Errorf("readline error : %s", err.Error())
		}
	}
	pass, err := readline.Password("Password: ")
	if err != nil {
		return nil, "", err
	}
	data := map[string]any{"email": user, "password": string(pass)}
	rawResp, err := models.Send("POST", State.APIURL+"/api/login", "", data)
	if err != nil {
		return nil, "", fmt.Errorf("error sending login request : %s", err.Error())
	}
	bodyBytes, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error reading answer from API : %s", err.Error())
	}
	var resp map[string]any
	if err = json.Unmarshal(bodyBytes, &resp); err != nil {
		return nil, "", fmt.Errorf("error parsing response : %s", err.Error())
	}
	status, ok := resp["status"].(bool)
	if !ok {
		return nil, "", fmt.Errorf("invalid response from API")
	}
	if !status {
		return nil, "", fmt.Errorf(resp["message"].(string))
	}
	account, accountOk := (resp["account"].(map[string]interface{}))
	token, tokenOk := account["token"].(string)
	userID, userIDOk := account["_id"].(string)
	if !accountOk || !tokenOk || !userIDOk {
		return nil, "", fmt.Errorf("invalid response from API")
	}
	return &User{user, userID}, token, nil
}
