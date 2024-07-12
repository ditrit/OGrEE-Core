package models

import (
	l "cli/logger"
	"fmt"
	pathutil "path"
	"time"
)

const (
	SITE = iota
	BLDG
	ROOM
	RACK
	DEVICE
	AC
	PWRPNL
	CABINET
	CORRIDOR
	GENERIC
	ROOMTMPL
	OBJTMPL
	BLDGTMPL
	GROUP
	STRAY_DEV
	DOMAIN
	TAG
	LAYER
	VIRTUALOBJ
)

type Entity struct {
	Category    string           `json:"category"`
	Description string           `json:"description"`
	Domain      string           `json:"domain"`
	CreatedDate *time.Time       `json:"createdDate,omitempty"`
	LastUpdated *time.Time       `json:"lastUpdated,omitempty"`
	Name        string           `json:"name"`
	Id          string           `json:"id,omitempty"`
	ParentId    string           `json:"parentId,omitempty"`
	Attributes  EntityAttributes `json:"attributes"`
}

func EntityToString(entity int) string {
	switch entity {
	case DOMAIN:
		return "domain"
	case SITE:
		return "site"
	case BLDG:
		return "building"
	case ROOM:
		return "room"
	case RACK:
		return "rack"
	case DEVICE:
		return "device"
	case AC:
		return "ac"
	case PWRPNL:
		return "panel"
	case STRAY_DEV:
		return "stray_device"
	case ROOMTMPL:
		return "room_template"
	case OBJTMPL:
		return "obj_template"
	case BLDGTMPL:
		return "bldg_template"
	case CABINET:
		return "cabinet"
	case GROUP:
		return "group"
	case CORRIDOR:
		return "corridor"
	case TAG:
		return "tag"
	case LAYER:
		return "layer"
	case GENERIC:
		return "generic"
	case VIRTUALOBJ:
		return "virtual_obj"
	default:
		return "INVALID"
	}
}

func EntityStrToInt(entity string) int {
	switch entity {
	case "domain":
		return DOMAIN
	case "site", "si":
		return SITE
	case "building", "bldg", "bd":
		return BLDG
	case "room", "ro":
		return ROOM
	case "rack", "rk":
		return RACK
	case "device", "dv", "dev":
		return DEVICE
	case "ac":
		return AC
	case "panel", "pn":
		return PWRPNL
	case "stray_device":
		return STRAY_DEV
	case "room_template":
		return ROOMTMPL
	case "obj_template":
		return OBJTMPL
	case "bldg_template":
		return BLDGTMPL
	case "cabinet", "cb":
		return CABINET
	case "group", "gr":
		return GROUP
	case "corridor", "co":
		return CORRIDOR
	case "tag":
		return TAG
	case "layer":
		return LAYER
	case "generic", "ge":
		return GENERIC
	case "vobj", "virtual_obj":
		return VIRTUALOBJ
	default:
		return -1
	}
}

func GetParentOfEntity(ent int) int {
	switch ent {
	case SITE, BLDG, ROOM, DEVICE:
		return ent - 1
	case ROOMTMPL, BLDGTMPL, OBJTMPL, GROUP:
		return -1
	case RACK, AC, PWRPNL, CABINET, CORRIDOR, GENERIC:
		return ROOM
	default:
		return -3
	}
}

func EntityCreationMustBeInformed(entity int) bool {
	return entity != TAG
}

func SetObjectBaseData(entity int, path string, data map[string]any) error {
	name := pathutil.Base(path)
	if name == "." || name == "" {
		l.GetWarningLogger().Println("Invalid path name provided for OCLI object creation")
		return fmt.Errorf("invalid path name provided for OCLI object creation")
	}
	data["name"] = name
	data["category"] = EntityToString(entity)
	data["description"] = ""
	if _, hasAttributes := data["attributes"].(map[string]any); !hasAttributes {
		data["attributes"] = map[string]any{}
	}
	return nil
}
