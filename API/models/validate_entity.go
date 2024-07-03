package models

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"p3/repository"
	u "p3/utils"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/elliotchance/pie/v2"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:embed schemas/*.json
//go:embed schemas/refs/*.json
var embeddfs embed.FS
var c *jsonschema.Compiler
var schemaTypes map[string]any

func init() {
	// Load JSON schemas
	c = jsonschema.NewCompiler()
	loadJsonSchemas("")
	loadJsonSchemas("refs/")
}

func loadJsonSchemas(schemaPrefix string) {
	var schemaPath = "schemas/"
	dir := strings.Trim(schemaPath+schemaPrefix, "/") // without trailing '/'
	entries, err := embeddfs.ReadDir((dir))
	if err != nil {
		println(err.Error())
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			file, err := embeddfs.Open(schemaPath + schemaPrefix + e.Name())
			if err != nil {
				continue
			}
			loadJsonSchema(schemaPrefix, e.Name(), file)
		}
	}
}

func loadJsonSchema(schemaPrefix, fileName string, file fs.File) {
	fullFileName := schemaPrefix + fileName
	if fileName == "types.json" {
		// Make two copies of the reader stream
		var buf bytes.Buffer
		tee := io.TeeReader(file, &buf)

		c.AddResource(fullFileName, tee)

		// Read and unmarshall types.json file
		typesBytes, _ := io.ReadAll(&buf)
		json.Unmarshal(typesBytes, &schemaTypes)

		// Remove types that do not have a "pattern" attribute
		schemaTypes = schemaTypes["definitions"].(map[string]any)
		for key, definition := range schemaTypes {
			if _, ok := definition.(map[string]any)["pattern"]; !ok {
				delete(schemaTypes, key)
			}
		}
	} else {
		c.AddResource(fullFileName, file)
	}
}

func validateDomain(entity int, obj, parent map[string]any) *u.Error {
	if entity == u.DOMAIN || !u.IsEntityHierarchical(entity) {
		return nil
	}
	if !CheckDomainExists(obj["domain"].(string)) {
		return &u.Error{Type: u.ErrNotFound,
			Message: "Domain not found: " + obj["domain"].(string)}
	}
	if parentDomain, ok := parent["domain"].(string); ok {
		if !DomainIsEqualOrChild(parentDomain, obj["domain"].(string)) {
			return &u.Error{Type: u.ErrBadFormat,
				Message: "Object domain is not equal or child of parent's domain"}
		}
	}
	return nil
}

func getParentSetId(entity int, obj map[string]any) (map[string]any, *u.Error) {
	var parent map[string]interface{}
	if u.IsEntityHierarchical(entity) {
		var err *u.Error
		parent, err = validateParent(u.EntityToString(entity), entity, obj)
		if err != nil {
			return parent, err
		} else if parent["id"] != nil {
			obj["id"] = parent["id"].(string) +
				u.HN_DELIMETER + obj["name"].(string)
		} else {
			obj["id"] = obj["name"].(string)
		}
	}
	return parent, nil
}

func validateParentId(entNum int, parentId any) (bool, *u.Error) {
	if entNum == u.SITE {
		// never has a parent
		return false, nil
	}
	// Check ParentID is valid
	if parentId == nil || parentId == "" {
		if entNum == u.DOMAIN || entNum == u.STRAYOBJ || entNum == u.VIRTUALOBJ {
			// allowed to not have a parent
			return false, nil
		}
		return false, &u.Error{Type: u.ErrBadFormat, Message: "ParentID is not valid"}
	}
	return true, nil
}

