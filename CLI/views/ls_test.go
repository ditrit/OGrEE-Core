package views_test

import (
	"cli/views"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLsShowsLayersAtTheEnd(t *testing.T) {
	printed, err := views.Ls([]map[string]any{
		{
			"name": "#racks",
		},
		{
			"id":   "BASIC.A.R1.A01",
			"name": "A01",
		},
	}, "", nil)
	assert.Nil(t, err)
	assert.Equal(t,
		`A01
#racks
`, printed,
	)
}

func TestLsWithRelativePathOrdersByThePath(t *testing.T) {
	printed, err := views.Ls([]map[string]any{
		{
			"id":   "BASIC.A.R1.A01",
			"name": "A01",
		},
		{
			"id":   "BASIC.A.R2.A01",
			"name": "A01",
		},
		{
			"id":   "BASIC.A.R1.B01",
			"name": "B01",
		},
		{
			"id":   "BASIC.B.R1.A01",
			"name": "A01",
		},
	}, "", &views.RelativePathArgs{FromPath: "/Physical/BASIC/#racks"})
	assert.Nil(t, err)
	assert.Equal(t,
		`A/R1/A01
A/R1/B01
A/R2/A01
B/R1/A01
`,
		printed,
	)
}

func TestLsWithRelativePathOnObjectsWithSlugsOrdersBySlug(t *testing.T) {
	printed, err := views.Ls([]map[string]any{
		{
			"slug": "tag2",
		},
		{
			"slug": "tag1",
		},
	}, "", &views.RelativePathArgs{FromPath: "/Physical/BASIC/#racks"})
	assert.Nil(t, err)
	assert.Equal(t,
		`tag1
tag2
`,
		printed,
	)
}

func TestLsWithFormatRemoveLayers(t *testing.T) {
	printed, err := views.LsWithFormat([]map[string]any{
		{
			"name": "#racks",
		},
		{
			"id":        "BASIC.A.R1.A01",
			"name":      "A01",
			"attribute": "2",
		},
		{
			"id":        "BASIC.A.R1.B01",
			"name":      "B01",
			"attribute": "1",
		},
	}, "", nil, []string{"attribute"})
	assert.Nil(t, err)
	assert.Equal(t,
		`A01    attribute: 2
B01    attribute: 1
`,
		printed,
	)
}

func TestLsWithFormatOrdersByAttributeAndRemoveLayers(t *testing.T) {
	printed, err := views.LsWithFormat([]map[string]any{
		{
			"name": "#racks",
		},
		{
			"id":        "BASIC.A.R1.A01",
			"name":      "A01",
			"attribute": "2",
		},
		{
			"id":        "BASIC.A.R1.B01",
			"name":      "B01",
			"attribute": "1",
		},
	}, "attribute", nil, []string{})
	assert.Nil(t, err)
	assert.Equal(t,
		`B01    attribute: 1
A01    attribute: 2
`,
		printed,
	)
}

func TestLsWithFormatWithRelativePathOrdersByAttribute(t *testing.T) {
	printed, err := views.LsWithFormat([]map[string]any{
		{
			"id":        "BASIC.A.R1.A01",
			"name":      "A01",
			"attribute": "2",
		},
		{
			"id":        "BASIC.A.R1.B01",
			"name":      "B01",
			"attribute": "1",
		},
	}, "attribute", &views.RelativePathArgs{FromPath: "/Physical/BASIC/#racks"}, []string{})
	assert.Nil(t, err)
	assert.Equal(t,
		`A/R1/B01    attribute: 1
A/R1/A01    attribute: 2
`,
		printed,
	)
}

func TestLsWithFormatWithoutSortAttrOrdersByName(t *testing.T) {
	printed, err := views.LsWithFormat([]map[string]any{
		{
			"id":        "BASIC.A.R1.A01",
			"name":      "A01",
			"attribute": "2",
		},
		{
			"id":        "BASIC.A.R1.B01",
			"name":      "B01",
			"attribute": "1",
		},
	}, "", nil, []string{"attribute"})
	assert.Nil(t, err)
	assert.Equal(t,
		`A01    attribute: 2
B01    attribute: 1
`,
		printed,
	)
}

func TestLsWithFormatWithRelativePathWithoutSortAttrOrdersById(t *testing.T) {
	printed, err := views.LsWithFormat([]map[string]any{
		{
			"id":        "BASIC.A.R1.A01",
			"name":      "A01",
			"attribute": "2",
		},
		{
			"id":        "BASIC.A.R1.B01",
			"name":      "B01",
			"attribute": "1",
		},
	}, "", &views.RelativePathArgs{FromPath: "/Physical/BASIC/#racks"}, []string{"attribute"})
	assert.Nil(t, err)
	assert.Equal(t,
		`A/R1/A01    attribute: 2
A/R1/B01    attribute: 1
`,
		printed,
	)
}
