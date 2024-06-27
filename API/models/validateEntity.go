package models

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
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
var types map[string]any

func init() {
	// Load JSON schemas
	c = jsonschema.NewCompiler()
	println("Loaded json schemas for validation:")
	loadJsonSchemas("")
	loadJsonSchemas("refs/")
	println()
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
			if err == nil {
				if e.Name() == "types.json" {
					// Make two copies of the reader stream
					var buf bytes.Buffer
					tee := io.TeeReader(file, &buf)

					print(schemaPrefix + e.Name() + " ")
					c.AddResource(schemaPrefix+e.Name(), tee)

					// Read and unmarshall types.json file
					typesBytes, _ := io.ReadAll(&buf)
					json.Unmarshal(typesBytes, &types)

					// Remove types that do not have a "pattern" attribute
					types = types["definitions"].(map[string]any)
					for key, definition := range types {
						if _, ok := definition.(map[string]any)["pattern"]; !ok {
							delete(types, key)
						}
					}
				} else {
					print(schemaPrefix + e.Name() + " ")
					c.AddResource(schemaPrefix+e.Name(), file)
				}
			}
		}
	}
}

func validateParent(ent string, entNum int, t map[string]interface{}) (map[string]interface{}, *u.Error) {
	if entNum == u.SITE {
		return nil, nil
	}

	//Check ParentID is valid
	if t["parentId"] == nil || t["parentId"] == "" {
		if entNum == u.DOMAIN || entNum == u.STRAYOBJ || entNum == u.VIRTUALOBJ {
			return nil, nil
		}
		return nil, &u.Error{Type: u.ErrBadFormat, Message: "ParentID is not valid"}
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
	fmt.Println(deviceData["attributes"].(map[string]any)["slot"])
	if deviceSlots, err := slotToValidSlice(deviceData["attributes"].(map[string]any)); err == nil {
		// check if requested slots exist in parent device
		countFound := 0
		if templateSlug, ok := parentData["attributes"].(map[string]any)["template"].(string); ok {
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
		if len(deviceSlots) != countFound {
			return &u.Error{Type: u.ErrInvalidValue,
				Message: "Invalid slot: parent does not have all the requested slots"}
		}
	} else if err != nil {
		return err
	}
	return nil
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
			errSlice := []string{}
			for _, schErr := range v.BasicOutput().Errors {
				// Check all types
				for _, definition := range types {
					pattern := definition.(map[string]any)["pattern"].(string)
					// If the pattern is in the error message
					if strings.Contains(schErr.Error, "does not match pattern "+quote(pattern)) || strings.Contains(schErr.Error, "does not match pattern "+pattern) {
						// Substitute it for the more user-friendly description
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

	// Extra checks
	// Check parent and domain for objects
	var parent map[string]interface{}
	if u.IsEntityHierarchical(entity) {
		var err *u.Error
		parent, err = validateParent(u.EntityToString(entity), entity, t)
		if err != nil {
			return err
		} else if parent["id"] != nil {
			t["id"] = parent["id"].(string) +
				u.HN_DELIMETER + t["name"].(string)
		} else {
			t["id"] = t["name"].(string)
		}
		//Check domain
		if entity != u.DOMAIN {
			if !CheckDomainExists(t["domain"].(string)) {
				return &u.Error{Type: u.ErrNotFound,
					Message: "Domain not found: " + t["domain"].(string)}
			}
			if parentDomain, ok := parent["domain"].(string); ok {
				if !DomainIsEqualOrChild(parentDomain, t["domain"].(string)) {
					return &u.Error{Type: u.ErrBadFormat,
						Message: "Object domain is not equal or child of parent's domain"}
				}
			}
		}
	}

	// Check attributes
	if entity == u.RACK || entity == u.GROUP || entity == u.CORRIDOR || entity == u.GENERIC ||
		entity == u.DEVICE || entity == u.VIRTUALOBJ {
		attributes := t["attributes"].(map[string]any)

		if pie.Contains(u.RoomChildren, entity) {
			// if entity is room children, verify that the id (name) is not repeated in other children
			idIsPresent, err := ObjectsHaveAttribute(
				u.SliceRemove(u.RoomChildren, entity),
				"id",
				t["id"].(string),
			)
			if err != nil {
				return err
			}

			if idIsPresent {
				return &u.Error{
					Type:    u.ErrBadFormat,
					Message: "Object name must be unique among corridors, racks and generic objects",
				}
			}
		}

		switch entity {
		case u.CORRIDOR:
			// Set the color manually based on temp. as specified by client
			if attributes["temperature"] == "warm" {
				attributes["color"] = "990000"
			} else if attributes["temperature"] == "cold" {
				attributes["color"] = "000099"
			}
		case u.GROUP:
			objects := attributes["content"].([]interface{})
			if len(objects) <= 1 && objects[0] == "" {
				return &u.Error{
					Type:    u.ErrBadFormat,
					Message: "objects separated by a comma must be on the payload",
				}
			}

			// Ensure objects are all unique
			if !pie.AreUnique(objects) {
				return &u.Error{
					Type:    u.ErrBadFormat,
					Message: "The group cannot have duplicate objects",
				}
			}

			// Ensure objects all exist
			orReq := bson.A{}
			for _, objectName := range objects {
				if strings.Contains(objectName.(string), u.HN_DELIMETER) {
					return &u.Error{
						Type:    u.ErrBadFormat,
						Message: "All group objects must be directly under the parent (no . allowed)",
					}
				}
				orReq = append(orReq, bson.M{"id": t["parentId"].(string) + u.HN_DELIMETER + objectName.(string)})
			}
			filter := bson.M{"$or": orReq}

			// If parent is rack, retrieve devices
			if parent["parent"].(string) == "rack" {
				count, err := repository.CountObjects(u.DEVICE, filter)
				if err != nil {
					return err
				}

				if count != len(objects) {
					return &u.Error{Type: u.ErrBadFormat,
						Message: "Unable to verify objects in specified group" +
							" please check and try again"}
				}
			} else if parent["parent"].(string) == "room" {
				// If parent is room, retrieve room children
				count, err := repository.CountObjectsManyEntities(u.RoomChildren, filter)
				if err != nil {
					return err
				}

				if count != len(objects) {
					return &u.Error{
						Type:    u.ErrBadFormat,
						Message: "Some object(s) could not be found. Please check and try again",
					}
				}
			}

			// Check if Group ID is unique
			entities := u.GetEntitiesByNamespace(u.Physical, t["id"].(string))
			for _, entStr := range entities {
				if entStr != u.EntityToString(u.GROUP) {
					// Get objects
					entData, err := GetManyObjects(entStr, bson.M{"id": t["id"]}, u.RequestFilters{}, "", nil)
					if err != nil {
						err.Message = "Error while check id unicity at " + entStr + ":" + err.Message
						return err
					}
					if len(entData) > 0 {
						return &u.Error{Type: u.ErrBadFormat,
							Message: "This group ID is not unique among " + entStr + "s"}
					}
				}
			}
		case u.DEVICE:
			if deviceSlots, err := slotToValidSlice(attributes); err == nil {
				// check if all requested slots are free
				idPattern := primitive.Regex{Pattern: "^" + t["parentId"].(string) +
					"(." + u.NAME_REGEX + "){1}$", Options: ""} // find siblings
				if siblings, err := GetManyObjects(u.EntityToString(u.DEVICE), bson.M{"id": idPattern},
					u.RequestFilters{}, "", nil); err != nil {
					return err
				} else {
					for _, obj := range siblings {
						if obj["name"] != t["name"] { // do not check itself
							if siblingSlots, err := slotToValidSlice(obj["attributes"].(map[string]any)); err == nil {
								for _, requestedSlot := range deviceSlots {
									if pie.Contains(siblingSlots, requestedSlot) {
										return &u.Error{Type: u.ErrBadFormat,
											Message: "Invalid slot: one or more requested slots are already in use"}
									}
								}
							}
						}
					}
				}
			} else if err != nil {
				return err
			}
		case u.VIRTUALOBJ:
			if attributes["vlinks"] != nil {
				// check if all vlinks point to valid objects
				for _, vlinkId := range attributes["vlinks"].([]any) {
					count, err := repository.CountObjectsManyEntities([]int{u.DEVICE, u.VIRTUALOBJ},
						bson.M{"id": strings.Split(vlinkId.(string), "#")[0]})
					if err != nil {
						return err
					}

					if count != 1 {
						return &u.Error{
							Type:    u.ErrBadFormat,
							Message: "One or more vlink objects could not be found. Note that it must be device or virtual obj",
						}
					}
				}
			}
		}
	} else if entity == u.LAYER && !doublestar.ValidatePattern(t["applicability"].(string)) {
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
