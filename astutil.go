package main

import (
	cmd "cli/controllers"
	u "cli/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
)

func checkTypesAreSame(x, y interface{}) bool {
	//println(reflect.TypeOf(x))
	return reflect.TypeOf(x) == reflect.TypeOf(y)
}

func checkTypeAreNumeric(x, y interface{}) bool {
	var xOK, yOK bool
	switch x.(type) {
	case int, float64, float32:
		xOK = true
	default:
		xOK = false
	}

	switch y.(type) {
	case int, float64, float32:
		yOK = true
	default:
		yOK = false
	}

	return xOK && yOK
}

func checkIfOrientation(x string) bool {
	switch x {
	case /*"EN", "NW", "WS", "SE", "NE", "SW",*/
		"-E-N", "-E+N", "+E-N", "+E+N", "+N+E",
		"+N-E", "-N-E", "-N+E",
		"-N-W", "-N+W", "+N-W", "+N+W",
		"-W-S", "-W+S", "+W-S", "+W+S",
		"-S-E", "-S+E", "+S-E", "+S+E",
		"+x+y", "+x-y", "-x-y", "-x+y",
		"+X+Y", "+X-Y", "-X-Y", "-X+Y":
		return true
	default:
		return false
	}
}

var floatType = reflect.TypeOf(float64(0))

func getFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

// Open a file and return the JSON in the file
// Used by EasyPost, EasyUpdate and Load Template
func fileToJSON(path string) map[string]interface{} {
	data := map[string]interface{}{}
	x, e := ioutil.ReadFile(path)
	if e != nil {
		if cmd.State.DebugLvl > cmd.NONE {
			println("Error while opening file! " + e.Error())
		}
		return nil
	}
	json.Unmarshal(x, &data)
	return data
}

// Iterates through x and executes the element if the
// element is a node
func evalMapNodes(x map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for i := range x {
		switch v := x[i].(type) {
		case node:
			val, err := v.execute()
			if err != nil {
				return nil, err
			}
			result[i] = val
		case map[string]interface{}:
			val, err := evalMapNodes(v)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}
	}
	return result, nil
}

// Generic function for evaluating []node and returning the desired array
func evalNodeArr[elt comparable](arr *[]node, x []elt) ([]elt, error) {
	for _, v := range *arr {
		val, e := v.execute()
		if e != nil {
			return nil, e
		}
		if _, ok := val.(elt); !ok {
			//do something here
			return nil, fmt.Errorf("Error unexpected element")
		}
		x = append(x, val.(elt))
	}
	return x, nil
}

// This func is for distinguishing template from sizeU
// for creating devices,
// distinguishing template from size when creating buildings,
// and template validity check for rooms,
// refer to:
// https://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#Create-a-Device
func checkIfTemplate(x interface{}, ent int) bool {
	var location string
	if s, ok := x.(string); ok {
		switch ent {
		case cmd.BLDG:
			location = "/Logical/BldgTemplates/" + s
		case cmd.ROOM:
			location = "/Logical/RoomTemplates/" + s
		default:
			location = "/Logical/ObjectTemplates/" + s
		}
		_, exists := cmd.CheckObject(location, true)
		return exists
	}
	return false
}

func resMap(x map[string]interface{}, ent string, isUpdate bool) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	attrs := make(map[string]string)

	for key := range x {
		val, ok := x[key].(string)
		if !ok {
			return nil, fmt.Errorf("Attribute should contain a string")
		}
		if isUpdate == true {
			res[key] = val
			continue
		}

		if u.IsNestedAttr(key, ent) {
			attrs[key] = val
		} else {
			res[key] = val
		}
	}
	if len(attrs) > 0 {
		res["attributes"] = attrs
	}
	return res, nil
}

// errResponder helper func for specialUpdateNode
// used for separator, pillar err msgs and parseAreas()
func errorResponder(attr, numElts string, multi bool) error {
	var errorMsg string
	if multi {
		errorMsg = "Invalid " + attr + " attributes provided." +
			" They must be arrays/lists/vectors with " + numElts + " elements."
	} else {
		errorMsg = "Invalid " + attr + " attribute provided." +
			" It must be an array/list/vector with " + numElts + " elements."
	}

	segment := " Please refer to the wiki or manual reference" +
		" for more details on how to create objects " +
		"using this syntax"

	return fmt.Errorf(errorMsg + segment)
}

func IsMapStrInf(x interface{}) bool {
	_, ok := x.(map[string]interface{})
	return ok
}

func IsInfArr(x interface{}) bool {
	_, ok := x.([]interface{})
	return ok
}

func IsString(x interface{}) bool {
	_, ok := x.(string)
	return ok
}

func IsStringArr(x interface{}) bool {
	_, ok := x.([]string)
	return ok
}

func IsStringValue(x interface{}, value string) bool {
	return x == value
}

func IsHexString(s string) bool {
	//Eliminate 'odd length' errors
	if len(s)%2 != 0 {
		s = "0" + s
	}

	_, err := hex.DecodeString(s)
	return err == nil
}

func IsBool(x interface{}) bool {
	_, ok := x.(bool)
	return ok
}

func IsInt(x interface{}) bool {
	_, ok := x.(int)
	return ok
}

func IsFloat(x interface{}) bool {
	_, ok := x.(float64)
	_, ok2 := x.(float32)
	return ok || ok2
}

func IsAmongValues(x interface{}, values *[]string) bool {
	for i := range *values {
		if x == (*values)[i] {
			return true
		}
	}
	return false
}
