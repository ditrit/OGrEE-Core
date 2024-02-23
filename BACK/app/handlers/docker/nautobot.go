package docker

import (
	"back-admin/models"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

// NAUTOBOT

func CreateNautobot(c *gin.Context) {
	var newNautobot models.Nautobot
	if err := c.BindJSON(&newNautobot); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Default values
	if newNautobot.Port == "" {
		newNautobot.Port = "8001"
	}

	nautobotDir := "nautobot-docker-compose"
	if _, err := os.Stat(nautobotDir); os.IsNotExist(err) {
		// Clone github repo
		println("Cloning nautobot git repo...")
		args := []string{"clone", "https://github.com/nautobot/nautobot-docker-compose.git"}
		cmd := exec.Command("git", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			c.IndentedJSON(http.StatusBadRequest, stderr.String())
			return
		}
	}

	// Modify docker compose file
	err := replaceTextInFile(nautobotDir+"/docker-compose.yml", []string{"8080:8080"}, []string{newNautobot.Port + ":8080"})
	if err != nil {
		panic(err)
	}

	// Create copy of .env file
	args := []string{"-f", "local.env.example", "local.env"}
	cmd := exec.Command("cp", args...)
	cmd.Dir = nautobotDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusBadRequest, stderr.String())
		return
	}

	// Modify .env file
	err = replaceTextInFile(nautobotDir+"/local.env",
		[]string{"NAUTOBOT_CREATE_SUPERUSER=false", "NAUTOBOT_SUPERUSER_NAME=admin", "NAUTOBOT_SUPERUSER_PASSWORD=admin"},
		[]string{"NAUTOBOT_CREATE_SUPERUSER=true", "NAUTOBOT_SUPERUSER_NAME=" + newNautobot.Username, "NAUTOBOT_SUPERUSER_PASSWORD=" + newNautobot.Password})
	if err != nil {
		panic(err)
	}

	println("Run docker (may take a long time...)")
	// Run docker
	args = []string{"compose", "up", "-d"}
	cmd = exec.Command("docker", args...)
	cmd.Dir = nautobotDir
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusBadRequest, stderr.String())
		return
	}
	println("Finished with docker")
	c.IndentedJSON(http.StatusOK, "all good")
}

func RemoveNautobot(c *gin.Context) {
	nautobotDir := "nautobot-docker-compose"
	println("Run docker (may take a long time...)")
	// Run docker
	args := []string{"compose", "down", "-v"}
	cmd := exec.Command("docker", args...)
	cmd.Dir = nautobotDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusBadRequest, stderr.String())
		return
	}
	println("Finished with docker")
	c.IndentedJSON(http.StatusOK, "all good")
}

func replaceTextInFile(fileName string, textsToReplace, replaceWith []string) error {
	input, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		for j, textToReplace := range textsToReplace {
			if strings.Contains(line, textToReplace) {
				lines[i] = strings.Replace(line, textToReplace, replaceWith[j], 1)
			}
		}
	}
	output := strings.Join(lines, "\n")
	err = os.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		return err
	}
	return nil
}
