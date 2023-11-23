package models

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
	ROOMTMPL
	OBJTMPL
	BLDGTMPL
	GROUP
	STRAY_DEV
	DOMAIN
	TAG
)

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
	case "device", "dv":
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
	default:
		return -1
	}
}

func GetParentOfEntity(ent int) int {
	switch ent {
	case SITE:
		return ent - 1
	case BLDG:
		return ent - 1
	case ROOM:
		return ent - 1
	case RACK:
		return ent - 1
	case DEVICE:
		return ent - 1
	case AC:
		return ROOM
	case PWRPNL:
		return ROOM
	case ROOMTMPL:
		return -1
	case BLDGTMPL:
		return -1
	case OBJTMPL:
		return -1
	case CABINET:
		return ROOM
	case GROUP:
		return -1
	case CORRIDOR:
		return ROOM
	default:
		return -3
	}
}
