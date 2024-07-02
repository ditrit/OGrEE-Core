package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"cli/models"
	test_utils "cli/test"
	"cli/utils"
	"encoding/json"
	"strings"
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
		generic,
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

var generic = map[string]any{
	"category": "generic",
	"id":       "BASIC.A.R1.table1",
	"name":     "table1",
	"parentId": "BASIC.A.R1",
	"attributes": map[string]any{
		"type": "table",
	},
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

var vobjCluster = map[string]any{
	"category": "virtual_obj",
	"attributes": map[string]any{
		"virtual_config": map[string]any{"role": "proxmox", "type": "cluster"},
	},
	"children": []any{},
	"id":       "cluster",
	"name":     "cluster",
	"parentId": "",
}

func layersSetup(t *testing.T) (controllers.Controller, *mocks.APIPort, *mocks.Ogree3DPort) {
	controller, mockAPI, mockOgree3d, clockMock := test_utils.NewControllerWithMocks(t)
	controllers.State.Hierarchy = controllers.BuildBaseTree(controller)

	clockMock.On("Now").Return(time.Now()).Maybe()

	return controller, mockAPI, mockOgree3d
}
func TestLsOnRoom(t *testing.T) {
	tests := []struct {
		name       string
		child      map[string]any
		objectName string
	}{
		{"ShowsRacksIfAnyObjectIsRack", rack1, models.RacksLayer.Name()},
		{"ShowsCorridorsIfAnyObjectIsCorridor", corridor, models.CorridorsLayer.Name()},
		{"ShowsGroupsIfAnyObjectIsGroup", roomGroup, models.GroupsLayer.Name()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)

			test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
			test_utils.MockGetObjectHierarchy(mockAPI, map[string]any{
				"category": "room",
				"children": []any{
					test_utils.CopyMap(tt.child),
				},
				"id":       "BASIC.A.R1",
				"name":     "R1",
				"parentId": "BASIC.A",
			})

			objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, nil)
			assert.Nil(t, err)
			assert.Len(t, objects, 2)
			utils.ContainsObjectNamed(t, objects, tt.child["name"].(string))
			utils.ContainsObjectNamed(t, objects, tt.objectName)
		})
	}
}

func TestLsOnARoomShowsGenericsAndGenericsByTypeIfAnyObjectIsGeneric(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "room",
		"children": []any{
			generic,
		},
		"id":       "BASIC.A.R1",
		"name":     "R1",
		"parentId": "BASIC.A",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 3)
	utils.ContainsObjectNamed(t, objects, "table1")
	utils.ContainsObjectNamed(t, objects, models.GenericsLayer.Name())
	utils.ContainsObjectNamed(t, objects, "#tables")
}

func TestLsOnARoomWithAllChildrenShowsAllLayers(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)

	objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 10)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, "B01")
	utils.ContainsObjectNamed(t, objects, "CO1")
	utils.ContainsObjectNamed(t, objects, "GRT")
	utils.ContainsObjectNamed(t, objects, "table1")
	utils.ContainsObjectNamed(t, objects, models.CorridorsLayer.Name())
	utils.ContainsObjectNamed(t, objects, models.GroupsLayer.Name())
	utils.ContainsObjectNamed(t, objects, models.RacksLayer.Name())
	utils.ContainsObjectNamed(t, objects, models.GenericsLayer.Name())
	utils.ContainsObjectNamed(t, objects, "#tables")
}

func TestLsOnARackShowsGroupsIfAnyObjectIsGroup(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "rack",
		"children": []any{rackGroup},
		"id":       "BASIC.A.R1.A01",
		"name":     "A01",
		"parentId": "BASIC.A.R1",
	})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/A01", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "GRrack")
	utils.ContainsObjectNamed(t, objects, models.GroupsLayer.Name())
}

