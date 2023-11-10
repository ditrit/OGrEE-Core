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

func TestWildcardStarStarReturnsAllObjects(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-**&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 10)
	unit.ContainsObject(t, objects, "wildcard-site-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1.wildcard-2")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1.wildcard-device-1")
	unit.ContainsObject(t, objects, "wildcard-site-2")
	unit.ContainsObject(t, objects, "wildcard-site-2.wildcard-building-3")
}

func TestWildcardPointStarStarReturnsAllChildrenRecursive(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 7)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1.wildcard-2")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1.wildcard-device-1")
}

func TestWildcardPointStarStarPointReturnsAllChildrenRecursive(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.**.wildcard-1&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 3)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
}

func TestWildcardStarStarAndStar(t *testing.T) {
	response, objects := e2e.GetObjects("id=wildcard-site-1.*.wildcard-1.**&namespace=physical.hierarchy")
	assert.Equal(t, http.StatusOK, response.Code)

	assert.Len(t, objects, 3)
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-1.wildcard-1.wildcard-1.wildcard-device-1")
	unit.ContainsObject(t, objects, "wildcard-site-1.wildcard-building-2.wildcard-1.wildcard-2")
}
