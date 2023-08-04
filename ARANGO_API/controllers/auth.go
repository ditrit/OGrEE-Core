package controllers

import (
	"arango-api/models"
	"arango-api/utils/token"
	"fmt"
	"os"
)
func CheckLogin(user models.LoginInput) (string,error){
	var u models.User
	u.Username = user.Username
	u.Password = user.Password
	apiUser := os.Getenv("API_USER")
	apiPassword := os.Getenv("API_PASSWORD")
	if apiUser == u.Username && apiPassword == u.Password{
		token,err := token.GenerateToken(1235)
		if err != nil {
			fmt.Println(err)
			return "",err
		}
		return token,nil
	}
	return "", fmt.Errorf("Bad username or password");

}