package ogreetypes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type Entity struct {
	Category      string           `json:"category"`
	Description   []string         `json:"description"`
	Domain        string           `json:"domain"`
	CreatedDate   *time.Time       `json:"createdDate,omitempty"`
	LastUpdated   *time.Time       `json:"lastUpdated,omitempty"`
	Name          string           `json:"name"`
	HierarchyName string           `json:"hierarchyName"`
	Id            string           `json:"id,omitempty"`
	ParentId      string           `json:"parentId"`
	Attributes    EntityAttributes `json:"attributes"`
}

type EntityAttributes interface{}

func (entity *Entity) UnmarshalJSON(data []byte) error {
	type EntityAlias Entity
	var tempEntity EntityAlias
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&tempEntity)
	if err != nil {
		return fmt.Errorf("error during header decoding : %s", err.Error())
	}
	switch tempEntity.Category {
	case "site":
		tempEntity.Attributes = &SiteAttributes{}
	case "building":
		tempEntity.Attributes = &BuildingAttributes{}
	case "room":
		tempEntity.Attributes = &RoomAttributes{}
	case "rack":
		tempEntity.Attributes = &RackAttributes{}
	case "device":
		tempEntity.Attributes = &DeviceAttributes{}
	case "domain":
		tempEntity.Attributes = &DomainAttributes{}
	case "stray_device":
		tempEntity.Attributes = &DeviceAttributes{}
	case "room_template":
		tempEntity.Attributes = &RoomTemplateAttributes{}
	case "obj_template":
		tempEntity.Attributes = &DeviceTemplateAttributes{}
	case "bldg_template":
		tempEntity.Attributes = &BuildingTemplateAttributes{}
	case "group":
		tempEntity.Attributes = &GroupAttributes{}
	case "corridor":
		tempEntity.Attributes = &CorridorAttributes{}
	default:
		return fmt.Errorf("unknown object category")
	}
	err = json.NewDecoder(bytes.NewReader(data)).Decode(&tempEntity)
	if err != nil {
		return fmt.Errorf("error during object decoding : %s", err.Error())
	}
	*entity = Entity(tempEntity)
	return nil
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
