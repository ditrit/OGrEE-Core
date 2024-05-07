package controllers_test

import (
	"cli/controllers"
	"cli/models"
	test_utils "cli/test"
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

// Test PWD
func TestPWD(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	controller.CD("/")
	location := controllers.PWD()
	assert.Equal(t, "/", location)

	test_utils.MockGetObject(mockAPI, rack1)
	path := "/Physical/" + strings.Replace(rack1["id"].(string), ".", "/", -1)
	err := controller.CD(path)
	assert.Nil(t, err)

	location = controllers.PWD()
	assert.Equal(t, path, location)
}

// Tests ObjectUrl
func TestObjectUrlInvalidPath(t *testing.T) {
	_, err := controllers.C.ObjectUrl("/invalid/path", 0)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid object path", err.Error())
}

func TestObjectUrlPaths(t *testing.T) {
	paths := map[string]any{
		models.StrayPath + "stray-object":                     "/api/stray-objects/stray-object",
		models.PhysicalPath + "BASIC/A":                       "/api/hierarchy-objects/BASIC.A",
		models.ObjectTemplatesPath + "my-template":            "/api/obj-templates/my-template",
		models.RoomTemplatesPath + "my-room-template":         "/api/room-templates/my-room-template",
		models.BuildingTemplatesPath + "my-building-template": "/api/bldg-templates/my-building-template",
		models.GroupsPath + "group1":                          "/api/groups/group1",
		models.TagsPath + "my-tag":                            "/api/tags/my-tag",
		models.LayersPath + "my-layer":                        "/api/layers/my-layer",
		models.DomainsPath + "domain1":                        "/api/domains/domain1",
		models.DomainsPath + "domain1/subdomain":              "/api/domains/domain1.subdomain",
	}

	for key, value := range paths {
		basePath, err := controllers.C.ObjectUrl(key, 0)
		assert.Nil(t, err)
		assert.Equal(t, value, basePath)
	}
}

// Tests ObjectUrlGeneric
func TestObjectUrlGenericInvalidPath(t *testing.T) {
	_, err := controllers.C.ObjectUrlGeneric("/invalid/path", 0, nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid object path", err.Error())
}

func TestObjectUrlGenericWithNoFilters(t *testing.T) {
	paths := []map[string]any{
		map[string]any{
			"basePath":  models.StrayPath,
			"objectId":  "stray-object",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "physical.stray",
		},
		map[string]any{
			"basePath":  models.PhysicalPath,
			"objectId":  "BASIC/A",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "physical.hierarchy",
		},
		map[string]any{
			"basePath":  models.ObjectTemplatesPath,
			"objectId":  "my-template",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.objtemplate",
		},
		map[string]any{
			"basePath":  models.RoomTemplatesPath,
			"objectId":  "my-room-template",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.roomtemplate",
		},
		map[string]any{
			"basePath":  models.BuildingTemplatesPath,
			"objectId":  "my-building-template",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.bldgtemplate",
		},
		map[string]any{
			"basePath":  models.GroupsPath,
			"objectId":  "group1",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "logical",
			"extraParams": map[string]any{
				"category": "group",
			},
		},
		map[string]any{
			"basePath":  models.TagsPath,
			"objectId":  "my-tag",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.tag",
		},
		map[string]any{
			"basePath":  models.LayersPath,
			"objectId":  "my-layer",
			"endpoint":  "/api/objects",
			"idName":    "slug",
			"namespace": "logical.layer",
		},
		map[string]any{
			"basePath":  models.DomainsPath,
			"objectId":  "domain1",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "organisational",
		},
		map[string]any{
			"basePath":  models.DomainsPath,
			"objectId":  "domain1/subdomain",
			"endpoint":  "/api/objects",
			"idName":    "id",
			"namespace": "organisational",
		},
	}
	for _, value := range paths {
		resultUrl, err := controllers.C.ObjectUrlGeneric(value["basePath"].(string)+value["objectId"].(string), 0, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, resultUrl)

		parsedUrl, _ := url.Parse(resultUrl)
		assert.Equal(t, value["endpoint"], parsedUrl.Path)
		assert.Equal(t, strings.Replace(value["objectId"].(string), "/", ".", -1), parsedUrl.Query().Get(value["idName"].(string)))
		assert.Equal(t, value["namespace"], parsedUrl.Query().Get("namespace"))

		if extraParams, ok := value["extraParams"]; ok {
			for k, v := range extraParams.(map[string]any) {
				assert.Equal(t, v, parsedUrl.Query().Get(k))
			}
		}
	}
}

func TestObjectUrlGenericWithNormalFilters(t *testing.T) {
	filters := map[string]string{
		"color": "00ED00",
	}
	id := "BASIC/A"
	resultUrl, err := controllers.C.ObjectUrlGeneric(models.PhysicalPath+id, 0, filters, nil)
	assert.Nil(t, err)
	assert.NotNil(t, resultUrl)

	parsedUrl, _ := url.Parse(resultUrl)
	assert.Equal(t, "/api/objects", parsedUrl.Path)
	assert.Equal(t, strings.Replace(id, "/", ".", -1), parsedUrl.Query().Get("id"))
	assert.Equal(t, "physical.hierarchy", parsedUrl.Query().Get("namespace"))
	assert.Equal(t, "00ED00", parsedUrl.Query().Get("color"))
}

func TestObjectUrlGenericWithFilterField(t *testing.T) {
	filters := map[string]string{
		"filter": "color=00ED00",
	}
	id := "BASIC/A"
	resultUrl, err := controllers.C.ObjectUrlGeneric(models.PhysicalPath+id, 0, filters, nil)
	assert.Nil(t, err)
	assert.NotNil(t, resultUrl)

	parsedUrl, _ := url.Parse(resultUrl)
	assert.Equal(t, "/api/objects/search", parsedUrl.Path)
	assert.Equal(t, strings.Replace(id, "/", ".", -1), parsedUrl.Query().Get("id"))
	assert.Equal(t, "physical.hierarchy", parsedUrl.Query().Get("namespace"))
}

// Tests GetSlot
func TestGetSlotWithNoTemplate(t *testing.T) {
	rack := map[string]any{
		"attributes": map[string]any{},
	}
	result, err := controllers.C.GetSlot(rack, "")
	assert.Nil(t, err)
	assert.Nil(t, result)

	rack["attributes"].(map[string]any)["template"] = ""
	result, err = controllers.C.GetSlot(rack, "")
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func TestGetSlotWithTemplateNonExistentSlot(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	template := map[string]any{
		"slug":        "rack-template",
		"description": "",
		"category":    "rack",
		"sizeWDHmm":   []any{605, 1200, 2003},
		"fbxModel":    "",
		"attributes": map[string]any{
			"vendor": "IBM",
			"model":  "9360-4PX",
		},
		"slots": []any{},
	}

	test_utils.MockGetObjTemplate(mockAPI, template)
	rack := map[string]any{
		"attributes": map[string]any{
			"template": "rack-template",
		},
	}
	_, err := controller.GetSlot(rack, "u02")
	assert.NotNil(t, err)
	assert.Equal(t, "the slot u02 does not exist", err.Error())
}

func TestGetSlotWithTemplateWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	slot := map[string]any{
		"location":   "u01",
		"type":       "u",
		"elemOrient": []any{33.3, -44.4, 107},
		"elemPos":    []any{58, 51, 44.45},
		"elemSize":   []any{482.6, 1138, 44.45},
		"mandatory":  "no",
		"labelPos":   "frontrear",
	}

	template := map[string]any{
		"slug":        "rack-template",
		"description": "",
		"category":    "rack",
		"sizeWDHmm":   []any{605, 1200, 2003},
		"fbxModel":    "",
		"attributes": map[string]any{
			"vendor": "IBM",
			"model":  "9360-4PX",
		},
		"slots": []any{
			slot,
		},
	}

	test_utils.MockGetObjTemplate(mockAPI, template)
	rack := map[string]any{
		"attributes": map[string]any{
			"template": "rack-template",
		},
	}
	result, err := controller.GetSlot(rack, "u01")
	assert.Nil(t, err)
	assert.Equal(t, slot["location"], result["location"])
}

// Tests UnsetAttribute
func TestUnsetAttributeObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockObjectNotFound(mockAPI, "/api/hierarchy-objects/BASIC.A.R1.A01")

	err := controller.UnsetAttribute("/Physical/BASIC/A/R1/A01", "color")
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestUnsetAttributeWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	rack := map[string]any{
		"category":    "rack",
		"id":          "BASIC.A.R1.A01",
		"name":        "A01",
		"parentId":    "BASIC.A.R1",
		"domain":      "test-domain",
		"description": "",
		"attributes": map[string]any{
			"height":     "47",
			"heightUnit": "U",
			"rotation":   `[45, 45, 45]`,
			"posXYZ":     `[4.6666666666667, -2, 0]`,
			"posXYUnit":  "m",
			"size":       `[1, 1]`,
			"sizeUnit":   "cm",
			"color":      "00ED00",
		},
	}
	updatedRack := test_utils.CopyMap(rack)
	delete(updatedRack["attributes"].(map[string]any), "color")
	delete(updatedRack, "id")

	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockPutObject(mockAPI, updatedRack, updatedRack)

	err := controller.UnsetAttribute("/Physical/BASIC/A/R1/A01", "color")
	assert.Nil(t, err)
}

// Tests UnsetInObj
func TestUnsetInObjInvalidIndex(t *testing.T) {
	controller, _, _ := layersSetup(t)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "color", -1)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Index out of bounds. Please provide an index greater than 0", err.Error())
}

func TestUnsetInObjObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockObjectNotFound(mockAPI, "/api/hierarchy-objects/BASIC.A.R1.A01")

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "color", 0)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "object not found", err.Error())
}

func TestUnsetInObjAttributeNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{}

	test_utils.MockGetObject(mockAPI, rack)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "color", 0)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Attribute :color was not found", err.Error())
}

func TestUnsetInObjAttributeNotAnArray(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{
		"color": "00ED00",
	}

	test_utils.MockGetObject(mockAPI, rack)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "color", 0)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Attribute is not an array", err.Error())
}

func TestUnsetInObjEmptyArray(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{
		"posXYZ": []any{},
	}

	test_utils.MockGetObject(mockAPI, rack)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "posXYZ", 0)
	assert.NotNil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "Cannot delete anymore elements", err.Error())
}

func TestUnsetInObjWorksWithNestedAttribute(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{
		"posXYZ": []any{1, 2, 3},
	}
	updatedRack := test_utils.CopyMap(rack1)
	updatedRack["attributes"] = map[string]any{
		"posXYZ": []any{1.0, 3.0},
	}
	delete(updatedRack, "children")

	test_utils.MockGetObject(mockAPI, rack)
	test_utils.MockPutObject(mockAPI, updatedRack, updatedRack)

	result, err := controller.UnsetInObj("/Physical/BASIC/A/R1/A01", "posXYZ", 1)
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func TestUnsetInObjWorksWithAttribute(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	template := map[string]any{
		"slug":            "small-room",
		"category":        "room",
		"axisOrientation": "+x+y",
		"sizeWDHm":        []any{9.6, 22.8, 3.0},
		"floorUnit":       "t",
		"technicalArea":   []any{5.0, 0.0, 0.0, 0.0},
		"reservedArea":    []any{3.0, 1.0, 1.0, 3.0},
		"colors": []any{
			map[string]any{
				"name":  "my-color1",
				"value": "00ED00",
			},
			map[string]any{
				"name":  "my-color2",
				"value": "ffffff",
			},
		},
	}
	updatedTemplate := test_utils.CopyMap(template)
	updatedTemplate["colors"] = slices.Delete(updatedTemplate["colors"].([]any), 1, 2)
	test_utils.MockPutObject(mockAPI, updatedTemplate, updatedTemplate)
	test_utils.MockGetRoomTemplate(mockAPI, template)

	result, err := controller.UnsetInObj(models.RoomTemplatesPath+"small-room", "colors", 1)
	assert.Nil(t, err)
	assert.Nil(t, result)
}

// Tests GetByAttr
func TestGetByAttrErrorWhenObjIsNotRack(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectHierarchy(mockAPI, chassis)

	err := controller.GetByAttr(models.PhysicalPath+"BASIC/A/R1/A01/chT", "colors")
	assert.NotNil(t, err)
	assert.Equal(t, "command may only be performed on rack objects", err.Error())
}

func TestGetByAttrErrorWhenObjIsRackWithSlotName(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["attributes"] = map[string]any{
		"slot": []any{
			map[string]any{
				"location":   "u01",
				"type":       "u",
				"elemOrient": []any{33.3, -44.4, 107},
				"elemPos":    []any{58, 51, 44.45},
				"elemSize":   []any{482.6, 1138, 44.45},
				"mandatory":  "no",
				"labelPos":   "frontrear",
				"color":      "@color1",
			},
		},
	}
	test_utils.MockGetObjectHierarchy(mockAPI, rack)

	err := controller.GetByAttr(models.PhysicalPath+"BASIC/A/R1/A01", "u01")
	assert.Nil(t, err)
}

func TestGetByAttrErrorWhenObjIsRackWithHeight(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	rack := test_utils.CopyMap(rack1)
	rack["height"] = "47"
	test_utils.MockGetObjectHierarchy(mockAPI, rack)

	err := controller.GetByAttr(models.PhysicalPath+"BASIC/A/R1/A01", 47)
	assert.Nil(t, err)
}

// Test UI (UIDelay, UIToggle, UIHighlight)
func TestUIDelay(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	// ogree3D.
	time := 15.0
	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": "delay",
			"data":    time,
		},
	}
	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.UIDelay(time)
	assert.Nil(t, err)
}

