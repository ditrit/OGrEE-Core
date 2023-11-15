package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/models"
	"cli/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var roomWithoutChildren = map[string]any{
	"category": "room",
	"children": []any{},
	"id":       "BASIC.A.R1",
	"name":     "R1",
	"parentId": "BASIC.A",
}

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
	controller, mockAPI, mockOgree3d, clockMock := newControllerWithMocks(t)
	controllers.State.Hierarchy = controllers.BuildBaseTree(controller)

	clockMock.On("Now").Return(time.Now()).Maybe()

	return controller, mockAPI, mockOgree3d
}

func TestLsOnASiteShowsRacksIfAnyObjectIsRack(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{
			copyMap(rack1),
		},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, models.RacksLayer.Name())
}

func TestLsOnASiteShowsCorridorsIfAnyObjectIsCorridor(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{
			corridor,
		},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "CO1")
	utils.ContainsObjectNamed(t, objects, models.CorridorsLayer.Name())
}

func TestLsOnASiteShowsGroupsIfAnyObjectIsGroup(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{
			roomGroup,
		},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "GRT")
	utils.ContainsObjectNamed(t, objects, models.GroupsLayer.Name())
}

func TestLsOnASiteWithAllChildrenShowsAllLayers(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 7)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, "B01")
	utils.ContainsObjectNamed(t, objects, "CO1")
	utils.ContainsObjectNamed(t, objects, "GRT")
	utils.ContainsObjectNamed(t, objects, models.CorridorsLayer.Name())
	utils.ContainsObjectNamed(t, objects, models.GroupsLayer.Name())
	utils.ContainsObjectNamed(t, objects, models.RacksLayer.Name())
}

func TestLsOnARackShowsGroupsIfAnyObjectIsGroup(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "rack",
		"children": []any{rackGroup},
		"id":       "BASIC.A.R1.A01",
		"name":     "A01",
		"parentId": "BASIC.A.R1",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/A01", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "GRrack")
	utils.ContainsObjectNamed(t, objects, models.GroupsLayer.Name())
}

func TestLsOnARackShowsOneLayerForEachTypeOfDevice(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, rack1)

	objects, err := controller.Ls("/Physical/BASIC/A/R1/A01", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 6)
	utils.ContainsObjectNamed(t, objects, "GRrack")
	utils.ContainsObjectNamed(t, objects, "chT")
	utils.ContainsObjectNamed(t, objects, "pdu")
	utils.ContainsObjectNamed(t, objects, "#chassis")
	utils.ContainsObjectNamed(t, objects, models.GroupsLayer.Name())
	utils.ContainsObjectNamed(t, objects, "#pdus")
}

func TestLsOnRacksLayerShowsRacks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, "B01")
}

func TestLsOnGroupLayerShowsGroups(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.*&namespace=logical", []any{roomGroup})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#groups", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "GRT")
}

func TestLsOnCorridorsLayerShowsCorridors(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=corridor&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{corridor})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#corridors", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "CO1")
}

func TestLsOnTypeLayerShowsDevicesOfThatType(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObjects(mockAPI, "category=device&id=BASIC.A.R1.A01.*&namespace=physical.hierarchy&type=chassis", []any{chassis})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/A01/#chassis", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "chT")
}

func TestLsOnLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks/A01", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 6)
	utils.ContainsObjectNamed(t, objects, "GRrack")
	utils.ContainsObjectNamed(t, objects, "chT")
	utils.ContainsObjectNamed(t, objects, "pdu")
	utils.ContainsObjectNamed(t, objects, "#chassis")
	utils.ContainsObjectNamed(t, objects, models.GroupsLayer.Name())
	utils.ContainsObjectNamed(t, objects, "#pdus")
}

func TestLsOnNestedLayerWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.A01.*&namespace=logical", []any{rackGroup})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks/A01/#groups", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "GRrack")
}

func TestGetOnRacksLayerGetsRacksAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Contains(t, objects, removeChildren(rack1))
	assert.Contains(t, objects, removeChildren(rack2))
}

func TestGetOnCorridorsLayerGetsCorridorsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=corridor&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{corridor})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#corridors", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(corridor))
}

func TestGetOnGroupLayerGetsGroupsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.*&namespace=logical", []any{roomGroup})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#groups", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(roomGroup))
}

func TestGetOnAllLayerGetsAllAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/*", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Contains(t, objects, removeChildren(rack1))
	assert.Contains(t, objects, removeChildren(rack2))
}

func TestGetOnWildcardLayerGetsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.A*&namespace=physical.hierarchy", []any{rack1})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/A*", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(rack1))
}

func TestGetOnLayerChildGetsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.A01&namespace=physical.hierarchy", []any{rack1})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/A01", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(rack1))
}

func TestGetOnNestedLayerGetsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObjects(mockAPI, "category=group&id=BASIC.A.R1.A01.*&namespace=logical", []any{rackGroup})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/A01/#groups", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, removeChildren(rackGroup))
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

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObject(mockAPI, rack1)

	err := controller.CD("/Physical/BASIC/A/R1/#racks/A01")
	assert.Nil(t, err)
	assert.Equal(t, controllers.State.CurrPath, "/Physical/BASIC/A/R1/A01")
}

func TestCdOnLayerGrandChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObject(mockAPI, chassis)

	err := controller.CD("/Physical/BASIC/A/R1/#racks/A01/chT")
	assert.Nil(t, err)
	assert.Equal(t, controllers.State.CurrPath, "/Physical/BASIC/A/R1/A01/chT")
}

func TestCdOnNestedLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObject(mockAPI, rackGroup)

	err := controller.CD("/Physical/BASIC/A/R1/#racks/A01/#groups/GRrack")
	assert.Nil(t, err)
	assert.Equal(t, controllers.State.CurrPath, "/Physical/BASIC/A/R1/A01/GRrack")
}

func TestSelectLayerSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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
	assert.Contains(t, selection, "/Physical/BASIC/A/R1/A01")
	assert.Contains(t, selection, "/Physical/BASIC/A/R1/B01")
}

func TestSelectGroupsLayerSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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
	assert.Contains(t, selection, "/Physical/BASIC/A/R1/GRT")
}

func TestSelectLayerAllSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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
	assert.Contains(t, selection, "/Physical/BASIC/A/R1/A01")
	assert.Contains(t, selection, "/Physical/BASIC/A/R1/B01")
}

func TestSelectLayerWildcardSelectsWildcard(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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
	assert.Contains(t, selection, "/Physical/BASIC/A/R1/A01")
}

func TestSelectLayerChildSelectsChild(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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
	assert.Contains(t, selection, "/Physical/BASIC/A/R1/A01")
}

func TestSelectNestedLayerSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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
	assert.Contains(t, selection, "/Physical/BASIC/A/R1/A01/GRrack")
}

func TestRemoveLayerRemovesAllObjectsOfTheLayer(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})
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

func TestDrawLayerDrawsAllObjectsOfTheLayer(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})
	mockGetObject(mockAPI, rack1)
	mockGetObject(mockAPI, rack2)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	mockOgree3D.On(
		"Inform", "Draw",
		0, map[string]any{"data": removeChildren(rack1), "type": "create"},
	).Return(nil)
	mockOgree3D.On(
		"Inform", "Draw",
		0, map[string]any{"data": removeChildren(rack2), "type": "create"},
	).Return(nil)

	err := controller.Draw("/Physical/BASIC/A/R1/#racks", 0, true)
	assert.Nil(t, err)
}

func TestDrawLayerWithDepthDrawsAllObjectsOfTheLayerAndChildren(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})
	mockGetObjectHierarchy(mockAPI, rack1)
	mockGetObjectHierarchy(mockAPI, rack2)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	mockOgree3D.On(
		"Inform", "Draw",
		0, map[string]any{"data": keepOnlyDirectChildren(rack1), "type": "create"},
	).Return(nil)
	mockOgree3D.On(
		"Inform", "Draw",
		0, map[string]any{"data": keepOnlyDirectChildren(rack2), "type": "create"},
	).Return(nil)

	err := controller.Draw("/Physical/BASIC/A/R1/#racks", 1, true)
	assert.Nil(t, err)
}

func TestUndrawLayerUndrawAllObjectsOfTheLayer(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjectsByEntity(mockAPI, "layers", []any{})
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})
	mockGetObject(mockAPI, rack1)
	mockGetObject(mockAPI, rack2)

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	mockOgree3D.On(
		"Inform", "Undraw",
		0, map[string]any{"data": "BASIC.A.R1.A01", "type": "delete"},
	).Return(nil)
	mockOgree3D.On(
		"Inform", "Undraw",
		0, map[string]any{"data": "BASIC.A.R1.B01", "type": "delete"},
	).Return(nil)

	err := controller.Undraw("/Physical/BASIC/A/R1/#racks")
	assert.Nil(t, err)
}

func TestTranslateApplicabilityReturnsErrorIfPathIsRoot(t *testing.T) {
	_, err := controllers.TranslateApplicability("/")
	assert.ErrorContains(t, err, "applicability must be an hierarchical path, found: /")
}

func TestTranslateApplicabilityReturnsErrorIfPathIsNotHierarchical(t *testing.T) {
	_, err := controllers.TranslateApplicability("/Logical/Tags")
	assert.ErrorContains(t, err, "applicability must be an hierarchical path, found: /Logical/Tags")
}

func TestTranslateApplicabilityTransformsPhysicalSlashIntoEmpty(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical")
	assert.Nil(t, err)
	assert.Equal(t, "", applicability)
}

func TestTranslateApplicabilityCleansPathOfLastSlash(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/")
	assert.Nil(t, err)
	assert.Equal(t, "", applicability)
}

func TestTranslateApplicabilityCleansPathOfSlashPointAtEnd(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/.")
	assert.Nil(t, err)
	assert.Equal(t, "", applicability)
}

func TestTranslateApplicabilityCleansPathOfSlashPoint(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/./BASIC")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC", applicability)
}

func TestTranslateApplicabilityTransformsPhysicalPathIntoID(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/BASIC/A")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC.A", applicability)
}

func TestTranslateApplicabilitySupportsPointPointAtTheEnd(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/BASIC/..")
	assert.Nil(t, err)
	assert.Equal(t, "", applicability)
}

func TestTranslateApplicabilitySupportsPointPoint(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/BASIC/../COMPLEX/R1")
	assert.Nil(t, err)
	assert.Equal(t, "COMPLEX.R1", applicability)
}

func TestTranslateApplicabilitySupportsStarAtTheEnd(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/*")
	assert.Nil(t, err)
	assert.Equal(t, "*", applicability)
}

func TestTranslateApplicabilitySupportsStarStarAtTheEnd(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/**")
	assert.Nil(t, err)
	assert.Equal(t, "**", applicability)
}

func TestTranslateApplicabilitySupportsStar(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/*/chT")
	assert.Nil(t, err)
	assert.Equal(t, "*.chT", applicability)
}

