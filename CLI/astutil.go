package main

import (
	cmd "cli/controllers"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
)

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

func valToFloat(val any, name string) (float64, error) {
	v, err := getFloat(val)
	if err != nil {
		return 0, fmt.Errorf("%s should be a number", name)
	}
	return v, nil
}

func nodeToFloat(n node, name string) (float64, error) {
	val, err := n.execute()
	if err != nil {
		return 0, err
	}
	return valToFloat(val, name)
}

func stringToNum(s string) (any, error) {
	intVal, err := strconv.Atoi(s)
	if err == nil {
		return intVal, nil
	}
	floatVal, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return floatVal, nil
	}
	return nil, fmt.Errorf("the string is not a number")
}

func nodeToNum(n node, name string) (any, error) {
	val, err := n.execute()
	if err != nil {
		return nil, err
	}
	stringVal, isString := val.(string)
	if isString {
		val, err = stringToNum(stringVal)
		if err != nil {
			return nil, fmt.Errorf("%s should be a number", name)
		}
	}
	return val, nil
}

func valToInt(val any, name string) (int, error) {
	stringVal, isString := val.(string)
	if isString {
		intVal, err := strconv.Atoi(stringVal)
		if err != nil {
			return 0, fmt.Errorf("%s should be an integer", name)
		}
		return intVal, nil
	}
	intVal, ok := val.(int)
	if !ok {
		return 0, fmt.Errorf("%s should be an integer", name)
	}
	return intVal, nil
}

func nodeToInt(n node, name string) (int, error) {
	val, err := n.execute()
	if err != nil {
		return 0, err
	}
	return valToInt(val, name)
}

func valToBool(val any, name string) (bool, error) {
	var err error
	stringVal, isString := val.(string)
	if isString {
		val, err = strconv.ParseBool(stringVal)
		if err != nil {
			return false, fmt.Errorf("%s should be a boolean", name)
		}
	}
	boolVal, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("%s should be a boolean", name)
	}
	return boolVal, nil
}

func nodeToBool(n node, name string) (bool, error) {
	val, err := n.execute()
	if err != nil {
		return false, err
	}
	return valToBool(val, name)
}

func valTo3dRotation(val any) ([]float64, error) {
	switch rotation := val.(type) {
	case []float64:
		return rotation, nil
	case string:
		switch rotation {
		case "front":
			return []float64{0, 0, 180}, nil
		case "rear":
			return []float64{0, 0, 0}, nil
		case "left":
			return []float64{0, 90, 0}, nil
		case "right":
			return []float64{0, -90, 0}, nil
		case "top":
			return []float64{90, 0, 0}, nil
		case "bottom":
			return []float64{-90, 0, 0}, nil
		}
	}
	return nil, fmt.Errorf(
		`rotation should be a vector3, or one of the following keywords :
		front, rear, left, right, top, bottom`)
}

func nodeTo3dRotation(n node) ([]float64, error) {
	val, err := n.execute()
	if err != nil {
		return nil, err
	}
	return valTo3dRotation(val)
}

func valToString(val any, name string) (string, error) {
	intVal, isInt := val.(int)
	if isInt {
		return strconv.Itoa(intVal), nil
	}
	stringVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%s should be a string", name)
	}
	return stringVal, nil
}

func nodeToString(n node, name string) (string, error) {
	val, err := n.execute()
	if err != nil {
		return "", err
	}
	return valToString(val, name)
}

func valToVec(val any, size int, name string) ([]float64, error) {
	vecVal, isVec := val.([]float64)
	if !isVec || (size >= 0 && len(vecVal) != size) {
		msg := fmt.Sprintf("%s should be a vector", name)
		if size != -1 {
			msg += strconv.Itoa(size)
		}
		return nil, fmt.Errorf(msg)
	}
	return vecVal, nil
}

func nodeToVec(n node, size int, name string) ([]float64, error) {
	val, err := n.execute()
	if err != nil {
		return nil, err
	}

	return valToVec(val, size, name)
}

func valToColor(color interface{}) (string, bool) {
	var colorStr string
	if IsString(color) || IsInt(color) || IsFloat(color) {
		if IsString(color) {
			colorStr = color.(string)
		}

		if IsInt(color) {
			colorStr = strconv.Itoa(color.(int))
		}

		if IsFloat(color) {
			colorStr = strconv.FormatFloat(color.(float64), 'f', -1, 64)
		}

		for len(colorStr) < 6 {
			colorStr = "0" + colorStr
		}

		if len(colorStr) != 6 {
			return "", false
		}

		if !IsHexString(colorStr) {
			return "", false
		}

	} else {
		return "", false
	}
	return colorStr, true
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

func checkIfTemplate(name string, ent int) bool {
	var location string
	switch ent {
	case cmd.BLDG:
		location = "/Logical/BldgTemplates/" + name
	case cmd.ROOM:
		location = "/Logical/RoomTemplates/" + name
	default:
		location = "/Logical/ObjectTemplates/" + name
	}
	_, err := cmd.Tree(location, 0)
	return err == nil
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

func IsInfArr(x interface{}) bool {
	_, ok := x.([]interface{})
	return ok
}

func IsString(x interface{}) bool {
	_, ok := x.(string)
	return ok
}

func IsHexString(s string) bool {
	//Eliminate 'odd length' errors
	if len(s)%2 != 0 {
		s = "0" + s
	}

	_, err := hex.DecodeString(s)
	return err == nil
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
