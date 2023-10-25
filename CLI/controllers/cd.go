package controllers

import (
	"cli/models"
	"errors"
)

func (controller Controller) CD(path string) error {
	if State.DebugLvl >= 3 {
		println("THE PATH: ", path)
	}

	if models.PathIsLayer(path) {
		return errors.New("it is not possible to cd into a layer")
	}

	_, err := controller.Tree(path, 0)
	if err != nil {
		return err
	}
	State.PrevPath = State.CurrPath
	State.CurrPath = models.PathRemoveLayer(path)
	return nil
}
