package controllers

import (

	"net/http"
	"ogree-bff/utils/token"

	"github.com/gin-gonic/gin"
)



func DeviceBindingObject(c *gin.Context,database string) {

	objAttr:= c.Param("objAttr")
	
	obj := c.Param("obj")

	//MONGO Check
	mongoURL, ok := c.Value("mongo").(string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get api connection"})
		return 
	}
	key := token.ExtractToken(c)
	mongoResp,err := Send("GET",mongoURL+"/api/objects/"+obj,"",key,nil)
	if err != nil {
		if mongoResp != nil {
			result := GetJSONBody(mongoResp)
			c.IndentedJSON(mongoResp.StatusCode,result.Message)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	mongoBody:=GetJSONBody(mongoResp)
	if mongoBody.StatusCode != http.StatusOK {
		c.IndentedJSON(mongoBody.StatusCode,mongoBody.Message)
		return
	}
	mongoDataResult := mongoBody.Message.(map[string]interface{})
	mongoResult := mongoDataResult["data"].(map[string]interface{})
	if mongoResult["category"] != "device"{
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message":obj +" is not a device"})
		return
	}
	mongoAttr := mongoResult["attributes"].(map[string]interface{})
	if mongoAttr[objAttr] == nil  && mongoResult[objAttr] == nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message":obj +" has not attributes :" + objAttr})
		return
	}
	var attributeNeed string
	if mongoResult[objAttr] != nil {
		attributeNeed = mongoResult[objAttr].(string)
	}else{
		attributeNeed = mongoAttr[objAttr].(string)
	}

	//ARANGO
	deviceAttr:= c.Param("deviceAttr")
	deviceURL, ok := c.Value(database).(string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get api connection"})
		return 
	}
	query := GetQueryString(c)
	if query == "" {
		query +="?"+deviceAttr+"="+attributeNeed
	}else{
		query +="&"+deviceAttr+"="+attributeNeed
	}

	deviceResp,err := Send("GET",deviceURL+"/api/devices",query,key,nil)
	if err != nil {
		if deviceResp != nil {
			result := GetJSONBody(deviceResp)
			c.IndentedJSON(deviceResp.StatusCode,result.Message)
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(deviceResp)
	c.IndentedJSON(deviceResp.StatusCode, result.Message)
	return


	
}