package controllers

import (
	"cli/utils"
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

func (controller Controller) UpdateInteract(path, attrName string, values []any, hasSharpe bool) error {
	if attrName != "labelFont" && len(values) != 1 {
		return fmt.Errorf("only 1 value expected")
	}
	switch attrName {
	case "displayContent", "alpha", "tilesName", "tilesColor", "U", "slots", "localCS":
		return controller.SetBooleanInteractAttribute(path, values, attrName, hasSharpe)
	case "label":
		return controller.SetLabel(path, values, hasSharpe)
	case "labelFont":
		return controller.SetLabelFont(path, values)
	case "labelBackground":
		return controller.SetLabelBackground(path, values)
	}
	return nil
}

func (controller Controller) SetLabel(path string, values []any, hasSharpe bool) error {
	value, err := utils.ValToString(values[0], "value")
	if err != nil {
		return err
	}
	return controller.InteractObject(path, "label", value, hasSharpe)
}

func (controller Controller) SetLabelFont(path string, values []any) error {
	msg := "The font can only be bold or italic" +
		" or be in the form of color@[colorValue]." +
		"\n\nFor more information please refer to: " +
		"\nhttps://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#interact-with-objects"

	switch len(values) {
	case 1:
		if values[0] != "bold" && values[0] != "italic" {
			return fmt.Errorf(msg)
		}
		return controller.InteractObject(path, "labelFont", values[0], false)
	case 2:
		if values[0] != "color" {
			return fmt.Errorf(msg)
		}
		c, ok := utils.ValToColor(values[1])
		if !ok {
			return fmt.Errorf("please provide a valid 6 length hex value for the color")
		}
		return controller.InteractObject(path, "labelFont", "color@"+c, false)
	default:
		return fmt.Errorf(msg)
	}
}

func (controller Controller) SetLabelBackground(path string, values []any) error {
	c, ok := utils.ValToColor(values[0])
	if !ok {
		return fmt.Errorf("please provide a valid 6 length hex value for the color")
	}
	return controller.InteractObject(path, "labelBackground", c, false)
}

func (controller Controller) SetBooleanInteractAttribute(path string, values []any, attrName string, hasSharpe bool) error {
	boolVal, err := utils.ValToBool(values[0], attrName)
	if err != nil {
		return err
	}
	return controller.InteractObject(path, attrName, boolVal, hasSharpe)
}
