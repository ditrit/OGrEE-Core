package utils

import "fmt"

const usersEndpoint = "/api/users"

var endpoints = map[string]string{
	"login":          "/api/login",
	"users":          usersEndpoint,
	"usersInstance":  usersEndpoint + "/%s",
	"usersBulk":      usersEndpoint + "/bulk",
	"changePassword": usersEndpoint + "/password/change",
	"resetPassword":  usersEndpoint + "/password/reset",
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