func TestUIToggle(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	// ogree3D.
	feature := "feature"
	enable := true
	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": feature,
			"data":    enable,
		},
	}

	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.UIToggle(feature, enable)
	assert.Nil(t, err)
}

func TestUIHighlightObjectNotFound(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	path := "/api/hierarchy-objects/BASIC.A.R1.A01"

	test_utils.MockObjectNotFound(mockAPI, path)

	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": "highlight",
			"data":    "BASIC.A.R1.A01",
		},
	}

	ogree3D.AssertNotCalled(t, "HandleUI", -1, data)
	err := controller.UIHighlight("/Physical/BASIC/A/R1/A01")
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestUIHighlightWorks(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": "highlight",
			"data":    rack1["id"],
		},
	}

	test_utils.MockGetObject(mockAPI, rack1)
	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.UIHighlight("/Physical/BASIC/A/R1/A01")
	assert.Nil(t, err)
}

func TestUIClearCache(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "ui",
		"data": map[string]interface{}{
			"command": "clearcache",
			"data":    "",
		},
	}

	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.UIClearCache()
	assert.Nil(t, err)
}

func TestCameraMove(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "camera",
		"data": map[string]interface{}{
			"command":  "move",
			"position": map[string]interface{}{"x": 0.0, "y": 1.0, "z": 2.0},
			"rotation": map[string]interface{}{"x": 0.0, "y": 0.0},
		},
	}

	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.CameraMove("move", []float64{0, 1, 2}, []float64{0, 0})
	assert.Nil(t, err)
}

