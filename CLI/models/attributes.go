package models

// Auxillary functions for parsing and validation of data
// before the CLI sends off to API

import (
	l "cli/logger"
	"fmt"
	"strconv"
	"strings"
)

type EntityAttributes map[string]any

var BldgBaseAttrs = EntityAttributes{
	"posXYUnit":  "m",
	"sizeUnit":   "m",
	"heightUnit": "m",
}

var RoomBaseAttrs = EntityAttributes{
	"floorUnit":  "t",
	"posXYUnit":  "m",
	"sizeUnit":   "m",
	"heightUnit": "m",
}

var RackBaseAttrs = EntityAttributes{
	"sizeUnit":   "cm",
	"heightUnit": "U",
}

var DeviceBaseAttrs = EntityAttributes{
	"orientation": "front",
	"sizeUnit":    "mm",
	"heightUnit":  "mm",
}

var GenericBaseAttrs = EntityAttributes{
	"sizeUnit":   "cm",
	"heightUnit": "cm",
}

var CorridorBaseAttrs = GenericBaseAttrs

var BaseAttrs = map[int]EntityAttributes{
	BLDG:     BldgBaseAttrs,
	ROOM:     RoomBaseAttrs,
	RACK:     RackBaseAttrs,
	DEVICE:   DeviceBaseAttrs,
	GENERIC:  GenericBaseAttrs,
	CORRIDOR: CorridorBaseAttrs,
}

const referToWikiMsg = " Please refer to the wiki or manual reference" +
	" for more details on how to create objects " +
	"using this syntax"

func SetPosAttr(ent int, attr EntityAttributes) error {
	switch ent {
	case BLDG, ROOM:
		return SetPosXY(attr)
	case RACK, CORRIDOR, GENERIC:
		return SetPosXYZ(attr)
	default:
		return fmt.Errorf("invalid entity for pos attribution")
	}
}

func SetPosXY(attr EntityAttributes) error {
	attr["posXY"] = SerialiseVector(attr, "posXY")
	if posXY, ok := attr["posXY"].([]float64); !ok || len(posXY) != 2 {
		l.GetErrorLogger().Println(
			"User gave invalid posXY value")
		return fmt.Errorf("invalid posXY attribute provided." +
			" \nIt must be an array/list/vector with 2 elements." +
			referToWikiMsg)
	}
	return nil
}

func SetPosXYZ(attr EntityAttributes) error {
	attr["posXYZ"] = SerialiseVector(attr, "posXYZ")
	if posXY, ok := attr["posXYZ"].([]float64); !ok || len(posXY) != 3 {
		l.GetErrorLogger().Println(
			"User gave invalid pos value")
		return fmt.Errorf("invalid pos attribute provided." +
			" \nIt must be an array/list/vector with 2 or 3 elements." +
			referToWikiMsg)
	}
	return nil
}

func SetSize(attr map[string]any) error {
	attr["size"] = SerialiseVector(attr, "size")
	if _, ok := attr["size"].([]any); !ok {
		if size, ok := attr["size"].([]float64); !ok || len(size) == 0 {
			l.GetErrorLogger().Println(
				"User gave invalid size value")
			return fmt.Errorf("invalid size attribute provided." +
				" \nIt must be an array/list/vector with 3 elements." +
				referToWikiMsg)

		}
	}
	return nil
}

func SerialiseVector(attr map[string]interface{}, want string) []float64 {
	if vector, ok := attr[want].([]float64); ok {
		if want == "size" && len(vector) == 3 {
			attr["height"] = vector[2]
			vector = vector[:len(vector)-1]
		} else if want == "posXYZ" && len(vector) == 2 {
			vector = append(vector, 0)
		}
		return vector
	} else {
		return []float64{}
	}
}

func SetDeviceSizeUIfExists(attr EntityAttributes) {
	if sizeU, ok := attr["sizeU"]; ok {
		//Convert block
		//And Set height
		if sizeUInt, ok := sizeU.(int); ok {
			attr["sizeU"] = sizeUInt
			attr["height"] = float64(sizeUInt) * 44.5
		} else if sizeUFloat, ok := sizeU.(float64); ok {
			attr["sizeU"] = sizeUFloat
			attr["height"] = sizeUFloat * 44.5
		}
	}
}

