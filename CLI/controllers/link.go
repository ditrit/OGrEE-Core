package controllers

import (
	"cli/models"
	"fmt"
	"net/http"
	"strings"
)

func (controller Controller) LinkObject(source string, destination string, attrs []string, values []any, slots []string) error {
	sourceUrl, err := controller.ObjectUrl(source, 0)
	if err != nil {
		return err
	}
	destPath, err := controller.SplitPath(destination)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(sourceUrl, "/api/stray_objects/") {
		return fmt.Errorf("only stray objects can be linked")
	}
	payload := map[string]any{"parentId": destPath.ObjectID}

	if slots != nil {
		if slots, err = models.ExpandStrVector(slots); err != nil {
			return err
		}
		payload["slot"] = slots
	}

	_, err = controller.API.Request("PATCH", sourceUrl+"/link", payload, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func (controller Controller) UnlinkObject(path string) error {
	sourceUrl, err := controller.ObjectUrl(path, 0)
	if err != nil {
		return err
	}
	_, err = controller.API.Request("PATCH", sourceUrl+"/unlink", nil, http.StatusOK)
	return err
}
