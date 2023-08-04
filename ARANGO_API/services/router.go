package services

import (
	"arango-api/handlers"
	"os"
	"arango-api/utils/token"
	"net/http"
	driver "github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
)

func DBMiddleware(db driver.Database, addr string) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Set("database", &db)
		c.Set("addr", &addr)
		c.Next()
	}
}

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

func InitRouter(db driver.Database, addr string) *gin.Engine {
	env := os.Getenv("ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(DBMiddleware(db, addr))
	proteted := router.Group("/api/v1")
	proteted.Use(JwtAuthMiddleware())
	proteted.GET("/Devices", handlers.GetDevices)
	proteted.POST("/Devices", handlers.PostDevices)
	proteted.DELETE("/Devices/:key", handlers.DeleteDevice)
	proteted.GET("/Devices/:key/Connected", handlers.GetDevicesConnectedTo)

	proteted.GET("/Connections", handlers.GetConnection)
	proteted.POST("/Connections", handlers.PostConnection)
	proteted.DELETE("/Connections/:key", handlers.DeleteConnection)

	proteted.GET("/Database", handlers.GetBDD)
	proteted.POST("/Database", handlers.ConnectBDD)

	router.POST("/api/v1/Login",handlers.Login)
	router.GET("/api/v1/health",func(c *gin.Context){
		c.String(http.StatusAccepted,"")
	})

	swagger := handlers.SwaggerHandler()
	router.Use(gin.WrapH(swagger))

	return router
}
