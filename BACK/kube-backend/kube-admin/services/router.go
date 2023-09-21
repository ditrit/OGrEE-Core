package services
import (
	"github.com/gin-gonic/gin"
	"kube-admin/auth"
	"kube-admin/handlers"
	"net/http"
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
func InitRouter() *gin.Engine{
	router := gin.Default()
	router.Use(CORSMiddleware())

	protected := router.Group("/api")
	protected.Use(auth.JwtAuthMiddleware()) // protected

	unprotected := router.Group("/api")
	unprotected.POST("/login", handlers.Login) // public endpoint
	//init healthcheck route
	unprotected.GET("/health", func(c *gin.Context) {
		c.String(http.StatusAccepted, "")
	})

	protected.GET("/apps", handlers.GetAllApps)
	protected.GET("/tenants", handlers.GetTenants)
	protected.GET("/tenants/:name", handlers.GetTenantPodsInfo)
	protected.DELETE("/tenants/:name", handlers.RemoveTenant)
	protected.POST("/tenants", handlers.AddTenant)
	protected.POST("/tenants/:name/logo", handlers.AddTenantLogo)
	protected.PUT("/tenants/:name", handlers.UpdateTenants)
	protected.POST("/tenants/:name/backup", handlers.BackupTenantDB)
	protected.GET("/containers/:name", handlers.GetContainerLogs)

	// protected.POST("/servers", handlers.CreateNewBackend)
	
	protected.POST("/tools/netbox", handlers.CreateNetbox)
	protected.DELETE("/tools/netbox", handlers.RemoveNetbox)
	protected.POST("/tools/netbox/dump", handlers.AddNetboxDump)
	protected.POST("/tools/netbox/import", handlers.ImportNetboxDump)

	swagger := handlers.SwaggerHandler()
	router.Use(gin.WrapH(swagger))

	return router
}