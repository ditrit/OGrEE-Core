package utils

import "fmt"

const usersEndpoint = "/api/users"
const domainsEndpoint = "/api/domains"
const objectsEndpoint = "/api/objects"
const hierarchyEdnpoint = "/api/hierarchy"
const entityEndpoint = "/api/%s"

var endpoints = map[string]string{
	"login":               "/api/login",
	"users":               usersEndpoint,
	"usersInstance":       usersEndpoint + "/%s",
	"usersBulk":           usersEndpoint + "/bulk",
	"changePassword":      usersEndpoint + "/password/change",
	"resetPassword":       usersEndpoint + "/password/reset",
	"entity":              entityEndpoint,
	"entityInstance":      entityEndpoint + "/%s",
	"entityAncestors":     entityEndpoint + "/%s/%s",
	"entityUnlink":        entityEndpoint + "/%s/unlink",
	"entityLink":          entityEndpoint + "/%s/link",
	"domains":             domainsEndpoint,
	"domainsBulk":         domainsEndpoint + "/bulk",
	"getObject":           objectsEndpoint,
	"complexFilterSearch": objectsEndpoint + "/search",
	"validateEntity":      "/api/validate/%s",
	"layersObjects":       "/api/layers/%s/objects",
	"tokenValid":          "/api/token/valid",
	"hierarchy":           hierarchyEdnpoint,
	"hierarchyAttributes": hierarchyEdnpoint + "/attributes",
	"tempunits":           "/api/tempunits/%s",
	"projects":            "/api/projects",
}

func GetEndpoint(endpointName string, pathParams ...any) string {
	endpoint, exists := endpoints[endpointName]
	if !exists {
		return ""
	}
	if len(pathParams) > 0 {
		endpoint = fmt.Sprintf(endpoint, pathParams...)
	}
	return endpoint
}
