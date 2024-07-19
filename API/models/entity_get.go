package models

import (
	"fmt"
	"maps"
	"p3/repository"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetObject(req bson.M, entityStr string, filters u.RequestFilters, userRoles map[string]Role) (map[string]interface{}, *u.Error) {
	object, err := repository.GetObject(req, entityStr, filters)

	if err != nil {
		return nil, err
	}

	//Remove _id
	object = fixID(object)

	entity := u.EntityStrToInt(entityStr)

	if shouldFillTags(entity, filters) {
		object = fillTags(object)
	}

	// Check permissions
	if u.IsEntityHierarchical(entity) {
		var domain string
		if entity == u.DOMAIN {
			domain = object["id"].(string)
		} else {
			domain = object["domain"].(string)
		}
		if userRoles != nil {
			if permission := CheckUserPermissions(userRoles, u.EntityStrToInt(entityStr), domain); permission == NONE {
				return nil, &u.Error{Type: u.ErrUnauthorized,
					Message: "User does not have permission to see this object"}
			} else if permission == READONLYNAME {
				object = FixReadOnlyName(object)
			}
		}
	}

	return object, nil
}

func GetManyObjects(entityStr string, req bson.M, filters u.RequestFilters, complexFilterExp string, userRoles map[string]Role) ([]map[string]interface{}, *u.Error) {
	ctx, cancel := u.Connect()
	var err error
	var c *mongo.Cursor

	var opts *options.FindOptions
	if len(filters.FieldsToShow) > 0 {
		compoundIndex := bson.D{bson.E{Key: "domain", Value: 1}, bson.E{Key: "id", Value: 1}}
		for _, field := range filters.FieldsToShow {
			if field != "domain" && field != "id" {
				compoundIndex = append(compoundIndex, bson.E{Key: field, Value: 1})
			}
		}
		opts = options.Find().SetProjection(compoundIndex)
	}
	err = repository.GetDateFilters(req, filters.StartDate, filters.EndDate)
	if err != nil {
		return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
	}

	if complexFilterExp != "" {
		if complexFilters, err := ComplexFilterToMap(complexFilterExp); err != nil {
			return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
		} else {
			err = getDatesFromComplexFilters(complexFilters)
			if err != nil {
				return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
			}
			u.ApplyWildcardsOnComplexFilter(complexFilters)
			maps.Copy(req, complexFilters)
		}
	}

	if opts != nil {
		c, err = repository.GetDB().Collection(entityStr).Find(ctx, req, opts)
	} else {
		c, err = repository.GetDB().Collection(entityStr).Find(ctx, req)
	}
	if err != nil {
		fmt.Println(err)
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	defer cancel()

	entity := u.EntityStrToInt(entityStr)
	data, e1 := ExtractCursor(c, ctx, entity, userRoles)
	if e1 != nil {
		fmt.Println(e1)
		return nil, &u.Error{Type: u.ErrInternal, Message: e1.Error()}
	}

	if shouldFillTags(entity, filters) {
		for i := range data {
			fillTags(data[i])
		}
	}

	return data, nil
}

func GetEntityCount(entity int) int64 {
	ent := u.EntityToString(entity)
	ctx, cancel := u.Connect()
	ans, e := repository.GetDB().Collection(ent).CountDocuments(ctx, bson.M{}, nil)
	if e != nil {
		println(e.Error())
		return -1
	}
	defer cancel()
	return ans
}

func GetStats() map[string]interface{} {
	ans := map[string]interface{}{}
	t := map[string]interface{}{}
	latestDocArr := []map[string]interface{}{}
	var latestTime interface{}

	for entity := range u.Entities {
		num := GetEntityCount(entity)
		if num == -1 {
			num = 0
		}

		ans["Number of "+u.EntityToString(entity)+"s:"] = num

		//Retrieve the latest updated document in each collection
		//and store into the latestDocArr array
		obj := map[string]interface{}{}
		filter := options.FindOne().SetSort(bson.M{"lastUpdated": -1})
		ctx, cancel := u.Connect()

		e := repository.GetDB().Collection(u.EntityToString(entity)).FindOne(ctx, bson.M{}, filter).Decode(&obj)
		if e == nil {
			latestDocArr = append(latestDocArr, obj)
		}
		defer cancel()
	}

	//Get the latest update out of latestDocArr
	value := -1
	for _, obj := range latestDocArr {
		if int(obj["lastUpdated"].(primitive.DateTime)) > value {
			value = int(obj["lastUpdated"].(primitive.DateTime))
			latestTime = obj["lastUpdated"]
		}
	}

	if latestTime == nil {
		latestTime = "N/A"
	}

	cmd := bson.D{{"dbStats", 1}, {"scale", 1024}}

	if e := CommandRunner(cmd).Decode(&t); e != nil {
		println(e.Error())
		return nil
	}

	ans["Number of Hierarchal Objects"] = t["collections"]
	ans["Last Job Timestamp"] = latestTime

	return ans
}
