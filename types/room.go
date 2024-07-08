package ogreetypes

import (
	"encoding/json"
	"time"
)

type AxisOrientation string

const (
	XY   AxisOrientation = "+x+y"
	XMY  AxisOrientation = "+x-y"
	MXY  AxisOrientation = "-x+y"
	MXMY AxisOrientation = "-x-y"
)

type SeparatorType string

const (
	WIREFRAME SeparatorType = "wireframe"
	PLAIN     SeparatorType = "plain"
)

type Separator struct {
	StartPosXYm Vector2       `json:"startPosXYm"`
	EndPosXYm   Vector2       `json:"endPosXYm"`
	Type        SeparatorType `json:"type"`
}

type Pillar struct {
	CenterXY Vector2 `json:"centerXY"`
	SizeXY   Vector2 `json:"sizeXY"`
	Rotation float64 `json:"rotation"`
}

type Tile struct {
	Color    string `json:"color"`
	Label    string `json:"label"`
	Location string `json:"location"`
	Name     string `json:"name"`
	Texture  string `json:"texture"`
}

type RoomTemplateAttributes struct {
	Center          *Vector2         `json:"center,omitempty"`
	Colors          map[string]Color `json:"colors,omitempty"`
	CreatedDate     *time.Time       `json:"createdDate,omitempty"`
	LastUpdated     *time.Time       `json:"lastUpdated,omitempty"`
	FloorUnit       FloorMetric      `json:"floorUnit"`
	Id              string           `json:"id,omitempty"`
	AxisOrientation AxisOrientation  `json:"axisOrientation"`
	ReservedArea    []int            `json:"reservedArea,omitempty"`
	TechnicalArea   []int            `json:"technicalArea,omitempty"`
	Separators      []Separator      `json:"separators,omitempty"`
	Pillars         []Pillar         `json:"pillars,omitempty"`
	SizeWDHm        Vector3          `json:"sizeWDHm"`
	Slug            Slug             `json:"slug"`
	TileAngle       *float64         `json:"tileAngle,omitempty"`
	Tiles           []Tile           `json:"tiles,omitempty"`
	Vertices        []Vector2        `json:"vertices,omitempty"`
}

type RoomAttributes struct {
	FloorUnit   FloorMetric        `json:"floorUnit"`
	Height      float64            `json:"height"`
	HeightUnit  MetricImperialUnit `json:"heightUnit"`
	Orientation AxisOrientation    `json:"axisOrientation"`
	Rotation    float64            `json:"rotation"`
	PosXY       Vector2            `json:"posXY"`
	PosXYUnit   MetricImperialUnit `json:"posXYUnit"`
	Size        Vector2            `json:"size"`
	SizeUnit    MetricImperialUnit `json:"sizeUnit"`
	Template    string             `json:"template,omitempty"`
}

type RoomAttributesAlias RoomAttributes

type RoomAttributesJson struct {
	RoomAttributesAlias
	PosXYAux Vector2Wrapper `json:"posXY"`
	SizeAux  Vector2Wrapper `json:"size"`
}

func (r RoomAttributes) MarshalJSON() ([]byte, error) {
	return json.Marshal(RoomAttributesJson{
		RoomAttributesAlias: RoomAttributesAlias(r),
		PosXYAux:            Vector2Wrapper{r.PosXY},
		SizeAux:             Vector2Wrapper{r.Size},
	})
}

func (r *RoomAttributes) UnmarshalJSON(data []byte) error {
	var rjson RoomAttributesJson
	if err := json.Unmarshal(data, &rjson); err != nil {
		return err
	}
	*r = RoomAttributes(rjson.RoomAttributesAlias)
	r.PosXY = rjson.PosXYAux.v
	r.Size = rjson.SizeAux.v
	return nil
}
