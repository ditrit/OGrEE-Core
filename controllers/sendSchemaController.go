package controllers

// Auxillary functions for parsing and validation of data
// before the CLI sends off to API

import (
	"bytes"
	l "cli/logger"
	"fmt"
	"strconv"
	"strings"
)

// Serialising size & posXY is inefficient but
// the team wants it for now
// "size":"[25,29.4,0]" -> "size": "{\"x\":25,\"y\":29.4,\"z\":0}"
func serialiseAttr(attr map[string]interface{}, want string) string {
	var newSize string
	if size, ok := attr[want].(string); ok {
		left := strings.Index(size, "[")
		right := strings.Index(size, "]")
		coords := []string{"x", "y", "z"}

		if left != -1 && right != -1 {
			var length int
			subStr := size[left+1 : right]
			nums := stringSplitter(subStr, ",", want)
			if nums == nil { //Error!
				return ""
			}
			//nums := strings.Split(subStr, ",")

			if len(nums) == 3 && want == "size" {
				length = 2
			} else {
				//Make posXYZ length 3 for racks
				if want == "posXYZ" && len(nums) == 2 {
					nums = append(nums, "0")
				}
				length = len(nums)
			}

			for idx := 0; idx < length; idx++ {
				newSize += "\"" + coords[idx] + "\":" + nums[idx]

				if idx < length-1 {
					newSize += ","
				}
			}
			newSize = "{" + newSize + "}"

			if len(nums) == 3 && want == "size" {
				attr["height"] = nums[2]
			}
		}
	}
	return newSize
}

// Same utility func as above but we have an arbitrary array
// and want to cast it to -> "size": "{\"x\":25,\"y\":29.4,\"z\":0}"
func serialiseAttr2(attr map[string]interface{}, want string) string {
	var newSize string
	if items, ok := attr[want].([]interface{}); ok {
		coords := []string{"x", "y", "z"}
		var length int

		if isValid := arrayVerifier(&items, want); !isValid {
			return ""
		}

		if len(items) == 3 && want == "size" {
			length = 2
		} else {
			//Make posXYZ length 3 for racks
			if want == "posXYZ" && len(items) == 2 {
				items = append(items, 0)
			}
			length = len(items)
		}

		for idx := 0; idx < length; idx++ {
			r := bytes.NewBufferString("")
			fmt.Fprintf(r, "%v ", items[idx])
			//itemStr :=
			newSize += "\"" + coords[idx] + "\":" + r.String()

			if idx < length-1 {
				newSize += ","
			}
		}
		newSize = "{" + newSize + "}"

		if len(items) == 3 && want == "size" {
			if _, ok := items[2].(int); ok {
				items[2] = strconv.Itoa(items[2].(int))
			} else if _, ok := items[2].(float64); ok {
				items[2] = strconv.FormatFloat(items[2].(float64), 'G', -1, 64)
			}
			attr["height"] = items[2]
		}
	}
	return newSize
}

// Auxillary function for serialiseAttr2
// to help ensure that the arbitrary arrays
// ([]interface{}) are valid before they get serialised
func arrayVerifier(x *[]interface{}, attribute string) bool {
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
			if EntityStrToInt(cat) != -1 {

				//Start the filtration
				for i := range x {
					if i == "attributes" {
						for idx := range x[i].(map[string]interface{}) {
							if IsAttrDrawable("", idx, x, true) == true {
								attrs[idx] = x[i].(map[string]interface{})[idx]
							}
						}
					} else {
						if IsAttrDrawable("", i, x, true) == true {
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

				reservedStr = "{\"left\":" + r4.String() + ",\"right\":" + r3.String() + ",\"top\":" + r1.String() + ",\"bottom\":" + r2.String() + "}"
				techStr = "{\"left\":" + t4.String() + ",\"right\":" + t3.String() + ",\"top\":" + t1.String() + ",\"bottom\":" + t2.String() + "}"
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
	switch x.(type) {
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
		xArr := x.([]interface{})
		var arrStr []string
		for i := range xArr {
			arrStr = append(arrStr, Stringify(xArr[i]))
		}
		return "[" + strings.Join(arrStr, ",") + "]"

	}
	return ""
}