func TestLsOnARackShowsOneLayerForEachTypeOfDevice(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, rack1)

	objects, err := controller.Ls("/Physical/BASIC/A/R1/A01", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 6)
	utils.ContainsObjectNamed(t, objects, "GRrack")
	utils.ContainsObjectNamed(t, objects, "chT")
	utils.ContainsObjectNamed(t, objects, "pdu")
	utils.ContainsObjectNamed(t, objects, "#chassis")
	utils.ContainsObjectNamed(t, objects, models.GroupsLayer.Name())
	utils.ContainsObjectNamed(t, objects, "#pdus")
}

func TestLsOnLogicalVObjsShowsOneLayerForEachTypeOfVObj(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetVirtualObjects(mockAPI, "limit=1", []any{vobjCluster})

	objects, err := controller.Ls("/Logical/VirtualObjects", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "cluster")
	utils.ContainsObjectNamed(t, objects, "#clusters")
}

func TestLsOnRacksLayerShowsRacks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.*&namespace=physical.hierarchy", map[string]any{"filter": "category=rack"}, []any{rack1, rack2})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks", map[string]string{}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, "B01")
}

func TestLs(t *testing.T) {
	tests := []struct {
		name                                     string
		mockGetObjectHierarchyResponse           map[string]any
		queryParams                              string
		filter                                   string
		mockGetObjectsWithComplexFiltersResponse []any
		lsPath                                   string
	}{
		{"OnGroupLayerShowsGroups", roomWithChildren, "id=BASIC.A.R1.*&namespace=physical.hierarchy", "category=group", []any{roomGroup}, "/Physical/BASIC/A/R1/#groups"},
		{"OnCorridorsLayerShowsCorridors", roomWithChildren, "id=BASIC.A.R1.*&namespace=physical.hierarchy", "category=corridor", []any{corridor}, "/Physical/BASIC/A/R1/#corridors"},
		{"OnGenericLayerShowsGeneric", roomWithChildren, "id=BASIC.A.R1.*&namespace=physical.hierarchy", "category=generic", []any{generic}, "/Physical/BASIC/A/R1/#generics"},
		{"OnDeviceTypeLayerShowsDevicesOfThatType", rack1, "id=BASIC.A.R1.A01.*&namespace=physical.hierarchy", "category=device&type=chassis", []any{chassis}, "/Physical/BASIC/A/R1/A01/#chassis"},
		{"OnGenericTypeLayerShowsDevicesOfThatType", roomWithChildren, "id=BASIC.A.R1.*&namespace=physical.hierarchy", "category=generic&type=table", []any{generic}, "/Physical/BASIC/A/R1/#tables"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)

			test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
			test_utils.MockGetObjectHierarchy(mockAPI, tt.mockGetObjectHierarchyResponse)
			test_utils.MockGetObjectsWithComplexFilters(mockAPI, tt.queryParams, map[string]any{"filter": tt.filter}, tt.mockGetObjectsWithComplexFiltersResponse)

			objects, err := controller.Ls(tt.lsPath, map[string]string{}, nil)
			assert.Nil(t, err)
			assert.Len(t, objects, len(tt.mockGetObjectsWithComplexFiltersResponse))
			for _, instance := range tt.mockGetObjectsWithComplexFiltersResponse {
				utils.ContainsObjectNamed(t, objects, instance.(map[string]any)["name"].(string))
			}
		})
	}
}

func TestLsOnLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectHierarchy(mockAPI, rack1)

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks/A01", nil, nil)
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

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectHierarchy(mockAPI, rack1)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.A01.*&namespace=physical.hierarchy", map[string]any{"filter": "category=group"}, []any{rackGroup})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#racks/A01/#groups", map[string]string{}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "GRrack")
}

func testComplexFiltersLayers(t *testing.T, mockQueryParams string, mockFilter string, mockListResponse []any, path string) {
	controller, mockAPI, _ := layersSetup(t)
	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, mockQueryParams, map[string]any{"filter": mockFilter}, mockListResponse)

	objects, _, err := controller.GetObjectsWildcard(path, map[string]string{}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, len(mockListResponse))
	for _, instance := range mockListResponse {
		assert.Contains(t, objects, test_utils.RemoveChildren(instance.(map[string]any)))
	}
}

