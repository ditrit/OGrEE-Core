package controllers

//This file has a collection of utility functions used in the
//controller package
import (
	"encoding/json"
	"fmt"
)

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
					argument = append(argument, nil)
				}
			}
			fmt.Printf(format, argument...)
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
