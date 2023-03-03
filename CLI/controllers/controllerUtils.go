package controllers

//This file has a collection of utility functions used in the
//controller package
//And const definitions used throughout the controllers package
import (
	"encoding/json"
	"fmt"
)

const (
	TENANT = iota
	SITE
	BLDG
	ROOM
	RACK
	DEVICE
	AC
	PWRPNL
	CABINET
	CORIDOR
	SENSOR
	ROOMTMPL
	OBJTMPL
	BLDGTMPL
	GROUP
	STRAY_DEV
	STRAYSENSOR
)

// Debug Level Declaration
const (
	NONE = iota
	ERROR
	WARNING
	INFO
	DEBUG
)

// Error Message Const
// TODO: Replace Const with Err Msg/Reporting Func
// that distinguishes API & CLI Errors
const APIErrorPrefix = "[Response From API] "
const RACKUNIT = .04445 //meter

// Display contents of []map[string]inf array
func DispMapArr(x []map[string]interface{}) {
	for idx := range x {
		println()
		println()
		println("OBJECT: ", idx)
		displayObject(x[idx])
		println()
	}
}

// displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func DispWithAttrs(objs *[]interface{}, attrs *[]string) {
	for _, objInf := range *objs {
		if obj, ok := objInf.(map[string]interface{}); ok {
			for _, a := range *attrs {
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

func DispfWithAttrs(formatx string, objs *[]interface{}, attrs *[]string) {
	//Convert user input format to workable format
	//input for Printf
	var format string
	formatx = `"` + formatx + `"`
	_, e := fmt.Sscanf(formatx, "%q", &format)
	if e != nil {
		println(e.Error())
		return
	}

	for _, objInf := range *objs {
		if obj, ok := objInf.(map[string]interface{}); ok {
			var argument []interface{}
			//var printer string
			for _, a := range *attrs {
				//Check if attr is in object
				if ok, nested := AttrIsInObj(obj, a); ok {
					if nested {
						argument = append(argument, obj["attributes"].(map[string]interface{})[a])
					} else {
						argument = append(argument, obj[a])
					}
				} else {
					argument = append(argument, "NULL")
				}
			}
			fmt.Printf(format, argument...)
			fmt.Printf("\n")
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

// Provides a mapping for stray
// and normal objects
func MapStrayString(x string) string {
	if x == "device" {
		return "stray-device"
	}
	if x == "sensor" {
		return "stray-sensor"
	}

	if x == "stray-device" {
		return "device"
	}
	if x == "stray-sensor" {
		return "sensor"
	}
	return "INVALID-MAP"
}

func MapStrayInt(x int) int {
	if x == DEVICE {
		return STRAY_DEV
	}
	if x == SENSOR {
		return STRAYSENSOR
	}

	if x == STRAY_DEV {
		return DEVICE
	}
	if x == STRAYSENSOR {
		return SENSOR
	}
	return -1
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
	case CORIDOR:
		return "corridor"
	case SENSOR:
		return "sensor"
	default:
		return "INVALID"
	}
}

func EntityStrToInt(entity string) int {
	switch entity {
	case "tenant", "tn":
		return TENANT
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
		return CORIDOR
	case "sensor", "sr":
		return SENSOR
	default:
		return -1
	}
}

func GetParentOfEntity(ent int) int {
	switch ent {
	case TENANT:
		return -1
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
	case CORIDOR:
		return ROOM
	case SENSOR:
		return -2
	default:
		return -3
	}
}
