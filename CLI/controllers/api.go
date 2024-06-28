package controllers

import (
	"bytes"
	"cli/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	baseURL   = "/api/"
	LayersURL = baseURL + "layers"
)

var API APIPort = &apiPortImpl{}

type APIPort interface {
	Request(method string, endpoint string, body map[string]any, expectedStatus int) (*Response, error)
}

type apiPortImpl struct{}

// Request
func (api *apiPortImpl) Request(method string, endpoint string, body map[string]any, expectedStatus int) (*Response, error) {
	URL := State.APIURL + endpoint
	httpResponse, err := Send(method, URL, GetKey(), body)
	if err != nil {
		return nil, err
	}
	response, err := ParseResponse(httpResponse)
	if err != nil {
		return nil, fmt.Errorf("on %s %s : %s", method, endpoint, err.Error())
	}
	if response.Status != expectedStatus {
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

func Send(method, URL, key string, data map[string]any) (*http.Response, error) {
	client := &http.Client{}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	return client.Do(req)
}

// Response handling
type Response struct {
	Status  int
	message string
	Body    map[string]any
}

func ParseResponse(response *http.Response) (*Response, error) {
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody := map[string]interface{}{}
	message := ""
	if len(bodyBytes) > 0 {
		err = json.Unmarshal(bodyBytes, &responseBody)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal json : \n%s", string(bodyBytes))
		}
		message, _ = responseBody["message"].(string)
	}
	return &Response{response.StatusCode, message, responseBody}, nil
}

// URL handling
func (controller Controller) ObjectUrl(pathStr string, depth int) (string, error) {
	path, err := controller.SplitPath(pathStr)
	if err != nil {
		return "", err
	}
	useGeneric := false

	var baseUrl string
	switch path.Prefix {
	case models.StrayPath:
		baseUrl = "/api/stray_objects"
	case models.PhysicalPath:
		baseUrl = "/api/hierarchy_objects"
	case models.ObjectTemplatesPath:
		baseUrl = "/api/obj_templates"
	case models.RoomTemplatesPath:
		baseUrl = "/api/room_templates"
	case models.BuildingTemplatesPath:
		baseUrl = "/api/bldg_templates"
	case models.GroupsPath:
		baseUrl = "/api/groups"
	case models.TagsPath:
		baseUrl = "/api/tags"
	case models.LayersPath:
		baseUrl = LayersURL
	case models.DomainsPath:
		baseUrl = "/api/domains"
	case models.VirtualObjsPath:
		if strings.Contains(path.ObjectID, ".Physical.") {
			baseUrl = "/api/objects"
			path.ObjectID = strings.Split(path.ObjectID, ".Physical.")[1]
			useGeneric = true
		} else {
			baseUrl = "/api/virtual_objs"
		}
	default:
		return "", fmt.Errorf("invalid object path")
	}

	params := url.Values{}
	if useGeneric {
		params.Add("id", path.ObjectID)
		if depth > 0 {
			params.Add("limit", strconv.Itoa(depth))
		}
	} else {
		baseUrl += "/" + path.ObjectID
		if depth > 0 {
			baseUrl += "/all"
			params.Add("limit", strconv.Itoa(depth))
		}
	}
	parsedUrl, _ := url.Parse(baseUrl)
	parsedUrl.RawQuery = params.Encode()
	return parsedUrl.String(), nil
}

func (controller Controller) ObjectUrlGeneric(pathStr string, depth int, filters map[string]string, recursive *RecursiveParams) (string, error) {
	params := url.Values{}
	path, err := controller.SplitPath(pathStr)
	if err != nil {
		return "", err
	}

	if recursive != nil {
		err = path.MakeRecursive(recursive.MinDepth, recursive.MaxDepth, recursive.PathEntered)
		if err != nil {
			return "", err
		}
	}

	if filters == nil {
		filters = map[string]string{}
	}

	isNodeLayerInVirtualPath := false
	if path.Layer != nil {
		path.Layer.ApplyFilters(filters)
		if path.Prefix == models.VirtualObjsPath && path.Layer.Name() == "#nodes" {
			isNodeLayerInVirtualPath = true
			filters["filter"] = strings.Replace(filters["filter"], "category=virtual_obj",
				"virtual_config.clusterId="+path.ObjectID[:len(path.ObjectID)-2], 1)
		}
	}

	switch path.Prefix {
	case models.StrayPath:
		params.Add("namespace", "physical.stray")
		params.Add("id", path.ObjectID)
	case models.PhysicalPath:
		params.Add("namespace", "physical.hierarchy")
		params.Add("id", path.ObjectID)
	case models.ObjectTemplatesPath:
		params.Add("namespace", "logical.objtemplate")
		params.Add("slug", path.ObjectID)
	case models.RoomTemplatesPath:
		params.Add("namespace", "logical.roomtemplate")
		params.Add("slug", path.ObjectID)
	case models.BuildingTemplatesPath:
		params.Add("namespace", "logical.bldgtemplate")
		params.Add("slug", path.ObjectID)
	case models.TagsPath:
		params.Add("namespace", "logical.tag")
		params.Add("slug", path.ObjectID)
	case models.LayersPath:
		params.Add("namespace", "logical.layer")
		params.Add("slug", path.ObjectID)
	case models.GroupsPath:
		params.Add("namespace", "logical")
		params.Add("category", "group")
		params.Add("id", path.ObjectID)
	case models.DomainsPath:
		params.Add("namespace", "organisational")
		params.Add("id", path.ObjectID)
	case models.VirtualObjsPath:
		if !isNodeLayerInVirtualPath {
			params.Add("category", "virtual_obj")
			if path.ObjectID != "Logical."+models.VirtualObjsNode+".*" {
				params.Add("id", path.ObjectID)
			}
		}
	default:
		return "", fmt.Errorf("invalid object path")
	}
	if depth > 0 {
		params.Add("limit", strconv.Itoa(depth))
	}

	endpoint := "/api/objects"
	for key, value := range filters {
		if key != "filter" {
			params.Set(key, value)
		} else {
			endpoint = "/api/objects/search"
		}
	}

	url, _ := url.Parse(endpoint)
	url.RawQuery = params.Encode()

	return url.String(), nil
}

func GetKey() string {
	return State.APIKEY
}
