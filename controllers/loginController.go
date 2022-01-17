package controllers

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

	"cli/readline"
)

func GetKey() string {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "apikey=") {
			return scanner.Text()[7:]
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
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
	}
	return ""
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
		"https://ogree.chibois.net/api/user",
		bytes.NewBuffer(buf))

	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != http.StatusCreated {
		println("Error while creating credentials on server! Now exiting")
		ErrorLogger.Println("Error while creating credentials on server! Now exiting")
		os.Exit(-1)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		readline.Line("Error: " + err.Error() + " Now Exiting")
		ErrorLogger.Println("Error while trying to read server response: ", err)
		os.Exit(-1)
	}
	json.Unmarshal(bodyBytes, &tp)
	key = (tp["account"].(map[string]interface{}))["token"].(string)

	os.Mkdir(".resources", 0755)
	os.WriteFile("./.resources/.env",
		[]byte("user="+user+"\n"+"apikey="+key),
		0666)

	InfoLogger.Println("Credentials created")
	return user, key
}

func CheckKeyIsValid(key string) bool {
	client := &http.Client{}

	req, _ := http.NewRequest("GET",
		"https://ogree.chibois.net/api/token/valid", nil)

	req.Header.Set("Authorization", "Bearer "+key)

	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != 200 {
		//readline.Line(e.Error())
		if resp != nil {
		readline.Line("Status code" + strconv.Itoa(resp.StatusCode))
		} else {
			println("Unable to connect to API")
		}

		return false
	}
	return true
}

func Login() (string, string) {
	var user, key string
	file, err := os.Open("./.resources/.env")
	if err != nil {
		InfoLogger.Println("Key not found, going generate..")
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
			ErrorLogger.Println(err)
		}

		if !CheckKeyIsValid(key) {
			println("Error while checking key. Now exiting")
			ErrorLogger.Println("Error while checking key. Now exiting")
			os.Exit(-1)
		}
	}
	defer file.Close()

	//println("Checking credentials...")
	//println(CheckKeyIsValid(key))

	user = (strings.Split(user, "@"))[0]
	InfoLogger.Println("Successfully Logged In")
	return user, key
}
