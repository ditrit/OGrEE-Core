package controllers

import (
	"cli/models"
	"fmt"
)

func (controller Controller) UIDelay(time float64) error {
	subdata := map[string]interface{}{"command": "delay", "data": time}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return controller.Ogree3D.Inform("HandleUI", -1, data)
}

func (controller Controller) UIToggle(feature string, enable bool) error {
	subdata := map[string]interface{}{"command": feature, "data": enable}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return controller.Ogree3D.Inform("HandleUI", -1, data)
}

func (controller Controller) UIHighlight(path string) error {
	obj, err := controller.GetObject(path)
	if err != nil {
		return err
	}

	subdata := map[string]interface{}{"command": "highlight", "data": obj["id"]}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return controller.Ogree3D.Inform("HandleUI", -1, data)
}

func (controller Controller) UIClearCache() error {
	subdata := map[string]interface{}{"command": "clearcache", "data": ""}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return controller.Ogree3D.Inform("HandleUI", -1, data)
}

func (controller Controller) CameraMove(command string, position []float64, rotation []float64) error {
	subdata := map[string]interface{}{"command": command}
	subdata["position"] = map[string]interface{}{"x": position[0], "y": position[1], "z": position[2]}
	subdata["rotation"] = map[string]interface{}{"x": rotation[0], "y": rotation[1]}
	data := map[string]interface{}{"type": "camera", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return controller.Ogree3D.Inform("HandleUI", -1, data)
}

func (controller Controller) CameraWait(time float64) error {
	subdata := map[string]interface{}{"command": "wait"}
	subdata["position"] = map[string]interface{}{"x": 0, "y": 0, "z": 0}
	subdata["rotation"] = map[string]interface{}{"x": 999, "y": time}
	data := map[string]interface{}{"type": "camera", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return controller.Ogree3D.Inform("HandleUI", -1, data)
}

func (controller Controller) FocusUI(path string) error {
	var id string
	if path != "" {
		obj, err := controller.GetObject(path)
		if err != nil {
			return err
		}
		category := models.EntityStrToInt(obj["category"].(string))
		if !models.IsPhysical(path) || category == models.SITE || category == models.BLDG || category == models.ROOM {
			msg := "You cannot focus on this object. Note you cannot" +
				" focus on Sites, Buildings and Rooms. " +
				"For more information please refer to the help doc  (man >)"
			return fmt.Errorf(msg)
		}
		id = obj["id"].(string)
	} else {
		id = ""
	}

	data := map[string]interface{}{"type": "focus", "data": id}
	err := controller.Ogree3D.Inform("FocusUI", -1, data)
	if err != nil {
		return err
	}

	if path != "" {
		return controller.CD(path)
	} else {
		fmt.Println("Focus is now empty")
	}

	return nil
}
