package ogreetypes

import (
	"encoding/json"
)

type BuildingTemplate struct {
	Slug     Slug      `json:"slug"`
	Center   Vector2   `json:"center"`
	SizeWDHm Vector3   `json:"sizeWDHm"`
	Vertices []Vector2 `json:"vertices"`
}

func (b BuildingTemplate) MarshalJSON() ([]byte, error) {
	type Alias BuildingTemplate
	return json.Marshal(struct {
		Category string `json:"category"`
		Alias
	}{
		Category: "building",
		Alias:    Alias(b),
	})
}

func (b *BuildingTemplate) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, b)
}

type Building struct {
	Header     `json:"-"`
	Height     float64            `json:"height"`
	HeightUnit MetricImperialUnit `json:"heightUnit"`
	PosXY      Vector2            `json:"-"`
	PosXYUnit  MetricImperialUnit `json:"posXYUnit"`
	Size       Vector2            `json:"-"`
	SizeUnit   MetricImperialUnit `json:"sizeUnit"`
	Rotation   float64            `json:"rotation"`
	Template   string             `json:"template,omitempty"`
}

type BuildingAlias Building

type BuildingJsonAttributes struct {
	BuildingAlias
	PosXYAux Vector2Wrapper `json:"posXY"`
	SizeAux  Vector2Wrapper `json:"size"`
}

type BuildingJson struct {
	Category string `json:"category"`
	Header
	Attributes BuildingJsonAttributes `json:"attributes"`
}

func (b Building) MarshalJSON() ([]byte, error) {
	return json.Marshal(BuildingJson{
		Category: "building",
		Header:   b.Header,
		Attributes: BuildingJsonAttributes{
			BuildingAlias: BuildingAlias(b),
			PosXYAux:      Vector2Wrapper{b.PosXY},
			SizeAux:       Vector2Wrapper{b.Size},
		},
	})
}

func (b *Building) UnmarshalJSON(data []byte) error {
	var bjson BuildingJson
	if err := json.Unmarshal(data, &bjson); err != nil {
		return err
	}
	*b = Building(bjson.Attributes.BuildingAlias)
	b.Header = bjson.Header
	b.PosXY = bjson.Attributes.PosXYAux.v
	b.Size = bjson.Attributes.SizeAux.v
	return nil
}
