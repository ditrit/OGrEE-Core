package models

import (
	"fmt"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TENANT = iota
	SITE
	BLDG
	ROOM
	RACK
	DEVICE
	SUBDEV
	SUBDEV1
)

func ValidateEntity(entity int, t map[string]interface{}) (map[string]interface{}, bool) {
	var objID primitive.ObjectID
	var err error
	if t["name"] == "" {
		return u.Message(false, "Name should be on payload"), false
	}

	if t["category"] == "" {
		return u.Message(false, "Category should be on the payload"), false
	}

	if t["domain"] == "" {
		return u.Message(false, "Domain should be on the payload"), false
	}

	//Check if Parent ID is valid
	//Tenants do not have Parents
	if entity > TENANT {

		objID, err = primitive.ObjectIDFromHex(t["parentId"].(string))
		if err != nil {
			return u.Message(false, "ParentID is not valid"), false
		}
		parent := u.EntityToString(entity - 1)

		ctx, cancel := u.Connect()
		if GetDB().Collection(parent).
			FindOne(ctx, bson.M{"_id": objID}).Err() != nil {
			println("We got: ", t["parentId"].(string))
			return u.Message(false, "SiteParentID should be correspond to tenant ID"), false

		}
		defer cancel()
	}

	if _, ok := t["attributes"]; !ok {
		return u.Message(false, "Attributes should be on the payload"), false
	} else {
		if v, ok := t["attributes"].(map[string]interface{}); !ok {
			return u.Message(false, "Attributes should be on the payload"), false
		} else {
			switch entity {
			case TENANT:
				if _, ok := v["color"]; !ok {
					return u.Message(false,
						"Color Attribute must be specified on the payload"), false
				}

			case SITE:
				switch v["orientation"] {
				case "EN", "NW", "WS", "SE":
				case "":
					return u.Message(false, "Orientation should be on the payload"), false

				default:
					return u.Message(false, "Orientation is invalid!"), false
				}

				if v["usableColor"] == "" {
					return u.Message(false, "Usable Color should be on the payload"), false
				}

				if v["reservedColor"] == "" {
					return u.Message(false, "Reserved Color should be on the payload"), false
				}

				if v["technicalColor"] == "" {
					return u.Message(false, "Technical Color should be on the payload"), false
				}

			case BLDG:
				if v["posXY"] == "" {
					return u.Message(false, "XY coordinates should be on payload"), false
				}

				if v["posXYU"] == "" {
					return u.Message(false, "PositionXYU string should be on the payload"), false
				}

				if v["posZ"] == "" {
					return u.Message(false, "Z coordinates should be on payload"), false
				}

				if v["posZU"] == "" {
					return u.Message(false, "PositionZU string should be on the payload"), false
				}

				if v["size"] == "" {
					return u.Message(false, "Invalid building size on the payload"), false
				}

				if v["sizeU"] == "" {
					return u.Message(false, "Building size string should be on the payload"), false
				}

				if v["height"] == "" {
					return u.Message(false, "Invalid Height on payload"), false
				}

				if v["heightU"] == "" {
					return u.Message(false, "Building Height string should be on the payload"), false
				}

			case ROOM:
				if v["posXY"] == "" {
					return u.Message(false, "XY coordinates should be on payload"), false
				}

				if v["posXYU"] == "" {
					return u.Message(false, "PositionXYU string should be on the payload"), false
				}

				if v["posZ"] == "" {
					return u.Message(false, "Z coordinates should be on payload"), false
				}

				if v["posZU"] == "" {
					return u.Message(false, "PositionZU string should be on the payload"), false
				}

				switch v["orientation"] {
				case "-E-N", "-E+N", "+E-N", "+E+N":
				case "-N-W", "-N+W", "+N-W", "+N+W":
				case "-W-S", "-W+S", "+W-S", "+W+S":
				case "-S-E", "-S+E", "+S-E", "+S+E":
				case "":
					return u.Message(false, "Orientation should be on the payload"), false

				default:
					return u.Message(false, "Orientation is invalid!"), false
				}

				if v["size"] == "" {
					return u.Message(false, "Invalid size on the payload"), false
				}

				if v["sizeU"] == "" {
					return u.Message(false, "Room size string should be on the payload"), false
				}

				if v["height"] == "" {
					return u.Message(false, "Invalid Height on payload"), false
				}

				if v["heightU"] == "" {
					return u.Message(false, "Room Height string should be on the payload"), false
				}
			case RACK:
				if v["posXY"] == "" {
					return u.Message(false, "XY coordinates should be on payload"), false
				}

				if v["posXYU"] == "" {
					return u.Message(false, "PositionXYU string should be on the payload"), false
				}

				switch v["orientation"] {
				case "front", "rear", "left", "right":
				case "":
					return u.Message(false, "Orientation should be on the payload"), false

				default:
					return u.Message(false, "Orientation is invalid!"), false
				}

				if v["size"] == "" {
					return u.Message(false, "Invalid size on the payload"), false
				}

				if v["sizeU"] == "" {
					return u.Message(false, "Rack size string should be on the payload"), false
				}

				if v["height"] == "" {
					return u.Message(false, "Invalid Height on payload"), false
				}

				if v["heightU"] == "" {
					return u.Message(false, "Rack Height string should be on the payload"), false
				}
			case DEVICE:
				switch v["orientation"] {
				case "front", "rear", "frontflipped", "rearflipped":
				case "":
					return u.Message(false, "Orientation should be on the payload"), false

				default:
					return u.Message(false, "Orientation is invalid!"), false
				}

				if v["size"] == "" {
					return u.Message(false, "Invalid size on the payload"), false
				}

				if v["sizeUnit"] == "" {
					return u.Message(false, "Rack size string should be on the payload"), false
				}

				if v["height"] == "" {
					return u.Message(false, "Invalid Height on payload"), false
				}

				if v["heightU"] == "" {
					return u.Message(false, "Rack Height string should be on the payload"), false
				}
			case SUBDEV, SUBDEV1:

				switch v["orientation"] {
				case "front", "rear", "frontflipped", "rearflipped":
				case "":
					return u.Message(false, "Orientation should be on the payload"), false

				default:
					return u.Message(false, "Orientation is invalid!"), false
				}

				if v["size"] == "" {
					return u.Message(false, "Invalid size on the payload"), false
				}

				if v["sizeUnit"] == "" {
					return u.Message(false, "Subdevice size string should be on the payload"), false
				}

				if v["height"] == "" {
					return u.Message(false, "Invalid Height on payload"), false
				}

				if v["heightU"] == "" {
					return u.Message(false, "Subdevice Height string should be on the payload"), false
				}
			}
		}
	}

	//Successfully validated the Object
	return u.Message(true, "success"), true
}

