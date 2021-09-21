package utils

//Builds json messages and
//returns json response

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
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

func Connect() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
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

func EntityToString(entity int) string {
	switch entity {
	case TENANT:
		return "tenant"
	case SITE:
		return "site"
	case BLDG:
		return "building"
	case ROOM:
		return "room"
	case RACK:
		return "rack"
	case DEVICE:
		return "device"
	case SUBDEV:
		return "subdevice"
	default:
		return "subdevice1"
	}
}

func EntityStrToInt(entity string) int {
	switch entity {
	case "tenant":
		return TENANT
	case "site":
		return SITE
	case "building", "bldg":
		return BLDG
	case "room":
		return ROOM
	case "rack":
		return RACK
	case "device":
		return DEVICE
	case "subdevice":
		return SUBDEV
	default:
		return SUBDEV1
	}
}
