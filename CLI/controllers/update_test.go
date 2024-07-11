package controllers_test

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"maps"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// region tags

func TestUpdateTagColor(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	oldSlug := "slug"
	path := models.TagsPath + oldSlug

	test_utils.MockGetObjectByEntity(mockAPI, "tags", map[string]any{
		"slug":        oldSlug,
		"description": "description",
		"color":       "aaaaaa",
	})

	dataUpdate := map[string]any{
		"color": "aaaaab",
	}
	dataUpdated := map[string]any{
		"slug":        oldSlug,
		"description": "description",
		"color":       "aaaaab",
	}

	test_utils.MockUpdateObject(mockAPI, dataUpdate, dataUpdated)
	_, err := controller.PatchObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

func TestUpdateTagSlug(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	oldSlug := "slug"
	newSlug := "new-slug"

	path := models.TagsPath + oldSlug

	test_utils.MockGetObjectByEntity(mockAPI, "tags", map[string]any{
		"slug":        oldSlug,
		"description": "description",
		"color":       "aaaaaa",
	})

	dataUpdate := map[string]any{
		"slug": newSlug,
	}
	dataUpdated := map[string]any{
		"slug":        newSlug,
		"description": "description",
		"color":       "aaaaaa",
	}

	test_utils.MockUpdateObject(mockAPI, dataUpdate, dataUpdated)
	_, err := controller.PatchObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

//endregion tags
// region device's sizeU

// Test an update of a device's sizeU with heightUnit == mm
func TestUpdateDeviceSizeUmm(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	path := models.PhysicalIDToPath("BASIC.A.R1.A01.chU1")

	device := test_utils.GetEntity("device", "chU1", "BASIC.A.R1.A01", "test")
	test_utils.MockGetObject(mockAPI, device)

	dataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU": 1,
		},
	}
	mockDataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU":  float64(1),
			"height": 44.45,
		},
	}
	device["attributes"].(map[string]any)["sizeU"] = 1
	device["attributes"].(map[string]any)["height"] = 44.45

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, device)
	_, err := controller.PatchObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

// Test an update of a device's sizeU with heightUnit == cm
func TestUpdateDeviceSizeUcm(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	path := models.PhysicalIDToPath("BASIC.A.R1.A01.chU1")

	device := test_utils.GetEntity("device", "chU1", "BASIC.A.R1.A01", "test")
	device["attributes"].(map[string]any)["heightUnit"] = "cm"

	test_utils.MockGetObject(mockAPI, device)

	dataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU": 1,
		},
	}
	mockDataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU":  float64(1),
			"height": 4.445,
		},
	}
	device["attributes"].(map[string]any)["sizeU"] = 1
	device["attributes"].(map[string]any)["height"] = 4.445

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, device)
	_, err := controller.PatchObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

// endregion sizeU
// region device's height

// Test an update of a device's height with heightUnit == mm
func TestUpdateDeviceheightmm(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	path := models.PhysicalIDToPath("BASIC.A.R1.A01.chU1")

	device := test_utils.GetEntity("device", "chU1", "BASIC.A.R1.A01", "test")
	test_utils.MockGetObject(mockAPI, device)

	dataUpdate := map[string]any{
		"attributes": map[string]any{
			"height": 44.45,
		},
	}
	mockDataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU":  float64(1),
			"height": 44.45,
		},
	}
	device["attributes"].(map[string]any)["sizeU"] = 1
	device["attributes"].(map[string]any)["height"] = 44.45

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, device)
	_, err := controller.PatchObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

// Test an update of a device's height with heightUnit == cm
func TestUpdateDeviceheightcm(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	path := models.PhysicalIDToPath("BASIC.A.R1.A01.chU1")

	device := test_utils.GetEntity("device", "chU1", "BASIC.A.R1.A01", "test")
	device["attributes"].(map[string]any)["heightUnit"] = "cm"
	test_utils.MockGetObject(mockAPI, device)

	dataUpdate := map[string]any{
		"attributes": map[string]any{
			"height": 4.445,
		},
	}
	mockDataUpdate := map[string]any{
		"attributes": map[string]any{
			"sizeU":  float64(1),
			"height": 4.445,
		},
	}
	device["attributes"].(map[string]any)["sizeU"] = 1
	device["attributes"].(map[string]any)["height"] = 4.445

	test_utils.MockUpdateObject(mockAPI, mockDataUpdate, device)
	_, err := controller.PatchObj(path, dataUpdate, false)
	assert.Nil(t, err)
}

// endregion
// region update attribute

func TestUpdateDeviceDescription(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	device := test_utils.CopyMap(chassis)
	device["description"] = "my old description"
	updatedDevice := test_utils.CopyMap(device)
	updatedDevice["description"] = "my new description"
	dataUpdate := map[string]any{"description": "my new description"}

	path := "/Physical/" + strings.Replace(device["id"].(string), ".", "/", -1)

	test_utils.MockGetObject(mockAPI, device)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedDevice)

	result, err := controller.UpdateDescription(path, "description", []any{updatedDevice["description"]})
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["description"], updatedDevice["description"])
}

