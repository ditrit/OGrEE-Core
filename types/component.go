package ogreetypes

type ComponentOrientation string

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
