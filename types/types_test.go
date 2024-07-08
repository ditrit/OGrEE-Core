package ogreetypes

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestBuilding(t *testing.T) {
	b := BuildingAttributes{
		Height: 0.1,
		PosXY:  Vector2{X: 0.1, Y: 0.2},
		Size:   Vector2{X: 0.1, Y: 0.2},
	}
	bBytes, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		t.Errorf(err.Error())
	}
	var bback BuildingAttributes
	err = json.Unmarshal(bBytes, &bback)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(b, bback) {
		t.Errorf("unmarshalled building does not match original building")
	}
}

func TestGroup(t *testing.T) {
	g := GroupAttributes{
		Content: "test",
	}
	gBytes, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		t.Errorf(err.Error())
	}
	var gback GroupAttributes
	err = json.Unmarshal(gBytes, &gback)
	if err != nil {
		t.Errorf(err.Error())
	}
	if !reflect.DeepEqual(g, gback) {
		t.Errorf("unmarshalled group does not match original group")
	}
}

func TestEntity(t *testing.T) {
	entity := Entity{
		Category:    "group",
		Description: []string{"test"},
		CreatedDate: nil,
		LastUpdated: nil,
		Name:        "test",
		Id:          "",
		ParentId:    "test",
		Attributes: &GroupAttributes{
			Content: "test",
		},
	}
	entityBytes, err := json.MarshalIndent(&entity, "", "  ")
	if err != nil {
		t.Errorf(err.Error())
	}
	var entityBack Entity
	err = json.Unmarshal(entityBytes, &entityBack)

	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if !reflect.DeepEqual(entity, entityBack) {
		t.Errorf("unmarshalled object does not match original object")
	}
}