func TestUpdateDeviceAttribute(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	device := test_utils.CopyMap(chassis)
	updatedDevice := test_utils.CopyMap(device)
	updatedDevice["attributes"].(map[string]any)["slot"] = []any{"slot1"}
	dataUpdate := map[string]any{"attributes": map[string]any{"slot": []string{"slot1"}}}

	path := "/Physical/" + strings.Replace(device["id"].(string), ".", "/", -1)

	test_utils.MockGetObject(mockAPI, device)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedDevice)

	result, err := controller.UpdateAttributes(path, "slot", updatedDevice["attributes"].(map[string]any)["slot"].([]any))
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["attributes"], updatedDevice["attributes"])
}

func TestUpdateGroupContent(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	group := test_utils.CopyMap(rackGroup)
	group["attributes"] = map[string]any{
		"content": "A,B",
	}
	newValue := "A,B,C"
	updatedGroup := test_utils.CopyMap(group)
	updatedGroup["attributes"].(map[string]any)["content"] = newValue
	dataUpdate := updatedGroup["attributes"].(map[string]any)

	path := "/Physical/" + strings.Replace(group["id"].(string), ".", "/", -1)

	test_utils.MockGetObject(mockAPI, group)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedGroup)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	result, err := controller.PatchObj(path, dataUpdate, false)
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["attributes"].(map[string]any)["content"].(string), newValue)
}

// endregion
// region update virtual

func TestAddVirtualConfig(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	device := test_utils.CopyMap(chassis)
	updatedDevice := test_utils.CopyMap(device)
	vconfig := map[string]any{"type": "node", "clusterId": "mycluster", "role": "proxmox"}
	updatedDevice["attributes"].(map[string]any)[controllers.VirtualConfigAttr] = vconfig
	dataUpdate := map[string]any{"attributes": map[string]any{controllers.VirtualConfigAttr: vconfig}}

	path := "/Physical/" + strings.Replace(device["id"].(string), ".", "/", -1)

	test_utils.MockGetObject(mockAPI, device)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedDevice)

	err := controller.UpdateObject(path, controllers.VirtualConfigAttr, []any{"node", "mycluster", "proxmox"})
	assert.Nil(t, err)
}

func TestUpdateVirtualConfigData(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	// original device
	device := test_utils.CopyMap(chassis)
	vconfig := map[string]any{"type": "node", "clusterId": "mycluster", "role": "proxmox"}
	device["attributes"].(map[string]any)[controllers.VirtualConfigAttr] = test_utils.CopyMap(vconfig)
	// updated device
	updatedDevice := test_utils.CopyMap(device)
	vconfig["type"] = "host"
	updatedDevice["attributes"].(map[string]any)[controllers.VirtualConfigAttr] = vconfig
	// update data
	dataUpdate := map[string]any{"attributes": updatedDevice["attributes"]}

	path := "/Physical/" + strings.Replace(device["id"].(string), ".", "/", -1)

	test_utils.MockGetObject(mockAPI, device)
	test_utils.MockGetObject(mockAPI, device)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedDevice)

	err := controller.UpdateObject(path, controllers.VirtualConfigAttr+".type", []any{"host"})
	assert.Nil(t, err)
}

func TestUpdateVirtualLink(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	vobj := test_utils.CopyMap(vobjCluster)
	updatedDevice := test_utils.CopyMap(vobj)
	updatedDevice["attributes"].(map[string]any)["vlinks"] = []any{"device"}
	dataUpdate := map[string]any{"attributes": map[string]any{"vlinks": []any{"device"}}}

	path := "/Logical/VirtualObjects/" + strings.Replace(vobj["id"].(string), ".", "/", -1)

	// Add vlink
	test_utils.MockGetVirtualObject(mockAPI, vobj)
	test_utils.MockGetVirtualObject(mockAPI, vobj)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedDevice)

	result, err := controller.UpdateVirtualLink(path, "vlinks+", "device")
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["attributes"], updatedDevice["attributes"])

	// Remove vlink
	test_utils.MockGetVirtualObject(mockAPI, updatedDevice)
	test_utils.MockGetVirtualObject(mockAPI, updatedDevice)
	dataUpdate = map[string]any{"attributes": map[string]any{"vlinks": []any{}}}
	test_utils.MockUpdateObject(mockAPI, dataUpdate, vobj)

	result, err = controller.UpdateVirtualLink(path, "vlinks-", "device")
	assert.Nil(t, err)
	assert.Equal(t, result["data"].(map[string]any)["attributes"], vobj["attributes"])
}

