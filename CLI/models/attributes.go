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
			" Please refer to the wiki or manual reference" +
			" for more details on how to create objects " +
			"using this syntax")
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
			" Please refer to the wiki or manual reference" +
			" for more details on how to create objects " +
			"using this syntax")
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
				" Please refer to the wiki or manual reference" +
				" for more details on how to create objects " +
				"using this syntax")

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
			if slots, err := ExpandStrVector(x); err != nil {
				return err
			} else {
				attr["slot"] = slots
			}
		}
	}
	return nil
}

// ExpandStrVector: allow usage of .. on device slot and group content vector
// converting [slot01..slot03] on [slot01,slot02,slot03]
func ExpandStrVector(slotVector []string) ([]string, error) {
	slots := []string{}
	for _, slot := range slotVector {
		if strings.Contains(slot, "..") {
			if len(slotVector) != 1 {
				return nil, fmt.Errorf("Invalid device syntax: .. can only be used in a single element vector")
			}
			parts := strings.Split(slot, "..")
			if len(parts) != 2 ||
				(parts[0][:len(parts[0])-1] != parts[1][:len(parts[1])-1]) {
				l.GetWarningLogger().Println("Invalid device syntax encountered")
				return nil, fmt.Errorf("Invalid device syntax: incorrect use of .. for slot")
			} else {
				start, errS := strconv.Atoi(string(parts[0][len(parts[0])-1]))
				end, errE := strconv.Atoi(string(parts[1][len(parts[1])-1]))
				if errS != nil || errE != nil {
					l.GetWarningLogger().Println("Invalid device syntax encountered")
					return nil, fmt.Errorf("Invalid device syntax: incorrect use of .. for slot")
				} else {
					prefix := parts[0][:len(parts[0])-1]
					for i := start; i <= end; i++ {
						slots = append(slots, prefix+strconv.Itoa(i))
					}
				}
			}
		} else {
			slots = append(slots, slot)
		}
	}
	return slots, nil
}
