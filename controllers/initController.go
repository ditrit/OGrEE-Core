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

	l "cli/logger"
	"cli/models"
	"cli/readline"
)

//Startup the go routine for listening
func TriggerListen(rl *readline.Instance) {
	go models.ListenForUnity(rl)
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