func TestGetOnRacksLayerGetsRacksAttributes(t *testing.T) {
	testComplexFiltersLayers(t, "id=BASIC.A.R1.*&namespace=physical.hierarchy", "(category=rack) & (category=rack)", []any{rack1, rack2}, "/Physical/BASIC/A/R1/#racks")
}

func TestGetOnCorridorsLayerGetsCorridorsAttributes(t *testing.T) {
	testComplexFiltersLayers(t, "id=BASIC.A.R1.*&namespace=physical.hierarchy", "(category=corridor) & (category=corridor)", []any{corridor}, "/Physical/BASIC/A/R1/#corridors")
}

func TestGetOnGroupLayerGetsGroupsAttributes(t *testing.T) {
	testComplexFiltersLayers(t, "id=BASIC.A.R1.*&namespace=physical.hierarchy", "(category=group) & (category=group)", []any{roomGroup}, "/Physical/BASIC/A/R1/#groups")
}

func TestGetOnAllLayerGetsAllAttributes(t *testing.T) {
	testComplexFiltersLayers(t, "id=BASIC.A.R1.*&namespace=physical.hierarchy", "(category=rack) & (category=rack)", []any{rack1, rack2}, "/Physical/BASIC/A/R1/#racks/*")
}

func TestGetOnWildcardLayerGetsAttributes(t *testing.T) {
	testComplexFiltersLayers(t, "id=BASIC.A.R1.A*&namespace=physical.hierarchy", "(category=rack) & (category=rack)", []any{rack1}, "/Physical/BASIC/A/R1/#racks/A*")
}

func TestGetOnLayerChildGetsAttributes(t *testing.T) {
	testComplexFiltersLayers(t, "id=BASIC.A.R1.A01&namespace=physical.hierarchy", "(category=rack) & (category=rack)", []any{rack1}, "/Physical/BASIC/A/R1/#racks/A01")
}

func TestGetOnNestedLayerGetsAttributes(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectHierarchy(mockAPI, rack1)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.A01.*&namespace=physical.hierarchy", map[string]any{"filter": "(category=group) & (category=group)"}, []any{rackGroup})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#racks/A01/#groups", map[string]string{}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Contains(t, objects, test_utils.RemoveChildren(rackGroup))
}

func TestTreeFails(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"OnLayer", "/Physical/BASIC/A/R1/#racks"},
		{"OnNestedLayer", "/Physical/BASIC/A/R1/#racks/A01/#groups"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, _, _ := layersSetup(t)
			_, err := controller.Tree(tt.path, 1)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, "it is not possible to tree a layer")
		})
	}
}

func TestTreeOnLayerChildWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectHierarchy(mockAPI, rack1)

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

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectHierarchy(mockAPI, rack1)
	test_utils.MockGetObjectHierarchy(mockAPI, rackGroup)

	node, err := controller.Tree("/Physical/BASIC/A/R1/#racks/A01/#groups/GRrack", 1)
	assert.Nil(t, err)
	assert.Equal(t, "GRrack", node.Name)
	assert.Len(t, node.Children, 0)
}

func TestCdFails(t *testing.T) {
	tests := []struct {
		name   string
		cdPath string
	}{
		{"OnLayer", "/Physical/BASIC/A/R1/#racks"},
		{"OnNestedLayer", "/Physical/BASIC/A/R1/#racks/A01/#groups"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, _, _ := layersSetup(t)

			err := controller.CD(tt.cdPath)
			assert.NotNil(t, err)
			assert.ErrorContains(t, err, "it is not possible to cd into a layer")
		})
	}
}