func CreateEntity(entity int, t map[string]interface{}) (map[string]interface{}, string) {

	if resp, ok := ValidateEntity(entity, t); !ok {
		return resp, "validate"
	}

	ctx, cancel := u.Connect()

	entStr := u.EntityToString(entity)
	res, e := GetDB().Collection(entStr).InsertOne(ctx, t)
	if e != nil {
		return u.Message(false,
				"Internal error while creating "+entStr+": "+e.Error()),
			e.Error()
	}
	defer cancel()

	t["id"] = res.InsertedID

	resp := u.Message(true, "success")
	resp["data"] = t
	return resp, ""
}

func GetEntity(entityID primitive.ObjectID, ent string) (map[string]interface{}, string) {
	t := map[string]interface{}{}

	ctx, cancel := u.Connect()
	e := GetDB().Collection(ent).FindOne(ctx, bson.M{"_id": entityID}).Decode(&t)
	if e != nil {
		return nil, e.Error()
	}
	defer cancel()
	return t, ""
}

func GetAllEntities(ent string) ([]map[string]interface{}, string) {
	data := make([]map[string]interface{}, 0)
	ctx, cancel := u.Connect()
	c, err := GetDB().Collection(ent).Find(ctx, bson.D{{}})
	if err != nil {
		fmt.Println(err)
		return nil, err.Error()
	}
	defer cancel()

	for c.Next(GetCtx()) {
		x := map[string]interface{}{}
		e := c.Decode(x)
		if e != nil {
			fmt.Println(err)
			return nil, err.Error()
		}
		data = append(data, x)
	}

	return data, ""
}

func DeleteEntity(entity string, id primitive.ObjectID) map[string]interface{} {
	ctx, cancel := u.Connect()
	c, _ := GetDB().Collection(entity).DeleteOne(ctx, bson.M{"_id": id})
	if c.DeletedCount == 0 {
		return u.Message(false, "There was an error in deleting the rack")
	}
	defer cancel()

	return u.Message(true, "success")
}

func UpdateEntity(ent string, id primitive.ObjectID, t *map[string]interface{}) (map[string]interface{}, string) {
	ctx, cancel := u.Connect()
	e := GetDB().Collection(ent).FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$set": *t}).Err()
	if e != nil {
		return u.Message(false, "failure: "+e.Error()), e.Error()
	}
	defer cancel()
	return u.Message(true, "success"), ""
}
