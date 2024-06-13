package utils

import (
	"cli/controllers"
	mocks "cli/mocks/controllers"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/stretchr/testify/mock"
)

func mockResponse(mockAPI *mocks.APIPort, httpMethod string, url string, body map[string]any, responseStatus int, response interface{}) {
	var requestBody interface{}
	if body != nil {
		requestBody = body
	} else {
		requestBody = mock.Anything
	}
	mockAPI.On(
		"Request", httpMethod,
		url,
		requestBody, responseStatus,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": response,
			},
		}, nil,
	).Once()
}

func mockResponseWithParams(mockAPI *mocks.APIPort, httpMethod string, urlPath string, queryParams string, body map[string]any, responseStatus int, response interface{}) {
	params, err := url.ParseQuery(queryParams)
	if err != nil {
		log.Fatalln(err.Error())
	}

	mockResponse(mockAPI, httpMethod, urlPath+"?"+params.Encode(), body, responseStatus, response)
}

func MockGetObjectHierarchy(mockAPI *mocks.APIPort, object map[string]any) {
	mockResponse(mockAPI, http.MethodGet, "/api/hierarchy_objects/"+object["id"].(string)+"/all?limit=1", nil, http.StatusOK, KeepOnlyDirectChildren(object))
}

func MockGetObject(mockAPI *mocks.APIPort, object map[string]any) {
	mockResponse(mockAPI, http.MethodGet, "/api/hierarchy_objects/"+object["id"].(string), nil, http.StatusOK, RemoveChildren(object))
}

func MockGetObjects(mockAPI *mocks.APIPort, queryParams string, result []any) {
	mockResponseWithParams(mockAPI, http.MethodGet, "/api/objects", queryParams, nil, http.StatusOK, RemoveChildrenFromList(result))
}

func MockGetObjectsWithComplexFilters(mockAPI *mocks.APIPort, queryParams string, body map[string]any, result []any) {
	mockResponseWithParams(mockAPI, http.MethodPost, "/api/objects/search", queryParams, body, http.StatusOK, RemoveChildrenFromList(result))
}

func MockGetObjectByEntity(mockAPI *mocks.APIPort, entity string, object map[string]any) {
	idOrSlug, idPresent := object["id"].(string)
	if !idPresent {
		idOrSlug = object["slug"].(string)
	}

	mockResponse(mockAPI, http.MethodGet, "/api/"+entity+"/"+idOrSlug, nil, http.StatusOK, RemoveChildren(object))
}

func MockDeleteObjects(mockAPI *mocks.APIPort, queryParams string, result []any) {
	mockResponseWithParams(mockAPI, http.MethodDelete, "/api/objects", queryParams, nil, http.StatusOK, RemoveChildrenFromList(result))
}

func MockDeleteObjectsWithComplexFilters(mockAPI *mocks.APIPort, queryParams string, body map[string]any, result []any) {
	mockResponseWithParams(mockAPI, http.MethodDelete, "/api/objects/search", queryParams, body, http.StatusOK, RemoveChildrenFromList(result))
}

func MockGetObjectsByEntity(mockAPI *mocks.APIPort, entity string, objects []any) {
	mockResponse(mockAPI, http.MethodGet, "/api/"+entity, nil, http.StatusOK, RemoveChildrenFromList(objects))
}

func MockCreateObject(mockAPI *mocks.APIPort, entity string, data map[string]any) {
	mockResponse(mockAPI, http.MethodPost, "/api/"+entity+"s", data, http.StatusCreated, data)
}

func MockUpdateObject(mockAPI *mocks.APIPort, dataUpdate map[string]any, dataUpdated map[string]any) {
	mockResponse(mockAPI, http.MethodPatch, mock.Anything, dataUpdate, http.StatusOK, dataUpdated)
}

func MockPutObject(mockAPI *mocks.APIPort, dataUpdate map[string]any, dataUpdated map[string]any) {
	mockResponse(mockAPI, http.MethodPut, mock.Anything, dataUpdate, http.StatusOK, dataUpdated)
}

func MockObjectNotFound(mockAPI *mocks.APIPort, path string) {
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

func MockGetObjTemplate(mockAPI *mocks.APIPort, template map[string]any) {
	mockResponse(mockAPI, http.MethodGet, "/api/obj_templates/"+template["slug"].(string), nil, http.StatusOK, template)
}

func MockGetRoomTemplate(mockAPI *mocks.APIPort, template map[string]any) {
	mockResponse(mockAPI, http.MethodGet, "/api/room_templates/"+template["slug"].(string), nil, http.StatusOK, template)
}