func testCd(t *testing.T, entity string, mockObjectsHierarchy []any, mockGetObjectResponse map[string]any, cdPath string, expectedPath string) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, entity, []any{})
	for _, object := range mockObjectsHierarchy {
		test_utils.MockGetObjectHierarchy(mockAPI, object.(map[string]any))
	}
	test_utils.MockGetObject(mockAPI, mockGetObjectResponse)

	err := controller.CD(cdPath)
	assert.Nil(t, err)
	assert.Equal(t, expectedPath, controllers.State.CurrPath)
}

func TestCdOnLayerChildWorks(t *testing.T) {
	testCd(t, "layers", []any{roomWithChildren}, rack1, "/Physical/BASIC/A/R1/#racks/A01", "/Physical/BASIC/A/R1/A01")
}

func TestCdOnLayerGrandChildWorks(t *testing.T) {
	testCd(t, "layers", []any{roomWithChildren}, chassis, "/Physical/BASIC/A/R1/#racks/A01/chT", "/Physical/BASIC/A/R1/A01/chT")
}

func TestCdOnNestedLayerChildWorks(t *testing.T) {
	testCd(t, "layers", []any{roomWithChildren, rack1}, rackGroup, "/Physical/BASIC/A/R1/#racks/A01/#groups/GRrack", "/Physical/BASIC/A/R1/A01/GRrack")
}

func TestSelect(t *testing.T) {
	tests := []struct {
		name                                     string
		queryParams                              string
		filter                                   string
		mockGetObjectsWithComplexFiltersResponse []any
		selectPath                               string
	}{
		{"SelectLayerSelectsAll", "id=BASIC.A.R1.*&namespace=physical.hierarchy", "category=rack", []any{rack1, rack2}, "/Physical/BASIC/A/R1/#racks"},
		{"SelectGroupsLayerSelectsAll", "id=BASIC.A.R1.*&namespace=physical.hierarchy", "category=group", []any{roomGroup}, "/Physical/BASIC/A/R1/#groups"},
		{"SelectLayerAllSelectsAll", "id=BASIC.A.R1.*&namespace=physical.hierarchy", "category=rack", []any{rack1, rack2}, "/Physical/BASIC/A/R1/#racks/*"},
		{"SelectLayerWildcardSelectsWildcard", "id=BASIC.A.R1.A*&namespace=physical.hierarchy", "category=rack", []any{rack1}, "/Physical/BASIC/A/R1/#racks/A*"},
		{"SelectLayerChildSelectsChild", "id=BASIC.A.R1.A01&namespace=physical.hierarchy", "category=rack", []any{rack1}, "/Physical/BASIC/A/R1/#racks/A01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, mockOgree3D := layersSetup(t)

			test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
			test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
			test_utils.MockGetObjectsWithComplexFilters(mockAPI, tt.queryParams, map[string]any{"filter": tt.filter}, tt.mockGetObjectsWithComplexFiltersResponse)
			instancesIds := []string{}
			for _, instance := range tt.mockGetObjectsWithComplexFiltersResponse {
				test_utils.MockGetObject(mockAPI, instance.(map[string]any))
				instancesIds = append(instancesIds, instance.(map[string]any)["id"].(string))
			}

			ogreeData, _ := json.Marshal(instancesIds)
			mockOgree3D.On(
				"InformOptional", "SetClipBoard",
				-1, map[string]any{"data": string(ogreeData), "type": "select"},
			).Return(nil)

			selection, err := controller.Select(tt.selectPath)
			assert.Nil(t, err)
			assert.Len(t, selection, len(tt.mockGetObjectsWithComplexFiltersResponse))
			for _, id := range instancesIds {
				// we have the path of each id present
				path := "/Physical/" + strings.Replace(id, ".", "/", -1)
				assert.Contains(t, selection, path)
			}
		})
	}
}

func TestSelectNestedLayerSelectsAll(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectHierarchy(mockAPI, rack1)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.A01.*&namespace=physical.hierarchy", map[string]any{"filter": "category=group"}, []any{rackGroup})
	test_utils.MockGetObject(mockAPI, rackGroup)

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
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockDeleteObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.*&namespace=physical.hierarchy", map[string]any{"filter": "category=rack"}, []any{rack1, rack2})

	controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

	_, err := controller.DeleteObj("/Physical/BASIC/A/R1/#racks")
	assert.Nil(t, err)
}

