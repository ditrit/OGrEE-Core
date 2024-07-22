package models

import (
	"p3/repository"
	u "p3/utils"
)

// GetSiteParentAttribute: search for the object of given ID,
// then search for its site parent and return its requested attribute
func GetSiteParentAttribute(id string, attribute string) (map[string]any, *u.Error) {
	data, err := repository.GetObjectSiteParent(id)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, &u.Error{Type: u.ErrNotFound, Message: "No object found with given id"}
	} else if attribute == "sitecolors" {
		resp := map[string]any{}
		for _, colorName := range []string{"reservedColor", "technicalColor", "usableColor"} {
			if color := data["attributes"].(map[string]interface{})[colorName]; color != nil {
				resp[colorName] = color
			} else {
				resp[colorName] = ""
			}
		}
		return resp, nil
	} else if attrValue := data["attributes"].(map[string]interface{})[attribute]; attrValue == nil {
		return nil, &u.Error{Type: u.ErrNotFound,
			Message: "Parent site has no temperatureUnit in attributes"}
	} else {
		return map[string]any{attribute: attrValue}, nil
	}
}
