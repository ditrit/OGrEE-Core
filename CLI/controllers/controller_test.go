package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
)

func newControllerWithMocks(t *testing.T) (controllers.Controller, *mocks.APIPort, *mocks.Ogree3DPort) {
	mockAPI := mocks.NewAPIPort(t)
	mockOgree3D := mocks.NewOgree3DPort(t)
	return controllers.Controller{
		API:     mockAPI,
		Ogree3D: mockOgree3D,
	}, mockAPI, mockOgree3D
}

func removeChildren(object map[string]any) map[string]any {
	objectCopy := copyMap(object)
	delete(objectCopy, "children")

	return objectCopy
}

func mockGetObjects(mockAPI *mocks.APIPort, queryParams string, result []any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/objects?"+queryParams,
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": removeChildrenFromList(result),
			},
		}, nil,
	).Once()
}

func removeChildrenFromList(objects []any) []any {
	result := []any{}
	for _, object := range objects {
		result = append(result, removeChildren(object.(map[string]any)))
	}

	return result
}

func copyMap(toCopy map[string]any) map[string]any {
	jsonMap, _ := json.Marshal(toCopy)

	var newMap map[string]any

	json.Unmarshal(jsonMap, &newMap)

	return newMap
}

func mockGetObjectByEntity(mockAPI *mocks.APIPort, entity string, object map[string]any) {
	idOrSlug, idPresent := object["id"].(string)
	if !idPresent {
		idOrSlug = object["slug"].(string)
	}

	mockAPI.On(
		"Request", http.MethodGet,
		"/api/"+entity+"/"+idOrSlug,
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": removeChildren(object),
			},
		}, nil,
	).Once()
}
