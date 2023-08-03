package services

import (
	"go-api/handlers"
	"os"
	driver "github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
)


func DBMiddleware(db driver.Database,addr string) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Set("database", &db)
		c.Set("addr", &addr)
		c.Next()
	}
}

func InitRouter(db driver.Database,addr string) *gin.Engine {
	env := os.Getenv("ENV")
	if env =="production"{
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()

	

	router.Use(DBMiddleware(db,addr))

	router.GET("/api/v1//Devices", handlers.GetDevices)
	router.POST("/api/v1//Devices", handlers.PostDevices)
	router.DELETE("/api/v1//Devices/:key", handlers.DeleteDevice)

	router.GET("/api/v1//Connections", handlers.GetConnection)
	router.POST("/api/v1//Connections", handlers.PostConnection)
	router.DELETE("/api/v1//Connections/:key", handlers.DeleteConnection)

	router.GET("/api/v1/Database", handlers.GetBDD)
	router.POST("/api/v1/Database", handlers.ConnectBDD)

	swagger := handlers.SwaggerHandler()
	router.Use(gin.WrapH(swagger))

	return router
}
