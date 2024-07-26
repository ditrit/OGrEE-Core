package models_test

import (
	"cli/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// region sizeU

func TestComputeFromSizeUmm(t *testing.T) {
	device := map[string]any{
		"attributes": map[string]any{
			"heightUnit": "mm",
		},
	}
	input := map[string]any{
		"attributes": map[string]any{
			"sizeU": 1,
		},
	}
	err := models.ComputeSizeUAndHeight(device, input)
	assert.Nil(t, err)

}

func TestComputeFromSizeUcm(t *testing.T) {
	device := map[string]any{
		"attributes": map[string]any{
			"heightUnit": "cm",
		},
	}
	input := map[string]any{
		"attributes": map[string]any{
			"sizeU": 1,
		},
	}
	err := models.ComputeSizeUAndHeight(device, input)
	assert.Nil(t, err)

}

func TestComputeFromSizeUFail(t *testing.T) {
	device := map[string]any{
		"attributes": map[string]any{
			"heightUnit": "banana",
		},
	}
	input := map[string]any{
		"attributes": map[string]any{
			"sizeU": 1,
		},
	}
	err := models.ComputeSizeUAndHeight(device, input)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "unknown heightUnit value")
}

// endregion

// region height

func TestComputeFromHeightmm(t *testing.T) {
	device := map[string]any{
		"attributes": map[string]any{
			"heightUnit": "mm",
		},
	}
	input := map[string]any{
		"attributes": map[string]any{
			"height": 44.45,
		},
	}
	err := models.ComputeSizeUAndHeight(device, input)
	assert.Nil(t, err)

}

func TestComputeFromHeightcm(t *testing.T) {
	device := map[string]any{
		"attributes": map[string]any{
			"heightUnit": "cm",
		},
	}
	input := map[string]any{
		"attributes": map[string]any{
			"height": 4.445,
		},
	}
	err := models.ComputeSizeUAndHeight(device, input)
	assert.Nil(t, err)

}

func TestComputeFromHeightFail(t *testing.T) {
	device := map[string]any{
		"attributes": map[string]any{
			"heightUnit": "banana",
		},
	}
	input := map[string]any{
		"attributes": map[string]any{
			"height": 1,
		},
	}
	err := models.ComputeSizeUAndHeight(device, input)
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "unknown heightUnit value")
}

// endregion

// region template

func TestSetDeviceSizeUFromTemplateWorks(t *testing.T) {
	deviceAttrs := map[string]any{}
	input := map[string]any{
		"attributes": map[string]any{
			"type": "chassis",
		},
	}
	err := models.SetDeviceSizeUFromTemplate(deviceAttrs, input, any(10000))
	assert.Nil(t, err)
	assert.Equal(t, 225, deviceAttrs["sizeU"])
}

// endregion
