package views

import (
	"cli/models"
	"encoding/json"
	"fmt"
	"sort"
)

func Object(path string, obj map[string]any) {
	if models.IsLayer(path) {
		obj[models.LayerApplicability] = models.PhysicalIDToPath(obj[models.LayerApplicability].(string))
	}

	DisplayJson("", obj)
}

func DisplayJson(indent string, jsonMap map[string]any) {
	defaultIndent := "    "
	// sort keys in alphabetical order
	keys := make([]string, 0, len(jsonMap))
	for k := range jsonMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// print map
	println("{")
	for _, key := range keys {
		if key == "attributes" {
			print(defaultIndent + "\"" + key + "\": ")
			DisplayJson(defaultIndent, jsonMap[key].(map[string]any))
		} else {
			print(indent + defaultIndent + "\"" + key + "\": ")
			if value, err := json.Marshal(jsonMap[key]); err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%v\n", string(value))
			}
		}
	}
	println(indent + "}")
}
