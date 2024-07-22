package models

import (
	"encoding/json"
	"errors"
	"p3/repository"
	u "p3/utils"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var AttrsWithInnerObj = []string{"pillars", "separators", "breakers"}

func UpdateObject(entityStr string, id string, updateData map[string]interface{}, isPatch bool, userRoles map[string]Role, isRecursive bool) (map[string]interface{}, *u.Error) {
	var idFilter bson.M
	if u.IsEntityNonHierarchical(u.EntityStrToInt(entityStr)) {
		idFilter = bson.M{"slug": id}
	} else {
		idFilter = bson.M{"id": id}
	}

	//Update timestamp requires first obj retrieval
	//there isn't any way for mongoDB to make a field
	//immutable in a document
	var oldObj map[string]any
	var err *u.Error
	if entityStr == u.HIERARCHYOBJS_ENT {
		oldObj, err = GetHierarchicalObjectById(id, u.RequestFilters{}, userRoles)
		if err == nil {
			entityStr = oldObj["category"].(string)
		}
	} else {
		oldObj, err = GetObject(idFilter, entityStr, u.RequestFilters{}, userRoles)
	}
	if err != nil {
		return nil, err
	}

	entity := u.EntityStrToInt(entityStr)

	// Check if permission is only readonly
	if u.IsEntityHierarchical(entity) && oldObj["description"] == nil {
		// Description is always present, unless GetEntity was called with readonly permission
		return nil, &u.Error{Type: u.ErrUnauthorized,
			Message: "User does not have permission to change this object"}
	}

	tags, tagsPresent := getTags(updateData)

	// Update old object data with patch data
	if isPatch {
		if tagsPresent {
			return nil, &u.Error{
				Type:    u.ErrBadFormat,
				Message: "Tags cannot be modified in this way, use tags+ and tags-",
			}
		}

		var formattedOldObj map[string]interface{}
		// Convert primitive.A and similar types
		bytes, _ := json.Marshal(oldObj)
		json.Unmarshal(bytes, &formattedOldObj)
		// Update old with new
		err := updateOldObjWithPatch(formattedOldObj, updateData)
		if err != nil {
			return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
		}

		updateData = formattedOldObj
		// Remove API set fields
		delete(updateData, "id")
		delete(updateData, "lastUpdated")
		delete(updateData, "createdDate")
	} else if tagsPresent {
		err := verifyTagList(tags)
		if err != nil {
			return nil, err
		}
	}

	result, err := WithTransaction(func(ctx mongo.SessionContext) (interface{}, error) {
		err = prepareUpdateObject(ctx, entity, id, updateData, oldObj, userRoles)
		if err != nil {
			return nil, err
		}

		mongoRes := repository.GetDB().Collection(entityStr).FindOneAndReplace(
			ctx,
			idFilter, updateData,
			options.FindOneAndReplace().SetReturnDocument(options.After),
		)
		if mongoRes.Err() != nil {
			return nil, mongoRes.Err()
		}

		if oldObj["id"] != updateData["id"] {
			// Changes to id should be propagated
			if err := repository.PropagateParentIdChange(
				ctx,
				oldObj["id"].(string),
				updateData["id"].(string),
				entity,
			); err != nil {
				return nil, err
			} else if entity == u.DOMAIN {
				if err := repository.PropagateDomainChange(ctx,
					oldObj["id"].(string),
					updateData["id"].(string),
				); err != nil {
					return nil, err
				}
			}
		}
		if u.IsEntityHierarchical(entity) && (oldObj["domain"] != updateData["domain"]) {
			if isRecursive {
				// Change domain of all children too
				if err := repository.PropagateDomainChangeToChildren(
					ctx,
					updateData["id"].(string),
					updateData["domain"].(string),
				); err != nil {
					return nil, err
				}
			} else {
				// Check if children domains are compatible
				if err := repository.CheckParentDomainChange(entity, updateData["id"].(string),
					updateData["domain"].(string)); err != nil {
					return nil, err
				}
			}
		}

		return mongoRes, nil
	})

	if err != nil {
		return nil, err
	}

	var updatedDoc map[string]interface{}
	result.(*mongo.SingleResult).Decode(&updatedDoc)

	return fixID(updatedDoc), nil
}

func prepareUpdateObject(ctx mongo.SessionContext, entity int, id string, updateData, oldObject map[string]any, userRoles map[string]Role) *u.Error {
	// Check user permissions in case domain is being updated
	if entity != u.DOMAIN && u.IsEntityHierarchical(entity) && (oldObject["domain"] != updateData["domain"]) {
		if perm := CheckUserPermissions(userRoles, entity, updateData["domain"].(string)); perm < WRITE {
			return &u.Error{Type: u.ErrUnauthorized, Message: "User does not have permission to change this object"}
		}
	}

	// tag list edition support
	err := addAndRemoveFromTags(entity, id, updateData)
	if err != nil {
		return err
	}

	// Ensure the update is valid
	err = ValidateEntity(entity, updateData)
	if err != nil {
		return err
	}

	updateData["lastUpdated"] = primitive.NewDateTimeFromTime(time.Now())
	updateData["createdDate"] = oldObject["createdDate"]
	delete(updateData, "parentId")

	if entity == u.TAG {
		// tag slug edition support
		if updateData["slug"].(string) != oldObject["slug"].(string) {
			err := repository.UpdateTagSlugInEntities(ctx, oldObject["slug"].(string), updateData["slug"].(string))
			if err != nil {
				return err
			}
		}

		err := updateTagImage(ctx, oldObject, updateData)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateOldObjWithPatch(old map[string]interface{}, patch map[string]interface{}) error {
	for k, v := range patch {
		switch patchValueCasted := v.(type) {
		case map[string]interface{}:
			if pie.Contains(AttrsWithInnerObj, k) {
				old[k] = v
			} else {
				switch oldValueCasted := old[k].(type) {
				case map[string]interface{}:
					err := updateOldObjWithPatch(oldValueCasted, patchValueCasted)
					if err != nil {
						return err
					}
				default:
					old[k] = v
				}
			}
		default:
			if k == "filter" && strings.HasPrefix(v.(string), "&") {
				v = "(" + old["filter"].(string) + ") " + v.(string)
			}
			old[k] = v
		}
	}

	return nil
}

// SwapEntity: use id to remove object from deleteEnt and then use data to create it in createEnt.
// Propagates id changes to children objects. For atomicity, all is done in a Mongo transaction.
func SwapEntity(createEnt, deleteEnt, id string, data map[string]interface{}, userRoles map[string]Role) *u.Error {
	if err := prepareCreateEntity(u.EntityStrToInt(createEnt), data, userRoles); err != nil {
		return err
	}

	_, err := WithTransaction(func(ctx mongo.SessionContext) (any, error) {
		// Create
		if _, err := repository.CreateObject(ctx, createEnt, data); err != nil {
			return nil, err
		}

		// Propagate
		if err := repository.PropagateParentIdChange(ctx, id, data["id"].(string),
			u.EntityStrToInt(data["category"].(string))); err != nil {
			return nil, err
		}

		// Delete
		if c, err := repository.GetDB().Collection(deleteEnt).DeleteOne(ctx, bson.M{"id": id}); err != nil {
			return nil, err
		} else if c.DeletedCount == 0 {
			return nil, errors.New("Error deleting object: not found")
		}

		return nil, nil
	})

	return err
}
