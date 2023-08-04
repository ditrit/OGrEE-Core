package database

import (
	//	driver "github.com/arangodb/go-driver"
	"arango-api/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InsertConnection(c *gin.Context, conn map[string]string) ([]interface{}, *models.ErrorMessage) {
	db, err := GetDBConn(c)
	if err != nil {
		return nil, err
	}

	// check if devices existed
	existed, err := DeviceExistedById(*db, conn["_from"])
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, &models.ErrorMessage{StatusCode: http.StatusNotFound, Message: "Device " + conn["_from"] + " not found"}
	}
	existed, err = DeviceExistedById(*db, conn["_to"])
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, &models.ErrorMessage{StatusCode: http.StatusNotFound, Message: "Device " + conn["_to"] + " not found"}
	}
	connStr, err := ParseToString(conn)

	if err != nil {
		return nil, err
	}
	querystring := "INSERT " + connStr + " INTO links RETURN NEW"

	result, err := ExecQuerry(*db, querystring)
	if err != nil {
		return nil, err
	}
	return result, nil
}