func TestCameraWait(t *testing.T) {
	controller, _, ogree3D := layersSetup(t)
	time := 15.0
	data := map[string]interface{}{
		"type": "camera",
		"data": map[string]interface{}{
			"command":  "wait",
			"position": map[string]interface{}{"x": 0, "y": 0, "z": 0},
			"rotation": map[string]interface{}{"x": 999, "y": time},
		},
	}

	ogree3D.On("Inform", "HandleUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.CameraWait(time)
	assert.Nil(t, err)
}

func TestFocusUIObjectNotFound(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)

	test_utils.MockObjectNotFound(mockAPI, "/api/hierarchy-objects/"+rack1["id"].(string))
	err := controller.FocusUI("/Physical/" + strings.Replace(rack1["id"].(string), ".", "/", -1))
	ogree3D.AssertNotCalled(t, "Inform", "mock.Anything", "mock.Anything", "mock.Anything")
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestFocusUIEmptyPath(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "focus",
		"data": "",
	}

	ogree3D.On("Inform", "FocusUI", -1, data).Return(nil).Once() // The inform should be called once
	err := controller.FocusUI("")
	mockAPI.AssertNotCalled(t, "Request", "GET", "mock.Anything", "mock.Anything", "mock.Anything")
	assert.Nil(t, err)
}

func TestFocusUIErrorWithRoom(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	errorMessage := "You cannot focus on this object. Note you cannot focus on Sites, Buildings and Rooms. "
	errorMessage += "For more information please refer to the help doc  (man >)"

	test_utils.MockGetObject(mockAPI, roomWithoutChildren)
	err := controller.FocusUI("/Physical/" + strings.Replace(roomWithoutChildren["id"].(string), ".", "/", -1))
	ogree3D.AssertNotCalled(t, "Inform", "mock.Anything", "mock.Anything", "mock.Anything")
	assert.NotNil(t, err)
	assert.Equal(t, errorMessage, err.Error())
}

