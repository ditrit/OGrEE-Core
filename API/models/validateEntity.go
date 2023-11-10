package models

import (
	"embed"
	"fmt"
	u "p3/utils"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.mongodb.org/mongo-driver/bson"
)

//go:embed schemas/*.json
//go:embed schemas/refs/*.json
var embeddfs embed.FS
var c *jsonschema.Compiler

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
				print(schemaPrefix + e.Name() + " ")
				c.AddResource(schemaPrefix+e.Name(), file)
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
		if entNum == u.DOMAIN || entNum == u.STRAYOBJ {
			return nil, nil
		}
		return nil, &u.Error{Type: u.ErrBadFormat, Message: "ParentID is not valid"}
	}
	req := bson.M{"id": t["parentId"].(string)}

	parent := map[string]interface{}{"parent": ""}
	// Anyone can have a stray parent
	stray, _ := GetObject(req, "stray_object", u.RequestFilters{}, nil)
	if stray != nil {
		parent["parent"] = "rack"
		parent["domain"] = stray["domain"]
		parent["id"] = stray["id"]
		return parent, nil
	}
	// If not, search specific possibilities
	switch entNum {
	case u.DEVICE:
		x, _ := GetObject(req, "rack", u.RequestFilters{}, nil)
		if x != nil {
			parent["parent"] = "rack"
			parent["domain"] = x["domain"]
			parent["id"] = x["id"]
			return parent, nil
		}

		y, _ := GetObject(req, "device", u.RequestFilters{}, nil)
		if y != nil {
			parent["parent"] = "device"
			parent["domain"] = y["domain"]
			parent["id"] = y["id"]
			return parent, nil
		}

		return nil, &u.Error{Type: u.ErrInvalidValue,
			Message: "ParentID should correspond to Existing ID"}

	case u.GROUP:
		w, _ := GetObject(req, "device", u.RequestFilters{}, nil)
		if w != nil {
			parent["parent"] = "device"
			parent["domain"] = w["domain"]
			parent["id"] = w["id"]
			return parent, nil
		}

		x, _ := GetObject(req, "rack", u.RequestFilters{}, nil)
		if x != nil {
			parent["parent"] = "rack"
			parent["domain"] = x["domain"]
			parent["id"] = x["id"]
			return parent, nil
		}

		y, _ := GetObject(req, "room", u.RequestFilters{}, nil)
		if y != nil {
			parent["parent"] = "room"
			parent["domain"] = y["domain"]
			parent["id"] = y["id"]
			return parent, nil
		}

		z, _ := GetObject(req, "building", u.RequestFilters{}, nil)
		if z != nil {
			parent["parent"] = "building"
			parent["domain"] = z["domain"]
			parent["id"] = z["id"]
			return parent, nil
		}

		return nil, &u.Error{Type: u.ErrInvalidValue,
			Message: "ParentID should correspond to Existing ID"}

	default:
		parentInt := u.GetParentOfEntityByInt(entNum)
		parentStr := u.EntityToString(parentInt)

		p, err := GetObject(req, parentStr, u.RequestFilters{}, nil)
		if len(p) > 0 {
			parent["parent"] = parentStr
			parent["domain"] = p["domain"]
			parent["id"] = p["id"]
			return parent, nil
		} else if err != nil {
			println("ENTITY VALUE: ", ent)
			println("We got Parent: ", parent, " with ID:", t["parentId"].(string))
			return nil, &u.Error{Type: u.ErrInvalidValue,
				Message: "ParentID should correspond to Existing ID"}
		}
	}
	return nil, nil
}

func validateJsonSchema(entity int, t map[string]interface{}) (bool, *u.Error) {
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
	/*
		TODO:
		Need to capture device if it is a parent
		and check that the device parent has a slot
		attribute
	*/
	if shouldFillTags(entity, u.RequestFilters{}) {
		t = fillTags(t)
	}

	// Validate JSON Schema
	if ok, err := validateJsonSchema(entity, t); !ok {
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
	if entity == u.RACK || entity == u.GROUP || entity == u.CORRIDOR {
		if _, ok := t["attributes"]; !ok {
			return &u.Error{Type: u.ErrBadFormat,
				Message: "Attributes should be on the payload"}
		} else {
			if v, ok := t["attributes"].(map[string]interface{}); !ok {
				return &u.Error{Type: u.ErrBadFormat,
					Message: "Attributes should be a JSON Dictionary"}
			} else {
				switch entity {
				case u.RACK:
					//Ensure the name is also unique among corridors
					req := bson.M{"name": t["name"].(string)}
					req["domain"] = t["domain"].(string)
					nameCheck, _ := GetManyObjects("corridor", req, u.RequestFilters{}, nil)
					if nameCheck != nil {
						if len(nameCheck) != 0 {
							if nameCheck != nil {
								println(nameCheck[0]["name"].(string))
							}
							return &u.Error{Type: u.ErrBadFormat,
								Message: "Rack name must be unique among corridors and racks"}
						}

					}

				case u.CORRIDOR:
					//Ensure the name is also unique among racks
					req := bson.M{"name": t["name"].(string)}
					req["domain"] = t["domain"].(string)
					nameCheck, _ := GetManyObjects("rack", req, u.RequestFilters{}, nil)
					if nameCheck != nil {
						if len(nameCheck) != 0 {
							return &u.Error{Type: u.ErrBadFormat,
								Message: "Corridor name must be unique among corridors and racks"}
						}
					}
					//Set the color manually based on temp. as specified by client
					if v["temperature"] == "warm" {
						v["color"] = "990000"
					} else if v["temperature"] == "cold" {
						v["color"] = "000099"
					}

				case u.GROUP:
					objects := strings.Split(v["content"].(string), ",")
					if len(objects) <= 1 {
						if objects[0] == "" {
							return &u.Error{Type: u.ErrBadFormat,
								Message: "objects separated by a comma must be" +
									" on the payload"}
						}
					}

					//Ensure objects are all unique
					if _, ok := EnsureUnique(objects); !ok {
						return &u.Error{Type: u.ErrBadFormat,
							Message: "The group cannot have duplicate objects"}
					}

					//Ensure objects all exist
					orReq := bson.A{}
					for i := range objects {
						orReq = append(orReq, bson.D{{"id", t["parentId"].(string) + u.HN_DELIMETER + objects[i]}})
					}
					filter := bson.M{"$or": orReq}

					//If parent is rack, retrieve devices
					if parent["parent"].(string) == "rack" {
						ans, err := GetManyObjects("device", filter, u.RequestFilters{}, nil)
						if err != nil {
							return err
						}
						if len(ans) != len(objects) {
							return &u.Error{Type: u.ErrBadFormat,
								Message: "Unable to verify objects in specified group" +
									" please check and try again"}
						}

					} else if parent["parent"].(string) == "room" {
						//If parent is room, retrieve corridors and racks
						corridors, err := GetManyObjects("corridor", filter, u.RequestFilters{}, nil)
						if err != nil {
							return err
						}

						racks, err := GetManyObjects("rack", filter, u.RequestFilters{}, nil)
						if err != nil {
							return err
						}
						if len(racks)+len(corridors) != len(objects) {
							return &u.Error{Type: u.ErrBadFormat,
								Message: "Some object(s) could be not be found. " +
									"Please check and try again"}
						}
					}
				}
			}
		}
	}

	//Successfully validated the Object
	return nil
}

// Auxillary Functions
func EnsureUnique(x []string) (string, bool) {
	dict := map[string]int{}
	for _, item := range x {
		dict[item]++
		if dict[item] > 1 {
			return item, false
		}
	}
	return "", true
}
