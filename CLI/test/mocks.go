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

func MockGetObjectHierarchy(mockAPI *mocks.APIPort, object map[string]any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/hierarchy-objects/"+object["id"].(string)+"/all?limit=1",
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": KeepOnlyDirectChildren(object),
			},
		}, nil,
	).Once()
}

func MockGetObject(mockAPI *mocks.APIPort, object map[string]any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/hierarchy-objects/"+object["id"].(string),
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": RemoveChildren(object),
			},
		}, nil,
	).Once()
}

func MockGetObjects(mockAPI *mocks.APIPort, queryParams string, result []any) {
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
				"data": RemoveChildrenFromList(result),
			},
		}, nil,
	).Once()
}

func MockGetObjectsWithComplexFilters(mockAPI *mocks.APIPort, queryParams string, body map[string]any, result []any) {
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
				"data": RemoveChildrenFromList(result),
			},
		}, nil,
	).Once()
}

func MockGetObjectByEntity(mockAPI *mocks.APIPort, entity string, object map[string]any) {
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
				"data": RemoveChildren(object),
			},
		}, nil,
	).Once()
}

func MockDeleteObjects(mockAPI *mocks.APIPort, queryParams string, result []any) {
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
				"data": RemoveChildrenFromList(result),
			},
		}, nil,
	).Once()
}

func MockDeleteObjectsWithComplexFilters(mockAPI *mocks.APIPort, queryParams string, body map[string]any, result []any) {
	params, err := url.ParseQuery(queryParams)
	if err != nil {
		log.Fatalln(err.Error())
	}

	mockAPI.On(
		"Request", http.MethodDelete,
		"/api/objects/search?"+params.Encode(),
		body, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": RemoveChildrenFromList(result),
			},
		}, nil,
	).Once()
}

func MockGetObjectsByEntity(mockAPI *mocks.APIPort, entity string, objects []any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/"+entity,
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": map[string]any{
					"objects": RemoveChildrenFromList(objects),
				},
			},
		}, nil,
	).Once()
}

func MockCreateObject(mockAPI *mocks.APIPort, entity string, data map[string]any) {
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

func MockUpdateObject(mockAPI *mocks.APIPort, dataUpdate map[string]any, dataUpdated map[string]any) {
	mockAPI.On("Request", http.MethodPatch, mock.Anything, dataUpdate, http.StatusOK).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": dataUpdated,
			},
		}, nil,
	)
}

func MockPutObject(mockAPI *mocks.APIPort, dataUpdate map[string]any, dataUpdated map[string]any) {
	mockAPI.On("Request", http.MethodPut, mock.Anything, dataUpdate, http.StatusOK).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": dataUpdated,
			},
		}, nil,
	)
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

func MockGetRoomTemplate(mockAPI *mocks.APIPort, template map[string]any) {
	mockAPI.On(
		"Request", http.MethodGet,
		"/api/room-templates/"+template["slug"].(string),
		mock.Anything, http.StatusOK,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": template,
			},
		}, nil,
	).Once()
}