func TestFocusUIWorks(t *testing.T) {
	controller, mockAPI, ogree3D := layersSetup(t)
	data := map[string]interface{}{
		"type": "focus",
		"data": rack1["id"],
	}

	ogree3D.On("Inform", "FocusUI", -1, data).Return(nil).Once() // The inform should be called once
	// Get Object will be called two times: Once in FocusUI and a second time in FocusUI->CD->Tree
	test_utils.MockGetObject(mockAPI, rack1)
	test_utils.MockGetObject(mockAPI, rack1)
	err := controller.FocusUI("/Physical/" + strings.Replace(rack1["id"].(string), ".", "/", -1))
	assert.Nil(t, err)
}

// Tests LinkObject
func TestLinkObjectErrorNotStaryObject(t *testing.T) {
	controller, _, _ := layersSetup(t)

	err := controller.LinkObject(models.PhysicalPath+"BASIC/A/R1/A01", models.PhysicalPath+"BASIC/A/R1/A01", []string{}, []any{}, []string{})
	assert.NotNil(t, err)
	assert.Equal(t, "only stray objects can be linked", err.Error())
}

func TestLinkObjectWithoutSlots(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	strayDevice := test_utils.CopyMap(chassis)
	delete(strayDevice, "id")
	delete(strayDevice, "parentId")
	response := map[string]any{"message": "successfully linked"}
	body := map[string]any{"parentId": "BASIC.A.R1.A01", "slot": []string{}}

	test_utils.MockUpdateObject(mockAPI, body, response)

	slots := []string{}
	attributes := []string{}
	values := []any{}
	for key, value := range strayDevice["attributes"].(map[string]any) {
		attributes = append(attributes, key)
		values = append(values, value)
	}
	err := controller.LinkObject(models.StrayPath+"chT", models.PhysicalPath+"BASIC/A/R1/A01", attributes, values, slots)
	assert.Nil(t, err)
}

func TestLinkObjectWithInvalidSlots(t *testing.T) {
	controller, _, _ := layersSetup(t)

	strayDevice := test_utils.CopyMap(chassis)
	delete(strayDevice, "id")
	delete(strayDevice, "parentId")

	slots := []string{"slot01..slot03", "slot4"}
	attributes := []string{}
	values := []any{}
	for key, value := range strayDevice["attributes"].(map[string]any) {
		attributes = append(attributes, key)
		values = append(values, value)
	}
	err := controller.LinkObject(models.StrayPath+"chT", models.PhysicalPath+"BASIC/A/R1/A01", attributes, values, slots)
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid device syntax: .. can only be used in a single element vector", err.Error())
}

func TestLinkObjectWithValidSlots(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	strayDevice := test_utils.CopyMap(chassis)
	delete(strayDevice, "id")
	delete(strayDevice, "parentId")
	response := map[string]any{"message": "successfully linked"}
	body := map[string]any{"parentId": "BASIC.A.R1.A01", "slot": []string{"slot01"}}

	test_utils.MockUpdateObject(mockAPI, body, response)

	slots := []string{"slot01"}
	attributes := []string{}
	values := []any{}
	for key, value := range strayDevice["attributes"].(map[string]any) {
		attributes = append(attributes, key)
		values = append(values, value)
	}
	err := controller.LinkObject(models.StrayPath+"chT", models.PhysicalPath+"BASIC/A/R1/A01", attributes, values, slots)
	assert.Nil(t, err)
}

// Tests UnlinkObject
func TestUnlinkObjectWithInvalidPath(t *testing.T) {
	controller, _, _ := layersSetup(t)

	err := controller.UnlinkObject("/invalid/path")
	assert.NotNil(t, err)
	assert.Equal(t, "invalid object path", err.Error())
}

func TestUnlinkObjectWithValidPath(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockUpdateObject(mockAPI, nil, map[string]any{"message": "successfully unlinked"})

	err := controller.UnlinkObject(models.PhysicalPath + "BASIC/A/R1/A01")
	assert.Nil(t, err)
}