func validateParent(ent string, entNum int, t map[string]interface{}) (map[string]interface{}, *u.Error) {
	if hasParentId, err := validateParentId(entNum, t["parentId"]); !hasParentId {
		return nil, err
	}

	// Anyone can have a stray parent
	if parent := getParent([]string{"stray_object"}, t); parent != nil {
		return parent, nil
	}

	// If not, search specific possibilities
	switch entNum {
	case u.DEVICE:
		if parent := getParent([]string{"rack", "device"}, t); parent != nil {
			if err := validateDeviceSlotExists(t, parent); err != nil {
				return nil, err
			}
			delete(parent, "attributes") // only used to check slots
			return parent, nil
		}

		return nil, &u.Error{Type: u.ErrInvalidValue,
			Message: "ParentID should correspond to existing rack or device ID"}

	case u.GROUP:
		if parent := getParent([]string{"rack", "room"}, t); parent != nil {
			return parent, nil
		}

		return nil, &u.Error{Type: u.ErrInvalidValue,
			Message: "Group parent should correspond to existing rack or room"}

	case u.VIRTUALOBJ:
		if parent := getParent([]string{"device", "virtual_obj"}, t); parent != nil {
			return parent, nil
		}

		return nil, &u.Error{Type: u.ErrInvalidValue,
			Message: "Group parent should correspond to existing device or virtual_obj"}
	default:
		parentStr := u.EntityToString(u.GetParentOfEntityByInt(entNum))
		if parent := getParent([]string{parentStr}, t); parent != nil {
			return parent, nil
		}

		return nil, &u.Error{Type: u.ErrInvalidValue,
			Message: fmt.Sprintf("ParentID should correspond to existing %s ID", parentStr)}
	}
}

func getParent(parentEntities []string, t map[string]any) map[string]any {
	parent := map[string]any{"parent": ""}
	req := bson.M{"id": t["parentId"].(string)}
	for _, parentEnt := range parentEntities {
		obj, _ := GetObject(req, parentEnt, u.RequestFilters{}, nil)
		if obj != nil {
			parent["parent"] = parentEnt
			parent["domain"] = obj["domain"]
			parent["id"] = obj["id"]
			if t["category"] == "device" {
				// need attributes to check slots
				parent["attributes"] = obj["attributes"]
			}
			return parent
		}
	}
	return nil
}

func validateDeviceSlotExists(deviceData map[string]interface{}, parentData map[string]interface{}) *u.Error {
	// get requested slots
	deviceSlots, err := slotToValidSlice(deviceData["attributes"].(map[string]any))
	if err != nil {
		return err
	}

	// check if requested slots exist in parent device
	countFound := 0
	if templateSlug, ok := parentData["attributes"].(map[string]any)["template"].(string); ok {
		// get parent slots from its template
		template, _ := GetObject(bson.M{"slug": templateSlug}, "obj_template", u.RequestFilters{}, nil)
		if ps, ok := template["slots"].(primitive.A); ok {
			parentSlots := []interface{}(ps)
			for _, parentSlot := range parentSlots {
				if pie.Contains(deviceSlots, parentSlot.(map[string]any)["location"].(string)) {
					countFound = countFound + 1
				}
			}
		}
	}

	// check if all was found
	if len(deviceSlots) != countFound {
		return &u.Error{Type: u.ErrInvalidValue,
			Message: "Invalid slot: parent does not have all the requested slots"}
	}

	return nil
}

func formatJsonSchemaErrors(errors []jsonschema.BasicError) []string {
	errSlice := []string{}
	for _, schErr := range errors {
		// Check all json schema defined types
		for _, definition := range schemaTypes {
			pattern := definition.(map[string]any)["pattern"].(string)
			// If the pattern is in the error message
			patternErrPrefix := "does not match pattern "
			if strings.Contains(schErr.Error, patternErrPrefix+quote(pattern)) || strings.Contains(schErr.Error, patternErrPrefix+pattern) {
				// Substitute it for the more user-friendly description given by the schema
				schErr.Error = "should be " + definition.(map[string]any)["descriptions"].(map[string]any)["en"].(string)
			}
		}
		if len(schErr.Error) > 0 && !strings.Contains(schErr.Error, "doesn't validate with") {
			if len(schErr.InstanceLocation) > 0 {
				errSlice = append(errSlice, schErr.InstanceLocation+" "+schErr.Error)
			} else {
				errSlice = append(errSlice, schErr.Error)
			}
		}
	}
	return errSlice
}

