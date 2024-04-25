package utils

import "fmt"

const usersEndpoint = "/api/users"
const domainsEndpoint = "/api/domains"

var endpoints = map[string]string{
	"login":               "/api/login",
	"users":               usersEndpoint,
	"usersInstance":       usersEndpoint + "/%s",
	"usersBulk":           usersEndpoint + "/bulk",
	"changePassword":      usersEndpoint + "/password/change",
	"resetPassword":       usersEndpoint + "/password/reset",
	"entity":              "/api/%s",
	"domains":             "/api/domains",
	"domainsBulk":         domainsEndpoint + "/bulk",
	"complexFilterSearch": "/api/objects/search",
	"validateEntity":      "/api/validate/%s",
	"layersObjects":       "/api/layers/%s/objects",
}

func GetEndpoint(endpointName string, pathParams ...string) string {
	endpoint, exists := endpoints[endpointName]
	if !exists {
		return ""
	}
	for _, pathParam := range pathParams {
		endpoint = fmt.Sprintf(endpoint, pathParam)
	}
	return endpoint
}
