package controllers

import (
	"cli/models"
	"fmt"
)

var API APIPort = &apiPortImpl{}

type APIPort interface {
	Request(method string, endpoint string, body map[string]any, expectedStatus int) (*Response, error)
}

type apiPortImpl struct{}

func (api *apiPortImpl) Request(method string, endpoint string, body map[string]any, expectedStatus int) (*Response, error) {
	URL := State.APIURL + endpoint
	httpResponse, err := models.Send(method, URL, GetKey(), body)
	if err != nil {
		return nil, err
	}
	response, err := ParseResponseClean(httpResponse)
	if err != nil {
		return nil, fmt.Errorf("on %s %s : %s", method, endpoint, err.Error())
	}
	if response.status != expectedStatus {
		msg := ""
		if State.DebugLvl >= DEBUG {
			msg += fmt.Sprintf("%s %s\n", method, URL)
		}
		msg += fmt.Sprintf("[Response From API] %s", response.message)
		errorsAny, ok := response.Body["errors"]
		if ok {
			errorsList := errorsAny.([]any)
			for _, err := range errorsList {
				msg += "\n    " + err.(string)
			}
		}
		return response, fmt.Errorf(msg)
	}
	return response, nil
}
