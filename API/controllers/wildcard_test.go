package controllers_test

import (
	"net/http"
	"p3/test/e2e"
	"p3/test/integration"
	"p3/test/unit"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	integration.CreateSite("wildcard-site-1")
	integration.CreateBuilding("wildcard-site-1", "wildcard-building-1")
	integration.CreateBuilding("wildcard-site-1", "wildcard-building-2")
	integration.CreateRoom("wildcard-site-1.wildcard-building-1", "wildcard-1")
	integration.CreateRack("wildcard-site-1.wildcard-building-1.wildcard-1", "wildcard-1")
	integration.CreateDevice("wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1", "wildcard-device-1")
	integration.CreateRoom("wildcard-site-1.wildcard-building-2", "wildcard-1")
	integration.CreateRack("wildcard-site-1.wildcard-building-2.wildcard-1", "wildcard-2")
	integration.CreateSite("wildcard-site-2")
	integration.CreateBuilding("wildcard-site-2", "wildcard-building-3")
}

func TestWildcardSomethingStarReturnsSites(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-*&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 2)
	unit.ContainsObject(t, objects, "wildcard-site-1")
	unit.ContainsObject(t, objects, "wildcard-site-2")
}

func TestWildcardPointStarReturnsAllDirectChildren(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.*&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 2)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2")
}

func TestWildcardPointStarSomethingReturnsAllDirectChildren(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.*.wildcard-1&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 2)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
}

func TestWildcardSomethingStarStarIsEquivalentToStar(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-**&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 2)
	unit.ContainsObject(t, objects, "wildcard-site-1")
	unit.ContainsObject(t, objects, "wildcard-site-2")
}

func TestWildcardPointStarStarIsEquivalentToStar(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 2)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2")
}

func TestWildcardPointStarStarPointReturnsAllChildrenRecursive(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**.wildcard-1&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 3)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
}

func TestWildcardStarStarPointStar(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**.*&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 7)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1.wildcard-device-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1.wildcard-2")
}

func TestWildcardStarStarPointStarSomething(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**.*1&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 5)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1.wildcard-device-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
}

func TestWildcardStarStarWithLimitStar(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**{0,2}.*&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 6)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1.wildcard-2")
}

func TestWildcardStarStarWithLimitStarLimits(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**{0,1}.*&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 4)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
}

func TestWildcardStarStarPointWithLimitSomething(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**{0,2}.wildcard-1&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 3)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
}

func TestWildcardStarStarPointWithLimitLimits(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**{0,1}.wildcard-1&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 2)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
}

func TestWildcardStarStarWithLimitSomethingStar(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**{0,2}.wildcard-building*&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 2)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2")
}

func TestWildcardStarStarPointWithInferiorLimit(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**{1,2}.wildcard-*&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 4)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1.wildcard-2")
}

func TestWildcardStarStarPointExactAmount(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**{1,1}.wildcard-*&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 2)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
}
