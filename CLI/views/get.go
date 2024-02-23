package views

import (
	"cli/models"
	"encoding/json"
	"log"
	"os"
)

func Object(path string, obj map[string]any) {
	if models.IsLayer(path) {
		obj[models.LayerApplicability] = models.PhysicalIDToPath(obj[models.LayerApplicability].(string))
	}

	DisplayJson(obj)
}

func DisplayJson(jsonMap map[string]any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(jsonMap); err != nil {
		log.Fatal(err)
	}
}