func TestTranslateApplicabilitySupportsStarStar(t *testing.T) {
	applicability, err := controllers.TranslateApplicability("/Physical/**/chT")
	assert.Nil(t, err)
	assert.Equal(t, "**.chT", applicability)
}

func TestTranslateApplicabilityEmptyReturnsCurrPath(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC/A"
	applicability, err := controllers.TranslateApplicability("")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC.A", applicability)
}

func TestTranslateApplicabilityPointReturnsCurrPath(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC/A"
	applicability, err := controllers.TranslateApplicability(".")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC.A", applicability)
}

func TestTranslateApplicabilityPointReturnsErrorIfCurrPathIsNotHierarchical(t *testing.T) {
	controllers.State.CurrPath = "/Logical/Tags"
	_, err := controllers.TranslateApplicability(".")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "applicability must be an hierarchical path, found: /Logical/Tags")
}

func TestTranslateApplicabilityPointReturnsEmptyIfCurrPathIsSlashPhysical(t *testing.T) {
	controllers.State.CurrPath = "/Physical"
	applicability, err := controllers.TranslateApplicability(".")
	assert.Nil(t, err)
	assert.Equal(t, "", applicability)
}

func TestTranslateApplicabilityPointPathReturnsCurrPathPlusPath(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC"
	applicability, err := controllers.TranslateApplicability("./A")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC.A", applicability)
}

func TestTranslateApplicabilityRelativePathReturnsCurrPathPlusPath(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC"
	applicability, err := controllers.TranslateApplicability("A")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC.A", applicability)
}

func TestTranslateApplicabilityRelativePathStarReturnsCurrPathPlusPath(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC"
	applicability, err := controllers.TranslateApplicability("*")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC.*", applicability)
}

func TestTranslateApplicabilityRelativePathReturnsErrorIfCurrPathIsNotHierarchical(t *testing.T) {
	controllers.State.CurrPath = "/Logical/Tags"
	_, err := controllers.TranslateApplicability("A")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "applicability must be an hierarchical path, found: /Logical/Tags/A")
}

