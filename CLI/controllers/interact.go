package controllers

import (
	"fmt"
	"strconv"
	"strings"
)

// Function called by InteractObject to extract the value of an attribute from the object
func valFromObj(obj map[string]any, val interface{}) (interface{}, error) {
	if value, ok := val.(string); ok {
		innerMap := obj["attributes"].(map[string]interface{})

		if _, ok := obj[value]; ok {
			if value == "description" {
				desc := obj["description"].([]interface{})
				val = ""
				//Combine entire the description array into a string
				for i := 0; i < len(desc); i++ {
					if i == 0 {
						val = desc[i].(string)
					} else {
						val = val.(string) + "\n" + desc[i].(string)
					}
				}
				return val, nil
			} else {
				val = obj[value]
			}
		} else if _, ok := innerMap[value]; ok {
			val = innerMap[value]
		} else {
			if strings.Contains(value, "description") {
				if desc, ok := obj["description"].([]interface{}); ok {
					if len(value) > 11 { //descriptionX format
						//split the number and description
						numStr := strings.Split(value, "description")[1]
						num, e := strconv.Atoi(numStr)
						if e != nil {
							return "", e
						}

						if num < 0 {
							return "", fmt.Errorf("Description index must be positive")
						}

						if num >= len(desc) {
							msg := "Description index is out of" +
								" range. The length for this object is: " +
								strconv.Itoa(len(desc))
							return "", fmt.Errorf(msg)
						}
						val = desc[num]
					} else {
						val = innerMap[value]
					}
				}
			} else {
				msg := "The specified attribute '" + val.(string) + "' does not exist" +
					" in the object. \nPlease view the object" +
					" (ie. $> get) and try again"
				return "", fmt.Errorf(msg)
			}
		}
	} else {
		return "", fmt.Errorf("The label value must be a string")
	}

	return val, nil
}

// Function called by InteractObject to separate the name of the attribute from the rest of the word
func splitAttrAndSuffix(input string) (string, string) {
	attr, suffix := "", ""

	for _, char := range input {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			attr += string(char)
		} else {
			suffix = input[len(attr):]
			break
		}
	}

	return attr, suffix
}

// Function called by update node for interact commands (ie label, labelFont)
func InteractObject(path string, keyword string, val interface{}, fromAttr bool) error {
	// First retrieve the object
	obj, err := C.GetObject(path)
	if err != nil {
		return err
	}

	// The label should be a string
	if str, ok := val.(string); ok {
		// Break the string into words
		words := strings.Fields(str)

		// Check if the val refers to an attribute field in the object
		// this means to retrieve value from object
		if fromAttr {
			attr, suffix := splitAttrAndSuffix(words[0])
			newVal, err := valFromObj(obj, attr)
			if err != nil {
				return err
			}
			words[0] = newVal.(string) + suffix
		}

		// Check all words
		for i, word := range words {
			// Substitute \n for <br> so 3D understands the linebreak
			for strings.Contains(word, "\\n") {
				j := strings.Index(word, "\\n")
				word = word[:j] + "<br>" + word[j+2:]
				words[i] = word
			}

			// While there is at least one attribute to be evaluated in the word
			for strings.Contains(word, "#") {
				j := strings.Index(word, "#")
				attr, suffix := splitAttrAndSuffix(word[j+1:])
				newVal, err := valFromObj(obj, attr)
				if err != nil {
					return err
				}

				word = word[:j] + newVal.(string) + suffix
				words[i] = word
			}

		}
		val = strings.Join(words, " ")

		data := map[string]interface{}{"id": obj["id"],
			"param": keyword, "value": val}
		ans := map[string]interface{}{"type": "interact", "data": data}

		//-1 since its not neccessary to check for filtering
		return Ogree3D.InformOptional("Interact", -1, ans)
	} else {
		return fmt.Errorf("The label value must be a string")
	}
}
