package controllers_test

import (
	"net/http"
	"p3/test/e2e"
	"p3/test/integration"
	"p3/test/unit"
	"testing"

	"github.com/stretchr/testify/assert"
)

const wildcardSite1 = "wildcard-site-1"
const wildcardBuilding1 = wildcardSite1 + ".wildcard-building-1"
const wildcardBuilding2 = wildcardSite1 + ".wildcard-building-2"
const wildcardRoom1 = wildcardBuilding1 + ".wildcard-1"
const wildcardRack1 = wildcardRoom1 + ".wildcard-1"
const wildcardDevice = wildcardRack1 + ".wildcard-device-1"
const wildcardRoom2 = wildcardBuilding2 + ".wildcard-1"
const wildcardRack2 = wildcardRoom2 + ".wildcard-2"
const wildcardSite2 = "wildcard-site-2"

func init() {
	integration.RequireCreateSite(wildcardSite1)
	integration.RequireCreateBuilding(wildcardSite1, "wildcard-building-1")
	integration.RequireCreateBuilding(wildcardSite1, "wildcard-building-2")
	integration.RequireCreateRoom(wildcardBuilding1, "wildcard-1")
	integration.RequireCreateRack(wildcardRoom1, "wildcard-1")
	integration.RequireCreateDevice(wildcardRack1, "wildcard-device-1")
	integration.RequireCreateRoom(wildcardBuilding2, "wildcard-1")
	integration.RequireCreateRack(wildcardRoom2, "wildcard-2")
	integration.RequireCreateSite(wildcardSite2)
	integration.RequireCreateBuilding(wildcardSite2, "wildcard-building-3")
}

func TestWildcard(t *testing.T) {
	tests := []struct {
		name        string
		queryParams string
		objectsId   []string
	}{
		{"SomethingStarReturnsSites", "id=wildcard-*&namespace=physical.hierarchy", []string{wildcardSite1, wildcardSite2}},
		{"PointStarReturnsAllDirectChildren", "id=wildcard-site-1.*&namespace=physical.hierarchy", []string{wildcardBuilding1, wildcardBuilding2}},
		{"PointStarSomethingReturnsAllDirectChildren", "id=wildcard-site-1.*.wildcard-1&namespace=physical.hierarchy", []string{wildcardRoom1, wildcardRoom2}},
		{"SomethingStarStarIsEquivalentToStar", "id=wildcard-**&namespace=physical.hierarchy", []string{wildcardSite1, wildcardSite2}},
		{"PointStarStarIsEquivalentToStar", "id=wildcard-site-1.**&namespace=physical.hierarchy", []string{wildcardBuilding1, wildcardBuilding2}},
		{"PointStarStarPointReturnsAllChildrenRecursive", "id=wildcard-site-1.**.wildcard-1&namespace=physical.hierarchy", []string{wildcardRoom1, wildcardRoom2, wildcardRack1}},
		{"StarStarPointStar", "id=wildcard-site-1.**.*&namespace=physical.hierarchy", []string{wildcardBuilding1, wildcardRoom1, wildcardRack1, wildcardDevice, wildcardBuilding2, wildcardRoom2, wildcardRack2}},
		{"StarStarPointStarSomething", "id=wildcard-site-1.**.*1&namespace=physical.hierarchy", []string{wildcardBuilding1, wildcardRoom1, wildcardRack1, wildcardDevice, wildcardRoom2}},
		{"StarStarWithLimitStar", "id=wildcard-site-1.**{0,2}.*&namespace=physical.hierarchy", []string{wildcardBuilding1, wildcardRoom1, wildcardRack1, wildcardBuilding2, wildcardRoom2, wildcardRack2}},
		{"StarStarWithLimitStarLimits", "id=wildcard-site-1.**{0,1}.*&namespace=physical.hierarchy", []string{wildcardBuilding1, wildcardRoom1, wildcardBuilding2, wildcardRoom2}},
		{"StarStarPointWithLimitSomething", "id=wildcard-site-1.**{0,2}.wildcard-1&namespace=physical.hierarchy", []string{wildcardRoom1, wildcardRack1, wildcardRoom2}},
		{"StarStarPointWithLimitLimits", "id=wildcard-site-1.**{0,1}.wildcard-1&namespace=physical.hierarchy", []string{wildcardRoom1, wildcardRoom2}},
		{"StarStarWithLimitSomethingStar", "id=wildcard-site-1.**{0,2}.wildcard-building*&namespace=physical.hierarchy", []string{wildcardBuilding1, wildcardBuilding2}},
		{"StarStarPointWithInferiorLimit", "id=wildcard-site-1.**{1,2}.wildcard-*&namespace=physical.hierarchy", []string{wildcardRoom1, wildcardRack1, wildcardRoom2, wildcardRack2}},
		{"StarStarPointExactAmount", "id=wildcard-site-1.**{1,1}.wildcard-*&namespace=physical.hierarchy", []string{wildcardRoom1, wildcardRoom2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, objects := e2e.GetObjects(tt.queryParams)
			assert.Equal(t, http.StatusOK, response.Code)

			assert.Len(t, objects, len(tt.objectsId))
			for _, id := range tt.objectsId {
				unit.ContainsObject(t, objects, id)
			}
		})
	}
}
