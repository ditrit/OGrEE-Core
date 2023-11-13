package views_test

import (
	"cli/views"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectsShowsLayersAtTheEnd(t *testing.T) {
	assert.Equal(t,
		`A01
#racks
`,
		views.Objects([]map[string]any{
			{
				"name": "#racks",
			},
			{
				"name": "A01",
			},
		}, false, ""),
	)
}

func TestObjectsWithRelativePathOrdersByThePath(t *testing.T) {
	assert.Equal(t,
		`A/R1/A01
A/R1/B01
A/R2/A01
B/R1/A01
`,
		views.Objects([]map[string]any{
			{
				"id": "BASIC.A.R1.A01",
			},
			{
				"id": "BASIC.A.R2.A01",
			},
			{
				"id": "BASIC.A.R1.B01",
			},
			{
				"id": "BASIC.B.R1.A01",
			},
		}, true, "/Physical/BASIC/#racks"),
	)
}

func TestSortedObjectsOrdersByAttributeAndRemoveLayers(t *testing.T) {
	printed, err := views.SortedObjects([]map[string]any{
		{
			"name": "#racks",
		},
		{
			"name":      "A01",
			"attribute": "2",
		},
		{
			"name":      "B01",
			"attribute": "1",
		},
	}, "attribute", []string{}, false, "")
	assert.Nil(t, err)
	assert.Equal(t,
		`B01    attribute: 1
A01    attribute: 2
`,
		printed,
	)
}

func TestSortedObjectsWithRelativePathOrdersByAttribute(t *testing.T) {
	printed, err := views.SortedObjects([]map[string]any{
		{
			"id":        "BASIC.A.R1.A01",
			"attribute": "2",
		},
		{
			"id":        "BASIC.A.R1.B01",
			"attribute": "1",
		},
	}, "attribute", []string{}, true, "/Physical/BASIC/#racks")
	assert.Nil(t, err)
	assert.Equal(t,
		`A/R1/B01    attribute: 1
A/R1/A01    attribute: 2
`,
		printed,
	)
}
