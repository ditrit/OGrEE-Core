package models

import (
	"p3/repository"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Alert struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

const WEB_ALERT = "web_alert"

// POST
func AddAlert(newAlert Alert) *u.Error {
	// Add the new alert
	ctx, cancel := u.Connect()
	_, err := repository.GetDB().Collection(WEB_ALERT).InsertOne(ctx, newAlert)
	if err != nil {
		println(err.Error())
		return &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}

	defer cancel()
	return nil
}

// GET
func GetAlerts() ([]Alert, *u.Error) {
	results := []Alert{}
	filter := bson.D{}
	ctx, cancel := u.Connect()
	cursor, err := repository.GetDB().Collection(WEB_ALERT).Find(ctx, filter)
	if err != nil {
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	} else if err = cursor.All(ctx, &results); err != nil {
		return nil, &u.Error{Type: u.ErrInternal, Message: err.Error()}
	}

	defer cancel()
	return results, nil
}

func GetAlert(id string) (Alert, *u.Error) {
	alert := &Alert{}
	filter := bson.M{"id": id}
	ctx, cancel := u.Connect()
	err := repository.GetDB().Collection(WEB_ALERT).FindOne(ctx, filter).Decode(alert)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return *alert, &u.Error{Type: u.ErrNotFound, Message: "Alert does not exist"}
		}
		return *alert, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}

	defer cancel()
	return *alert, nil
}

// DELETE
func DeleteAlert(alertId string) *u.Error {
	ctx, cancel := u.Connect()
	res, err := repository.GetDB().Collection(WEB_ALERT).DeleteOne(ctx, bson.M{"id": alertId})
	defer cancel()

	if err != nil {
		return &u.Error{Type: u.ErrDBError, Message: err.Error()}
	} else if res.DeletedCount <= 0 {
		return &u.Error{Type: u.ErrNotFound,
			Message: "Alert not found"}
	}
	return nil
}
