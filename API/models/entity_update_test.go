package models_test

import (
	"p3/models"
	"p3/test/integration"
	u "p3/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

var device map[string]any

func init() {
	siName := "site"
	bdName := siName + ".building"
	roName := bdName + ".room"
	rkName := roName + ".rack"

	integration.RequireCreateSite(siName)
	integration.RequireCreateBuilding(siName, "building")
	integration.RequireCreateRoom(bdName, "room")
	integration.RequireCreateRack(roName, "rack")
	device = integration.RequireCreateDevice(rkName, "device")
}

// region generic

func TestUpdateGenericWorks(t *testing.T) {
	generic := integration.RequireCreateGeneric("", "update-object-1")

	_, err := models.UpdateObject(
		u.EntityToString(u.GENERIC),
		generic["id"].(string),
		map[string]any{
			"attributes": map[string]any{
				"type": "table",
			},
		},
		true,
		integration.ManagerUserRoles,
		false,
	)
	assert.Nil(t, err)
}

// endregion generic

// region device's sizeU & height

func TestUpdateDeviceSizeUAndHeightmm(t *testing.T) {
	_, err := models.UpdateObject(
		u.EntityToString(u.DEVICE),
		device["id"].(string),
		map[string]any{
			"attributes": map[string]any{
				"sizeU":      1,
				"height":     44.45,
				"heightUnit": "mm",
			},
		},
		true,
		integration.ManagerUserRoles,
		false,
	)
	assert.Nil(t, err)
}
func TestUpdateDeviceSizeUAndHeightcm(t *testing.T) {
	_, err := models.UpdateObject(
		u.EntityToString(u.DEVICE),
		device["id"].(string),
		map[string]any{
			"attributes": map[string]any{
				"sizeU":      1,
				"height":     4.445,
				"heightUnit": "cm",
			},
		},
		true,
		integration.ManagerUserRoles,
		false,
	)
	assert.Nil(t, err)
}

func TestUpdateDeviceSizeUAndHeightmmError(t *testing.T) {
	_, err := models.UpdateObject(
		u.EntityToString(u.DEVICE),
		device["id"].(string),
		map[string]any{
			"attributes": map[string]any{
				"sizeU":      12,
				"height":     44.45,
				"heightUnit": "mm",
			},
		},
		true,
		integration.ManagerUserRoles,
		false,
	)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "sizeU and height are not consistent")
}

// endregion device's sizeU & height
