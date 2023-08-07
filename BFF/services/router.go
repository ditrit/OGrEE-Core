package services

import (
	"ogree-bff/handlers"
	"ogree-bff/utils/token"
	"net/http"
	"ogree-bff/models"

//	driver "github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
)



func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.TokenValid(c)
		if err != nil {
			c.IndentedJSON(http.StatusUnauthorized,gin.H{"message":"Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func InitRouter(apiList []models.API,env string ) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	proteted := router.Group("/api/v1")
	proteted.Use(JwtAuthMiddleware())

	router.POST("/api/v1/Login",handlers.Login)
	router.GET("/api/v1/health",func(c *gin.Context){
		c.String(http.StatusAccepted,"")
	})

	swagger := handlers.SwaggerHandler()
	router.Use(gin.WrapH(swagger))

	return router
}
