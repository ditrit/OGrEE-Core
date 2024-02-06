package controllers_test

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/mock"
)

func newControllerWithMocks(t *testing.T) (controllers.Controller, *mocks.APIPort, *mocks.Ogree3DPort, *mocks.ClockPort) {
	mockAPI := mocks.NewAPIPort(t)
	mockOgree3D := mocks.NewOgree3DPort(t)
	mockClock := mocks.NewClockPort(t)
	return controllers.Controller{
		API:     mockAPI,
		Ogree3D: mockOgree3D,
		Clock:   mockClock,
	}, mockAPI, mockOgree3D, mockClock
}

func mockGetObjectHierarchy(mockAPI *mocks.APIPort, object map[string]any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/hierarchy-objects/"+object["id"].(string)+"/all?limit=1",
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": keepOnlyDirectChildren(object),
			},
		}, nil,
	).Once()
}

func keepOnlyDirectChildren(object map[string]any) map[string]any {
	objectCopy := copyMap(object)

	for _, child := range objectCopy["children"].([]any) {
		delete(child.(map[string]any), "children")
	}

	return objectCopy
}

func mockGetObject(mockAPI *mocks.APIPort, object map[string]any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/hierarchy-objects/"+object["id"].(string),
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": removeChildren(object),
			},
		}, nil,
	).Once()
}

func emptyChildren(object map[string]any) map[string]any {
	objectCopy := copyMap(object)
	objectCopy["children"] = []any{}

	return objectCopy
}

func removeChildren(object map[string]any) map[string]any {
	objectCopy := copyMap(object)
	delete(objectCopy, "children")

	return objectCopy
}

func mockGetObjects(mockAPI *mocks.APIPort, queryParams string, result []any) {
	params, err := url.ParseQuery(queryParams)
	if err != nil {
		log.Fatalln(err.Error())
	}

	mockAPI.On(
		"Request", http.MethodGet,
		"/api/objects?"+params.Encode(),
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": removeChildrenFromList(result),
			},
		}, nil,
	).Once()
}

func mockGetObjectsWithComplexFilters(mockAPI *mocks.APIPort, queryParams string, body map[string]any, result []any) {
	params, err := url.ParseQuery(queryParams)
	if err != nil {
		log.Fatalln(err.Error())
	}

	mockAPI.On(
		"Request", http.MethodPost,
		"/api/objects/search?"+params.Encode(),
		body, http.StatusOK,
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

func mockDeleteObjects(mockAPI *mocks.APIPort, queryParams string, result []any) {
	params, err := url.ParseQuery(queryParams)
	if err != nil {
		log.Fatalln(err.Error())
	}

	mockAPI.On(
		"Request", http.MethodDelete,
		"/api/objects?"+params.Encode(),
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": removeChildrenFromList(result),
			},
		}, nil,
	).Once()
}

func mockGetObjectsByEntity(mockAPI *mocks.APIPort, entity string, objects []any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/"+entity,
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": map[string]any{
					"objects": removeChildrenFromList(objects),
				},
			},
		}, nil,
	).Once()
}

func mockCreateObject(mockAPI *mocks.APIPort, entity string, data map[string]any) {
	mockAPI.On(
		"Request", http.MethodPost,
		"/api/"+entity+"s",
		data, http.StatusCreated,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": data,
			},
		}, nil,
	).Once()
}

func mockUpdateObject(mockAPI *mocks.APIPort, dataUpdate map[string]any, dataUpdated map[string]any) {
	mockAPI.On("Request", http.MethodPatch, mock.Anything, dataUpdate, http.StatusOK).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": dataUpdated,
			},
		}, nil,
	)
}

func mockObjectNotFound(mockAPI *mocks.APIPort, path string) {
	mockAPI.On(
		"Request", http.MethodGet,
		path,
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Status: http.StatusNotFound,
		}, errors.New("not found"),
	).Once()
}

func mockGetObjTemplate(mockAPI *mocks.APIPort, template map[string]any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/obj-templates/"+template["slug"].(string),
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": template,
			},
		}, nil,
	).Once()
}
