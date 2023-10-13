package controllers

//This file has a collection of utility functions used in the
//controller package
//And const definitions used throughout the controllers package
import (
	"cli/models"
	"encoding/json"
	"fmt"
	"path"
	"strings"
)

const (
	SITE = iota
	BLDG
	ROOM
	RACK
	DEVICE
	AC
	PWRPNL
	CABINET
	CORRIDOR
	SENSOR
	ROOMTMPL
	OBJTMPL
	BLDGTMPL
	GROUP
	STRAY_DEV
	STRAYSENSOR
	DOMAIN
	TAG
)

// Debug Level Declaration
const (
	NONE = iota
	ERROR
	WARNING
	INFO
	DEBUG
)

const RACKUNIT = .04445 //meter

// displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func DispWithAttrs(objs []interface{}, attrs []string) {
	for _, objInf := range objs {
		if obj, ok := objInf.(map[string]interface{}); ok {
			for _, a := range attrs {
				//Check if attr is in object
				if ok, nested := AttrIsInObj(obj, a); ok {
					if nested {
						fmt.Print("\t"+a+":",
							obj["attributes"].(map[string]interface{})[a])
					} else {
						fmt.Print("\t"+a+":", obj[a])
					}
				} else {
					fmt.Print("\t" + a + ": NULL")
				}
			}
			fmt.Printf("\tName:%s\n", obj["name"].(string))
		}
	}
}

// Returns true/false if exists and true/false if attr
// is in "attributes" maps
func AttrIsInObj(obj map[string]interface{}, attr string) (bool, bool) {
	if _, ok := obj[attr]; ok {
		return ok, false
	}

	if hasAttr, _ := AttrIsInObj(obj, "attributes"); hasAttr == true {
		if objAttributes, ok := obj["attributes"].(map[string]interface{}); ok {
			_, ok := objAttributes[attr]
			return ok, true
		}
	}

	return false, false
}

func EntityToString(entity int) string {
	switch entity {
	case DOMAIN:
		return "domain"
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
	case STRAY_DEV:
		return "stray_device"
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
	case TAG:
		return "tag"
	default:
		return "INVALID"
	}
}

func EntityStrToInt(entity string) int {
	switch entity {
	case "domain":
		return DOMAIN
	case "site", "si":
		return SITE
	case "building", "bldg", "bd":
		return BLDG
	case "room", "ro":
		return ROOM
	case "rack", "rk":
		return RACK
	case "device", "dv":
		return DEVICE
	case "ac":
		return AC
	case "panel", "pn":
		return PWRPNL
	case "stray_device":
		return STRAY_DEV
	case "room_template":
		return ROOMTMPL
	case "obj_template":
		return OBJTMPL
	case "bldg_template":
		return BLDGTMPL
	case "cabinet", "cb":
		return CABINET
	case "group", "gr":
		return GROUP
	case "corridor", "co":
		return CORRIDOR
	case "sensor", "sr":
		return SENSOR
	case "tag":
		return TAG
	default:
		return -1
	}
}

func GetParentOfEntity(ent int) int {
	switch ent {
	case SITE:
		return ent - 1
	case BLDG:
		return ent - 1
	case ROOM:
		return ent - 1
	case RACK:
		return ent - 1
	case DEVICE:
		return ent - 1
	case AC:
		return ROOM
	case PWRPNL:
		return ROOM
	case ROOMTMPL:
		return -1
	case BLDGTMPL:
		return -1
	case OBJTMPL:
		return -1
	case CABINET:
		return ROOM
	case GROUP:
		return -1
	case CORRIDOR:
		return ROOM
	case SENSOR:
		return -2
	default:
		return -3
	}
}

func RequestAPI(method string, endpoint string, body map[string]any, expectedStatus int) (*Response, error) {
	URL := State.APIURL + endpoint
	httpResponse, err := models.Send(method, URL, GetKey(), body)
	if err != nil {
		return nil, err
	}
	response, err := ParseResponseClean(httpResponse)
	if err != nil {
		return nil, fmt.Errorf("on %s %s : %s", method, endpoint, err.Error())
	}
	if response.status != expectedStatus {
		msg := ""
		if State.DebugLvl >= DEBUG {
			msg += fmt.Sprintf("%s %s\n", method, URL)
		}
		msg += fmt.Sprintf("[Response From API] %s", response.message)
		errorsAny, ok := response.body["errors"]
		if ok {
			errorsList := errorsAny.([]any)
			for _, err := range errorsList {
				msg += "\n    " + err.(string)
			}
		}
		return response, fmt.Errorf(msg)
	}
	return response, nil
}

func TranslatePath(p string) string {
	if p == "" {
		p = "."
	}
	if p == "_" {
		return "_"
	}
	if p == "-" {
		return State.PrevPath
	}
	var output_words []string
	if p[0] != '/' {
		output_words = strings.Split(State.CurrPath, "/")[1:]
		if len(output_words) == 1 && output_words[0] == "" {
			output_words = output_words[0:0]
		}
	} else {
		p = p[1:]
	}
	input_words := strings.Split(p, "/")
	for _, word := range input_words {
		if word == "." {
			continue
		} else if word == ".." {
			if len(output_words) > 0 {
				output_words = output_words[:len(output_words)-1]
			}
		} else {
			output_words = append(output_words, word)
		}
	}
	if len(output_words) > 0 {
		if output_words[0] == "P" {
			output_words[0] = "Physical"
		} else if output_words[0] == "L" {
			output_words[0] = "Logical"
		} else if output_words[0] == "O" {
			output_words[0] = "Organisation"
		}
	}
	return path.Clean("/" + strings.Join(output_words, "/"))
}

type ErrorWithInternalError struct {
	UserError     error
	InternalError error
}

func (err ErrorWithInternalError) Error() string {
	return err.UserError.Error() + " caused by " + err.InternalError.Error()
}
