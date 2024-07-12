package models

import (
	"cli/utils"
	"fmt"
)

// Room inner attributes
type Pillar struct {
	CenterXY []float64 `json:"centerXY"`
	SizeXY   []float64 `json:"sizeXY"`
	Rotation float64   `json:"rotation"`
}

func ValuesToPillar(values []any) (string, Pillar, error) {
	if len(values) != 4 {
		return "", Pillar{}, fmt.Errorf("4 values (name, centerXY, sizeXY, rotation) expected to add a pillar")
	}
	name, err := utils.ValToString(values[0], "name")
	if err != nil {
		return name, Pillar{}, err
	}
	centerXY, err := utils.ValToVec(values[1], 2, "centerXY")
	if err != nil {
		return name, Pillar{}, err
	}
	sizeXY, err := utils.ValToVec(values[2], 2, "sizeXY")
	if err != nil {
		return name, Pillar{}, err
	}
	rotation, err := utils.ValToFloat(values[3], "rotation")
	if err != nil {
		return name, Pillar{}, err
	}
	return name, Pillar{centerXY, sizeXY, rotation}, nil
}

type Separator struct {
	StartPos []float64 `json:"startPosXYm"`
	EndPos   []float64 `json:"endPosXYm"`
	Type     string    `json:"type"`
}

func ValuesToSeparator(values []any) (string, Separator, error) {
	if len(values) != 4 {
		return "", Separator{}, fmt.Errorf("4 values (name, startPos, endPos, type) expected to add a separator")
	}
	name, err := utils.ValToString(values[0], "name")
	if err != nil {
		return name, Separator{}, err
	}
	startPos, err := utils.ValToVec(values[1], 2, "startPos")
	if err != nil {
		return name, Separator{}, err
	}
	endPos, err := utils.ValToVec(values[2], 2, "endPos")
	if err != nil {
		return name, Separator{}, err
	}
	sepType, err := utils.ValToString(values[3], "separator type")
	if err != nil {
		return name, Separator{}, err
	}
	return name, Separator{startPos, endPos, sepType}, nil
}

func ApplyRoomTemplateAttributes(attr, tmpl map[string]any) {
	//Copy Room specific attributes
	utils.CopyMapVal(attr, tmpl, "technicalArea")
	if _, ok := attr["technicalArea"]; ok {
		attr["technical"] = attr["technicalArea"]
		delete(attr, "technicalArea")
	}

	utils.CopyMapVal(attr, tmpl, "reservedArea")
	if _, ok := attr["reservedArea"]; ok {
		attr["reserved"] = attr["reservedArea"]
		delete(attr, "reservedArea")
	}

	for _, attrName := range []string{"axisOrientation", "separators",
		"pillars", "floorUnit", "tiles", "rows", "aisles",
		"vertices", "colors", "tileAngle"} {
		utils.CopyMapVal(attr, tmpl, attrName)
	}
}
