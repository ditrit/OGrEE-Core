package services

import (
	"back-admin/auth"
	"back-admin/handlers/docker"
	"back-admin/handlers/kube"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT,DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func InitRouter(isKube bool) *gin.Engine {
	router := gin.Default()
	router.Use(CORSMiddleware())

	unprotected := router.Group("/api")
	unprotected.POST("/login", kube.Login) // public endpoint
	unprotected.GET("/health", func(c *gin.Context) {
		c.String(http.StatusAccepted, "")
	})
	unprotected.GET("/version", func(c *gin.Context) {
		response := map[string]any{"version": "1.0", "isKubernetes": isKube}
		c.IndentedJSON(http.StatusOK, response)
	})

	protected := router.Group("/api")
	protected.Use(auth.JwtAuthMiddleware()) // protected

	if isKube {
		protected.GET("/apps", kube.GetAllApps)
		protected.GET("/tenants", kube.GetTenants)
		protected.GET("/tenants/:name", kube.GetTenantPodsInfo)
		protected.DELETE("/tenants/:name", kube.RemoveTenant)
		protected.POST("/tenants", kube.AddTenant)
		protected.POST("/tenants/:name/logo", kube.AddTenantLogo)
		protected.PUT("/tenants/:name", kube.UpdateTenants)
		protected.POST("/tenants/:name/backup", kube.BackupTenantDB)
		protected.GET("/containers/:name", kube.GetContainerLogs)

		// protected.POST("/servers", handlers.CreateNewBackend)

		protected.POST("/tools/netbox", kube.CreateNetbox)
		protected.DELETE("/tools/netbox", kube.RemoveNetbox)
		protected.POST("/tools/netbox/dump", kube.AddNetboxDump)
		protected.POST("/tools/netbox/import", kube.ImportNetboxDump)
	} else {
		protected.GET("/apps", docker.GetAllApps)
		protected.GET("/tenants", docker.GetTenants)
		protected.GET("/tenants/:name", docker.GetTenantDockerInfo)
		protected.DELETE("/tenants/:name", docker.RemoveTenant)
		protected.POST("/tenants", docker.AddTenant)
		protected.POST("/tenants/:name/logo", docker.AddTenantLogo)
		protected.PUT("/tenants/:name", docker.UpdateTenant)
		protected.POST("/tenants/:name/backup", docker.BackupTenantDB)
		protected.POST("/tenants/:name/stop", docker.StopStartTentant)
		protected.POST("/tenants/:name/start", docker.StopStartTentant)
		protected.GET("/containers/:name", docker.GetContainerLogs)
		protected.POST("/servers", docker.CreateNewBackend)
		protected.POST("/tools/netbox", docker.CreateNetbox)
		protected.DELETE("/tools/netbox", docker.RemoveNetbox)
		protected.POST("/tools/netbox/dump", docker.AddNetboxDump)
		protected.POST("/tools/netbox/import", docker.ImportNetboxDump)
		protected.POST("/tools/opendcim", docker.CreateOpenDcim)
		protected.DELETE("/tools/opendcim", docker.RemoveOpenDcim)
		protected.POST("/tools/nautobot", docker.CreateNautobot)
		protected.DELETE("/tools/nautobot", docker.RemoveNautobot)
	}

	swagger := kube.SwaggerHandler()
	router.Use(gin.WrapH(swagger))

	return router
}
