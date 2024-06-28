package ogreetypes

import (
	"encoding/json"
)

type BuildingTemplateAttributes struct {
	Slug     Slug      `json:"slug"`
	Center   Vector2   `json:"center"`
	SizeWDHm Vector3   `json:"sizeWDHm"`
	Vertices []Vector2 `json:"vertices"`
}

type BuildingAttributes struct {
	Height     float64            `json:"height"`
	HeightUnit MetricImperialUnit `json:"heightUnit"`
	PosXY      Vector2            `json:"-"`
	PosXYUnit  MetricImperialUnit `json:"posXYUnit"`
	Size       Vector2            `json:"-"`
	SizeUnit   MetricImperialUnit `json:"sizeUnit"`
	Rotation   float64            `json:"rotation"`
	Template   string             `json:"template,omitempty"`
}

type BuildingAttributesAlias BuildingAttributes

type BuildingAttributesJson struct {
	BuildingAttributesAlias
	PosXYAux Vector2Wrapper `json:"posXY"`
	SizeAux  Vector2Wrapper `json:"size"`
}

func (b BuildingAttributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(BuildingAttributesJson{
		BuildingAttributesAlias: BuildingAttributesAlias(b),
		PosXYAux:                Vector2Wrapper{b.PosXY},
		SizeAux:                 Vector2Wrapper{b.Size},
	})
}

func (b *BuildingAttributes) UnmarshalJSON(data []byte) error {
	var bjson BuildingAttributesJson
	if err := json.Unmarshal(data, &bjson); err != nil {
		return err
	}
	*b = BuildingAttributes(bjson.BuildingAttributesAlias)
	b.PosXY = bjson.PosXYAux.v
	b.Size = bjson.SizeAux.v
	return nil
}