func TestDraw(t *testing.T) {
	tests := []struct {
		name  string
		depth int
	}{
		{"LayerDrawsAllObjectsOfTheLayer", 0},
		{"LayerWithDepthDrawsAllObjectsOfTheLayerAndChildren", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, mockOgree3D := layersSetup(t)

			test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
			test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
			test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.*&namespace=physical.hierarchy", map[string]any{"filter": "category=rack"}, []any{rack1, rack2})
			dataRack1 := test_utils.RemoveChildren(rack1)
			dataRack2 := test_utils.RemoveChildren(rack2)
			if tt.depth == 0 {
				test_utils.MockGetObject(mockAPI, rack1)
				test_utils.MockGetObject(mockAPI, rack2)
			} else {
				test_utils.MockGetObjectHierarchy(mockAPI, rack1)
				test_utils.MockGetObjectHierarchy(mockAPI, rack2)
				dataRack1 = test_utils.KeepOnlyDirectChildren(rack1)
				dataRack2 = test_utils.KeepOnlyDirectChildren(rack2)
			}

			controllers.State.ObjsForUnity = controllers.SetObjsForUnity([]string{"all"})

			mockOgree3D.On(
				"Inform", "Draw",
				0, map[string]any{"data": dataRack1, "type": "create"},
			).Return(nil)
			mockOgree3D.On(
				"Inform", "Draw",
				0, map[string]any{"data": dataRack2, "type": "create"},
			).Return(nil)

			err := controller.Draw("/Physical/BASIC/A/R1/#racks", tt.depth, true)
			assert.Nil(t, err)
		})
	}
}

