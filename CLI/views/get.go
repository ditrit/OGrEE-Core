package views

import (
	"cli/models"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/elliotchance/pie/v2"
)

func Object(path string, obj map[string]any) {
	if models.IsLayer(path) {
		obj[models.LayerApplicability] = models.PhysicalIDToPath(obj[models.LayerApplicability].(string))
	}

	DisplayJson("", obj)
}

func DisplayJson(indent string, jsonMap map[string]any) {
	keysWithObjectsValue := []string{"attributes", "breakers", "pillars", "separators", "virtual_config"}
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
		thisLevelIndent := indent + defaultIndent
		if pie.Contains(keysWithObjectsValue, key) {
			print(thisLevelIndent + "\"" + key + "\": ")
			DisplayJson(thisLevelIndent, jsonMap[key].(map[string]any))
		} else {
			print(thisLevelIndent + "\"" + key + "\": ")
			if value, err := json.Marshal(jsonMap[key]); err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("%v\n", string(value))
			}
		}
	}
	println(indent + "}")
}