func SetDeviceSlotOrPosU(attr EntityAttributes) error {
	//Process the posU/slot attribute
	if x, ok := attr["posU/slot"].([]string); ok && len(x) > 0 {
		delete(attr, "posU/slot")
		if posU, err := strconv.Atoi(x[0]); len(x) == 1 && err == nil {
			attr["posU"] = posU
		} else {
			if slots, err := CheckExpandStrVector(x); err != nil {
				return err
			} else {
				attr["slot"] = slots
			}
		}
	}
	return nil
}

// CheckExpandStrVector: allow usage of .. on device slot and group content vector
// converting [slot01..slot03] on [slot01,slot02,slot03]
func CheckExpandStrVector(slotVector []string) ([]string, error) {
	slots := []string{}
	for _, slot := range slotVector {
		if strings.Contains(slot, "..") {
			if len(slotVector) != 1 {
				return nil, fmt.Errorf("Invalid device syntax: .. can only be used in a single element vector")
			}
			return expandStrToVector(slot)
		} else {
			slots = append(slots, slot)
		}
	}
	return slots, nil
}

func expandStrToVector(slot string) ([]string, error) {
	slots := []string{}
	errMsg := "Invalid device syntax: incorrect use of .. for slot"
	parts := strings.Split(slot, "..")
	if len(parts) != 2 ||
		(parts[0][:len(parts[0])-1] != parts[1][:len(parts[1])-1]) {
		l.GetWarningLogger().Println(errMsg)
		return nil, fmt.Errorf(errMsg)
	} else {
		start, errS := strconv.Atoi(string(parts[0][len(parts[0])-1]))
		end, errE := strconv.Atoi(string(parts[1][len(parts[1])-1]))
		if errS != nil || errE != nil {
			l.GetWarningLogger().Println(errMsg)
			return nil, fmt.Errorf(errMsg)
		} else {
			prefix := parts[0][:len(parts[0])-1]
			for i := start; i <= end; i++ {
				slots = append(slots, prefix+strconv.Itoa(i))
			}
			return slots, nil
		}
	}
}

// Validate for cmd [room]:areas=[r1,r2,r3,r4]@[t1,t2,t3,t4]
func SetRoomAreas(values []any) (map[string]any, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("2 values (reserved, technical) expected to set room areas")
	}
	areas := map[string]any{"reserved": values[0], "technical": values[1]}
	reserved, hasReserved := areas["reserved"].([]float64)
	if !hasReserved {
		return nil, ErrorResponder("reserved", "4", false)
	}
	tech, hasTechnical := areas["technical"].([]float64)
	if !hasTechnical {
		return nil, ErrorResponder("technical", "4", false)
	}

	if len(reserved) == 4 && len(tech) == 4 {
		return areas, nil
	} else {
		if len(reserved) != 4 && len(tech) == 4 {
			return nil, ErrorResponder("reserved", "4", false)
		} else if len(tech) != 4 && len(reserved) == 4 {
			return nil, ErrorResponder("technical", "4", false)
		} else { //Both invalid
			return nil, ErrorResponder("reserved and technical", "4", true)
		}
	}
}

// errResponder helper func for specialUpdateNode
// used for separator, pillar err msgs and validateRoomAreas()
func ErrorResponder(attr, numElts string, multi bool) error {
	var errorMsg string
	if multi {
		errorMsg = "Invalid " + attr + " attributes provided." +
			" They must be arrays/lists/vectors with " + numElts + " elements."
	} else {
		errorMsg = "Invalid " + attr + " attribute provided." +
			" It must be an array/list/vector with " + numElts + " elements."
	}

	segment := " Please refer to the wiki or manual reference" +
		" for more details on how to create objects " +
		"using this syntax"

	return fmt.Errorf(errorMsg + segment)
}
