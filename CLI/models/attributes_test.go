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
	slots, err := models.ExpandStrVector([]string{"slot1..slot3", "slot4"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid device syntax: .. can only be used in a single element vector")

	slots, err = models.ExpandStrVector([]string{"slot1..slot3..slot7"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid device syntax: incorrect use of .. for slot")

	slots, err = models.ExpandStrVector([]string{"slot1..slots3"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid device syntax: incorrect use of .. for slot")

	slots, err = models.ExpandStrVector([]string{"slot1..slotE"})
	assert.Nil(t, slots)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "Invalid device syntax: incorrect use of .. for slot")

	slots, err = models.ExpandStrVector([]string{"slot1..slot3"})
	assert.Nil(t, err)
	assert.NotNil(t, slots)
	assert.EqualValues(t, []string{"slot1", "slot2", "slot3"}, slots)

	slots, err = models.ExpandStrVector([]string{"slot1", "slot3"})
	assert.Nil(t, err)
	assert.NotNil(t, slots)
	assert.EqualValues(t, []string{"slot1", "slot3"}, slots)
}
