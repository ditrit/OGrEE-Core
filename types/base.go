package ogreetypes

import (
	"encoding/json"
	"fmt"
	"time"
)

type Vector2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Vector2Wrapper struct {
	v Vector2
}

func (v Vector2Wrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("{\"x\":%f,\"y\":%f}", v.v.X, v.v.Y))
}

func (v *Vector2Wrapper) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), &v.v)
}

type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Vector3Wrapper struct {
	v Vector3
}

func (v Vector3Wrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("{\"x\":%f,\"y\":%f,\"z\":%f}", v.v.X, v.v.Y, v.v.Z))
}

func (v *Vector3Wrapper) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(s), &v.v)
}

type Color string

type MetricImperialUnit string

const (
	MetricMM MetricImperialUnit = "mm"
	MetricCM MetricImperialUnit = "cm"
	MetricM  MetricImperialUnit = "m"
	MetricF  MetricImperialUnit = "f"
)

type FloorMetric string

const (
	FloorM FloorMetric = "m"
	FloorT FloorMetric = "t"
	FloorF FloorMetric = "f"
)

type Slug string

type Header struct {
	Description []string   `json:"description"`
	Domain      string     `json:"domain"`
	CreatedDate *time.Time `json:"createdDate,omitempty"`
	LastUpdated *time.Time `json:"lastUpdated,omitempty"`
	Name        string     `json:"name"`
	Id          string     `json:"id,omitempty"`
	ParentId    string     `json:"parentId"`
}
