package models

import (
	"p3/repository"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

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
