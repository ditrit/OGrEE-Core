package utils

//Builds json messages and
//returns json response

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
)

var BuildHash string
var BuildTree string
var BuildTime string
var GitCommitDate string

const (
	DOMAIN = iota
	// hierarchical root entities
	STRAYOBJ
	SITE
	// hierarchical entities with mandatory parent
	BLDG
	ROOM
	RACK
	DEVICE
	AC
	CABINET
	CORRIDOR
	PWRPNL
	GROUP
	// logical non hierarchical entities
	ROOMTMPL
	OBJTMPL
	BLDGTMPL
	TAG
)

type Namespace string

const (
	Any            Namespace = ""
	Physical       Namespace = "physical"
	Organisational Namespace = "organisational"
	Logical        Namespace = "logical"
)

const HN_DELIMETER = "."           // hierarchyName path delimiter
const NAME_REGEX = "\\w(\\w|\\-)*" // accepted regex for names that compose ids
const RESET_TAG = "RESET"          // used as email to identify a reset token

type RequestFilters struct {
	FieldsToShow []string `schema:"fieldOnly"`
	StartDate    string   `schema:"startDate"`
	EndDate      string   `schema:"endDate"`
	Limit        string   `schema:"limit"`
}

type HierarchyFilters struct {
	Namespace      Namespace `schema:"namespace"`
	StartDate      string    `schema:"startDate"`
	EndDate        string    `schema:"endDate"`
	Limit          string    `schema:"limit"`
	WithCategories bool      `schema:"withcategories"`
}

// Error definitions
type ErrType int

const (
	ErrUnauthorized ErrType = iota
	ErrForbidden
	ErrDuplicate
	ErrBadFormat
	ErrInvalidValue
	ErrDBError
	ErrInternal
	ErrNotFound
	WarnShouldChangePass
)

type Error struct {
	Type    ErrType
	Message string
	Details []string
}

func (err Error) Error() string {
	return err.Message
}

func GetBuildDate() string {
	return BuildTime
}

func GetCommitDate() string {
	return GitCommitDate
}

func GetBuildHash() string {
	return BuildHash
}

func GetBuildTree() string {
	return BuildTree
}

func Connect() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func Message(message string) map[string]interface{} {
	return map[string]interface{}{"message": message}
}

func RespDataWrapper(message string, data interface{}) map[string]interface{} {
	return map[string]interface{}{"message": message, "data": data}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	json.NewEncoder(w).Encode(data)
	w.Header().Add("Content-Type", "application/json")
}

func RespondWithError(w http.ResponseWriter, err *Error) {
	errMap := map[string]interface{}{"message": err.Message}
	if len(err.Details) > 0 {
		errMap["errors"] = err.Details
	}
	w.WriteHeader(ErrTypeToStatusCode(err.Type))
	json.NewEncoder(w).Encode(errMap)
	w.Header().Add("Content-Type", "application/json")
}

