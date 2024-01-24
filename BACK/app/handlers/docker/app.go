package docker

import (
	"back-admin/models"
	"net/http"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
)

var tmplt *template.Template
var apptmplt *template.Template
var servertmplt *template.Template
var netboxtmplt *template.Template
var opendcimtmplt *template.Template
var DEPLOY_DIR string
var DOCKER_DIR string

func init() {
	DEPLOY_DIR = os.Getenv("DEPLOY_DIR")
	if DEPLOY_DIR == "" {
		DEPLOY_DIR = "../../deploy/"
	}
	DOCKER_DIR = DEPLOY_DIR + "docker/"
	// hashedPassword, _ := bcrypt.GenerateFromPassword(
	// 	[]byte("password"), bcrypt.DefaultCost)
	// println(string(hashedPassword))
	tmpltPrefixPath := "handlers/docker/"
	tmplt = template.Must(template.ParseFiles(tmpltPrefixPath + "backend-assets/docker-env-template.txt"))
	apptmplt = template.Must(template.ParseFiles(tmpltPrefixPath + "flutter-assets/flutter-env-template.txt"))
	servertmplt = template.Must(template.ParseFiles(tmpltPrefixPath + "backend-assets/template.service"))
	netboxtmplt = template.Must(template.ParseFiles(tmpltPrefixPath + "tools-assets/netbox-docker-template.txt"))
	opendcimtmplt = template.Must(template.ParseFiles(tmpltPrefixPath + "tools-assets/opendcim-env-template.txt"))
}

func GetAllApps(c *gin.Context) {
	response := make(map[string]interface{})
	response["tenants"] = getTenantsFromJSON()
	response["tools"] = []models.ContainerInfo{}
	if netbox, err := getDockerInfo("netbox"); err != nil {
		println(err.Error())
	} else {
		response["tools"] = netbox
	}
	if opendcim, err := getDockerInfo("opendcim"); err != nil {
		println(err.Error())
	} else {
		response["tools"] = append(response["tools"].([]models.ContainerInfo), opendcim...)
	}
	c.IndentedJSON(http.StatusOK, response)
}
