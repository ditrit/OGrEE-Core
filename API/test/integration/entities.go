package integration

import (
	"log"
	"p3/models"
	"p3/utils"
)

var managerUserRoles = map[string]models.Role{
	models.ROOT_DOMAIN: models.Manager,
}

func createObject(entity int, obj map[string]interface{}) map[string]any {
	createdObj, err := models.CreateEntity(
		entity,
		obj,
		managerUserRoles,
	)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return createdObj
}

func CreateSite(name string) map[string]any {
	return createObject(
		utils.SITE,
		map[string]any{
			"attributes": map[string]any{
				"reservedColor":  "AAAAAA",
				"technicalColor": "D0FF78",
				"usableColor":    "5BDCFF",
			},
			"category":    "site",
			"description": []any{name},
			"domain":      TestDBName,
			"name":        name,
		},
	)
}

func CreateBuilding(siteID, name string) map[string]any {
	return createObject(
		utils.BLDG,
		map[string]any{
			"attributes": map[string]any{
				"height":     "5",
				"heightUnit": "m",
				"posXY":      "{\"x\":50 ,\"y\":0}",
				"posXYUnit":  "m",
				"size":       "{\"x\":49 ,\"y\":46.6}",
				"sizeUnit":   "m",
				"rotation":   "30.5",
			},
			"category":    "building",
			"description": []any{name},
			"domain":      TestDBName,
			"name":        name,
			"parentId":    siteID,
		},
	)
}

func CreateRoom(buildingID, name string) map[string]any {
	return createObject(
		utils.ROOM,
		map[string]any{
			"attributes": map[string]any{
				"floorUnit":       "t",
				"height":          "2.8",
				"heightUnit":      "m",
				"axisOrientation": "+x+y",
				"rotation":        "-90",
				"posXY":           "{\"x\":0,\"y\":0}",
				"posXYUnit":       "m",
				"size":            "{\"x\": -13,\"y\":-2.9}",
				"sizeUnit":        "m",
				"template":        "",
			},
			"category":    "room",
			"description": []any{name},
			"domain":      TestDBName,
			"name":        name,
			"parentId":    buildingID,
		},
	)
}

func CreateRack(roomID, name string) map[string]any {
	return createObject(
		utils.RACK,
		map[string]any{
			"attributes": map[string]any{
				"height":     "47",
				"heightUnit": "U",
				"rotation":   "{\"x\":45 ,\"y\":45 ,\"z\":45}",
				"posXYUnit":  "m",
				"posXYZ":     "{\"x\":4.6666666666667 ,\"y\": -2 ,\"z\":0}",
				"size":       "{\"x\":80 ,\"y\":100.532442}",
				"sizeUnit":   "cm",
				"template":   "",
			},
			"category":    "rack",
			"description": []any{name},
			"domain":      TestDBName,
			"name":        name,
			"parentId":    roomID,
		},
	)
}

func CreateDevice(parentID, name string) map[string]any {
	return createObject(
		utils.DEVICE,
		map[string]any{
			"attributes": map[string]any{
				"TDP":         "",
				"TDPmax":      "",
				"fbxModel":    "https://github.com/test.fbx",
				"height":      "40.1",
				"heightUnit":  "mm",
				"model":       "TNF2LTX",
				"orientation": "front",
				"partNumber":  "0303XXXX",
				"size":        "{\"x\":388.4, \"y\":205.9}",
				"sizeUnit":    "mm",
				"slot":        "slot6",
				"template":    "huawei-xxxxxx",
				"type":        "blade",
				"vendor":      "Huawei",
				"weightKg":    1.81,
			},
			"category":    "device",
			"description": []any{name},
			"domain":      TestDBName,
			"name":        name,
			"parentId":    parentID,
		},
	)
}
