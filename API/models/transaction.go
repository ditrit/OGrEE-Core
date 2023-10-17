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
		return callback(ctx)
	}

	result, err := session.WithTransaction(ctx, callbackWrapper)
	if err != nil {
		if errCasted, ok := err.(*u.Error); ok {
			if errCasted != nil {
				return nilT, errCasted
			}

			return castResult[T](result), nil
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
