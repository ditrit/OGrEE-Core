package utils

//Builds json messages and
//returns json response

import (
	"container/list"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

type ShellState struct {
	currPath      string
	sessionBuffer list.List
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
}

func UpdateSessionState(ln *string) {
	State.sessionBuffer.PushBack(*ln)
}

//Function is an abstraction of a normal exit
func Exit() {
	writeHistoryOnExit(&State.sessionBuffer)
	os.Exit(0)
}
