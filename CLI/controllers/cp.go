package controllers

import (
	"cli/models"
	"errors"
)

var ErrObjectCantBeCopied = errors.New("source object can not be copied")

func (controller Controller) Cp(source, dest string) error {
	object, err := controller.GetObject(source)
	if err != nil {
		return err
	}

	if !models.IsLayer(source) {
		return ErrObjectCantBeCopied
	}

	var destPath string

	if models.IsLayer(dest) {
		destPathObj, err := controller.SplitPath(dest)
		if err != nil {
			return err
		}

		object["slug"] = destPathObj.ObjectID
		destPath = dest
	} else {
		object["slug"] = dest
		destPath = models.LayersPath + dest
	}

	delete(object, "createdDate")
	delete(object, "lastUpdated")

	err = controller.PostObj(models.LAYER, models.EntityToString(models.LAYER), object, destPath)
	if err != nil {
		return err
	}

	return nil
}