func ValidateJsonSchema(entity int, t map[string]interface{}) (bool, *u.Error) {
	// Get JSON schema
	var schemaName string
	switch entity {
	case u.AC, u.CABINET, u.PWRPNL:
		schemaName = "base_schema.json"
	case u.STRAYOBJ:
		schemaName = "stray_schema.json"
	default:
		schemaName = u.EntityToString(entity) + "_schema.json"
	}

	sch, err := c.Compile(schemaName)
	if err != nil {
		return false, &u.Error{Type: u.ErrInternal, Message: err.Error()}
	}

	// Validate JSON Schema
	if err := sch.Validate(t); err != nil {
		switch v := err.(type) {
		case *jsonschema.ValidationError:
			fmt.Println(t)
			println(v.GoString())
			// Format errors array
			errSlice := formatJsonSchemaErrors(v.BasicOutput().Errors)
			return false, &u.Error{Type: u.ErrBadFormat,
				Message: "JSON body doesn't validate with the expected JSON schema",
				Details: errSlice}
		}
		return false, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
	} else {
		println("JSON Schema: all good, validated!")
		return true, nil
	}
}

func ValidateEntity(entity int, t map[string]interface{}) *u.Error {
	if shouldFillTags(entity, u.RequestFilters{}) {
		t = fillTags(t)
	}

	// Validate JSON Schema
	if ok, err := ValidateJsonSchema(entity, t); !ok {
		return err
	}

	// Check parent and domain for objects
	var parent map[string]interface{}
	parent, err := getParentSetId(entity, t)
	if err != nil {
		return err
	}
	if err := validateDomain(entity, t, parent); err != nil {
		return err
	}

	// Check ID unique for some entities
	if err := checkIdUnique(entity, t["id"]); err != nil {
		return err
	}

	// Check attributes
	if pie.Contains(u.EntitiesWithAttributeCheck, entity) {
		if err := validateAttributes(entity, t, parent); err != nil {
			return err
		}
	}

	// Layer extra check
	if entity == u.LAYER && !doublestar.ValidatePattern(t["applicability"].(string)) {
		return &u.Error{
			Type:    u.ErrBadFormat,
			Message: "Layer applicability pattern is not valid",
		}
	}

	//Successfully validated the Object
	return nil
}

// Returns true if at least 1 objects of type "entities" have the "value" for the "attribute".
func ObjectsHaveAttribute(entities []int, attribute, value string) (bool, *u.Error) {
	for _, entity := range entities {
		count, err := repository.CountObjects(entity, bson.M{attribute: value})
		if err != nil {
			return false, err
		}

		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}

func slotToValidSlice(attributes map[string]any) ([]string, *u.Error) {
	slotAttr := attributes["slot"]
	if pa, ok := slotAttr.(primitive.A); ok {
		slotAttr = []interface{}(pa)
	}
	if arr, ok := slotAttr.([]interface{}); ok {
		if len(arr) < 1 {
			return []string{}, &u.Error{Type: u.ErrInvalidValue,
				Message: "Invalid slot: must be a vector [] with at least one element"}
		}
		slotSlice := make([]string, len(arr))
		for i := range arr {
			slotSlice[i] = arr[i].(string)
		}
		return slotSlice, nil
	} else { // no slot provided (just posU is valid)
		return []string{}, nil
	}
}

// Returns single-quoted string
func quote(s string) string {
	s = fmt.Sprintf("%q", s)
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return "'" + s[1:len(s)-1] + "'"
}

// ID is guaranteed to be unique for each entity by mongo
// but some entities need some extra checks
func checkIdUnique(entity int, id any) *u.Error {
	// Check if Room Child ID is unique among all room children
	if pie.Contains(u.RoomChildren, entity) {
		if err := checkIdUniqueAmongEntities(u.SliceRemove(u.RoomChildren, entity),
			id.(string)); err != nil {
			return err
		}
	}
	// Check if Group ID is unique
	if entity == u.GROUP {
		entities := u.GetEntitiesById(u.Physical, id.(string))
		if err := checkIdUniqueAmongEntities(u.EntitiesStrToInt(entities),
			id.(string)); err != nil {
			return err
		}
	}
	return nil
}

func checkIdUniqueAmongEntities(entities []int, id string) *u.Error {
	idIsPresent, err := ObjectsHaveAttribute(
		entities,
		"id",
		id,
	)
	if err != nil {
		return err
	}

	if idIsPresent {
		return &u.Error{
			Type:    u.ErrBadFormat,
			Message: "This object ID is not unique",
		}
	}
	return nil
}
