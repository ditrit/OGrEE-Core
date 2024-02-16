package controllers

import (
	"fmt"
	"strings"
)

// Function called by InteractObject to extract the value of an attribute from the object
func valFromObj(obj map[string]any, val interface{}) (interface{}, error) {
	if value, ok := val.(string); ok {
		innerMap := obj["attributes"].(map[string]interface{})

		if _, ok := obj[value]; ok {
			val = obj[value]
		} else if _, ok := innerMap[value]; ok {
			val = innerMap[value]
		} else {
			msg := "The specified attribute '" + val.(string) + "' does not exist" +
				" in the object. \nPlease view the object" +
				" (ie. $> get) and try again"
			return "", fmt.Errorf(msg)
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
func (controller Controller) InteractObject(path string, keyword string, val interface{}, fromAttr bool) error {
	// First retrieve the object
	obj, err := controller.GetObject(path)
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
	} else if keyword == "label" {
		return fmt.Errorf("The label value must be a string")
	}

	data := map[string]interface{}{"id": obj["id"], "param": keyword, "value": val}
	ans := map[string]interface{}{"type": "interact", "data": data}

	//-1 since its not neccessary to check for filtering
	return Ogree3D.InformOptional("Interact", -1, ans)
}
