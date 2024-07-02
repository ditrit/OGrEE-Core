package controllers

//This file has a collection of utility functions used in the
//controller package
//And const definitions used throughout the controllers package
import (
	"encoding/json"
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
const VIRTUALCONFIG = "virtual_config"

// displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

// Returns true/false if exists and true/false if attr
// is in "attributes" maps
func AttrIsInObj(obj map[string]interface{}, attr string) (bool, bool) {
	if _, ok := obj[attr]; ok {
		return ok, false
	}

	if hasAttr, _ := AttrIsInObj(obj, "attributes"); hasAttr {
		if objAttributes, ok := obj["attributes"].(map[string]interface{}); ok {
			_, ok := objAttributes[attr]
			return ok, true
		}
	}

	return false, false
}

type ErrorWithInternalError struct {
	UserError     error
	InternalError error
}

func (err ErrorWithInternalError) Error() string {
	return err.UserError.Error() + " caused by " + err.InternalError.Error()
}

// Utility functions
func determineStrKey(x map[string]interface{}, possible []string) string {
	for idx := range possible {
		if _, ok := x[possible[idx]]; ok {
			return possible[idx]
		}
	}
	return "" //The code should not reach this point!
}
