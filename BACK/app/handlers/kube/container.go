package kube

import (
	"back-admin/services/k8s"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetContainerLogs(c *gin.Context) {
	name := c.Param("name")
	ns := strings.Split(name, "_")[0]
	podName := strings.Split(name, "_")[1]

	if response, err := k8s.GetContainerLogs(ns, podName); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"logs": response})
	}
	return
}
