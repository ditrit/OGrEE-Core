package models

import (
	"embed"
	"fmt"
	u "p3/utils"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func validateParent(ent string, entNum int, t map[string]interface{}) (map[string]interface{}, bool) {
	if entNum == u.SITE {
		return nil, true
	}

	//Check ParentID is valid
	if t["parentId"] == nil || t["parentId"] == "" {
		if entNum == u.DOMAIN || entNum == u.STRAYDEV {
			return nil, true
		}
		return u.Message(false, "ParentID is not valid"), false
	}
	objID, err := primitive.ObjectIDFromHex(t["parentId"].(string))
	var req primitive.M
	if err == nil {
		// parentId given with ID
		req = bson.M{"_id": objID}
	} else {
		// parentId given with hierarchyName
		req = bson.M{"hierarchyName": t["parentId"].(string)}
	}

	parent := map[string]interface{}{"parent": ""}
	switch entNum {
	case u.DEVICE:
		x, _ := GetEntity(req, "rack", u.RequestFilters{}, nil)
		if x != nil {
			parent["parent"] = "rack"
			parent["hierarchyName"] = getHierarchyName(x)
			return parent, true
		}

		y, _ := GetEntity(req, "device", u.RequestFilters{}, nil)
		if y != nil {
			parent["parent"] = "device"
			parent["hierarchyName"] = getHierarchyName(y)
			return parent, true
		}

		return u.Message(false,
			"ParentID should be correspond to Existing ID"), false

	case u.SENSOR, u.GROUP:
		w, _ := GetEntity(req, "device", u.RequestFilters{}, nil)
		if w != nil {
			parent["parent"] = "device"
			parent["hierarchyName"] = getHierarchyName(w)
			return parent, true
		}

		x, _ := GetEntity(req, "rack", u.RequestFilters{}, nil)
		if x != nil {
			parent["parent"] = "rack"
			parent["hierarchyName"] = getHierarchyName(x)
			return parent, true
		}

		y, _ := GetEntity(req, "room", u.RequestFilters{}, nil)
		if y != nil {
			parent["parent"] = "room"
			parent["hierarchyName"] = getHierarchyName(y)
			return parent, true
		}

		z, _ := GetEntity(req, "building", u.RequestFilters{}, nil)
		if z != nil {
			parent["parent"] = "building"
			parent["hierarchyName"] = getHierarchyName(z)
			return parent, true
		}

		return u.Message(false,
			"ParentID should be correspond to Existing ID"), false

	case u.STRAYDEV, u.STRAYSENSOR:
		if t["parentId"] != nil && t["parentId"] != "" {
			if pid, ok := t["parentId"].(string); ok {
				ID, _ := primitive.ObjectIDFromHex(pid)

				p, err := GetEntity(bson.M{"_id": ID}, "stray_device", u.RequestFilters{}, nil)
				if len(p) > 0 {
					parent["parent"] = "stray_device"
					parent["hierarchyName"] = getHierarchyName(p)
					return parent, true
				} else if err != "" {
					return u.Message(false,
						"ParentID should be an Existing ID or null"), false
				}
			} else {
				return u.Message(false,
					"ParentID should be an Existing ID or null"), false
			}
		}

	default:
		parentInt := u.GetParentOfEntityByInt(entNum)
		parentStr := u.EntityToString(parentInt)

		p, err := GetEntity(req, parentStr, u.RequestFilters{}, nil)
		if len(p) > 0 {
			parent["parent"] = parentStr
			parent["hierarchyName"] = getHierarchyName(p)
			return parent, true
		} else if err != "" {
			println("ENTITY VALUE: ", ent)
			println("We got Parent: ", parent, " with ID:", t["parentId"].(string))
			return u.Message(false,
				"ParentID should correspond to Existing ID"), false
		}
	}
	return nil, true
}

func getHierarchyName(parent map[string]interface{}) string {
	if parent["hierarchyName"] != nil {
		return parent["hierarchyName"].(string)
	} else {
		return parent["name"].(string)
	}
}

func validateJsonSchema(entity int, t map[string]interface{}) (map[string]interface{}, bool) {
	// Get JSON schema
	var schemaName string
	switch entity {
	case u.AC, u.CABINET, u.PWRPNL:
		schemaName = "base_schema.json"
	case u.STRAYDEV, u.STRAYSENSOR:
		schemaName = "stray_schema.json"
	default:
		schemaName = u.EntityToString(entity) + "_schema.json"
	}

	sch, err := c.Compile(schemaName)
	if err != nil {
		return u.Message(false, err.Error()), false
	}

	// Validate JSON Schema
	if err := sch.Validate(t); err != nil {
		switch v := err.(type) {
		case *jsonschema.ValidationError:
			fmt.Println(t)
			println(v.GoString())
			resp := u.Message(false, "JSON body doesn't validate with the expected JSON schema")
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
			resp["errors"] = errSlice
			return resp, false
		}
		return u.Message(false, err.Error()), false
	} else {
		println("JSON Schema: all good, validated!")
		return nil, true
	}
}

func ValidateEntity(entity int, t map[string]interface{}) (map[string]interface{}, bool) {
	/*
		TODO:
		Need to capture device if it is a parent
		and check that the device parent has a slot
		attribute
	*/

	// Validate JSON Schema
	if resp, err := validateJsonSchema(entity, t); !err {
		return resp, false
	}

	// Extra checks
	// Check parent and domain for objects
	var parent map[string]interface{}
	if entity != u.BLDGTMPL && entity != u.ROOMTMPL && entity != u.OBJTMPL {
		var ok bool
		parent, ok = validateParent(u.EntityToString(entity), entity, t)
		if !ok {
			return parent, ok
		} else if parent["hierarchyName"] != nil {
			t["hierarchyName"] = parent["hierarchyName"].(string) + u.HN_DELIMETER + t["name"].(string)
		} else {
			t["hierarchyName"] = t["name"].(string)
		}
		//Check domain
		if entity != u.DOMAIN {
			if !CheckDomainExists(t["domain"].(string)) {
				return u.Message(false, "Domain not found: "+t["domain"].(string)), false
			}
		}
	}

	// Check attributes
	if entity == u.RACK || entity == u.GROUP || entity == u.CORRIDOR {
		if _, ok := t["attributes"]; !ok {
			return u.Message(false, "Attributes should be on the payload"), false
		} else {
			if v, ok := t["attributes"].(map[string]interface{}); !ok {
				return u.Message(false, "Attributes should be a JSON Dictionary"), false
			} else {
				switch entity {
				case u.RACK:
					//Ensure the name is also unique among corridors
					req := bson.M{"name": t["name"].(string)}
					req["domain"] = t["domain"].(string)
					nameCheck, _ := GetManyEntities("corridor", req, u.RequestFilters{}, nil)
					if nameCheck != nil {
						if len(nameCheck) != 0 {
							msg := "Rack name must be unique among corridors and racks"
							if nameCheck != nil {
								println(nameCheck[0]["name"].(string))
							}
							return u.Message(false, msg), false
						}

					}

				case u.CORRIDOR:
					//Ensure the 2 racks are valid
					racks := strings.Split(v["content"].(string), ",")
					if len(racks) != 2 {
						msg := "2 racks separated by a comma must be on the payload"
						return u.Message(false, msg), false
					}

					//Trim Spaces because they mess up
					//the retrieval of objects from DB
					racks[0] = strings.TrimSpace(racks[0])
					racks[1] = strings.TrimSpace(racks[1])

					//Ensure the name is also unique among racks
					req := bson.M{"name": t["name"].(string)}
					req["domain"] = t["domain"].(string)
					nameCheck, _ := GetManyEntities("rack", req, u.RequestFilters{}, nil)
					if nameCheck != nil {
						if len(nameCheck) != 0 {
							msg := "Corridor name must be unique among corridors and racks"
							return u.Message(false, msg), false
						}

					}

					//Fetch the 2 racks and ensure they exist
					filter := bson.M{"_id": t["parentId"], "name": racks[0]}
					orReq := bson.A{bson.D{{"name", racks[0]}}, bson.D{{"name", racks[1]}}}

					filter = bson.M{"parentId": t["parentId"], "$or": orReq}
					ans, e := GetManyEntities("rack", filter, u.RequestFilters{}, nil)
					if e != "" {
						msg := "The racks you specified were not found." +
							" Please verify your input and try again"
						println(e)
						return u.Message(false, msg), false
					}

					if len(ans) != 2 {
						//Request possibly mentioned same racks
						//thus giving length of 1
						if !(len(ans) == 1 && racks[0] == racks[1]) {

							//Figure out the rack name that wasn't found
							var notFound string
							if racks[0] != ans[0]["name"].(string) {
								notFound = racks[0]
							} else {
								notFound = racks[1]
							}
							msg := "Unable to get the rack: " + notFound + ". Please check your inventory and try again"
							println("LENGTH OF u.RACK CHECK:", len(ans))
							println("CORRIDOR PARENTID: ", t["parentId"].(string))
							return u.Message(false, msg), false
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
							msg := "objects separated by a comma must be" +
								" on the payload"
							return u.Message(false, msg), false
						}

					}

					//Ensure objects are all unique
					if _, ok := EnsureUnique(objects); !ok {
						msg := "The group cannot have duplicate objects"
						return u.Message(false, msg), false
					}

					//Ensure objects all exist
					orReq := bson.A{}
					for i := range objects {
						orReq = append(orReq, bson.D{{"name", objects[i]}})
					}
					filter := bson.M{"parentId": t["parentId"], "$or": orReq}

					//If parent is rack, retrieve devices
					if parent["parent"].(string) == "rack" {
						ans, ok := GetManyEntities("device", filter, u.RequestFilters{}, nil)
						if ok != "" {
							return u.Message(false, ok), false
						}
						if len(ans) != len(objects) {
							msg := "Unable to verify objects in specified group" +
								" please check and try again"
							return u.Message(false, msg), false
						}

					} else if parent["parent"].(string) == "room" {

						//If parent is room, retrieve corridors and racks
						corridors, e1 := GetManyEntities("corridor", filter, u.RequestFilters{}, nil)
						if e1 != "" {
							return u.Message(false, e1), false
						}

						racks, e2 := GetManyEntities("rack", filter, u.RequestFilters{}, nil)
						if e2 != "" {
							return u.Message(false, e1), false
						}
						if len(racks)+len(corridors) != len(objects) {
							msg := "Some object(s) could be not be found. " +
								"Please check and try again"
							return u.Message(false, msg), false
						}
					}

				}
			}
		}
	}

	//Successfully validated the Object
	return u.Message(true, "success"), true
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
