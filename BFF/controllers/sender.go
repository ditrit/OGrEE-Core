package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ogree-bff/models"
	"ogree-bff/utils/token"

	"github.com/gin-gonic/gin"
)

func GetQueryString(c *gin.Context) (string) {
	query := ""
	for key,value := range c.Request.URL.Query() {
		if query == ""{
			query+="?"+key+"="+value[0]
		}else{
			query+="&"+key+"="+value[0]
		}
	}
	return query
}

func GetPath(c *gin.Context) (string) {
	return c.Request.URL.Path
}

func Send(method, URL, query ,key string, data interface{}) (*http.Response,error) {
	client := &http.Client{}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil,err
	}
	req, err := http.NewRequest(method, URL+query, bytes.NewBuffer(dataJSON))
	if err != nil {
		return nil,err
	}
	if(key != ""){
		req.Header.Set("Authorization", "Bearer "+key)
	}

	return client.Do(req)
}

func GetJSONBody(resp *http.Response) models.Message {
	defer resp.Body.Close()
	var responseBody interface{}
	body, err := io.ReadAll(resp.Body)
	
	if err != nil {
		return models.Message{StatusCode: http.StatusInternalServerError,Message: err.Error()}
	}
	err = json.Unmarshal(body, &responseBody)
  	if err != nil {
    	return models.Message{StatusCode: http.StatusInternalServerError,Message: err.Error()}
  	}
	return models.Message{StatusCode: http.StatusAccepted,Message: responseBody}
}

func Get(c *gin.Context, api string){
	url, ok := c.Value(api).(string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get api connection"})
		return 
	}
	key := token.ExtractToken(c)
	query := GetQueryString(c)
	path := GetPath(c)
	resp,err := Send("GET",url+path,query,key,nil)
	if err != nil && resp == nil{
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(resp)
	c.IndentedJSON(resp.StatusCode, result.Message)
	return
}

func Post(c *gin.Context, api string){
	url, ok := c.Value(api).(string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get api connection"})
		return 
	}
	var data interface{}
	// Call BindJSON to bind the received JSON to
	if err := c.ShouldBindJSON(&data); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	
	key := token.ExtractToken(c)
	path := GetPath(c)
	resp,err := Send("POST",url+path,"",key,data)
	fmt.Println(err)
	if err != nil && resp == nil{
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(resp)
	c.IndentedJSON(resp.StatusCode, result.Message)
}

func Delete(c *gin.Context, api string){
	url, ok := c.Value(api).(string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get api connection"})
		return 
	}
	key := token.ExtractToken(c)
	path := GetPath(c)
	resp,err := Send("DELETE",url+path,"",key,nil)
	if err != nil && resp == nil{
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(resp)
	c.IndentedJSON(resp.StatusCode, result.Message)
}

func Patch(c *gin.Context, api string){
	url, ok := c.Value(api).(string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get api connection"})
		return 
	}
	var data interface{}
	// Call BindJSON to bind the received JSON to
	if err := c.ShouldBindJSON(&data); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	key := token.ExtractToken(c)
	path := GetPath(c)
	resp,err := Send("PATCH",url+path,"",key,data)
	if err != nil && resp == nil{
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(resp)
	c.IndentedJSON(resp.StatusCode, result.Message)
}

func Put(c *gin.Context, api string){
	url, ok := c.Value(api).(string)
	if !ok {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Failed to get api connection"})
		return 
	}
	var data interface{}
	// Call BindJSON to bind the received JSON to
	if err := c.ShouldBindJSON(&data); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	key := token.ExtractToken(c)
	path := GetPath(c)
	resp,err := Send("PUT",url+path,"",key,data)
	if err != nil  && resp == nil{
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	result := GetJSONBody(resp)
	c.IndentedJSON(resp.StatusCode, result.Message)
}