func TestUndrawLayerUndrawAllObjectsOfTheLayer(t *testing.T) {
	controller, mockAPI, mockOgree3D := layersSetup(t)

	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.*&namespace=physical.hierarchy", map[string]any{"filter": "category=rack"}, []any{rack1, rack2})
	test_utils.MockGetObject(mockAPI, rack1)
	test_utils.MockGetObject(mockAPI, rack2)

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

func testApplicability(t *testing.T, path map[string]string, applicability string, expectedApplicability string, errorString string) {
	previousPath, hasPrevPath := path["previousPath"]
	currPath, hasCurrPath := path["currPath"]
	if hasPrevPath {
		controllers.State.PrevPath = previousPath
	} else if hasCurrPath {
		controllers.State.CurrPath = currPath
	}
	applicability, err := controllers.TranslateApplicability(applicability)
	if len(errorString) > 0 {
		assert.NotNil(t, err)
		assert.ErrorContains(t, err, errorString)
	} else {
		assert.Nil(t, err)
		assert.Equal(t, expectedApplicability, applicability)
	}
}

func TestTranslateApplicabilityReturnsErrorIfPathIsRoot(t *testing.T) {
	testApplicability(t, map[string]string{}, "/", "", "applicability must be an hierarchical path, found: /")
}

func TestTranslateApplicabilityReturnsErrorIfPathIsNotHierarchical(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Logical/Tags", "", "applicability must be an hierarchical path, found: /Logical/Tags")
}

func TestTranslateApplicabilityTransformsPhysicalSlashIntoEmpty(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical", "", "")
}

func TestTranslateApplicabilityCleansPathOfLastSlash(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical", "", "")
}

func TestTranslateApplicabilityCleansPathOfSlashPointAtEnd(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/.", "", "")
}

func TestTranslateApplicabilityCleansPathOfSlashPoint(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/./BASIC", "BASIC", "")
}

func TestTranslateApplicabilityTransformsPhysicalPathIntoID(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/BASIC/A", "BASIC.A", "")
}

func TestTranslateApplicabilitySupportsPointPointAtTheEnd(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/BASIC/..", "", "")
}

func TestTranslateApplicabilitySupportsPointPoint(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/BASIC/../COMPLEX/R1", "COMPLEX.R1", "")
}

func TestTranslateApplicabilitySupportsStarAtTheEnd(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/*", "*", "")
}

func TestTranslateApplicabilitySupportsStarStarAtTheEnd(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/**", "**", "")
}

func TestTranslateApplicabilitySupportsStar(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/*/chT", "*.chT", "")
}

func TestTranslateApplicabilitySupportsStarStar(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/**/chT", "**.chT", "")
}

func TestTranslateApplicabilityEmptyReturnsCurrPath(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC/A"}, "", "BASIC.A", "")
}

func TestTranslateApplicabilityPointReturnsCurrPath(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC/A"}, ".", "BASIC.A", "")
}

func TestTranslateApplicabilityPointReturnsErrorIfCurrPathIsNotHierarchical(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Logical/Tags"}, ".", "", "applicability must be an hierarchical path, found: /Logical/Tags")
}

func TestTranslateApplicabilityPointReturnsEmptyIfCurrPathIsSlashPhysical(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical"}, ".", "", "")
}

func TestTranslateApplicabilityPointPathReturnsCurrPathPlusPath(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC"}, "./A", "BASIC.A", "")
}

func TestTranslateApplicabilityRelativePathReturnsCurrPathPlusPath(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC"}, "A", "BASIC.A", "")
}

func TestTranslateApplicabilityRelativePathStarReturnsCurrPathPlusPath(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC"}, "*", "BASIC.*", "")
}

func TestTranslateApplicabilityRelativePathReturnsErrorIfCurrPathIsNotHierarchical(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Logical/Tags"}, "A", "", "applicability must be an hierarchical path, found: /Logical/Tags/A")
}

func TestTranslateApplicabilityPointPointReturnsBeforeCurrPath(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC/A"}, "..", "BASIC", "")
}

func TestTranslateApplicabilityPointPointReturnsEmptyIfBeforeCurrPathIsPhysical(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC"}, "..", "", "")
}

func TestTranslateApplicabilityPointPointReturnsErrorIfBeforeCurrPathIsNotHierarchical(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical"}, "..", "", "applicability must be an hierarchical path, found: /")
}

func TestTranslateApplicabilityPointPointPathReturnsCurrPathPlusPath(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC"}, "../COMPLEX", "COMPLEX", "")
}

func TestTranslateApplicabilityPointPointTwoTimes(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical/BASIC/R1"}, "../../COMPLEX", "COMPLEX", "")
}

func TestTranslateApplicabilityMinusReturnsPrevPath(t *testing.T) {
	testApplicability(t, map[string]string{"previousPath": "/Physical/BASIC"}, "-", "BASIC", "")
}

func TestTranslateApplicabilityMinusPathReturnsPrevPathPlusPath(t *testing.T) {
	testApplicability(t, map[string]string{"previousPath": "/Physical"}, "-/BASIC", "BASIC", "")
}

func TestTranslateApplicabilityUnderscorePathReturnsCurrPathPlusUnderscore(t *testing.T) {
	testApplicability(t, map[string]string{"currPath": "/Physical"}, "_", "_", "")
}

func TestTranslateApplicabilityReturnsErrorIfPatternIsNotValid(t *testing.T) {
	testApplicability(t, map[string]string{}, "/Physical/[", "", "applicability pattern is not valid")
}

func TestLsNotShowLayerIfNotMatch(t *testing.T) {
	tests := []struct {
		name          string
		applicability string
	}{
		{"WithoutStar", "BASIC.A.R2"},
		{"WithStar", "BASIC.*"},
		{"WithDoubleStar", "BASIC.B.**"},
		{"WithDoubleStarAndMore", "BASIC.**.chT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)
			test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{
				map[string]any{
					"slug":                    "test",
					models.LayerApplicability: tt.applicability,
					models.LayerFilters:       "any = yes",
				},
			})
			test_utils.MockGetObjectHierarchy(mockAPI, roomWithoutChildren)

			objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, nil)
			assert.Nil(t, err)
			assert.Len(t, objects, 0)
		})
	}
}

