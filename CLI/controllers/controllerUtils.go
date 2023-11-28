package controllers

//This file has a collection of utility functions used in the
//controller package
//And const definitions used throughout the controllers package
import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
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

func TranslatePath(p string, acceptSelection bool) string {
	if p == "" {
		p = "."
	}
	if p == "_" && acceptSelection {
		return "_"
	}
	if p == "-" {
		return State.PrevPath
	}
	var output_words []string
	if p[0] != '/' {
		outputBase := State.CurrPath
		if p[0] == '-' {
			outputBase = State.PrevPath
		}

		output_words = strings.Split(outputBase, "/")[1:]
		if len(output_words) == 1 && output_words[0] == "" {
			output_words = output_words[0:0]
		}
	} else {
		p = p[1:]
	}
	input_words := strings.Split(p, "/")
	for i, word := range input_words {
		if word == "." || (i == 0 && word == "-") {
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
