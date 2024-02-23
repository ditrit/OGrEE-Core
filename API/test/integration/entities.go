package integration

import (
	"log"
	"p3/models"
	"p3/utils"
	"strings"
)

var ManagerUserRoles = map[string]models.Role{
	models.ROOT_DOMAIN: models.Manager,
}

func createObject(entity int, obj map[string]interface{}, require bool) (map[string]any, *utils.Error) {
	createdObj, err := models.CreateEntity(
		entity,
		obj,
		ManagerUserRoles,
	)

	if require && err != nil {
		log.Fatalln(err.Error())
	}

	return createdObj, err
}

func internalCreateSite(name string, require bool) (map[string]any, *utils.Error) {
	return createObject(
		utils.SITE,
		map[string]any{
			"attributes": map[string]any{
				"reservedColor":  "AAAAAA",
				"technicalColor": "D0FF78",
				"usableColor":    "5BDCFF",
			},
			"category":    "site",
			"description": name,
			"domain":      TestDBName,
			"name":        name,
		},
		require,
	)
}

func RequireCreateSite(name string) map[string]any {
	obj, _ := internalCreateSite(name, true)
	return obj
}

func CreateSite(name string) (map[string]any, *utils.Error) {
	return internalCreateSite(name, false)
}

func internalCreateBuilding(siteID, name string, require bool) (map[string]any, *utils.Error) {
	if siteID == "" {
		site := RequireCreateSite(name + "-site")
		siteID = site["id"].(string)
	}

	return createObject(
		utils.BLDG,
		map[string]any{
			"attributes": map[string]any{
				"height":     "5",
				"heightUnit": "m",
				"posXY":      "[50, 0]",
				"posXYUnit":  "m",
				"size":       "[49, 46.6]",
				"sizeUnit":   "m",
				"rotation":   "30.5",
			},
			"category":    "building",
			"description": name,
			"domain":      TestDBName,
			"name":        name,
			"parentId":    siteID,
		},
		require,
	)
}

func RequireCreateBuilding(siteID, name string) map[string]any {
	obj, _ := internalCreateBuilding(siteID, name, true)
	return obj
}

func CreateBuilding(siteID, name string) (map[string]any, *utils.Error) {
	return internalCreateBuilding(siteID, name, false)
}

func internalCreateRoom(buildingID, name string, require bool) (map[string]any, *utils.Error) {
	if buildingID == "" {
		building := RequireCreateBuilding("", name+"-building")
		buildingID = building["id"].(string)
	}

	return createObject(
		utils.ROOM,
		map[string]any{
			"attributes": map[string]any{
				"floorUnit":       "t",
				"height":          "2.8",
				"heightUnit":      "m",
				"axisOrientation": "+x+y",
				"rotation":        "-90",
				"posXY":           "[0, 0]",
				"posXYUnit":       "m",
				"size":            "[-13, -2.9]",
				"sizeUnit":        "m",
				"template":        "",
			},
			"category":    "room",
			"description": name,
			"domain":      TestDBName,
			"name":        name,
			"parentId":    buildingID,
		},
		require,
	)
}

func RequireCreateRoom(buildingID, name string) map[string]any {
	obj, _ := internalCreateRoom(buildingID, name, true)
	return obj
}

func CreateRoom(buildingID, name string) (map[string]any, *utils.Error) {
	return internalCreateRoom(buildingID, name, false)
}

func internalCreateRack(roomID, name string, require bool) (map[string]any, *utils.Error) {
	if roomID == "" {
		room := RequireCreateRoom("", name+"-room")
		roomID = room["id"].(string)
	}

	return createObject(
		utils.RACK,
		map[string]any{
			"attributes": map[string]any{
				"height":     "47",
				"heightUnit": "U",
				"rotation":   "[45, 45, 45]",
				"posXYUnit":  "m",
				"posXYZ":     "[4.6666666666667, -2, 0]",
				"size":       "[80, 100.532442]",
				"sizeUnit":   "cm",
				"template":   "",
			},
			"category":    "rack",
			"description": name,
			"domain":      TestDBName,
			"name":        name,
			"parentId":    roomID,
		},
		require,
	)
}

