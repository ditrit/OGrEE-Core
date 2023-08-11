package services

import (
	"net/http"
	"ogree-bff/handlers"
	"ogree-bff/models"
	"ogree-bff/utils/token"

	//	driver "github.com/arangodb/go-driver"
	"github.com/gin-gonic/gin"
)

func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.TokenValid(c)
		if err != nil {
			c.IndentedJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func APIMiddleware(apiList []models.API) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, api := range apiList {
			c.Set(api.Name, api.URL)

		}
		c.Next()

	}

}
func initDevices(protected, unprotected *gin.RouterGroup) {

	protected.GET("/deviceComp/:entity", handlers.GetDevices)
	protected.GET("/deviceComp/:entity/ConnecteTo/:id", handlers.GetDevicesConnectedTo)
	protected.POST("/deviceComp/:entity", handlers.CreateDevices)
	protected.DELETE("/deviceComp/:entity/:id", handlers.DeleteDevice)

	protected.GET("/Connections", handlers.GetConnections)
	protected.POST("/Connections", handlers.CreateConnections)
	protected.DELETE("/Connections/:id", handlers.DeleteConnections)

	protected.GET("/deviceComp/:entity/:obj/:objAttr/:deviceAttr", handlers.GetDeviceBindingObject)
}

func initAuth(protected, unprotected *gin.RouterGroup) {

	protected.GET("/token/valid", handlers.ValidToken)
	protected.POST("/users/password/reset", handlers.ResetUserPassword)
	protected.POST("/users/password/change", handlers.ModifyUserPassword)

	unprotected.POST("/login", handlers.Login)
	unprotected.POST("/users/password/forgot", handlers.UserForgotPassword)
}

func initOrganization(protected, unprotected *gin.RouterGroup) {
	protected.POST("/users", handlers.CreateAccount)
	protected.POST("/users/bulk", handlers.CreateBulk)
	protected.GET("/users", handlers.GetAllAccounts)
	protected.DELETE("/users/:user", handlers.RemoveAccount)
	protected.PATCH("/users/:user", handlers.ModifyUserRoles)
	protected.POST("/domains/bulk", handlers.CreateBulkDomain)
	protected.GET("/hierarchy/domains", handlers.GetCompleteDomainHierarchy)
}

func initFlutterApp(protected, unprotected *gin.RouterGroup) {
	protected.POST("/projects", handlers.CreateProjects)
	protected.GET("/projects", handlers.GetProjects)
	protected.DELETE("/projects/*id", handlers.DeleteProjects)
	protected.PUT("/projects/*id", handlers.UpdateProjects)

}
func initAbout(protected, unprotected *gin.RouterGroup) {
	protected.GET("/stats", handlers.GetStats)
	protected.POST("/versions", handlers.GetAPIVersion)
}

func initObjects(protected, unprotected *gin.RouterGroup) {

	protected.GET("/:entity", handlers.GetAllEntities)
	protected.POST("/:entity", handlers.CreateObject)

	protected.GET("/objects/*hierarchyName", handlers.GetGenericObject)

	protected.GET("/:entity/*id", handlers.GetEntity)
	protected.DELETE("/:entity/*id", handlers.DeleteObject)
	protected.PATCH("/:entity/*id", handlers.PartialUpdateObject)
	protected.PUT("/:entity/*id", handlers.UpdateObject)

	protected.GET("/tempunits/*id", handlers.GetTempUnit)

	//protected.GET("/:entity/:id/:subent",handlers.GetEntitiesOfAncestor)

	protected.GET("/hierarchy", handlers.GetCompleteHierarchy)
	protected.GET("/hierarchy/attributes", handlers.GetCompleteHierarchyAttrs)

	//protected.GET("/:entity/:id/:HierarchalPath",handlers.GetEntitiesUsingNamesOfParents)

	protected.POST("/validate/*entity", handlers.ValidateObject)

}

func InitRouter(apiList []models.API, env string) *gin.Engine {
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	protected := router.Group("/api")
	protected.Use(JwtAuthMiddleware())
	protected.Use(APIMiddleware(apiList))

	unprotected := router.Group("/api")
	unprotected.Use(APIMiddleware(apiList))

	initDevices(protected, unprotected)
	initAuth(protected, unprotected)
	initOrganization(protected, unprotected)
	initObjects(protected, unprotected)
	initAbout(protected, unprotected)

	//init healthcheck route
	unprotected.GET("/health", func(c *gin.Context) {
		c.String(http.StatusAccepted, "")
	})

	swagger := handlers.SwaggerHandler()
	router.Use(gin.WrapH(swagger))

	return router
}