func TestLsShowLayerIfMatch(t *testing.T) {
	tests := []struct {
		name          string
		applicability string
	}{
		{"PerfectMatchWithoutStar", "BASIC.A.R1"},
		{"MatchWithOneStar", "BASIC.A.*"},
		{"MatchWithSomethingStar", "BASIC.A.R*"},
		{"MatchWithDoubleStar", "BASIC.**"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller, mockAPI, _ := layersSetup(t)
			test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{
				map[string]any{
					"slug":                    "test",
					models.LayerApplicability: tt.applicability,
					models.LayerFilters:       "any = yes",
				},
			})
			test_utils.MockGetObjectHierarchy(mockAPI, roomWithoutChildren)

			objects, err := controller.Ls("/Physical/BASIC/A/R1", nil, nil)
			assert.Nil(t, err)
			assert.Len(t, objects, 1)
			utils.ContainsObjectNamed(t, objects, "#test")
		})
	}
}

func TestLsShowLayerIfPerfectMatchOnPhysical(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "",
			models.LayerFilters:       "any = yes",
		},
	})
	test_utils.MockGetObjectsByEntity(mockAPI, "sites", []any{})

	objects, err := controller.Ls("/Physical", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "Stray")
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsShowLayerIfPerfectMatchOnPhysicalChild(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC",
			models.LayerFilters:       "any = yes",
		},
	})
	test_utils.MockGetObjectHierarchy(mockAPI, map[string]any{
		"category": "site",
		"children": []any{},
		"id":       "BASIC",
		"name":     "BASIC",
		"parentId": "",
	})

	objects, err := controller.Ls("/Physical/BASIC", nil, nil)
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
	test_utils.MockGetObjectsByEntity(mockAPI, "sites", []any{site})

	_, err := controller.Tree("/Physical", 1)
	assert.Nil(t, err)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC",
			models.LayerFilters:       "any = yes",
		},
	})
	test_utils.MockGetObjectHierarchy(mockAPI, site)

	objects, err := controller.Ls("/Physical/BASIC", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsShowLayerIfMatchWithDoubleStarAndMore(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{
		map[string]any{
			"slug":                    "test",
			models.LayerApplicability: "BASIC.**.A01",
			models.LayerFilters:       "any = yes",
		},
	})
	test_utils.MockGetObjectHierarchy(mockAPI, test_utils.EmptyChildren(rack1))

	objects, err := controller.Ls("/Physical/BASIC/A/R1/A01", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsReturnsLayerCreatedAfterLastUpdate(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})

	objects, err := controller.Ls("/Logical/Layers", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)

	test_utils.MockCreateObject(mockAPI, "layer", map[string]any{
		"slug":                    "test",
		models.LayerFilters:       "key = value",
		models.LayerApplicability: "BASIC.A.R1",
	})
	err = controller.CreateLayer("test", "/Physical/BASIC/A/R1", "key = value")
	assert.Nil(t, err)

	objects, err = controller.Ls("/Logical/Layers", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "test", objects[0]["name"])
}

func TestLsReturnsLayerCreatedAndUpdatedAfterLastUpdate(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{})

	objects, err := controller.Ls("/Logical/Layers", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)

	testLayer := map[string]any{
		"slug":                    "test",
		models.LayerFilters:       "key = value",
		models.LayerApplicability: "BASIC.A.R1",
	}

	test_utils.MockCreateObject(mockAPI, "layer", testLayer)
	err = controller.CreateLayer("test", "/Physical/BASIC/A/R1", "key = value")
	assert.Nil(t, err)

	test_utils.MockGetObjectByEntity(mockAPI, "layers", testLayer)
	test_utils.MockUpdateObject(mockAPI, map[string]any{
		models.LayerFilters: "& (category = device)",
	}, map[string]any{
		"slug":                    "test",
		models.LayerFilters:       "(key = value) & (category = device)",
		models.LayerApplicability: "BASIC.A.R1",
	})

	err = controller.UpdateLayer("/Logical/Layers/test", models.LayerFiltersAdd, "category = device")
	assert.Nil(t, err)

	objects, err = controller.Ls("/Logical/Layers", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	assert.Equal(t, "test", objects[0]["name"])

	test_utils.MockGetObjectHierarchy(mockAPI, roomWithoutChildren)

	objects, err = controller.Ls("/Physical/BASIC/A/R1", nil, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 1)
	utils.ContainsObjectNamed(t, objects, "#test")
}

