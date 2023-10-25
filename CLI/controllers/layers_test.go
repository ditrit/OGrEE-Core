package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

var roomWithChildren = map[string]any{
	"category": "room",
	"children": []any{
		rack1,
		rack2,
		corridor,
		roomGroup,
	},
	"id":       "BASIC.A.R1",
	"name":     "R1",
	"parentId": "BASIC.A",
}

var rack1 = map[string]any{
	"category": "rack",
	"children": []any{chassis, rackGroup, pdu},
	"id":       "BASIC.A.R1.A01",
	"name":     "A01",
	"parentId": "BASIC.A.R1",
}

var rackGroup = map[string]any{
	"category": "group",
	"children": []any{},
	"id":       "BASIC.A.R1.A01.GRrack",
	"name":     "GRrack",
	"parentId": "BASIC.A.R1.A01",
}

var rack2 = map[string]any{
	"category": "rack",
	"children": []any{},
	"id":       "BASIC.A.R1.B01",
	"name":     "B01",
	"parentId": "BASIC.A.R1",
}

var chassis = map[string]any{
	"category": "device",
	"attributes": map[string]any{
		"type": "chassis",
	},
	"children": []any{},
	"id":       "BASIC.A.R1.A01.chT",
	"name":     "chT",
	"parentId": "BASIC.A.R1.A01",
}

var pdu = map[string]any{
	"category": "device",
	"attributes": map[string]any{
		"type": "pdu",
	},
	"children": []any{},
	"id":       "BASIC.A.R1.A01.pdu",
	"name":     "pdu",
	"parentId": "BASIC.A.R1.A01",
}

var corridor = map[string]any{
	"category": "corridor",
	"children": []any{},
	"id":       "BASIC.A.R1.CO1",
	"name":     "CO1",
	"parentId": "BASIC.A.R1",
}

var roomGroup = map[string]any{
	"category": "group",
	"children": []any{},
	"id":       "BASIC.A.R1.GRT",
	"name":     "GRT",
	"parentId": "BASIC.A.R1",
}

func layersSetup(t *testing.T) (controllers.Controller, *mocks.APIPort, *mocks.Ogree3DPort) {
	controllers.State.Hierarchy = controllers.BuildBaseTree()
	return newControllerWithMocks(t)
}

func TestLsOnASiteShowsRacksIfAnyObjectIsRack(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{
			copyMap(rack1),
		},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, "A01", objects[0]["name"])
	assert.Equal(t, models.RacksLayer.Name, objects[1]["name"])
}

func TestLsOnASiteShowsCorridorsIfAnyObjectIsCorridor(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{
			corridor,
		},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, "CO1", objects[0]["name"])
	assert.Equal(t, models.CorridorsLayer.Name, objects[1]["name"])
}

func TestLsOnASiteShowsGroupsIfAnyObjectIsGroup(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{
			roomGroup,
		},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, "GRT", objects[0]["name"])
	assert.Equal(t, models.GroupsLayer.Name, objects[1]["name"])
}

func TestLsOnASiteWithAllChildrenShowsAllLayers(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 7)
	assert.Equal(t, "A01", objects[0]["name"])
	assert.Equal(t, "B01", objects[1]["name"])
	assert.Equal(t, "CO1", objects[2]["name"])
	assert.Equal(t, "GRT", objects[3]["name"])
	assert.Equal(t, models.CorridorsLayer.Name, objects[4]["name"])
	assert.Equal(t, models.GroupsLayer.Name, objects[5]["name"])
	assert.Equal(t, models.RacksLayer.Name, objects[6]["name"])
}

func TestLsOnARackShowsGroupsIfAnyObjectIsGroup(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "rack",
		"children": []any{rackGroup},
		"id":       "BASIC.A.R1.A01",
		"name":     "A01",
		"parentId": "BASIC.A.R1",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/A01", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, "GRrack", objects[0]["name"])
	assert.Equal(t, models.GroupsLayer.Name, objects[1]["name"])
}

func TestLsOnRacksLayerShowsRacks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, "A01", objects[0]["name"])
	assert.Equal(t, "B01", objects[1]["name"])
}

func TestLsOnGroupLayerShowsGroups(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.*&namespace=logical", []any{roomGroup})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#groups", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "GRT", objects[0]["name"])
}

func TestLsOnCorridorsLayerShowsCorridors(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=corridor&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{corridor})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#corridors", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "CO1", objects[0]["name"])
}

func TestLsOnLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks/A01", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 4)
	assert.Equal(t, "GRrack", objects[0]["name"])
	assert.Equal(t, "chT", objects[1]["name"])
	assert.Equal(t, "pdu", objects[2]["name"])
	assert.Equal(t, models.GroupsLayer.Name, objects[3]["name"])
}

func TestLsOnNestedLayerWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.A01.*&namespace=logical", []any{rackGroup})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks/A01/#groups", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "GRrack", objects[0]["name"])
}

func TestGetOnRacksLayerGetsRacksAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, removeChildren(rack1), objects[0])
	assert.Equal(t, removeChildren(rack2), objects[1])
}

func TestGetOnCorridorsLayerGetsCorridorsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=corridor&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{corridor})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#corridors")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, removeChildren(corridor), objects[0])
}

func TestGetOnGroupLayerGetsGroupsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.*&namespace=logical", []any{roomGroup})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#groups")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, removeChildren(roomGroup), objects[0])
}

func TestGetOnAllLayerGetsAllAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/*")
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, removeChildren(rack1), objects[0])
	assert.Equal(t, removeChildren(rack2), objects[1])
}

func TestGetOnWildcardLayerGetsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.A*&namespace=physical.hierarchy", []any{rack1})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/A*")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, removeChildren(rack1), objects[0])
}

func TestGetOnLayerChildGetsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.A01&namespace=physical.hierarchy", []any{rack1})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/A01")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, removeChildren(rack1), objects[0])
}

func TestGetOnNestedLayerGetsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.A01.*&namespace=logical", []any{rackGroup})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/A01/#groups")
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, removeChildren(rackGroup), objects[0])
}

func TestTreeOnLayerFails(t *testing.T) {
	controller, _, _ := layersSetup(t)

	_, err := controller.Tree("/Physical/BASIC/A/R1/#racks", 1)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "it is not possible to tree a layer")
}

func TestTreeOnNestedLayerFails(t *testing.T) {
	controller, _, _ := layersSetup(t)

	_, err := controller.Tree("/Physical/BASIC/A/R1/#racks/A01/#groups", 1)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "it is not possible to tree a layer")
}

func TestTreeOnLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)

	node, err := controller.Tree("/Physical/BASIC/A/R1/#racks/A01", 1)
	assert.Nil(t, err)
	assert.Equal(t, "A01", node.Name)
	assert.Len(t, node.Children, 3)
	assert.NotNil(t, node.Children["GRrack"])
	assert.NotNil(t, node.Children["chT"])
	assert.NotNil(t, node.Children["pdu"])
}

func TestTreeOnNestedLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObjectHierarchy(mockAPI, rackGroup)

	node, err := controller.Tree("/Physical/BASIC/A/R1/#racks/A01/#groups/GRrack", 1)
	assert.Nil(t, err)
	assert.Equal(t, "GRrack", node.Name)
	assert.Len(t, node.Children, 0)
}

func TestCdOnLayerFails(t *testing.T) {
	controller, _, _ := layersSetup(t)

	err := controller.CD("/Physical/BASIC/A/R1/#racks")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "it is not possible to cd into a layer")
}

func TestCdOnNestedLayerFails(t *testing.T) {
	controller, _, _ := layersSetup(t)

	err := controller.CD("/Physical/BASIC/A/R1/#racks/A01/#groups")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "it is not possible to cd into a layer")
}

func TestCdOnLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObject(mockAPI, rack1)

	err := controller.CD("/Physical/BASIC/A/R1/#racks/A01")
	assert.Nil(t, err)
	assert.Equal(t, controllers.State.CurrPath, "/Physical/BASIC/A/R1/A01")
}

func TestCdOnLayerGrandChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObject(mockAPI, chassis)

	err := controller.CD("/Physical/BASIC/A/R1/#racks/A01/chT")
	assert.Nil(t, err)
	assert.Equal(t, controllers.State.CurrPath, "/Physical/BASIC/A/R1/A01/chT")
}

func TestCdOnNestedLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObject(mockAPI, rackGroup)

	err := controller.CD("/Physical/BASIC/A/R1/#racks/A01/#groups/GRrack")
	assert.Nil(t, err)
	assert.Equal(t, controllers.State.CurrPath, "/Physical/BASIC/A/R1/A01/GRrack")
}

func TestSelectLayerSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})
	mockGetObject(mockAPI, rack1)
	mockGetObject(mockAPI, rack2)

	mockOgree3D.On(
		"InformOptional", "SetClipBoard",
		-1, map[string]any{"data": "[\"BASIC.A.R1.A01\",\"BASIC.A.R1.B01\"]", "type": "select"},
	).Return(nil)

	selection, err := controller.Select("/Physical/BASIC/A/R1/#racks")
	assert.Nil(t, err)
	assert.Len(t, selection, 2)
	assert.Equal(t, "/Physical/BASIC/A/R1/A01", selection[0])
	assert.Equal(t, "/Physical/BASIC/A/R1/B01", selection[1])
}

func TestSelectGroupsLayerSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.*&namespace=logical", []any{roomGroup})
	mockGetObject(mockAPI, roomGroup)

	mockOgree3D.On(
		"InformOptional", "SetClipBoard",
		-1, map[string]any{"data": "[\"BASIC.A.R1.GRT\"]", "type": "select"},
	).Return(nil)

	selection, err := controller.Select("/Physical/BASIC/A/R1/#groups")
	assert.Nil(t, err)
	assert.Len(t, selection, 1)
	assert.Equal(t, "/Physical/BASIC/A/R1/GRT", selection[0])
}

func TestSelectLayerAllSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})
	mockGetObject(mockAPI, rack1)
	mockGetObject(mockAPI, rack2)

	mockOgree3D.On(
		"InformOptional", "SetClipBoard",
		-1, map[string]any{"data": "[\"BASIC.A.R1.A01\",\"BASIC.A.R1.B01\"]", "type": "select"},
	).Return(nil)

	selection, err := controller.Select("/Physical/BASIC/A/R1/#racks/*")
	assert.Nil(t, err)
	assert.Len(t, selection, 2)
	assert.Equal(t, "/Physical/BASIC/A/R1/A01", selection[0])
	assert.Equal(t, "/Physical/BASIC/A/R1/B01", selection[1])
}

func TestSelectLayerWildcardSelectsWildcard(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.A*&namespace=physical.hierarchy", []any{rack1})
	mockGetObject(mockAPI, rack1)

	mockOgree3D.On(
		"InformOptional", "SetClipBoard",
		-1, map[string]any{"data": "[\"BASIC.A.R1.A01\"]", "type": "select"},
	).Return(nil)

	selection, err := controller.Select("/Physical/BASIC/A/R1/#racks/A*")
	assert.Nil(t, err)
	assert.Len(t, selection, 1)
	assert.Equal(t, "/Physical/BASIC/A/R1/A01", selection[0])
}

func TestSelectLayerChildSelectsChild(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.A01&namespace=physical.hierarchy", []any{rack1})
	mockGetObject(mockAPI, rack1)

	mockOgree3D.On(
		"InformOptional", "SetClipBoard",
		-1, map[string]any{"data": "[\"BASIC.A.R1.A01\"]", "type": "select"},
	).Return(nil)

	selection, err := controller.Select("/Physical/BASIC/A/R1/#racks/A01")
	assert.Nil(t, err)
	assert.Len(t, selection, 1)
	assert.Equal(t, "/Physical/BASIC/A/R1/A01", selection[0])
}

func TestSelectNestedLayerSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.A01.*&namespace=logical", []any{rackGroup})
	mockGetObject(mockAPI, rackGroup)

	mockOgree3D.On(
		"InformOptional", "SetClipBoard",
		-1, map[string]any{"data": "[\"BASIC.A.R1.A01.GRrack\"]", "type": "select"},
	).Return(nil)

	selection, err := controller.Select("/Physical/BASIC/A/R1/#racks/A01/#groups")
	assert.Nil(t, err)
	assert.Len(t, selection, 1)
	assert.Equal(t, "/Physical/BASIC/A/R1/A01/GRrack", selection[0])
}

func TestRemoveLayerRemovesAllObjectsOfTheLayer(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockDeleteObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	mockOgree3D.On(
		"InformOptional", "DeleteObj",
		-1, map[string]any{"data": "BASIC.A.R1.A01", "type": "delete"},
	).Return(nil)
	mockOgree3D.On(
		"InformOptional", "DeleteObj",
		-1, map[string]any{"data": "BASIC.A.R1.B01", "type": "delete"},
	).Return(nil)

	_, err := controller.DeleteObj("/Physical/BASIC/A/R1/#racks")
	assert.Nil(t, err)
}
