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
	"strings"
	"time"
)

var BuildHash string
var BuildTree string
var BuildTime string
var GitCommitDate string

const (
	SITE = iota
	BLDG
	ROOM
	RACK
	DEVICE
	AC
	CABINET
	CORRIDOR
	PWRPNL
	SENSOR
	GROUP
	ROOMTMPL
	OBJTMPL
	BLDGTMPL
	STRAYDEV
	DOMAIN
	STRAYSENSOR
)

type RequestFilters struct {
	FieldsToShow []string `schema:"fieldOnly"`
	StartDate    string   `schema:"startDate"`
	EndDate      string   `schema:"endDate"`
	Limit        string   `schema:"limit"`
}

func GetBuildDate() string {
	return BuildTime
}

func GetCommitDate() string {
	return GitCommitDate
}

func GetBuildHash() string {
	return BuildHash
}

func GetBuildTree() string {
	return BuildTree
}

func Connect() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"status": status, "message": message}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	json.NewEncoder(w).Encode(data)
	w.Header().Add("Content-Type", "application/json")
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

func ParamsParse(link *url.URL, objType int) map[string]interface{} {
	q, _ := url.ParseQuery(link.RawQuery)
	values := make(map[string]interface{})

	//Building Attribute query varies based on
	//object type
	for key, _ := range q {
		if key != "fieldOnly" && key != "startDate" && key != "endDate" {
			if objType != ROOMTMPL && objType != OBJTMPL &&
				objType != BLDGTMPL { //Non template objects
				switch key {
				case "id", "name", "category", "parentID",
					"description", "domain", "parentid", "parentId",
					"hierarchyName", "createdDate", "lastUpdated":
					values[key] = q.Get(key)
				default:
					values["attributes."+key] = q.Get(key)
				}
			} else { //Template objects
				//Not sure how to search FBX TEMPLATES
				//For now it is disabled
				switch key {
				case "description", "slug", "category", "sizeWDHmm", "fbxModel":
					values[key] = q.Get(key)
				default:
					values["attributes."+key] = q.Get(key)
				}
			}
		}
	}
	return values
}

func EntityToString(entity int) string {
	switch entity {
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
	case AC:
		return "ac"
	case PWRPNL:
		return "panel"
	case DOMAIN:
		return "domain"
	case STRAYDEV:
		return "stray_device"
	case STRAYSENSOR:
		return "stray_sensor"
	case ROOMTMPL:
		return "room_template"
	case OBJTMPL:
		return "obj_template"
	case BLDGTMPL:
		return "bldg_template"
	case CABINET:
		return "cabinet"
	case GROUP:
		return "group"
	case CORRIDOR:
		return "corridor"
	case SENSOR:
		return "sensor"
	default:
		return "INVALID"
	}
}

func EntityStrToInt(entity string) int {
	switch entity {
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
	case "ac":
		return AC
	case "panel":
		return PWRPNL
	case "domain":
		return DOMAIN
	case "stray_device":
		return STRAYDEV
	case "stray_sensor":
		return STRAYSENSOR
	case "room_template":
		return ROOMTMPL
	case "obj_template":
		return OBJTMPL
	case "bldg_template":
		return BLDGTMPL
	case "cabinet":
		return CABINET
	case "group":
		return GROUP
	case "corridor":
		return CORRIDOR
	case "sensor":
		return SENSOR
	default:
		return -1
	}
}

func HierachyNameToEntity(name string) []int {
	resp := []int{STRAYDEV} // it can always be a stray
	switch strings.Count(name, ".") {
	case 0:
		resp = append(resp, SITE)
	case 1:
		resp = append(resp, BLDG)
	case 2:
		resp = append(resp, ROOM)
	case 3:
		resp = append(resp, RACK, GROUP, AC, CORRIDOR, PWRPNL, CABINET)
	case 4:
		resp = append(resp, DEVICE, GROUP)
	default:
		resp = append(resp, DEVICE)
	}

	return resp

}

func GetParentOfEntityByInt(entity int) int {
	switch entity {
	case AC, PWRPNL, CABINET, CORRIDOR:
		return ROOM
	case SENSOR:
		return -2
	case ROOMTMPL, OBJTMPL, BLDGTMPL, GROUP, STRAYDEV:
		return -1
	default:
		return entity - 1
	}
}

// Helper functions
func StrSliceContains(slice []string, elem string) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}
	return false
}
