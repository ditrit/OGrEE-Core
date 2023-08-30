package main

import (
	"flag"
	"net/http"
	"ogree_app_backend/auth"
	"os"
	"strconv"
	"text/template"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var tmplt *template.Template
var apptmplt *template.Template
var servertmplt *template.Template
var netboxtmplt *template.Template
var opendcimtmplt *template.Template
var DEPLOY_DIR string
var DOCKER_DIR string

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}
	DEPLOY_DIR = os.Getenv("DEPLOY_DIR")
	if DEPLOY_DIR == "" {
		DEPLOY_DIR = "../../deploy/"
	}
	DOCKER_DIR = DEPLOY_DIR + "docker/"
	// hashedPassword, _ := bcrypt.GenerateFromPassword(
	// 	[]byte("password"), bcrypt.DefaultCost)
	// println(string(hashedPassword))
	tmplt = template.Must(template.ParseFiles("backend-assets/docker-env-template.txt"))
	apptmplt = template.Must(template.ParseFiles("flutter-assets/flutter-env-template.txt"))
	servertmplt = template.Must(template.ParseFiles("backend-assets/template.service"))
	netboxtmplt = template.Must(template.ParseFiles("tools-assets/netbox-docker-template.txt"))
	opendcimtmplt = template.Must(template.ParseFiles("tools-assets/opendcim-env-template.txt"))
}

func main() {
	port := flag.Int("port", 8081, "an int")
	flag.Parse()
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"X-Requested-With", "Content-Type", "Authorization", "Origin", "Accept"}
	router.Use(cors.New(corsConfig))

	router.POST("/api/login", login) // public endpoint

	router.Use(auth.JwtAuthMiddleware()) // protected
	router.GET("/api/apps", getAllApps)
	// tenants
	router.GET("/api/tenants", getTenants)
	router.GET("/api/tenants/:name", getTenantDockerInfo)
	router.DELETE("/api/tenants/:name", removeTenant)
	router.POST("/api/tenants", addTenant)
	router.POST("/api/tenants/:name/logo", addTenantLogo)
	router.PUT("/api/tenants/:name", updateTenant)
	router.POST("/api/tenants/:name/backup", backupTenantDB)
	router.GET("/api/containers/:name", getContainerLogs)
	// server
	router.POST("/api/servers", createNewBackend)
	// netbox
	router.POST("/api/tools/netbox", createNetbox)
	router.DELETE("/api/tools/netbox", removeNetbox)
	router.POST("/api/tools/netbox/dump", addNetboxDump)
	router.POST("/api/tools/netbox/import", importNetboxDump)
	// opendcim
	router.POST("/api/tools/opendcim", createOpenDcim)
	router.DELETE("/api/tools/opendcim", removeOpenDcim)

	router.Run(":" + strconv.Itoa(*port))

}

func login(c *gin.Context) {
	var userIn user
	if err := c.BindJSON(&userIn); err != nil {
		println("ERROR:")
		println(err.Error())
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	} else {
		// Check credentials
		if userIn.Email != "admin" ||
			bcrypt.CompareHashAndPassword([]byte(os.Getenv("ADM_PASSWORD")), []byte(userIn.Password)) != nil {
			println("Credentials error")
			c.IndentedJSON(http.StatusForbidden, gin.H{"error": "Invalid credentials"})
			return
		}

		println("Generate")
		// Generate token
		token, err := auth.GenerateToken(userIn.Email)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Respond
		response := make(map[string]map[string]string)
		response["account"] = make(map[string]string)
		response["account"]["Email"] = userIn.Email
		response["account"]["token"] = token
		response["account"]["isTenant"] = "true"
		c.IndentedJSON(http.StatusOK, response)
	}
}

func getAllApps(c *gin.Context) {
	response := make(map[string]interface{})
	response["tenants"] = getTenantsFromJSON()
	response["tools"] = []container{}
	if netbox, err := getDockerInfo("netbox"); err != nil {
		println(err.Error())
	} else {
		response["tools"] = netbox
	}
	if opendcim, err := getDockerInfo("opendcim"); err != nil {
		println(err.Error())
	} else {
		response["tools"] = append(response["tools"].([]container), opendcim...)
	}
	c.IndentedJSON(http.StatusOK, response)
}
