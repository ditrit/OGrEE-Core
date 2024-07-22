package models

import (
	"encoding/json"
	"errors"
	"fmt"
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
	// Update timestamp requires, first, obj retrieval
	oldObj, err := GetObjectById(id, entityStr, u.RequestFilters{}, userRoles)
	if err != nil {
		return nil, err
	} else if entityStr == u.HIERARCHYOBJS_ENT {
		// overwrite category
		entityStr = oldObj["category"].(string)
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
		println("is PATCH")
		if patchData, err := preparePatch(tagsPresent, updateData, oldObj); err != nil {
			return nil, err
		} else {
			updateData = patchData
		}
	} else if tagsPresent {
		if err := verifyTagList(tags); err != nil {
			return nil, err
		}
	}

	fmt.Println(updateData)
	result, err := UpdateTransaction(entity, id, isRecursive, updateData, oldObj, userRoles)
	if err != nil {
		return nil, err
	}

	var updatedDoc map[string]interface{}
	result.(*mongo.SingleResult).Decode(&updatedDoc)

	return fixID(updatedDoc), nil
}

func UpdateTransaction(entity int, id string, isRecursive bool, updateData, oldObj map[string]any, userRoles map[string]Role) (any, *u.Error) {
	entityStr := u.EntityToString(entity)
	return WithTransaction(func(ctx mongo.SessionContext) (interface{}, error) {
		err := prepareUpdateObject(ctx, entity, id, updateData, oldObj, userRoles)
		if err != nil {
			return nil, err
		}

		idFilter := GetIdReqByEntity(entityStr, id)
		mongoRes := repository.GetDB().Collection(entityStr).FindOneAndReplace(
			ctx,
			idFilter, updateData,
			options.FindOneAndReplace().SetReturnDocument(options.After),
		)
		if mongoRes.Err() != nil {
			return nil, mongoRes.Err()
		}

		if err := propagateUpdateChanges(ctx, entity, oldObj, updateData, isRecursive); err != nil {
			return nil, err
		}

		return mongoRes, nil
	})
}

func preparePatch(tagsPresent bool, updateData, oldObj map[string]any) (map[string]any, *u.Error) {
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
	fmt.Println(updateData)
	return updateData, nil
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

func propagateUpdateChanges(ctx mongo.SessionContext, entity int, oldObj, updateData map[string]any, isRecursive bool) error {
	if oldObj["id"] != updateData["id"] {
		// Changes to id should be propagated
		if err := repository.PropagateParentIdChange(
			ctx,
			oldObj["id"].(string),
			updateData["id"].(string),
			entity,
		); err != nil {
			return err
		} else if entity == u.DOMAIN {
			if err := repository.PropagateDomainChange(ctx,
				oldObj["id"].(string),
				updateData["id"].(string),
			); err != nil {
				return err
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
				return err
			}
		} else {
			// Check if children domains are compatible
			if err := repository.CheckParentDomainChange(entity, updateData["id"].(string),
				updateData["domain"].(string)); err != nil {
				return err
			}
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
