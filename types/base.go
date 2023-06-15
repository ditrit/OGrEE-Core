package ogreetypes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Object interface {
	json.Marshaler
	json.Unmarshaler
}

func ParseObject(input io.Reader) (Object, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(input)
	inputBytes := buf.Bytes()
	var genericObj map[string]any
	err := json.NewDecoder(bytes.NewReader(inputBytes)).Decode(&genericObj)
	if err != nil {
		return nil, fmt.Errorf("error during header decoding : %s", err.Error())
	}
	category, ok := genericObj["category"].(string)
	if !ok {
		return nil, fmt.Errorf("category field is not a string")
	}
	var obj Object
	switch category {
	case "site":
		obj = &Site{}
	case "building":
		obj = &Building{}
	case "room":
		obj = &Room{}
	case "rack":
		obj = &Rack{}
	case "device":
		obj = &Device{}
	case "domain":
		obj = &Domain{}
	case "stray_device":
		obj = &Device{}
	case "room_template":
		obj = &RoomTemplate{}
	case "obj_template":
		obj = &DeviceTemplate{}
	case "bldg_template":
		obj = &BuildingTemplate{}
	case "group":
		obj = &Group{}
	case "corridor":
		obj = &Corridor{}
	default:
		return nil, fmt.Errorf("unknown object category")
	}
	err = json.NewDecoder(bytes.NewReader(inputBytes)).Decode(obj)
	if err != nil {
		return nil, fmt.Errorf("error during object decoding : %s", err.Error())
	}
	return obj, nil
}

type Vector2 struct {
	X float64
	Y float64
}

func (v Vector2) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{v.X, v.Y})
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
	X float64
	Y float64
	Z float64
}

func (v Vector3) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{v.X, v.Y, v.Z})
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
type FloorMetric string
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
