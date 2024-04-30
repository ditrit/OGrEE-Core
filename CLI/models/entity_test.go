package models_test

import (
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

var entityStrings = map[int][]string{
	models.DOMAIN:    []string{"domain"},
	models.SITE:      []string{"site", "si"},
	models.BLDG:      []string{"building", "bldg", "bd"},
	models.ROOM:      []string{"room", "ro"},
	models.RACK:      []string{"rack", "rk"},
	models.DEVICE:    []string{"device", "dv", "dev"},
	models.AC:        []string{"ac"},
	models.PWRPNL:    []string{"panel", "pn"},
	models.STRAY_DEV: []string{"stray_device"},
	models.ROOMTMPL:  []string{"room_template"},
	models.OBJTMPL:   []string{"obj_template"},
	models.BLDGTMPL:  []string{"bldg_template"},
	models.CABINET:   []string{"cabinet", "cb"},
	models.GROUP:     []string{"group", "gr"},
	models.CORRIDOR:  []string{"corridor", "co"},
	models.TAG:       []string{"tag"},
	models.LAYER:     []string{"layer"},
	models.GENERIC:   []string{"generic", "ge"},
}

func TestEntityToString(t *testing.T) {
	invalidValue := 100
	for key, values := range entityStrings {
		assert.Equal(t, values[0], models.EntityToString(key))
	}
	assert.Equal(t, "INVALID", models.EntityToString(invalidValue))
}

func TestEntityStrToInt(t *testing.T) {
	invalidValue := "invalid_value"
	for key, values := range entityStrings {
		for _, value := range values {
			assert.Equal(t, key, models.EntityStrToInt(value))
		}
	}
	assert.Equal(t, -1, models.EntityStrToInt(invalidValue))
}

func TestGetParentOfEntity(t *testing.T) {
	negativeOneCases := []int{models.ROOMTMPL, models.BLDGTMPL, models.OBJTMPL, models.GROUP}
	for _, value := range negativeOneCases {
		assert.Equal(t, -1, models.GetParentOfEntity(value))
	}
	hierarchyCases := []int{models.SITE, models.BLDG, models.ROOM, models.DEVICE}
	for _, value := range hierarchyCases {
		assert.Equal(t, value-1, models.GetParentOfEntity(value))
	}
	roomCases := []int{models.RACK, models.AC, models.PWRPNL, models.CABINET, models.CORRIDOR, models.GENERIC}
	for _, value := range roomCases {
		assert.Equal(t, models.ROOM, models.GetParentOfEntity(value))
	}
	invalidValue := 100
	assert.Equal(t, -3, models.GetParentOfEntity(invalidValue))
}
