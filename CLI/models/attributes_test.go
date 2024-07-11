package models_test

import (
	l "cli/logger"
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	l.InitLogs()
}

func TestExpandSlotVector(t *testing.T) {
	slots, err := models.CheckExpandStrVector([]string{"slot1..slot3", "slot4"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid device syntax: .. can only be used in a single element vector")

	slots, err = models.CheckExpandStrVector([]string{"slot1..slot3..slot7"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid device syntax: incorrect use of .. for slot")

	slots, err = models.CheckExpandStrVector([]string{"slot1..slots3"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid device syntax: incorrect use of .. for slot")

	slots, err = models.CheckExpandStrVector([]string{"slot1..slotE"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "invalid device syntax: incorrect use of .. for slot")

	slots, err = models.CheckExpandStrVector([]string{"slot1..slot3"})
	assert.Nil(t, err)
	assert.NotNil(t, slots)
	assert.EqualValues(t, []string{"slot1", "slot2", "slot3"}, slots)

	slots, err = models.CheckExpandStrVector([]string{"slot1", "slot3"})
	assert.Nil(t, err)
	assert.NotNil(t, slots)
	assert.EqualValues(t, []string{"slot1", "slot3"}, slots)
}

func TestErrorResponder(t *testing.T) {
	err := models.ErrorResponder("reserved", "4", false)
	assert.ErrorContains(t, err, "Invalid reserved attribute provided. It must be an array/list/vector with 4 elements. Please refer to the wiki or manual reference for more details on how to create objects using this syntax")

	err = models.ErrorResponder("reserved", "4", true)
	assert.ErrorContains(t, err, "Invalid reserved attributes provided. They must be arrays/lists/vectors with 4 elements. Please refer to the wiki or manual reference for more details on how to create objects using this syntax")
}
