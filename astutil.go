package main

import (
	cmd "cli/controllers"
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

//Open a file and return the JSON in the file
//Used by EasyPost, EasyUpdate and Load Template
func fileToJSON(path string) map[string]interface{} {
	data := map[string]interface{}{}
	x, e := ioutil.ReadFile(path)
	if e != nil {
		println("Error while opening file! " + e.Error())
		return nil
	}
	json.Unmarshal(x, &data)
	return data
}

func formActualPath(x string) string {
	if x == "" || x == "." {
		return cmd.State.CurrPath
	} else if string(x[0]) == "/" {
		return x

	} else {
		return cmd.State.CurrPath + "/" + x
	}
}

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

//This func is for distinguishing template from sizeU
//in the OCLI syntax for creating devices
//refer to:
//https://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#Create-a-Device
func checkIfTemplate(x interface{}) bool {
	if s, ok := x.(string); ok {
		if m, _ := cmd.GetObject("/Logical/ObjectTemplates/"+s, true); m != nil {
			return true
		}
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
		switch ent {
		case "sensor", "group":
			switch key {
			case "id", "name", "category", "parentID",
				"description", "domain", "type",
				"parentid", "parentId":
				res[key] = val

			default:
				attrs[key] = val
			}
		case "room_template":
			switch key {
			case "id", "slug", "orientation", "separators",
				"tiles", "colors", "rows", "sizeWDHm",
				"technicalArea", "reservedArea":
				res[key] = val

			default:
				attrs[key] = val
			}
		case "obj_template":
			switch key {
			case "id", "slug", "description", "category",
				"slots", "colors", "components", "sizeWDHmm",
				"fbxModel":
				res[key] = val

			default:
				attrs[key] = val
			}
		default:
			switch key {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
				res[key] = val

			default:
				attrs[key] = val
			}
		}
	}
	if len(attrs) > 0 {
		res["attributes"] = attrs
	}
	return res, nil
}
