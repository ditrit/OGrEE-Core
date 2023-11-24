package controllers

import (
	"fmt"
	"strings"
)

func (controller Controller) Select(path string) ([]string, error) {
	paths, err := controller.UnfoldPath(path)
	if err != nil {
		return nil, err
	}

	if len(paths) == 1 && paths[0] == path {
		err = controller.CD(paths[0])
		if err != nil {
			return nil, err
		}
	}

	return controller.SetClipBoard(paths)
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
