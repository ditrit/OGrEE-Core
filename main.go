package main

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

// Adding TAB completion support
//https://thoughtbot.com/blog/tab-completion-in-gnu-readline
import (
	"bufio"
	"bytes"
	"cli/models"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/joho/godotenv"
)

func BeginInterpreter(str *string) {
	lex := NewLexer(strings.NewReader(*str))
	e := yyParse(lex)
	println("\nReturn Code: ", e)
	return
}

func createCredentials() (string, string) {
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
		readline.
			Line("Error while creating credentials on server! Now exiting")
		os.Exit(-1)
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		readline.Line("Error: " + err.Error() + " Now Exiting")
		os.Exit(-1)
	}
	json.Unmarshal(bodyBytes, &tp)
	key = (tp["account"].(map[string]interface{}))["token"].(string)

	os.WriteFile("./.resources/.env",
		[]byte("user="+user+"\n"+"apikey="+key),
		0666)

	return user, key
}

func checkKeyIsValid(key string) bool {
	client := &http.Client{}

	req, _ := http.NewRequest("GET",
		"https://ogree.chibois.net/api/token/valid", nil)

	req.Header.Set("Authorization", "Bearer "+key)

	resp, e := client.Do(req)
	if e != nil || resp.StatusCode != 200 {
		readline.Line(e.Error())
		readline.Line("Status code" + strconv.Itoa(resp.StatusCode))
		return false
	}
	return true
}

func addHistory(rl *readline.Instance) {
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

func main() {
	var user, key string

	e := godotenv.Load(".resources/.env")
	if e != nil {
		user, key = createCredentials()
	} else {
		user = os.Getenv("user")
		key = os.Getenv("apikey")
		if !checkKeyIsValid(key) {
			readline.Line("Error while checking key. Now exiting")
			os.Exit(-1)
		}
	}

	user = (strings.Split(user, "@"))[0]
	rl, err := readline.New(user + "@" + "OGRE3D:$> ")
	if err != nil {
		panic(err)
	}

	defer rl.Close()
	models.InitState()
	addHistory(rl)
	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		BeginInterpreter(&line)
		models.UpdateSessionState(&line)
	}
}
