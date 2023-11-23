package controllers

import (
	"cli/models"
	"fmt"
	"strings"
)

func (controller Controller) Select(path string) ([]string, error) {
	var selection []string
	var err error

	if strings.Contains(path, "*") || models.PathHasLayer(path) {
		_, selection, err = controller.GetObjectsWildcard(path)
	} else if path != "" {
		selection = []string{path}
		err = controller.CD(path)
	}

	if err != nil {
		return nil, err
	}

	return controller.SetClipBoard(selection)
}

func (controller Controller) SetClipBoard(x []string) ([]string, error) {
	State.ClipBoard = x
	var data map[string]interface{}

	if len(x) == 0 { //This means deselect
		data = map[string]interface{}{"type": "select", "data": "[]"}
		err := controller.Ogree3D.InformOptional("SetClipBoard", -1, data)
		if err != nil {
			return nil, fmt.Errorf("cannot reset clipboard : %s", err.Error())
		}
	} else {
		//Verify paths
		arr := []string{}
		for _, val := range x {
			obj, err := controller.GetObject(val)
			if err != nil {
				return nil, err
			}
			id, ok := obj["id"].(string)
			if ok {
				arr = append(arr, id)
			}
		}
		serialArr := "[\"" + strings.Join(arr, "\",\"") + "\"]"
		data = map[string]interface{}{"type": "select", "data": serialArr}
		err := controller.Ogree3D.InformOptional("SetClipBoard", -1, data)
		if err != nil {
			return nil, fmt.Errorf("cannot set clipboard : %s", err.Error())
		}
	}
	return State.ClipBoard, nil
}
