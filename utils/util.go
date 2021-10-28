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
	AC
	PWRPNL
	WALL
	CABINET
	AISLE
	TILE
	GROUP
	CORIDOR
	ROOMTMPL
	OBJTMPL
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

func ParamsParse(link *url.URL, objType int) map[string]interface{} {
	q, _ := url.ParseQuery(link.RawQuery)
	values := make(map[string]interface{})

	//Building Attribute query varies based on
	//object type
	for key, _ := range q {
		if objType < ROOMTMPL { //Non template objects
			switch key {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
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
	return values
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
	case AC:
		return "ac"
	case PWRPNL:
		return "panel"
	case WALL:
		return "wall"
	case ROOMTMPL:
		return "room_template"
	case OBJTMPL:
		return "obj_template"
	case CABINET:
		return "cabinet"
	case AISLE:
		return "aisle"
	case TILE:
		return "tile"
	case GROUP:
		return "group"
	case CORIDOR:
		return "corridor"
	default:
		return "INVALID"
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
	case "subdevice1":
		return SUBDEV1
	case "ac":
		return AC
	case "panel":
		return PWRPNL
	case "wall":
		return WALL
	case "room_template":
		return ROOMTMPL
	case "obj_template":
		return OBJTMPL
	case "cabinet":
		return CABINET
	case "aisle":
		return AISLE
	case "tile":
		return TILE
	case "group":
		return GROUP
	case "corridor":
		return CORIDOR
	default:
		return -1
	}
}

func GetParentOfEntityByInt(entity int) int {
	switch entity {
	case AC, PWRPNL, WALL, CABINET, TILE, AISLE, CORIDOR, GROUP:
		return ROOM
	case ROOMTMPL, OBJTMPL:
		return -1
	default:
		return entity - 1
	}
}

//func GetParentOfEntityByStr(entity string) int {
//	switch entity {
//	case AC,PWRPNL,WALL:
//		return "room"
//	default:
//		return
//	}
//}
