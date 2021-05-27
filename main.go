package main

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program
import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/joho/godotenv"
)

type ShellState struct {
	currPath string
}

var State ShellState

func BeginInterpreter(str *string) {
	lex := NewLexer(strings.NewReader(*str))
	e := yyParse(lex)
	println("Return Code: ", e)
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
		[]byte("user="+user+"\n"+"apikey="+key+"\\0"),
		0666)

	return user, key
}

func main() {
	var user /*, pass */ string

	e := godotenv.Load(".resources/.env")
	if e != nil {
		user, _ = createCredentials()
	} else {
		user = os.Getenv("user")
		//pass = os.Getenv("apikey")
	}

	rl, err := readline.New(user + "@" + "OGRE3D:$> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		BeginInterpreter(&line)
	}

}