func RequireCreateRack(roomID, name string) map[string]any {
	obj, _ := internalCreateRack(roomID, name, true)
	return obj
}

func CreateRack(roomID, name string) (map[string]any, *utils.Error) {
	return internalCreateRack(roomID, name, false)
}

func internalCreateCorridor(roomID, name string, require bool) (map[string]any, *utils.Error) {
	if roomID == "" {
		room := RequireCreateRoom("", name+"-room")
		roomID = room["id"].(string)
	}

	return createObject(
		utils.CORRIDOR,
		map[string]any{
			"attributes": map[string]any{
				"color":       "000099",
				"content":     "B11,C19",
				"temperature": "cold",
				"height":      "470",
				"heightUnit":  "cm",
				"rotation":    "[45, 45, 45]",
				"posXYUnit":   "m",
				"posXYZ":      "[4.6666666666667, -2, 0]",
				"size":        "[80, 100.532442]",
				"sizeUnit":    "cm",
			},
			"category":    "corridor",
			"description": "corridor",
			"domain":      TestDBName,
			"name":        name,
			"parentId":    roomID,
		},
		require,
	)
}

func RequireCreateCorridor(roomID, name string) map[string]any {
	obj, _ := internalCreateCorridor(roomID, name, true)
	return obj
}

func CreateCorridor(roomID, name string) (map[string]any, *utils.Error) {
	return internalCreateCorridor(roomID, name, false)
}

func internalCreateGeneric(roomID, name string, require bool) (map[string]any, *utils.Error) {
	if roomID == "" {
		room := RequireCreateRoom("", name+"-room")
		roomID = room["id"].(string)
	}

	return createObject(
		utils.GENERIC,
		map[string]any{
			"attributes": map[string]any{
				"height":     "47",
				"heightUnit": "cm",
				"rotation":   "[45, 45, 45]",
				"posXYZ":     "[4.6666666666667, -2, 0]",
				"posXYUnit":  "m",
				"size":       "[80, 100.532442]",
				"shape":      "cube",
				"sizeUnit":   "cm",
				"template":   "",
				"type":       "box",
			},
			"category":    "generic",
			"description": name,
			"domain":      TestDBName,
			"name":        name,
			"parentId":    roomID,
		},
		require,
	)
}

func RequireCreateGeneric(roomID, name string) map[string]any {
	obj, _ := internalCreateGeneric(roomID, name, true)
	return obj
}

func CreateGeneric(roomID, name string) (map[string]any, *utils.Error) {
	return internalCreateGeneric(roomID, name, false)
}

func internalCreateDevice(parentID, name string, require bool) (map[string]any, *utils.Error) {
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
				"size":        "[388.4, 205.9]",
				"sizeUnit":    "mm",
				"template":    "huawei-xxxxxx",
				"type":        "blade",
				"vendor":      "Huawei",
				"weightKg":    1.81,
			},
			"category":    "device",
			"description": name,
			"domain":      TestDBName,
			"name":        name,
			"parentId":    parentID,
		},
		require,
	)
}

func RequireCreateDevice(parentID, name string) map[string]any {
	obj, _ := internalCreateDevice(parentID, name, true)
	return obj
}

func CreateDevice(parentID, name string) (map[string]any, *utils.Error) {
	return internalCreateDevice(parentID, name, false)
}

func internalCreateGroup(parentID, name string, content []string, require bool) (map[string]any, *utils.Error) {
	return createObject(
		utils.GROUP,
		map[string]any{
			"attributes": map[string]any{
				"content": strings.Join(content, ","),
			},
			"category":    "group",
			"description": name,
			"domain":      TestDBName,
			"name":        name,
			"parentId":    parentID,
		},
		require,
	)
}

func RequireCreateGroup(parentID, name string, content []string) map[string]any {
	obj, _ := internalCreateGroup(parentID, name, content, true)
	return obj
}

func CreateGroup(parentID, name string, content []string) (map[string]any, *utils.Error) {
	return internalCreateGroup(parentID, name, content, false)
}
