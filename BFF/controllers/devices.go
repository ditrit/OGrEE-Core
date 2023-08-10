package controllers

import (
	"fmt"
	"net/http"
	"ogree-bff/utils/token"

	"github.com/gin-gonic/gin"
)

func DeviceBindingObject(c *gin.Context) {
	entity := c.Param("entity")

	deviceURL, ok := c.Value(entity).(string)
	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": entity + " has not database"})
		return
	}
	objAttr := c.Param("objAttr")

	obj := c.Param("obj")

	//MONGO Check
	mongoURL, ok := c.Value("objects").(string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get api connection"})
		return
	}
	key := token.ExtractToken(c)
	mongoResp, err := Send("GET", mongoURL+"/api/objects/"+obj, "", key, nil)
	if err != nil {
		if mongoResp != nil {
			result := GetJSONBody(mongoResp)
			c.IndentedJSON(mongoResp.StatusCode, result.Message)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	mongoBody := GetJSONBody(mongoResp)
	if mongoBody.StatusCode != http.StatusOK {
		c.IndentedJSON(mongoBody.StatusCode, mongoBody.Message)
		return
	}
	mongoDataResult := mongoBody.Message.(map[string]interface{})
	mongoResult := mongoDataResult["data"].(map[string]interface{})
	if mongoResult["category"] != "device" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": obj + " is not a device"})
		return
	}
	mongoAttr := mongoResult["attributes"].(map[string]interface{})
	if mongoAttr[objAttr] == nil && mongoResult[objAttr] == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": obj + " has not attributes :" + objAttr})
		return
	}
	var attributeNeed string
	if mongoResult[objAttr] != nil {
		attributeNeed = mongoResult[objAttr].(string)
	} else {
		attributeNeed = mongoAttr[objAttr].(string)
	}

	//ARANGO
	deviceAttr := c.Param("deviceAttr")

	query := GetQueryString(c)
	if query == "" {
		query += "?" + deviceAttr + "=" + attributeNeed
	} else {
		query += "&" + deviceAttr + "=" + attributeNeed
	}

	deviceResp, err := Send("GET", deviceURL+"/api/devices", query, key, nil)
	if err != nil {
		if deviceResp != nil {
			result := GetJSONBody(deviceResp)
			c.IndentedJSON(deviceResp.StatusCode, result.Message)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(deviceResp)
	c.IndentedJSON(deviceResp.StatusCode, result.Message)
	return

}


func GetDevice(c *gin.Context,url,methode string){

	key := token.ExtractToken(c)
	query := GetQueryString(c)
	fmt.Println(c.Request.URL.Query(),query)
	deviceResp, err := Send(methode, url, query, key, nil)
	if err != nil {
		if deviceResp != nil {
			result := GetJSONBody(deviceResp)
			c.IndentedJSON(deviceResp.StatusCode, result.Message)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(deviceResp)
	c.IndentedJSON(deviceResp.StatusCode, result.Message)
	return
}

func PostDevice(c *gin.Context,url, method string){

	var data interface{}
	// Call BindJSON to bind the received JSON to
	if err := c.ShouldBindJSON(&data); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	
	key := token.ExtractToken(c)
	resp,err := Send(method,url,"",key,data)
	fmt.Println(err)
	if err != nil && resp == nil{
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(resp)
	c.IndentedJSON(resp.StatusCode, result.Message)
}
