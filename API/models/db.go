package models

import (
	"context"
	"fmt"
	"p3/repository"
	u "p3/utils"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CommandRunner(cmd interface{}) *mongo.SingleResult {
	ctx, cancel := u.Connect()
	result := repository.GetDB().RunCommand(ctx, cmd, nil)
	defer cancel()
	return result
}

func GetDBName() string {
	name := repository.GetDB().Name()

	//Remove the preceding 'ogree' at beginning of name
	if strings.Index(name, "ogree") == 0 {
		name = name[5:] //5=len('ogree')
	}
	return name
}

func ExtractCursor(c *mongo.Cursor, ctx context.Context, entity int, userRoles map[string]Role) ([]map[string]interface{}, error) {
	ans := []map[string]interface{}{}
	for c.Next(ctx) {
		x := map[string]interface{}{}
		err := c.Decode(x)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		// Remove _id
		x = fixID(x)
		// Check permissions
		if u.IsEntityHierarchical(entity) && userRoles != nil {
			permission := CheckUserPermissionsWithObject(userRoles, entity, x)
			if permission == READONLYNAME {
				x = FixReadOnlyName(x)
			}
			if permission >= READONLYNAME {
				ans = append(ans, x)
			}
		} else {
			ans = append(ans, x)
		}
	}
	return ans, nil
}

func WithTransaction[T any](callback func(mongo.SessionContext) (T, error)) (T, *u.Error) {
	ctx, cancel := u.Connect()
	defer cancel()

	var nilT T

	// Start a session and run the callback to update db
	session, err := repository.GetClient().StartSession()
	if err != nil {
		return nilT, &u.Error{Type: u.ErrDBError, Message: "Unable to start session: " + err.Error()}
	}
	defer session.EndSession(ctx)

	callbackWrapper := func(ctx mongo.SessionContext) (any, error) {
		result, err := callback(ctx)
		if err != nil {
			// support returning u.Error even if nil
			if errCasted, ok := err.(*u.Error); ok {
				if errCasted != nil {
					return nilT, errCasted // u.Error not nil -> return u.Error not nil
				}

				return result, nil // u.Error nil -> return error nil
			}

			return result, err // error not nil -> return error not nil
		}

		return result, nil // error nil -> return error nil
	}

	result, err := session.WithTransaction(ctx, callbackWrapper)
	if err != nil {
		if errCasted, ok := err.(*u.Error); ok {
			return nilT, errCasted
		}

		return nilT, &u.Error{Type: u.ErrDBError, Message: "Unable to complete transaction: " + err.Error()}
	}

	return castResult[T](result), nil
}

func castResult[T any](result any) T {
	if result == nil {
		var nilT T
		return nilT
	}

	return result.(T)
}

func GetIdReqByEntity(entityStr, id string) primitive.M {
	var idFilter primitive.M
	if u.IsEntityNonHierarchical(u.EntityStrToInt(entityStr)) {
		idFilter = bson.M{"slug": id}
	} else {
		idFilter = bson.M{"id": id}
	}
	return idFilter
}
