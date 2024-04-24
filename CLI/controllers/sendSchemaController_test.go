package controllers_test

import (
	"cli/controllers"
	l "cli/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	l.InitLogs()
}

func TestMergeMaps(t *testing.T) {
	x := map[string]any{
		"a": "10",
		"b": "11",
	}
	y := map[string]any{
		"b": "25",
		"c": "40",
	}
	testMap := copyMap(x)
	controllers.MergeMaps(testMap, y, false)
	assert.Contains(t, testMap, "a")
	assert.Contains(t, testMap, "b")
	assert.Contains(t, testMap, "c")
	assert.Equal(t, x["a"], testMap["a"])
	assert.Equal(t, x["b"], testMap["b"])
	assert.Equal(t, y["c"], testMap["c"])

	testMap = copyMap(x)
	controllers.MergeMaps(testMap, y, true)
	assert.Contains(t, testMap, "a")
	assert.Contains(t, testMap, "b")
	assert.Contains(t, testMap, "c")
	assert.Equal(t, x["a"], testMap["a"])
	assert.Equal(t, y["b"], testMap["b"])
	assert.Equal(t, y["c"], testMap["c"])
}

func TestGenerateFilteredJson(t *testing.T) {
	controllers.State.DrawableJsons = map[string]map[string]any{
		"rack": map[string]any{
			"name":        true,
			"parentId":    true,
			"category":    true,
			"description": false,
			"domain":      true,
			"attributes": map[string]any{
				"color": true,
			},
		},
	}
	object := map[string]any{
		"name":        "rack",
		"parentId":    "site.building.room",
		"category":    "rack",
		"description": "",
		"domain":      "domain",
		"attributes": map[string]any{
			"color": "aaaaaa",
		},
	}

	filteredObject := controllers.GenerateFilteredJson(object)

	assert.Contains(t, filteredObject, "name")
	assert.Contains(t, filteredObject, "parentId")
	assert.Contains(t, filteredObject, "category")
	assert.Contains(t, filteredObject, "domain")
	assert.NotContains(t, filteredObject, "description")
	assert.Contains(t, filteredObject, "attributes")
	assert.Contains(t, filteredObject["attributes"], "color")
}

func TestStringify(t *testing.T) {
	assert.Equal(t, "text", controllers.Stringify("text"))
	assert.Equal(t, "35", controllers.Stringify(35))
	assert.Equal(t, "35", controllers.Stringify(35.0))
	assert.Equal(t, "true", controllers.Stringify(true))
	assert.Equal(t, "hello,world", controllers.Stringify([]string{"hello", "world"}))
	assert.Equal(t, "[45,21]", controllers.Stringify([]float64{45, 21}))
	assert.Equal(t, "[hello,5,[world,450]]", controllers.Stringify([]any{"hello", 5, []any{"world", 450}}))
	assert.Equal(t, "", controllers.Stringify(map[string]any{"hello": 5}))
}

func TestExpandSlotVector(t *testing.T) {
	slots, err := controllers.ExpandSlotVector([]string{"slot1..slot3", "slot4"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid device syntax: .. can only be used in a single element vector")

	slots, err = controllers.ExpandSlotVector([]string{"slot1..slot3..slot7"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid device syntax: incorrect use of .. for slot")

	slots, err = controllers.ExpandSlotVector([]string{"slot1..slots3"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid device syntax: incorrect use of .. for slot")

	slots, err = controllers.ExpandSlotVector([]string{"slot1..slotE"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid device syntax: incorrect use of .. for slot")

	slots, err = controllers.ExpandSlotVector([]string{"slot1..slot3"})
	assert.Nil(t, err)
	assert.NotNil(t, slots)
	assert.EqualValues(t, []string{"slot1", "slot2", "slot3"}, slots)

	slots, err = controllers.ExpandSlotVector([]string{"slot1", "slot3"})
	assert.Nil(t, err)
	assert.NotNil(t, slots)
	assert.EqualValues(t, []string{"slot1", "slot3"}, slots)
}
