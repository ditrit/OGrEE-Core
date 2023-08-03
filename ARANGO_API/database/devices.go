package database

import (
	driver "github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
	"go-api/models"
	"net/http"

)


func DeviceExistedById(db driver.Database,id string) (bool, *models.ErrorMessage) {

	querystring := "FOR devices IN devices "
	querystring += "FILTER devices.`_id` == \""+id+"\" "
	querystring += "RETURN devices"
	
	created, err := ExecQuerry(db, querystring);
	if err != nil {
		return false, err
	}
	if len(created)!=0{
		return true, nil
	}
	return false,nil
}
func DeviceExisted(db driver.Database,device map[string]string) (bool, *models.ErrorMessage) {

	querystring := "FOR devices IN devices "
	querystring += "FILTER devices.`_name` == \""+device["_name"]+"\" "
	querystring += "&& devices.`group_name` == \""+ device["group_name"]+"\" "
	querystring += "&& devices.`created` == \""+ device["created"]+"\" "
	querystring += "RETURN devices"
	
	created, err := ExecQuerry(db, querystring);
	if err != nil {
		return false, err
	}
	if len(created)!=0{
		return true, nil
	}
	return false,nil
}

func InsertDevices(c *gin.Context, device map[string]string) ([]interface{}, *models.ErrorMessage) {
	db, err := GetDBConn(c)
	if err != nil {
		return nil, err
	}

	// check if devices existed
	existed,err := DeviceExisted(*db, device)
	if err != nil {
		return nil, err
	}
	if existed{
		return nil,&models.ErrorMessage{StatusCode: http.StatusFound,Message:"Device already existed"}
	}
	deviceStr,err := ParseToString(device)

	
	if err != nil {
        return nil, err
    }
	querystring := "INSERT "+deviceStr+" INTO devices RETURN NEW"

	result, err := ExecQuerry(*db, querystring);
	if err != nil {
		return nil, err
	}
	return result,nil
}
