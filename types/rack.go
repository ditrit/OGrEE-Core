package ogreetypes

import (
	"encoding/json"
	"time"
)

type RackOrientation string

const (
	RackFront RackOrientation = "front"
	RackRear  RackOrientation = "rear"
	RackLeft  RackOrientation = "left"
	RackRight RackOrientation = "right"
)

type RackUnit string

const (
	RackMM RackUnit = "mm"
	RackCM RackUnit = "cm"
	RackM  RackUnit = "m"
	RackU  RackUnit = "u"
	RackOU RackUnit = "ou"
	RackF  RackUnit = "f"
)

type LabelPosition string

const (
	LabelFront     LabelPosition = "front"
	LabelRear      LabelPosition = "rear"
	LabelFrontrear LabelPosition = "frontrear"
	LabelTop       LabelPosition = "top"
	LabelRight     LabelPosition = "right"
	LabelLeft      LabelPosition = "left"
)

type ComponentOrientation string

const (
	ComponentHorizontal ComponentOrientation = "horizontal"
	ComponentVertical   ComponentOrientation = "vertical"
)

type Component struct {
	Factor     string                `json:"factor,omitempty"`
	Color      *Color                `json:"color,omitempty"`
	ElemOrient *ComponentOrientation `json:"elemOrient,omitempty"`
	ElemPos    Vector3               `json:"elemPos"`
	ElemSize   Vector3               `json:"elemSize"`
	LabelPos   LabelPosition         `json:"labelPos"`
	Location   string                `json:"location"`
	Type       string                `json:"type"`
}

type RackTemplate struct {
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

func (r RackTemplate) MarshalJSON() ([]byte, error) {
	type Alias RackTemplate
	return json.Marshal(struct {
		category string
		Alias
	}{
		category: "rack",
		Alias:    Alias(r),
	})
}

type Rack struct {
	Header
	Height      float64            `json:"height"`
	HeightUnit  RackUnit           `json:"heightUnit"`
	Orientation RackOrientation    `json:"orientation"`
	PosXYZ      Vector3            `json:"-"`
	PosXYUnit   FloorMetric        `json:"posXYUnit"`
	Size        Vector2            `json:"-"`
	SizeUnit    MetricImperialUnit `json:"sizeUnit"`
	Template    string             `json:"template"`
}

type RackAlias Rack

type RackJsonAttributes struct {
	RackAlias
	PosXYZAux Vector3Wrapper `json:"posXYZ"`
	SizeAux   Vector2Wrapper `json:"size"`
}

type RackJson struct {
	Category string `json:"category"`
	Header
	Attributes RackJsonAttributes `json:"attributes"`
}

func (r Rack) MarshalJSON() ([]byte, error) {
	return json.Marshal(RackJson{
		Category: "rack",
		Header:   r.Header,
		Attributes: RackJsonAttributes{
			RackAlias: RackAlias(r),
			PosXYZAux: Vector3Wrapper{r.PosXYZ},
			SizeAux:   Vector2Wrapper{r.Size},
		},
	})
}

func (r *Rack) UnmarshalJSON(data []byte) error {
	var rjson RackJson
	if err := json.Unmarshal(data, &rjson); err != nil {
		return err
	}
	*r = Rack(rjson.Attributes.RackAlias)
	r.Header = rjson.Header
	r.PosXYZ = rjson.Attributes.PosXYZAux.v
	r.Size = rjson.Attributes.SizeAux.v
	return nil
}
