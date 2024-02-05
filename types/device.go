package ogreetypes

import (
	"encoding/json"
	"time"
)

type DeviceOrientation string

type DeviceTemplateAttributes struct {
	Attributes  map[string]string `json:"attributes"`
	Colors      map[string]Color  `json:"colors,omitempty"`
	Components  []Component       `json:"components"`
	CreatedDate *time.Time        `json:"createdDate,omitempty"`
	LastUpdated *time.Time        `json:"lastUpdated,omitempty"`
	Description string            `json:"description"`
	FbxModel    string            `json:"fbxModel"`
	Id          string            `json:"id,omitempty"`
	SizeWDHmm   Vector3           `json:"sizeWDHmm"`
	Slug        Slug              `json:"slug"`
	Slots       []Component       `json:"slots"`
}

type DeviceAttributes struct {
	FbxModel    string             `json:"fbxModel,omitempty"`
	Height      float64            `json:"height"`
	HeightUnit  RackUnit           `json:"heightUnit"`
	Orientation DeviceOrientation  `json:"orientation"`
	Size        Vector2            `json:"-"`
	SizeUnit    MetricImperialUnit `json:"sizeUnit"`
	Slot        string             `json:"slot,omitempty"`
	Template    string             `json:"template"`
	Type        string             `json:"type,omitempty"`
	PosU        *int               `json:"posU,omitempty"`
	SizeU       *int               `json:"sizeU,omitempty"`
}

type DeviceAttributesAlias DeviceAttributes

type DeviceAttributesJson struct {
	DeviceAttributesAlias
	SizeAux Vector2Wrapper `json:"size"`
}

func (d DeviceAttributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(DeviceAttributesJson{
		DeviceAttributesAlias: DeviceAttributesAlias(d),
		SizeAux:               Vector2Wrapper{d.Size},
	})
}

func (d *DeviceAttributes) UnmarshalJSON(data []byte) error {
	var djson DeviceAttributesJson
	if err := json.Unmarshal(data, &djson); err != nil {
		return err
	}
	*d = DeviceAttributes(djson.DeviceAttributesAlias)
	d.Size = djson.SizeAux.v
	return nil
}
