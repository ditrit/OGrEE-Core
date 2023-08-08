package handlers

import (
	"net/http"
	"arango-api/models"
	"arango-api/controllers"
  	"github.com/gin-gonic/gin"
)
// swagger:operation POST /login Login LoginToApi
// Login to api
// ---
// responses:
//   '200':
//     description: successful
//     schema:
//       items:
//         "$ref": "#/definitions/SuccessLogin"
//   '500':
//     description: Error
//     schema:
//       items:
//         "$ref": "#/definitions/ErrorResponse"
func Login(c *gin.Context){
	var input models.LoginInput

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := controllers.CheckLogin(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username or password is incorrect."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token":token})

}