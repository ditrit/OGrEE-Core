package models

import (
	"encoding/json"
	"io/ioutil"
	u "p3/utils"
	"testing"
)

func TestValidateJsonSchema(t *testing.T) {
	testingEntities := []int{u.SITE, u.BLDG, u.ROOM, u.RACK, u.DEVICE, u.GROUP, u.BLDGTMPL}
	for _, entInt := range testingEntities {
		entStr := u.EntityToString(entInt)
		println("*** Testing " + entStr)
		var obj map[string]interface{}
		data, e := ioutil.ReadFile("schemas/" + entStr + "_schema.json")
		if e != nil {
			t.Error(e.Error())
		}
		json.Unmarshal(data, &obj)
		resp, ok := validateJsonSchema(entInt, obj["examples"].([]interface{})[0].(map[string]interface{}))
		if !ok {
			t.Errorf("Error validating json schema: %s", resp)
		}
	}
}
