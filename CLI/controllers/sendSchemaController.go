package controllers

// Auxillary functions for parsing and validation of data
// before the CLI sends off to API

import (
	"bytes"
	l "cli/logger"
	"cli/models"
	"fmt"
	"strconv"
	"strings"
)

func serialiseVector(attributes map[string]interface{}, attributeName string) {
	if _, isString := attributes[attributeName].(string); isString {
		attributes[attributeName] = serialiseStringVector(attributes, attributeName)
	} else {
		attributes[attributeName] = serialiseFloatVector(attributes, attributeName)
	}
}

func serialiseStringVector(attr map[string]interface{}, want string) string {
	var newSize string

	size, ok := attr[want].(string)
	if !ok {
		return ""
	}

	left := strings.Index(size, "[")
	right := strings.Index(size, "]")
	if left == -1 || right == -1 {
		return ""
	}

	nums := stringSplitter(size[left+1:right], ",", want)
	if nums == nil {
		return ""
	}
	//nums := strings.Split(subStr, ",")

	length := len(nums)
	if length == 3 && want == "size" {
		length = 2
	} else if want == "posXYZ" && length == 2 {
		nums = append(nums, "0")
		length++
	}

	newSize = "[" + strings.Join(nums[:length], ", ") + "]"

	if len(nums) == 3 && want == "size" {
		if attr["shape"] == "sphere" || attr["shape"] == "cylinder" {
			attr["diameter"] = nums[2]
			if attr["shape"] == "cylinder" {
				attr["height"] = nums[1]
			}
			return ""
		} else {
			attr["height"] = nums[2]
		}
	}

	return newSize
}

func serialiseFloatVector(attr map[string]interface{}, want string) string {
	var newSize string

	items, ok := attr[want].([]float64)
	if !ok || !arrayVerifier(&items, want) {
		return ""
	}

	length := len(items)
	if length == 3 && want == "size" {
		length = 2
	} else if want == "posXYZ" && length == 2 {
		items = append(items, 0)
		length++
	}

	var itemStrs []string
	for idx := 0; idx < len(items); idx++ {
		itemStrs = append(itemStrs, strconv.FormatFloat(items[idx], 'G', -1, 64))
	}

	newSize = "[" + strings.Join(itemStrs[:length], ", ") + "]"

	if len(items) == 3 && want == "size" {
		if attr["shape"] == "sphere" || attr["shape"] == "cylinder" {
			attr["diameter"] = itemStrs[2]
			if attr["shape"] == "cylinder" {
				attr["height"] = itemStrs[1]
			}
			return ""
		} else {
			attr["height"] = itemStrs[2]
		}
	}

	return newSize
}

// Auxillary function for serialiseAttr2
// to help ensure that the arbitrary arrays
// ([]interface{}) are valid before they get serialised
func arrayVerifier(x *[]float64, attribute string) bool {
	switch attribute {
	case "size":
		return len(*x) == 3
	case "posXY":
		return len(*x) == 2
	case "posXYZ":
		return len(*x) == 2 || len(*x) == 3
	}
	return false
}

// Auxillary function for serialiseAttr
// to help verify the posXY and size attributes
// have correct lengths before they get serialised
func stringSplitter(want, separator, attribute string) []string {
	arr := strings.Split(want, separator)
	switch attribute {
	case "posXYZ":
		if len(arr) != 2 && len(arr) != 3 {
			return nil
		}
	case "posXY":
		if len(arr) != 2 {
			return nil
		}
	case "size":
		if len(arr) != 3 {
			return nil
		}
	}
	return arr
}

func MergeMaps(x, y map[string]interface{}, overwrite bool) {
	for i := range y {
		//Conflict case
		if _, ok := x[i]; ok {
			if overwrite {
				l.GetWarningLogger().Println("Conflict while merging maps")
				if State.DebugLvl > 1 {
					println("Conflict while merging data, resorting to overwriting!")
				}

				x[i] = y[i]
			}
		} else {
			x[i] = y[i]
		}

	}
}

// This func is used for when the user wants to filter certain
// attributes from being sent/displayed to Unity viewer client
func GenerateFilteredJson(x map[string]interface{}) map[string]interface{} {
	ans := map[string]interface{}{}
	attrs := map[string]interface{}{}
	if catInf, ok := x["category"]; ok {
		if cat, ok := catInf.(string); ok {
			if models.EntityStrToInt(cat) != -1 {

				//Start the filtration
				for i := range x {
					if i == "attributes" {
						for idx := range x[i].(map[string]interface{}) {
							if IsCategoryAttrDrawable(x["category"].(string), idx) {
								attrs[idx] = x[i].(map[string]interface{})[idx]
							}
						}
					} else {
						if IsCategoryAttrDrawable(x["category"].(string), i) {
							ans[i] = x[i]
						}
					}
				}
				if len(attrs) > 0 {
					ans["attributes"] = attrs
				}
				return ans
			}
		}
	}
	return x //Nothing will be filtered
}

// Helper func is used to check if sizeU is numeric
// this is necessary since the OCLI command for creating a device
// needs to distinguish if the parameter is a valid sizeU or template
func checkNumeric(x interface{}) bool {
	switch x.(type) {
	case int, float64, float32:
		return true
	default:
		return false
	}
}

// Hack function for the reserved and technical areas
// which copies that room areas function in ast.go
// [room]:areas=[r1,r2,r3,r4]@[t1,t2,t3,t4]
func parseReservedTech(x map[string]interface{}) map[string]interface{} {
	var reservedStr string
	var techStr string
	if reserved, ok := x["reserved"].([]interface{}); ok {
		if tech, ok := x["technical"].([]interface{}); ok {
			if len(reserved) == 4 && len(tech) == 4 {
				r4 := bytes.NewBufferString("")
				fmt.Fprintf(r4, "%v", reserved[3].(float64))
				r3 := bytes.NewBufferString("")
				fmt.Fprintf(r3, "%v", reserved[2].(float64))
				r2 := bytes.NewBufferString("")
				fmt.Fprintf(r2, "%v", reserved[1].(float64))
				r1 := bytes.NewBufferString("")
				fmt.Fprintf(r1, "%v", reserved[0].(float64))

				t4 := bytes.NewBufferString("")
				fmt.Fprintf(t4, "%v", tech[3].(float64))
				t3 := bytes.NewBufferString("")
				fmt.Fprintf(t3, "%v", tech[2].(float64))
				t2 := bytes.NewBufferString("")
				fmt.Fprintf(t2, "%v", tech[1].(float64))
				t1 := bytes.NewBufferString("")
				fmt.Fprintf(t1, "%v", tech[0].(float64))
				// [front/top, back/bottom, right, left]
				reservedStr = "[" + r1.String() + ", " + r2.String() + ", " + r3.String() + ", " + r4.String() + "]"
				techStr = "[" + t1.String() + ", " + t2.String() + ", " + t3.String() + ", " + t4.String() + "]"
				x["reserved"] = reservedStr
				x["technical"] = techStr
			}
		}
	}
	return x
}

// Helper func that safely deletes a string key in a map
func DeleteAttr(x map[string]interface{}, key string) {
	if _, ok := x[key]; ok {
		delete(x, key)
	}
}

// Helper func that safely copies a value in a map
func CopyAttr(dest, source map[string]interface{}, key string) bool {
	if _, ok := source[key]; ok {
		dest[key] = source[key]
		return true
	}
	return false
}

// Used for update commands to ensure all data sent to API
// are in string format
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
