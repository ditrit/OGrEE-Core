package models

import (
	u "p3/utils"
	"strings"
)

// Returns the json schema content of the requested id (schema name)
func GetSchemaFile(id string) ([]byte, *u.Error) {
	var schemaPath = "schemas/"

	if strings.HasSuffix(id, ".json") {
		if id == "types.json" || id == "base.json" {
			schemaPath = schemaPath + "refs/"
		}
		file, err := embeddfs.ReadFile(schemaPath + id)
		if err == nil {
			return file, nil
		} else {
			// not found
			println(err.Error())
			return nil, &u.Error{Type: u.ErrNotFound, Message: "Requested file not found"}
		}
	} else {
		//only json accepted
		return nil, &u.Error{Type: u.ErrInvalidValue, Message: "Only files with .json extension"}
	}
}
