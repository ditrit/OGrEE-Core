package kube

import (
	"back-admin/models"
	"back-admin/services/k8s"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// swagger:operation GET /apps APP GetAllApps
// Get AllApps
// ---
// produces:
// - application/json
// security:
//   - Bearer: []
//
// responses:
//
//	'200':
//	    description: ok
//	'400':
//	    description: Bad request
//	'500':
//	    description: Internal server error
func GetAllApps(c *gin.Context) {

	ns, err := k8s.GetNamespace()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	var response Response
	for _, n := range ns {
		var tenant models.Tenant
		tenant.Name = n
		pods, err := k8s.GetPods("ogree-" + n)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		tenant.HasWeb = false
		for _, pod := range pods {
			if pod.Name == "ogree-"+n+"_app" {
				tenant.HasWeb = true
				tenant.WebUrl = "app." + n + "." + os.Getenv("HOST")
				tenant.WebPort = "80"
			}
		}
		tenant.ApiUrl = "api." + n + "." + os.Getenv("HOST")
		tenant.ApiPort = "80"

		response.Tenants = append(response.Tenants, tenant)

	}
	response.Tools = []models.ContainerInfo{}
	if k8s.NetboxCreated() {
		if pods, err := k8s.GetNetbox(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		} else {
			pods.Ports = "dcim." + os.Getenv("HOST") + ":80"
			response.Tools = append(response.Tools, pods)
		}

	}
	c.IndentedJSON(http.StatusOK, response)
}

type Response struct {
	Tenants []models.Tenant        `json:"tenants"`
	Tools   []models.ContainerInfo `json:"tools"`
}