// Tests IsEntityDrawable
func TestIsEntityDrawableObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockObjectNotFound(mockAPI, "/api/hierarchy-objects/BASIC.A.R1.A01")

	isDrawable, err := controller.IsEntityDrawable(models.PhysicalPath + "BASIC/A/R1/A01")
	assert.False(t, isDrawable)
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestIsEntityDrawable(t *testing.T) {
	tests := []struct {
		name               string
		drawableObjects    []int
		expectedIsDrawable bool
	}{
		{"CategoryIsNotDrawable", []int{models.EntityStrToInt("device")}, false},
		{"CategoryIsDrawable", []int{models.EntityStrToInt("rack")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)
			controllers.State.DrawableObjs = tt.drawableObjects

			test_utils.MockGetObject(mockAPI, rack1)

			isDrawable, err := controller.IsEntityDrawable(models.PhysicalPath + "BASIC/A/R1/A01")
			assert.Equal(t, tt.expectedIsDrawable, isDrawable)
			assert.Nil(t, err)
		})
	}
}

// Tests IsAttrDrawable (and IsCategoryAttrDrawable)
func TestIsAttrDrawableObjectNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	path := "/api/hierarchy-objects/BASIC.A.R1.A01"

	test_utils.MockObjectNotFound(mockAPI, path)

	isAttrDrawable, err := controller.IsAttrDrawable(models.PhysicalPath+"BASIC/A/R1/A01", "color")
	assert.False(t, isAttrDrawable)
	assert.NotNil(t, err)
	assert.Equal(t, "object not found", err.Error())
}

func TestIsAttrDrawableTemplateJsonIsNil(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)
	controllers.State.DrawableObjs = []int{models.EntityStrToInt("rack")}

	controllers.State.DrawableJsons = map[string]map[string]any{
		"rack": nil,
	}

	test_utils.MockGetObject(mockAPI, rack1)

	isAttrDrawable, err := controller.IsAttrDrawable(models.PhysicalPath+"BASIC/A/R1/A01", "color")
	assert.True(t, isAttrDrawable)
	assert.Nil(t, err)
}

func TestIsAttrDrawable(t *testing.T) {
	tests := []struct {
		name                 string
		attributeDrawable    string
		attributeNonDrawable string
	}{
		{"SpecialAttribute", "name", "description"},
		{"SpecialAttribute", "color", "height"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)
			controllers.State.DrawableObjs = []int{models.EntityStrToInt("rack")}

			controllers.State.DrawableJsons = test_utils.GetTestDrawableJson()

			test_utils.MockGetObject(mockAPI, rack1)
			isAttrDrawable, err := controller.IsAttrDrawable(models.PhysicalPath+"BASIC/A/R1/A01", tt.attributeDrawable)
			assert.True(t, isAttrDrawable)
			assert.Nil(t, err)

			test_utils.MockGetObject(mockAPI, rack1)
			isAttrDrawable, err = controller.IsAttrDrawable(models.PhysicalPath+"BASIC/A/R1/A01", tt.attributeNonDrawable)
			assert.False(t, isAttrDrawable)
			assert.Nil(t, err)
		})
	}
}

// Tests CreateUser
func TestCreateUserInvalidEmail(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockAPI.On(
		"Request", "POST",
		"/api/users",
		"mock.Anything", 201,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"message": "A valid email address is required",
			},
			Status: 400,
		}, errors.New("[Response From API] A valid email address is required"),
	).Once()

	err := controller.CreateUser("email", "manager", "*")
	assert.NotNil(t, err)
	assert.Equal(t, "[Response From API] A valid email address is required", err.Error())
}

func TestCreateUserWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockAPI.On("Request", "POST",
		"/api/users",
		"mock.Anything", 201,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"message": "Account has been created",
			},
		}, nil,
	).Once()

	err := controller.CreateUser("email@email.com", "manager", "*")
	assert.Nil(t, err)
}

// Tests AddRole
func TestAddRoleUserNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockAPI.On("Request", "GET", "/api/users", "mock.Anything", 200).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": []any{},
			},
		}, nil,
	).Once()

	err := controller.AddRole("email@email.com", "manager", "*")
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestAddRoleWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockAPI.On("Request", "GET", "/api/users", "mock.Anything", 200).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": []any{
					map[string]any{
						"_id":   "507f1f77bcf86cd799439011",
						"email": "email@email.com",
					},
				},
			},
		}, nil,
	).Once()

	mockAPI.On("Request", "PATCH", "/api/users/507f1f77bcf86cd799439011", "mock.Anything", 200).Return(
		&controllers.Response{
			Body: map[string]any{
				"message": "successfully updated user roles",
			},
		}, nil,
	).Once()

	err := controller.AddRole("email@email.com", "manager", "*")
	assert.Nil(t, err)
}
