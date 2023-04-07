package ogreetypes

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestBuilding(t *testing.T) {
	b := Building{
		Header: Header{
			Description: []string{"test"},
			CreatedDate: nil,
			LastUpdated: nil,
			Name:        "test",
			Id:          "",
			ParentId:    "test",
		},
		Height: 0.1,
		PosXY:  Vector2{X: 0.1, Y: 0.2},
		Size:   Vector2{X: 0.1, Y: 0.2},
	}
	bytes, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		t.Errorf(err.Error())
	}
	var bback Building
	err = json.Unmarshal(bytes, &bback)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(b, bback) {
		t.Errorf("unmarshalled building does not match original building")
	}
}

func TestGroup(t *testing.T) {
	g := Group{
		Header: Header{
			Description: []string{"test"},
			CreatedDate: nil,
			LastUpdated: nil,
			Name:        "test",
			Id:          "",
			ParentId:    "test",
		},
		Content: "test",
	}
	bytes, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		t.Errorf(err.Error())
	}
	var gback Group
	err = json.Unmarshal(bytes, &gback)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(g, gback) {
		t.Errorf("unmarshalled group does not match original group")
	}
}