func TestTranslateApplicabilityPointPointReturnsBeforeCurrPath(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC/A"
	applicability, err := controllers.TranslateApplicability("..")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC", applicability)
}

func TestTranslateApplicabilityPointPointReturnsEmptyIfBeforeCurrPathIsPhysical(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC"
	applicability, err := controllers.TranslateApplicability("..")
	assert.Nil(t, err)
	assert.Equal(t, "", applicability)
}

func TestTranslateApplicabilityPointPointReturnsErrorIfBeforeCurrPathIsNotHierarchical(t *testing.T) {
	controllers.State.CurrPath = "/Physical"
	_, err := controllers.TranslateApplicability("..")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "applicability must be an hierarchical path, found: /")
}

func TestTranslateApplicabilityPointPointPathReturnsCurrPathPlusPath(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC"
	applicability, err := controllers.TranslateApplicability("../COMPLEX")
	assert.Nil(t, err)
	assert.Equal(t, "COMPLEX", applicability)
}

func TestTranslateApplicabilityPointPointTwoTimes(t *testing.T) {
	controllers.State.CurrPath = "/Physical/BASIC/R1"
	applicability, err := controllers.TranslateApplicability("../../COMPLEX")
	assert.Nil(t, err)
	assert.Equal(t, "COMPLEX", applicability)
}

func TestTranslateApplicabilityMinusReturnsPrevPath(t *testing.T) {
	controllers.State.PrevPath = "/Physical/BASIC"
	applicability, err := controllers.TranslateApplicability("-")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC", applicability)
}

func TestTranslateApplicabilityMinusPathReturnsPrevPathPlusPath(t *testing.T) {
	controllers.State.PrevPath = "/Physical"
	applicability, err := controllers.TranslateApplicability("-/BASIC")
	assert.Nil(t, err)
	assert.Equal(t, "BASIC", applicability)
}

func TestTranslateApplicabilityUnderscorePathReturnsCurrPathPlusUnderscore(t *testing.T) {
	controllers.State.CurrPath = "/Physical"
	applicability, err := controllers.TranslateApplicability("_")
	assert.Nil(t, err)
	assert.Equal(t, "_", applicability)
}

func TestTranslateApplicabilityReturnsErrorIfPatternIsNotValid(t *testing.T) {
	_, err := controllers.TranslateApplicability("/Physical/[")
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "applicability pattern is not valid")
}