func TestLsOnLayerUpdatedAfterLastUpdateDoesUpdatedFilter(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	testLayer := map[string]any{
		"slug":                    "test",
		models.LayerFilters:       "category = rack",
		models.LayerApplicability: "BASIC.A.R1",
	}

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{testLayer})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithoutChildren)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.*&namespace=physical.hierarchy", map[string]any{"filter": "category = rack"}, []any{})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#test", map[string]string{}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)

	test_utils.MockGetObjectByEntity(mockAPI, "layers", testLayer)
	test_utils.MockUpdateObject(mockAPI, map[string]any{
		models.LayerFilters: "& (category = device)",
	}, map[string]any{
		"slug":                    "test",
		models.LayerFilters:       "category = device",
		models.LayerApplicability: "BASIC.A.R1",
	})

	err = controller.UpdateLayer("/Logical/Layers/test", models.LayerFiltersAdd, "category = device")
	assert.Nil(t, err)

	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.*&namespace=physical.hierarchy", map[string]any{"filter": "category = device"}, []any{})

	objects, err = controller.Ls("/Physical/BASIC/A/R1/#test", map[string]string{}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 0)
}

func TestLsOnUserDefinedLayerAppliesFilters(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	testLayer := map[string]any{
		"slug":                    "test",
		models.LayerFilters:       "category = rack",
		models.LayerApplicability: "BASIC.A.R1",
	}

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{testLayer})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.*&namespace=physical.hierarchy", map[string]any{"filter": "category = rack"}, []any{rack1, rack2})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#test", map[string]string{}, nil)
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "A01")
	utils.ContainsObjectNamed(t, objects, "B01")
}

func TestLsRecursiveOnLayerListLayerRecursive(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	devices := map[string]any{
		"slug":                    "devices",
		models.LayerFilters:       "category = device",
		models.LayerApplicability: "BASIC.A.R1",
	}

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{devices})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.**.*&namespace=physical.hierarchy", map[string]any{"filter": "category = device"}, []any{chassis, pdu})

	objects, err := controller.Ls("/Physical/BASIC/A/R1/#devices", map[string]string{}, &controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth})
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	utils.ContainsObjectNamed(t, objects, "chT")
	utils.ContainsObjectNamed(t, objects, "pdu")
}

func TestGetRecursiveOnLayerReturnsLayerRecursive(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	devices := map[string]any{
		"slug":                    "devices",
		models.LayerFilters:       "category = device",
		models.LayerApplicability: "BASIC.A.R1",
	}

	test_utils.MockGetObjectsByEntity(mockAPI, "layers", []any{devices})
	test_utils.MockGetObjectHierarchy(mockAPI, roomWithChildren)
	test_utils.MockGetObjectsWithComplexFilters(mockAPI, "id=BASIC.A.R1.**.*&namespace=physical.hierarchy", map[string]any{"filter": "category = device"}, []any{chassis, pdu})

	objects, _, err := controller.GetObjectsWildcard("/Physical/BASIC/A/R1/#devices", nil, &controllers.RecursiveParams{MaxDepth: models.UnlimitedDepth})
	assert.Nil(t, err)
	assert.Len(t, objects, 2)
	assert.Contains(t, objects, test_utils.RemoveChildren(chassis))
	assert.Contains(t, objects, test_utils.RemoveChildren(pdu))
}
