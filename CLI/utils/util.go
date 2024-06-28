package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

func ExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exe)
}

var floatType = reflect.TypeOf(float64(0))

func GetFloat(unk interface{}) (float64, error) {
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		return 0, fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float(), nil
}

func ValToFloat(val any, name string) (float64, error) {
	stringVal, isString := val.(string)
	if isString {
		floatVal, err := strconv.ParseFloat(stringVal, 64)
		if err != nil {
			return 0, fmt.Errorf("%s should be a number", name)
		}
		return floatVal, nil
	}
	v, err := GetFloat(val)
	if err != nil {
		return 0, fmt.Errorf("%s should be a number", name)
	}
	return v, nil
}

func StringToNum(s string) (any, error) {
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

func ValToNum(val any, name string) (any, error) {
	stringVal, isString := val.(string)
	if isString {
		numVal, err := StringToNum(stringVal)
		if err != nil {
			return nil, fmt.Errorf("%s should be a number", name)
		}
		return numVal, nil
	}
	return val, nil
}

func ValToInt(val any, name string) (int, error) {
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

func ValToBool(val any, name string) (bool, error) {
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

func ValTo3dRotation(val any) ([]float64, error) {
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

var ErrShouldBeAString = errors.New("should be a string")

func ValToString(val any, name string) (string, error) {
	intVal, isInt := val.(int)
	if isInt {
		return strconv.Itoa(intVal), nil
	}
	stringVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%s %w", name, ErrShouldBeAString)
	}
	return stringVal, nil
}

func ValToVec(val any, size int, name string) ([]float64, error) {
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

func ValToColor(color interface{}) (string, bool) {
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

func IsNumeric(x interface{}) bool {
	return IsInt(x) || IsFloat(x)
}

func CompareVals(val1 any, val2 any) (bool, bool) {
	val1Float, err1 := ValToFloat(val1, "")
	val2Float, err2 := ValToFloat(val2, "")
	if err1 == nil && err2 == nil {
		return val1Float < val2Float, true
	}
	val1String, isString1 := val1.(string)
	val2String, isString2 := val2.(string)
	if isString1 && isString2 {
		return val1String < val2String, true
	}
	return false, false
}

func NameOrSlug(obj map[string]any) string {
	slug, okSlug := obj["slug"].(string)
	if okSlug {
		return slug
	}

	name, okName := obj["name"].(string)
	if okName {
		return name
	}
	panic("child has no name/slug")
}

func GetValFromObj(obj map[string]any, key string) (any, bool) {
	val, ok := obj[key]
	if ok {
		return val, true
	}
	attributes, ok := obj["attributes"].(map[string]any)
	if !ok {
		return nil, false
	}
	val, ok = attributes[key]
	if !ok {
		return nil, false
	}
	return val, true
}

// Helper func that safely copies a value in a map
func CopyMapVal(dest, source map[string]interface{}, key string) bool {
	if _, ok := source[key]; ok {
		dest[key] = source[key]
		return true
	}
	return false
}

// Convert []interface{} array to
// []map[string]interface{} array
func AnyArrToMapArr(x []interface{}) []map[string]interface{} {
	ans := []map[string]interface{}{}
	for i := range x {
		ans = append(ans, x[i].(map[string]interface{}))
	}
	return ans
}

func Stringify(x interface{}) string {
	switch xArr := x.(type) {
	case string:
		return x.(string)
	case int:
		return strconv.Itoa(x.(int))
	case float32, float64:
		return strconv.FormatFloat(float64(x.(float64)), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x.(bool))
	case []string:
		return strings.Join(x.([]string), ",")
	case []interface{}:
		var arrStr []string
		for i := range xArr {
			arrStr = append(arrStr, Stringify(xArr[i]))
		}
		return "[" + strings.Join(arrStr, ",") + "]"
	case []float64:
		var arrStr []string
		for i := range xArr {
			arrStr = append(arrStr, Stringify(xArr[i]))
		}
		return "[" + strings.Join(arrStr, ",") + "]"
	}
	return ""
}

func MergeMaps(x, y map[string]interface{}, overwrite bool) {
	for i := range y {
		//Conflict case
		if _, ok := x[i]; ok {
			if overwrite {
				x[i] = y[i]
			}
		} else {
			x[i] = y[i]
		}

	}
}