func TestLsNowShowLayerIfNotMatch(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.A.R2",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsShowLayerIfPerfectMatch(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.A.R1",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsShowLayerIfPerfectMatchOnPhysical(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectsByEntity(mockAPI, "sites", []any{})

	objects, err := controller.Ls("/Physical", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "Stray")
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsShowLayerIfPerfectMatchOnPhysicalChild(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "site",
		"children": []any{},
		"id":       "BASIC",
		"name":     "BASIC",
		"parentId": "",
	})

	objects, err := controller.Ls("/Physical/BASIC", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsShowLayerIfPerfectMatchOnPhysicalChildWhenItsCached(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	site := map[string]any{
		"category": "site",
		"children": []any{},
		"id":       "BASIC",
		"name":     "BASIC",
		"parentId": "",
	}
	mockGetObjectsByEntity(mockAPI, "sites", []any{site})

	_, err := controller.Tree("/Physical", 1)
	assert.Nil(t, err)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, site)

	objects, err := controller.Ls("/Physical/BASIC", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsShowLayerIfMatchWithStar(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.A.*",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsShowLayerIfMatchWithSomethingStar(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.A.R*",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "#test", objects[0]["name"])
}

func TestLsNotShowLayerIfNotMatchWithStar(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.*",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsShowLayerIfMatchWithDoubleStar(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.**",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsShowLayerIfMatchWithDoubleStarAndMore(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.**.A01",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, emptyChildren(rack1))

	objects, err := controller.Ls("/Physical/BASIC/A/R1/A01", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsNotShowLayerIfNotMatchWithDoubleStar(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.B.**",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsNotShowLayerIfDoubleStarIsAtTheEndAndZeroFoldersAreFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.A.R1.**",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, "")
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsNotShowLayerIfNotMatchWithDoubleStarAndMore(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.**.chT",
			models.LayerFilters:       map[string]any{"any": "yes"},
		},
	})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsReturnsLayerCreatedAfterLastUpdate(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})

	objects, err := controller.Ls("/Logical/Layers", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)

	mockCreateObject(mockAPI, "layer", map[string]any{
		"slug":                    "test",
		models.LayerFilters:       map[string]any{},
		models.LayerApplicability: "BASIC.A.R1",
	})
	err = controller.CreateLayer("test", "/Physical/BASIC/A/R1")
	assert.Nil(t, err)

	objects, err = controller.Ls("/Logical/Layers", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "test", objects[0]["name"])
}

func TestLsReturnsLayerCreatedAndUpdatedAfterLastUpdate(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockGetObjectsByEntity(mockAPI, "layers", []any{})

	objects, err := controller.Ls("/Logical/Layers", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)

	testLayer := map[string]any{
		"slug":                    "test",
		models.LayerFilters:       map[string]any{},
		models.LayerApplicability: "BASIC.A.R1",
	}

	mockCreateObject(mockAPI, "layer", testLayer)
	err = controller.CreateLayer("test", "/Physical/BASIC/A/R1")
	assert.Nil(t, err)

	mockGetObjectByEntity(mockAPI, "layers", testLayer)
	mockUpdateObject(mockAPI, map[string]any{
		models.LayerFilters: map[string]any{"category": "device"},
	}, map[string]any{
		"slug":                    "test",
		models.LayerFilters:       map[string]any{"category": "device"},
		models.LayerApplicability: "BASIC.A.R1",
	})
	err = controller.UpdateLayer("/Logical/Layers/test", "category", "device")
	assert.Nil(t, err)

	objects, err = controller.Ls("/Logical/Layers", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "test", objects[0]["name"])

	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err = controller.Ls("/Physical/BASIC/A/R1", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsOnLayerUpdatedAfterLastUpdateDoesUpdatedFilter(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	testLayer := map[string]any{
		"slug": "test",
		models.LayerFilters: map[string]any{
			"category": "rack",
		},
		models.LayerApplicability: "BASIC.A.R1",
	}

	mockGetObjectsByEntity(mockAPI, "layers", []any{testLayer})
	mockGetObjectHierarchy(mockAPI, roomWithoutChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#test", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)

	mockGetObjectByEntity(mockAPI, "layers", testLayer)
	mockUpdateObject(mockAPI, map[string]any{
		models.LayerFilters: map[string]any{"category": "device"},
	}, map[string]any{
		"slug":                    "test",
		models.LayerFilters:       map[string]any{"category": "device"},
		models.LayerApplicability: "BASIC.A.R1",
	})
	err = controller.UpdateLayer("/Logical/Layers/test", "category", "device")
	assert.Nil(t, err)

	mockGetObjects(mockAPI, "category=device&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{})

	objects, err = controller.Ls("/Physical/BASIC/A/R1/#test", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsOnUserDefinedLayerAppliesFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	testLayer := map[string]any{
		"slug": "test",
		models.LayerFilters: map[string]any{
			"category": "rack",
		},
		models.LayerApplicability: "BASIC.A.R1",
	}

	mockGetObjectsByEntity(mockAPI, "layers", []any{testLayer})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=rack&id=BASIC.A.R1.*&namespace=physical.hierarchy", []any{rack1, rack2})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#test", nil, false)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, "B01")
}

func TestLsRecursiveOnLayerListLayerRecursive(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	devices := map[string]any{
		"slug": "devices",
		models.LayerFilters: map[string]any{
			"category": "device",
		},
		models.LayerApplicability: "BASIC.A.R1",
	}

	mockGetObjectsByEntity(mockAPI, "layers", []any{devices})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=device&id=BASIC.A.R1.**&namespace=physical.hierarchy", []any{chassis, pdu})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#devices", nil, true)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "chT")
	utils.ContainsObjectNamed(t, objects, "pdu")
}

func TestGetRecursiveOnLayerReturnsLayerRecursive(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	devices := map[string]any{
		"slug": "devices",
		models.LayerFilters: map[string]any{
			"category": "device",
		},
		models.LayerApplicability: "BASIC.A.R1",
	}

	mockGetObjectsByEntity(mockAPI, "layers", []any{devices})
	mockGetObjectHierarchy(mockAPI, roomWithChildren)
	mockGetObjects(mockAPI, "category=device&id=BASIC.A.R1.**&namespace=physical.hierarchy", []any{chassis, pdu})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#devices", nil, true)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Contains(t, objects, removeChildren(chassis))
	assert.Contains(t, objects, removeChildren(pdu))
}
