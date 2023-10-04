package main

import (
	cmd "cli/controllers"
	"cli/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func nodeToFloat(n node, name string) (float64, error) {
	val, err := n.execute()
	if err != nil {
		return 0, err
	}
	return utils.ValToFloat(val, name)
}

func nodeToNum(n node, name string) (any, error) {
	val, err := n.execute()
	if err != nil {
		return nil, err
	}
	return utils.ValToNum(val, name)
}

func nodeToInt(n node, name string) (int, error) {
	val, err := n.execute()
	if err != nil {
		return 0, err
	}
	return utils.ValToInt(val, name)
}

func nodeToBool(n node, name string) (bool, error) {
	val, err := n.execute()
	if err != nil {
		return false, err
	}
	return utils.ValToBool(val, name)
}

func nodeTo3dRotation(n node) ([]float64, error) {
	val, err := n.execute()
	if err != nil {
		return nil, err
	}
	return utils.ValTo3dRotation(val)
}

// Transforms a node into a slug:
//  1. Transform into a string
//  2. Transform into lower cases only
//  3. Validate the format of a slug using regex (lowercase letters, numbers and hyphens only)
func nodeToSlug(n node, name string) (string, error) {
	nodeString, err := nodeToString(n, name)
	if err != nil {
		return "", err
	}

	return stringToSlug(nodeString)
}

func stringToSlug(slug string) (string, error) {
	slug = strings.ToLower(slug)

	match, err := regexp.MatchString("^[a-z0-9-_]+$", slug)
	if err != nil {
		return "", err
	}

	if !match {
		return "", errors.New("slugs must have letters, numbers and hyphens only")
	}

	return slug, nil
}

func nodeToString(n node, name string) (string, error) {
	val, err := n.execute()
	if err != nil {
		return "", err
	}
	return utils.ValToString(val, name)
}

func nodeToVec(n node, size int, name string) ([]float64, error) {
	val, err := n.execute()
	if err != nil {
		return nil, err
	}

	return utils.ValToVec(val, size, name)
}

func nodeToColorString(colorNode node) (string, error) {
	colorInf, err := colorNode.execute()
	if err != nil {
		return "", err
	}

	color, ok := utils.ValToColor(colorInf)
	if !ok {
		return "", fmt.Errorf("Please provide a valid 6 digit Hex value for the color")
	}

	return color, nil
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
