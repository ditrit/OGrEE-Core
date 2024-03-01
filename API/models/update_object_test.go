package models_test

import (
	"p3/models"
	"p3/test/integration"
	u "p3/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
