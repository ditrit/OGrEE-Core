package kube

import (
	"kube-admin/models"
	"kube-admin/services/k8s"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

func CreateNetbox(c *gin.Context) {
	var newNetbox models.Netbox
	if err := c.BindJSON(&newNetbox); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	if err := k8s.CreateNamespace("netbox-ogree"); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	host := "dcim." + os.Getenv("HOST")
	command := "helm install netbox"
	command += " --set image.tag=v3.5-2.6.1"
	command += " --set superuser.name=" + newNetbox.Username
	command += " --set superuser.password=" + newNetbox.Password
	command += " --set postgresql.image.tag=15.4.0"
	command += " --set service.type=NodePort"
	command += " --set 'csrf.trustedOrigins={https://" + host + "+,http://" + host + "}'"
	command += " --set cors.originAllowAll=true"
	command += " bootc/netbox -n netbox-ogree"

	cmd := exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Create ingress connection
	command = "helm install netbox-ing netbox/netbox --set host=" + host + " -n netbox-ogree"
	cmd = exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "Netbox Created at netbox." + os.Getenv("HOST")})

}

func RemoveNetbox(c *gin.Context) {
	if err := k8s.DeleteNamespace("netbox-ogree"); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"success": "Successfully removed Netbox"})
	}
}

func AddNetboxDump(c *gin.Context) {
	// Load file
	formFile, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	// Save file
	err = c.SaveUploadedFile(formFile, "netbox/dump.sql")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	c.String(http.StatusOK, "")
}

func ImportNetboxDump(c *gin.Context) {

	command := "kubectl cp ./netbox/drop.sql netbox-ogree/netbox-postgresql-0:/tmp/."
	cmd := exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	command = "kubectl cp ./netbox/dump.sql netbox-ogree/netbox-postgresql-0:/tmp/."
	cmd = exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	command = "kubectl scale --replicas=0 -n netbox-ogree deployment/netbox "
	cmd = exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	command = "kubectl exec -n netbox-ogree netbox-postgresql-0 -- /bin/sh -c 'PGPASSWORD=$POSTGRES_PASSWORD psql -U netbox -d postgres < tmp/drop.sql'"
	cmd = exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	command = "kubectl exec -n netbox-ogree netbox-postgresql-0 -- /bin/sh -c  'PGPASSWORD=$POSTGRES_PASSWORD psql -U netbox -d netbox < tmp/dump.sql'"
	cmd = exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	command = "kubectl scale --replicas=1 -n netbox-ogree deployment/netbox "
	cmd = exec.Command("bash", "-c", command)
	if err := cmd.Run(); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.String(http.StatusOK, "")

}