// endregion
// region update inner attr object

func TestUpdateRackBreakerData(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	// original device
	rack := test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
	path := "/Physical/site/building/room/rack"
	breakers := map[string]any{"break1": map[string]any{"powerpanel": "panel1"}}
	rack["attributes"].(map[string]any)[controllers.BreakerAttr+"s"] = test_utils.CopyMap(breakers)
	// updated device
	updatedRack := test_utils.CopyMap(rack)
	breakers["break1"].(map[string]any)["powerpanel"] = "panel2"
	updatedRack["attributes"].(map[string]any)[controllers.BreakerAttr+"s"] = breakers
	// update data
	dataUpdate := map[string]any{"attributes": updatedRack["attributes"]}

	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockUpdateObject(mockAPI, dataUpdate, updatedRack)

	err := controller.UpdateObject(path, controllers.BreakerAttr+"s.break1.powerpanel", []any{"panel2"})
	assert.Nil(t, err)
}

func TestAddInnerAtrObjWorks(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	tests := []struct {
		name          string
		addFunction   func(string, string, []any) error
		attr          string
		values        []any
		newAttributes map[string]any
	}{
		{"AddRoomSeparator", controller.UpdateObject, controllers.SeparatorAttr, []any{"mySeparator", []float64{1., 2.}, []float64{1., 2.}, "wireframe"}, map[string]interface{}{"separators": map[string]interface{}{"mySeparator": models.Separator{StartPos: []float64{1, 2}, EndPos: []float64{1, 2}, Type: "wireframe"}}}},
		{"AddRoomPillar", controller.UpdateObject, controllers.PillarAttr, []any{"myPillar", []float64{1., 2.}, []float64{1., 2.}, 2.5}, map[string]interface{}{"pillars": map[string]interface{}{"myPillar": models.Pillar{CenterXY: []float64{1, 2}, SizeXY: []float64{1, 2}, Rotation: 2.5}}}},
		{"AddRackBraker", controller.UpdateObject, controllers.BreakerAttr, []any{"myBreaker", "powerpanel"}, map[string]interface{}{"breakers": map[string]interface{}{"myBreaker": models.Breaker{Powerpanel: "powerpanel"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetObj map[string]any
			var target string
			if tt.attr == controllers.BreakerAttr {
				targetObj = test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
				target = "/Physical/site/building/room/rack"
			} else {
				targetObj = test_utils.GetEntity("room", "room", "site.building", "domain")
				target = "/Physical/site/building/room"
			}

			test_utils.MockGetObject(mockAPI, targetObj)
			test_utils.MockGetObject(mockAPI, targetObj)

			targetObj["attributes"] = tt.newAttributes
			test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": tt.newAttributes}, targetObj)

			err := tt.addFunction(target, tt.attr+"s+", tt.values)
			assert.Nil(t, err)
		})
	}
}

func TestAddInnerAtrObjTargetError(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	tests := []struct {
		name         string
		addFunction  func(string, string, []any) (map[string]any, error)
		attr         string
		values       []any
		errorMessage string
	}{
		{"AddRoomSeparator", controller.AddInnerAtrObj, "separator", []any{"mySeparator"}, "this attribute can only be added to rooms"},
		{"AddRoomPillar", controller.AddInnerAtrObj, "pillar", []any{"myPillar"}, "this attribute can only be added to rooms"},
		{"AddRackBraker", controller.AddInnerAtrObj, "breaker", []any{"myBreaker"}, "this attribute can only be added to racks"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetObj map[string]any
			var target string
			if tt.attr != controllers.BreakerAttr { // inverted compared to the right one
				targetObj = test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
				target = "/Physical/site/building/room/rack"
			} else {
				targetObj = test_utils.GetEntity("room", "room", "site.building", "domain")
				target = "/Physical/site/building/room"
			}

			test_utils.MockGetObject(mockAPI, targetObj)
			obj, err := tt.addFunction(tt.attr, target, tt.values)

			assert.Nil(t, obj)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestAddInnerAtrObjFormatError(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	tests := []struct {
		name         string
		addFunction  func(string, string, []any) (map[string]any, error)
		attr         string
		values       []any
		errorMessage string
	}{
		{"AddRoomSeparator", controller.AddInnerAtrObj, "separator", []any{"mySeparator"}, "4 values (name, startPos, endPos, type) expected to add a separator"},
		{"AddRoomPillar", controller.AddInnerAtrObj, "pillar", []any{"myPillar"}, "4 values (name, centerXY, sizeXY, rotation) expected to add a pillar"},
		{"AddRackBraker", controller.AddInnerAtrObj, "breaker", []any{"myBreaker"}, "at least 2 values (name and powerpanel) expected to add a breaker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetObj map[string]any
			var target string
			if tt.attr == controllers.BreakerAttr {
				targetObj = test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
				target = "/Physical/site/building/room/rack"
			} else {
				targetObj = test_utils.GetEntity("room", "room", "site.building", "domain")
				target = "/Physical/site/building/room"
			}
			test_utils.MockGetObject(mockAPI, targetObj)

			obj, err := tt.addFunction(tt.attr, target, tt.values)

			assert.Nil(t, obj)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestDeleteInnerAtrObjWithError(t *testing.T) {
	tests := []struct {
		name          string
		attributeName string
		separatorName string
		errorMessage  string
	}{
		{"InvalidArgument", "other", "separator", "others separator does not exist"},
		{"SeparatorDoesNotExist", "separator", "mySeparator", "separators mySeparator does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

			room := test_utils.GetEntity("room", "room", "site.building", "domain")
			test_utils.MockGetObject(mockAPI, room)
			obj, err := controller.DeleteInnerAttrObj("/Physical/site/building/room", tt.attributeName+"s", tt.separatorName)

			assert.Nil(t, obj)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestDeleteInnerAtrObjWorks(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)
	tests := []struct {
		name        string
		addFunction func(string, string, []any) (map[string]any, error)
		attr        string
		// values        []any
		currentAttributes map[string]any
	}{
		{"DeleteRoomSeparator", controller.AddInnerAtrObj, controllers.SeparatorAttr, map[string]interface{}{"separators": map[string]interface{}{"myseparator": models.Separator{StartPos: []float64{1, 2}, EndPos: []float64{1, 2}, Type: "wireframe"}}}},
		{"DeleteRoomPillar", controller.AddInnerAtrObj, controllers.PillarAttr, map[string]interface{}{"pillars": map[string]interface{}{"mypillar": models.Pillar{CenterXY: []float64{1, 2}, SizeXY: []float64{1, 2}, Rotation: 2.5}}}},
		{"DeleteRackBraker", controller.AddInnerAtrObj, controllers.BreakerAttr, map[string]interface{}{"breakers": map[string]interface{}{"mybreaker": models.Breaker{Powerpanel: "powerpanel"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targetObj map[string]any
			var target string
			if tt.attr == controllers.BreakerAttr {
				targetObj = test_utils.GetEntity("rack", "rack", "site.building.room", "domain")
				target = "/Physical/site/building/room/rack"
			} else {
				targetObj = test_utils.GetEntity("room", "room", "site.building", "domain")
				target = "/Physical/site/building/room"
			}

			maps.Copy(targetObj["attributes"].(map[string]any), tt.currentAttributes)
			// room["attributes"].(map[string]any)["pillars"] = map[string]interface{}{"myPillar": models.Pillar{CenterXY: []float64{1, 2}, SizeXY: []float64{1, 2}, Rotation: 2.5}}

			updatedTarget := maps.Clone(targetObj)
			updatedTarget["attributes"] = map[string]any{tt.attr + "s": map[string]interface{}{}}

			test_utils.MockGetObject(mockAPI, targetObj)
			test_utils.MockGetObject(mockAPI, targetObj)
			test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": map[string]interface{}{tt.attr + "s": map[string]interface{}{}}}, updatedTarget)
			err := controller.UpdateObject(target, tt.attr+"s-", []any{"my" + tt.attr})

			assert.Nil(t, err)
		})
	}
}

// endregion
// region room areas

func TestSetRoomAreas(t *testing.T) {
	controller, mockAPI, _, _ := test_utils.NewControllerWithMocks(t)

	room := test_utils.GetEntity("room", "room", "site.building", "domain")

	roomResponse := test_utils.GetEntity("room", "room", "site.building", "domain")
	test_utils.MockGetObject(mockAPI, room)

	roomResponse["attributes"] = map[string]any{
		"reserved":  []float64{1, 2, 3, 4},
		"technical": []float64{1, 2, 3, 4},
	}
	test_utils.MockUpdateObject(mockAPI, map[string]interface{}{"attributes": map[string]interface{}{"reserved": []float64{1, 2, 3, 4}, "technical": []float64{1, 2, 3, 4}}}, roomResponse)

	reservedArea := []float64{1, 2, 3, 4}
	technicalArea := []float64{1, 2, 3, 4}
	value, err := controller.UpdateRoomAreas("/Physical/site/building/room", []any{reservedArea, technicalArea})

	assert.Nil(t, err)
	assert.NotNil(t, value)
}

// endregion

func TestAddToMap(t *testing.T) {
	newMap, replaced := controllers.AddToMap[int](map[string]any{"a": 3}, "b", 10)

	assert.Equal(t, map[string]any{"a": 3, "b": 10}, newMap)
	assert.False(t, replaced)

	newMap, replaced = controllers.AddToMap[int](newMap, "b", 15)
	assert.Equal(t, map[string]any{"a": 3, "b": 15}, newMap)
	assert.True(t, replaced)
}
