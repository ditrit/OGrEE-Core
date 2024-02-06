package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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

func ObjectAttr(obj map[string]any, attr string) (any, bool) {
	val, ok := obj[attr]
	if ok {
		return val, true
	}
	attributes, ok := obj["attributes"].(map[string]any)
	if !ok {
		return nil, false
	}
	val, ok = attributes[attr]
	if !ok {
		return nil, false
	}
	return val, true
}

func ComplexFilterToMap(complexFilter string) map[string]any {
	// Split the input string into individual filter expressions
	chars := []string{"(", ")", "&", "|"}
	for _, char := range chars {
		complexFilter = strings.ReplaceAll(complexFilter, char, " "+char+" ")
	}
	return complexExpressionToMap(strings.Fields(complexFilter))
}

func complexExpressionToMap(expressions []string) map[string]any {
	// Find the rightmost operator (AND, OR) outside of parentheses
	parenCount := 0
	for i := len(expressions) - 1; i >= 0; i-- {
		switch expressions[i] {
		case "(":
			parenCount++
		case ")":
			parenCount--
		case "&":
			if parenCount == 0 {
				return map[string]any{"$and": []map[string]any{
					complexExpressionToMap(expressions[:i]),
					complexExpressionToMap(expressions[i+1:]),
				}}
			}
		case "|":
			if parenCount == 0 {
				return map[string]any{"$or": []map[string]any{
					complexExpressionToMap(expressions[:i]),
					complexExpressionToMap(expressions[i+1:]),
				}}
			}
		}
	}

	// If there are no operators outside of parentheses, look for the innermost pair of parentheses
	for i := 0; i < len(expressions); i++ {
		if expressions[i] == "(" {
			start, end := i+1, i+1
			for parenCount := 1; end < len(expressions) && parenCount > 0; end++ {
				switch expressions[end] {
				case "(":
					parenCount++
				case ")":
					parenCount--
				}
			}
			return complexExpressionToMap(append(expressions[:start-1], expressions[start:end-1]...))
		}
	}

	// Base case: single filter expression
	re := regexp.MustCompile(`^([\w-.]+)\s*(<=|>=|<|>|!=|=)\s*([\w-.]+)$`)

	ops := map[string]string{"<=": "$lte", ">=": "$gte", "<": "$lt", ">": "$gt", "!=": "$ne", "=": "$eq"}

	if len(expressions) <= 3 {
		expression := strings.Join(expressions[:], "")

		if match := re.FindStringSubmatch(expression); match != nil {
			switch match[1] {
			case "startDate":
				// if match[2] != "=" {
				// 	fmt.Println("Error: Invalid filter expression")
				// 	return map[string]any{"error": "invalid filter expression"}
				// }
				// startDate, e := time.Parse("2006-01-02", match[3])
				// if e != nil {
				// 	fmt.Println("Error:", e.Error())
				// 	return map[string]any{"error": e.Error()}
				// }
				// return map[string]any{"lastUpdated": map[string]any{"$gte": primitive.NewDateTimeFromTime(startDate)}}
				return map[string]any{"lastUpdated": map[string]any{"$gte": match[3]}}
			case "endDate":
				// if match[2] != "=" {
				// 	fmt.Println("Error: Invalid filter expression")
				// 	return map[string]any{"error": "invalid filter expression"}
				// }
				// endDate, e := time.Parse("2006-01-02", match[3])
				// endDate = endDate.Add(time.Hour * 24)
				// if e != nil {
				// 	fmt.Println("Error:", e.Error())
				// 	return map[string]any{"error": e.Error()}
				// }
				// return map[string]any{"lastUpdated": map[string]any{"$lte": primitive.NewDateTimeFromTime(endDate)}}
				return map[string]any{"lastUpdated": map[string]any{"$lte": match[3]}}
			case "id", "name", "category", "description", "domain", "createdDate", "lastUpdated", "slug":
				return map[string]any{match[1]: map[string]any{ops[match[2]]: match[3]}}
			default:
				return map[string]any{"attributes." + match[1]: map[string]any{ops[match[2]]: match[3]}}
			}
		}
	}

	fmt.Println("Error: Invalid filter expression")
	return map[string]any{"error": "invalid filter expression"}
}
