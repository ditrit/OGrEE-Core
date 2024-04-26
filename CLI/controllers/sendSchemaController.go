package controllers

// Auxillary functions for parsing and validation of data
// before the CLI sends off to API

import (
	l "cli/logger"
	"cli/models"
	"fmt"
	"strconv"
	"strings"
)

func serialiseVector(attr map[string]interface{}, want string) []float64 {
	if vector, ok := attr[want].([]float64); ok {
		if want == "size" && len(vector) == 3 {
			attr["height"] = vector[2]
			vector = vector[:len(vector)-1]
		} else if want == "posXYZ" && len(vector) == 2 {
			vector = append(vector, 0)
		}
		return vector
	} else {
		return []float64{}
	}
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

// ExpandStrVector: allow usage of .. on device slot and group content vector
// converting [slot01..slot03] on [slot01,slot02,slot03]
func ExpandStrVector(slotVector []string) ([]string, error) {
	slots := []string{}
	for _, slot := range slotVector {
		if strings.Contains(slot, "..") {
			if len(slotVector) != 1 {
				return nil, fmt.Errorf("Invalid device syntax: .. can only be used in a single element vector")
			}
			parts := strings.Split(slot, "..")
			if len(parts) != 2 ||
				(parts[0][:len(parts[0])-1] != parts[1][:len(parts[1])-1]) {
				l.GetWarningLogger().Println("Invalid device syntax encountered")
				return nil, fmt.Errorf("Invalid device syntax: incorrect use of .. for slot")
			} else {
				start, errS := strconv.Atoi(string(parts[0][len(parts[0])-1]))
				end, errE := strconv.Atoi(string(parts[1][len(parts[1])-1]))
				if errS != nil || errE != nil {
					l.GetWarningLogger().Println("Invalid device syntax encountered")
					return nil, fmt.Errorf("Invalid device syntax: incorrect use of .. for slot")
				} else {
					prefix := parts[0][:len(parts[0])-1]
					for i := start; i <= end; i++ {
						slots = append(slots, prefix+strconv.Itoa(i))
					}
				}
			}
		} else {
			slots = append(slots, slot)
		}
	}
	return slots, nil
}
