package controllers

import (
	"fmt"
	"strconv"
)

// Unity UI will draw already existing objects
// by retrieving the hierarchy. 'force' bool is useful
// for scripting where the user can 'force' input if
// the num objects to draw surpasses threshold
func (controller Controller) Draw(path string, depth int, force bool) error {
	paths, err := controller.UnfoldPath(path)
	if err != nil {
		return err
	}

	for _, path := range paths {
		obj, err := controller.GetObjectWithChildren(path, depth)
		if err != nil {
			return err
		}

		count := objectCounter(obj)
		okToGo := true
		if count > State.DrawThreshold && !force {
			msg := "You are about to send " + strconv.Itoa(count) +
				" objects to the Unity 3D client. " +
				"Do you want to continue ? (y/n)\n"
			(*State.Terminal).Write([]byte(msg))
			(*State.Terminal).SetPrompt(">")
			ans, _ := (*State.Terminal).Readline()
			if ans != "y" && ans != "Y" {
				okToGo = false
			}
		} else if force {
			okToGo = true
		} else if !force && count > State.DrawThreshold {
			okToGo = false
		}
		if okToGo {
			data := map[string]interface{}{"type": "create", "data": obj}
			//0 to include the JSON filtration
			unityErr := controller.Ogree3D.Inform("Draw", 0, data)
			if unityErr != nil {
				return unityErr
			}
		}
	}

	return nil
}

func objectCounter(parent map[string]interface{}) int {
	count := 0
	if parent != nil {
		count += 1
		if _, ok := parent["children"]; ok {
			if arr, ok := parent["children"].([]interface{}); ok {
				for _, childInf := range arr {
					if child, ok := childInf.(map[string]interface{}); ok {
						count += objectCounter(child)
					}
				}
			}
			if arr, ok := parent["children"].([]map[string]interface{}); ok {
				for _, child := range arr {
					count += objectCounter(child)
				}
			}
		}
	}
	return count
}

func (controller Controller) Undraw(path string) error {
	paths, err := controller.UnfoldPath(path)
	if err != nil {
		return err
	}

	for _, path := range paths {
		var id string
		if path == "" {
			id = ""
		} else {
			obj, err := controller.GetObject(path)
			if err != nil {
				return err
			}
			var ok bool
			id, ok = obj["id"].(string)
			if !ok {
				return fmt.Errorf("this object has no id")
			}
		}

		data := map[string]interface{}{"type": "delete", "data": id}

		err := controller.Ogree3D.Inform("Undraw", 0, data)
		if err != nil {
			return err
		}
	}

	return nil
}