func ErrLog(message, funcname, details string, r *http.Request) {
	f, err := os.OpenFile("resources/debug.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	ip := r.RemoteAddr

	log.SetOutput(f)
	log.Println(message + " FOR FUNCTION: " + funcname)
	log.Println("FROM IP: " + ip)
	log.Println(details)
}

func ParamsParse(link *url.URL, entityType int) map[string]interface{} {
	q, _ := url.ParseQuery(link.RawQuery)
	values := make(map[string]interface{})

	// Building Attribute query varies based on entityType
	for key, value := range q {
		if key != "fieldOnly" && key != "startDate" && key != "endDate" && key != "limit" {
			if IsEntityHierarchical(entityType) {
				switch key {
				case "id", "name", "category", "parentID",
					"description", "domain", "parentid", "parentId",
					"createdDate", "lastUpdated":
					values[key] = value
				default:
					values["attributes."+key] = value
				}
			} else {
				//Not sure how to search FBX TEMPLATES
				//For now it is disabled
				switch key {
				case "description", "slug", "category", "sizeWDHmm", "fbxModel":
					values[key] = value
				default:
					values["attributes."+key] = value
				}
			}
		}
	}
	return values
}

func ErrTypeToStatusCode(errType ErrType) int {
	switch errType {
	case ErrForbidden:
		return http.StatusForbidden
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrDuplicate, ErrBadFormat:
		return http.StatusBadRequest
	case ErrDBError, ErrInternal:
		return http.StatusInternalServerError
	case ErrNotFound:
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

var EntityTypeList = []int{
	DOMAIN,
	STRAYOBJ, SITE,
	BLDG, ROOM, RACK, DEVICE, AC, CABINET, CORRIDOR, PWRPNL, GROUP,
	ROOMTMPL, OBJTMPL, BLDGTMPL, TAG,
}

var EntitiesWithTags = []int{
	STRAYOBJ, SITE, BLDG, ROOM, RACK, DEVICE, AC, CABINET, CORRIDOR, PWRPNL, GROUP,
}

func EntityHasTags(entityType int) bool {
	return pie.Contains(EntitiesWithTags, entityType)
}

func IsEntityHierarchical(entity int) bool {
	return !IsEntityNonHierarchical(entity)
}

func IsEntityNonHierarchical(entity int) bool {
	return entity == BLDGTMPL || entity == ROOMTMPL || entity == OBJTMPL || entity == TAG
}

func EntityToString(entity int) string {
	switch entity {
	case SITE:
		return "site"
	case BLDG:
		return "building"
	case ROOM:
		return "room"
	case RACK:
		return "rack"
	case DEVICE:
		return "device"
	case AC:
		return "ac"
	case PWRPNL:
		return "panel"
	case DOMAIN:
		return "domain"
	case STRAYOBJ:
		return "stray_object"
	case ROOMTMPL:
		return "room_template"
	case OBJTMPL:
		return "obj_template"
	case BLDGTMPL:
		return "bldg_template"
	case CABINET:
		return "cabinet"
	case GROUP:
		return "group"
	case CORRIDOR:
		return "corridor"
	case TAG:
		return "tag"
	default:
		return "INVALID"
	}
}

func EntityStrToInt(entity string) int {
	switch entity {
	case "site":
		return SITE
	case "building", "bldg":
		return BLDG
	case "room":
		return ROOM
	case "rack":
		return RACK
	case "device":
		return DEVICE
	case "ac":
		return AC
	case "panel":
		return PWRPNL
	case "domain":
		return DOMAIN
	case "stray_object":
		return STRAYOBJ
	case "room_template":
		return ROOMTMPL
	case "obj_template":
		return OBJTMPL
	case "bldg_template":
		return BLDGTMPL
	case "cabinet":
		return CABINET
	case "group":
		return GROUP
	case "corridor":
		return CORRIDOR
	case "tag":
		return TAG
	default:
		return -1
	}
}

func HierarchyNameToEntity(name string) []int {
	resp := []int{STRAYOBJ} // it can always be a stray
	switch strings.Count(name, HN_DELIMETER) {
	case 0:
		resp = append(resp, SITE)
	case 1:
		resp = append(resp, BLDG)
	case 2:
		resp = append(resp, ROOM)
	case 3:
		resp = append(resp, RACK, GROUP, AC, CORRIDOR, PWRPNL, CABINET)
	case 4:
		resp = append(resp, DEVICE, GROUP)
	default:
		resp = append(resp, DEVICE)
	}

	return resp
}

func NamespaceToString(namespace Namespace) string {
	// switch namespace {
	// case Physical:
	// 	return "Physical"
	// case Organisational:
	// 	return "Organisational"
	// case Logical:
	// 	return "Logical"
	// }
	// return ""
	ref := reflect.ValueOf(namespace)
	return ref.String()
}

func GetEntitiesByNamespace(namespace Namespace) []int {
	var entities []int
	switch namespace {
	case Physical:
		for entity := STRAYOBJ; entity <= GROUP; entity++ {
			entities = append(entities, entity)
		}
	case Organisational:
		entities = append(entities, DOMAIN)
	case Logical:
		for entity := GROUP; entity <= TAG; entity++ {
			entities = append(entities, entity)
		}
	default:
		// All collections
		return EntityTypeList
	}

	return entities
}

func GetParentOfEntityByInt(entity int) int {
	switch entity {
	case DOMAIN:
		return DOMAIN
	case AC, PWRPNL, CABINET, CORRIDOR:
		return ROOM
	case ROOMTMPL, OBJTMPL, BLDGTMPL, TAG, GROUP, STRAYOBJ:
		return -1
	default:
		return entity - 1
	}
}

// Helper functions
func StrSliceContains(slice []string, elem string) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}
	return false
}
