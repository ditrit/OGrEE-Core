package models

import (
	"p3/repository"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

func WithTransaction(callback func(mongo.SessionContext) (any, error)) (interface{}, *u.Error) {
	ctx, cancel := u.Connect()
	defer cancel()

	// Start a session and run the callback to update db
	session, err := repository.GetClient().StartSession()
	if err != nil {
		return nil, &u.Error{Type: u.ErrDBError, Message: "Unable to start session: " + err.Error()}
	}
	defer session.EndSession(ctx)

	result, err := session.WithTransaction(ctx, callback)
	if err != nil {
		if errCasted, ok := err.(*u.Error); ok {
			if errCasted != nil {
				return nil, errCasted
			}

			return result, nil
		}

		return nil, &u.Error{Type: u.ErrDBError, Message: "Unable to complete transaction: " + err.Error()}
	}

	return result, nil
}
