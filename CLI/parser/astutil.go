package parser

import (
	cmd "cli/controllers"
	"cli/models"
	"cli/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
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

// errResponder helper func for specialUpdateNode
// used for separator, pillar err msgs and validateAreas()
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

func filtersToMapString(filters map[string]node) (map[string]string, error) {
	filtersString := map[string]string{}

	for key := range filters {
		filterVal, err := filters[key].execute()
		if err != nil {
			return nil, err
		}
		filtersString[key] = filterVal.(string)
	}

	return filtersString, nil
}

type recursiveArgs struct {
	isRecursive bool
	minDepth    string
	maxDepth    string
}

func (args *recursiveArgs) toParams(path string) (*cmd.RecursiveParams, error) {
	if !args.isRecursive {
		return nil, nil
	}

	minDepth, err := stringToIntOr(args.minDepth, 0)
	if err != nil {
		return nil, err
	}

	maxDepth, err := stringToIntOr(args.maxDepth, models.UnlimitedDepth)
	if err != nil {
		return nil, err
	}

	return &cmd.RecursiveParams{
		PathEntered: path,
		MinDepth:    minDepth,
		MaxDepth:    maxDepth,
	}, nil
}

func stringToIntOr(value string, defaultValue int) (int, error) {
	if value != "" {
		return strconv.Atoi(value)
	}

	return defaultValue, nil
}

func addSizeOrTemplate(sizeOrTemplate node, attributes map[string]any, entity int) error {
	size, err := nodeToSize(sizeOrTemplate)
	if err == nil {
		attributes["size"] = size
		return nil
	}

	template, err := nodeToString(sizeOrTemplate, "template")
	if err != nil {
		if errors.Is(err, utils.ErrShouldBeAString) {
			return errors.New("vector3 (size) or string (template) expected")
		}

		return err
	}

	attributes["template"] = template

	return nil
}

func nodeToSize(sizeNode node) ([]float64, error) {
	return nodeToVec(sizeNode, 3, "size")
}

func nodeToPosXYZ(positionNode node) ([]float64, error) {
	position, err := nodeToVec(positionNode, -1, "position")
	if err != nil {
		return nil, err
	}

	if len(position) != 2 && len(position) != 3 {
		return nil, fmt.Errorf("position should be a vector2 or a vector3")
	}

	return position, nil
}
