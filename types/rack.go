package ogreetypes

import (
	"encoding/json"
	"time"
)

type RackOrientation string
type RackUnit string
type LabelPosition string

type RackTemplateAttributes struct {
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

type RackAttributes struct {
	Height      float64            `json:"height"`
	HeightUnit  RackUnit           `json:"heightUnit"`
	Orientation RackOrientation    `json:"orientation"`
	PosXYZ      Vector3            `json:"-"`
	PosXYUnit   FloorMetric        `json:"posXYUnit"`
	Size        Vector2            `json:"-"`
	SizeUnit    MetricImperialUnit `json:"sizeUnit"`
	Template    string             `json:"template"`
}

type RackAttributesAlias RackAttributes

type RackAttributesJson struct {
	RackAttributesAlias
	PosXYZAux Vector3Wrapper `json:"posXYZ"`
	SizeAux   Vector2Wrapper `json:"size"`
}

func (r RackAttributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(RackAttributesJson{
		RackAttributesAlias: RackAttributesAlias(r),
		PosXYZAux:           Vector3Wrapper{r.PosXYZ},
		SizeAux:             Vector2Wrapper{r.Size},
	})
}

func (r *RackAttributes) UnmarshalJSON(data []byte) error {
	var rjson RackAttributesJson
	if err := json.Unmarshal(data, &rjson); err != nil {
		return err
	}
	*r = RackAttributes(rjson.RackAttributesAlias)
	r.PosXYZ = rjson.PosXYZAux.v
	r.Size = rjson.SizeAux.v
	return nil
}
