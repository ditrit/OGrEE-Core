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
	proteted := router.Group("/api")
	//proteted.Use(JwtAuthMiddleware())
	proteted.GET("/:devices", handlers.GetDevices)
	proteted.POST("/:devices", handlers.PostDevices)
	proteted.DELETE(":devices/:key", handlers.DeleteDevice)
	proteted.GET(":devices/ConnecteTo/:key", handlers.GetDevicesConnectedTo)

	proteted.GET("/Connections", handlers.GetConnection)
	proteted.POST("/Connections", handlers.PostConnection)
	proteted.DELETE("/Connections/:key", handlers.DeleteConnection)

	router.GET("/api/health",func(c *gin.Context){
		c.String(http.StatusAccepted,"")
	})

	swagger := handlers.SwaggerHandler()
	router.Use(gin.WrapH(swagger))

	return router
}
