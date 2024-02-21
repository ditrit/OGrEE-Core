package docker

import (
	"back-admin/models"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func GetTenants(c *gin.Context) {
	response := make(map[string][]models.Tenant)
	response["tenants"] = getTenantsFromJSON()
	c.IndentedJSON(http.StatusOK, response)
}

func getTenantsFromJSON() []models.Tenant {
	if _, err := os.Stat("tenants.json"); errors.Is(err, os.ErrNotExist) {
		// tenants.json does not exist, create it
		var file, e = os.Create("tenants.json")
		if e != nil {
			panic(e.Error())
		} else {
			file.WriteString("[]")
			file.Sync()
			defer file.Close()
			return []models.Tenant{}
		}
	}
	data, e := os.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []models.Tenant
	json.Unmarshal(data, &listTenants)
	fmt.Println(listTenants)
	return listTenants
}

func GetTenantDockerInfo(c *gin.Context) {
	name := c.Param("name")
	println(name)
	if response, err := getDockerInfo(name); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	} else {
		c.IndentedJSON(http.StatusOK, response)
	}
}

func getDockerInfo(name string) ([]models.ContainerInfo, error) {
	println(name)
	cmd := exec.Command("docker", "ps", "--all", "--format", "\"{{json .}}\"")
	cmd.Dir = DOCKER_DIR
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if output, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return nil, errors.New(stderr.String())
	} else {
		response := []models.ContainerInfo{}
		s := bufio.NewScanner(bytes.NewReader(output))
		for s.Scan() {
			var dc models.ContainerInfo
			jsonOutput := s.Text()
			jsonOutput, _ = strings.CutPrefix(jsonOutput, "\"")
			jsonOutput, _ = strings.CutSuffix(jsonOutput, "\"")
			// fmt.Println(jsonOutput)
			if err := json.Unmarshal([]byte(jsonOutput), &dc); err != nil {
				//handle error
				fmt.Println(err.Error())
			}
			// fmt.Println(dc)
			if name == "netbox" {
				if strings.Contains(dc.Name, "netbox-1") {
					response = append(response, dc)
				}
			} else if name == "opendcim" {
				if strings.Contains(dc.Name, "opendcim-webapp") {
					response = append(response, dc)
				}
			} else if name == "nautobot" {
				if strings.Contains(dc.Name, "nautobot-1") {
					response = append(response, dc)
				}
			} else if match, _ := regexp.MatchString("^"+name+"_", dc.Name); match {
				response = append(response, dc)
			}
		}
		if s.Err() != nil {
			// handle scan error
			fmt.Println(s.Err().Error())
			return nil, s.Err()
		}

		fmt.Println(response)
		return response, nil
	}
}

func GetContainerLogs(c *gin.Context) {
	name := c.Param("name")
	println(name)
	cmd := exec.Command("docker", "logs", name, "--tail", "1000")
	cmd.Dir = DOCKER_DIR
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if output, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusInternalServerError, stderr.String())
		return
	} else {
		response := map[string]string{}
		response["logs"] = string(output)

		c.IndentedJSON(http.StatusOK, response)
	}
}

func AddTenant(c *gin.Context) {
	data, e := os.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []models.Tenant
	json.Unmarshal(data, &listTenants)

	// Call BindJSON to bind the received JSON
	var newTenant models.Tenant
	if err := c.BindJSON(&newTenant); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else {
		if err := dockerCreateTenant(newTenant, c); err != "" {
			c.String(http.StatusInternalServerError, err)
			return
		}

		// Add to local json and respond
		// newTenant.CustomerPassword = ""
		listTenants = append(listTenants, newTenant)
		data, _ := json.MarshalIndent(listTenants, "", "  ")
		_ = os.WriteFile("tenants.json", data, 0755)
		c.String(http.StatusOK, "Tenant created!")
	}

}

func dockerCreateTenant(newTenant models.Tenant, c *gin.Context) string {
	tenantLower := strings.ToLower(newTenant.Name)
	// Image tagging
	if newTenant.ImageTag == "" {
		newTenant.ImageTag = "main"
	}

	// Docker compose prepare
	args := []string{"compose", "-p", tenantLower}
	if newTenant.HasWeb {
		args = append(args, "--profile")
		args = append(args, "web")
		// Create flutter assets folder
		newTenant.AssetsDir = "./app-deploy/" + tenantLower
		addAppAssets(newTenant, DOCKER_DIR+newTenant.AssetsDir)
	} else {
		// docker does not accept it empty, even if it wont be created
		newTenant.AssetsDir = DOCKER_DIR
	}
	if newTenant.HasDoc {
		args = append(args, "--profile")
		args = append(args, "doc")
	}
	args = append(args, "--env-file")
	envFilename := tenantLower + ".env"
	args = append(args, envFilename)
	args = append(args, "up")
	args = append(args, "--build")
	args = append(args, "-d")

	// Create tenantName.env
	file, _ := os.Create(DOCKER_DIR + envFilename)
	err := tmplt.Execute(file, newTenant)
	if err != nil {
		return "Error creating .env: " + err.Error()
	}
	file.Close()

	println("Run docker (may take a long time...)")

	cmd := exec.Command("docker", args...)
	cmd.Dir = DOCKER_DIR
	errStr := ""
	if err := streamExecuteCmd(cmd, c); err != nil {
		errStr = "Error running docker: " + err.Error()
	}

	print("Finished with docker")
	return errStr
}

func streamExecuteCmd(cmd *exec.Cmd, c *gin.Context) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout
	scanner := bufio.NewScanner(stdout)

	err = cmd.Start()
	if err != nil {
		return err
	}

	c.Stream(func(w io.Writer) bool {
		if scanner.Scan() {
			msg := string(scanner.Bytes())
			println(msg)
			c.SSEvent("progress", msg)
			return true
		}
		return false
	})

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func addAppAssets(newTenant models.Tenant, assestsDit string) {
	// Create flutter assets folder with .env
	err := os.MkdirAll(assestsDit, 0755)
	if err != nil && !strings.Contains(err.Error(), "already") {
		println(err.Error())
	}
	file, err := os.Create(assestsDit + "/.env")
	if err != nil {
		println(err.Error())
	}
	err = apptmplt.Execute(file, newTenant)
	if err != nil {
		println(err.Error())
	}
	file.Close()

	// Add default logo if none already present
	userLogo := assestsDit + "/logo.png"
	defaultLogo := "handlers/docker/flutter-assets/logo.png"
	if _, err := os.Stat(userLogo); err == nil {
		println("Logo already exists")
	} else {
		println("Setting logo by default")
		source, err := os.Open(defaultLogo)
		if err != nil {
			println("Error opening default logo")
		}
		defer source.Close()
		destination, err := os.Create(userLogo)
		if err != nil {
			println("Error creating tenant logo file")
		}
		defer destination.Close()
		_, err = io.Copy(destination, source)
		if err != nil {
			println("Error creating tenant logo")
		}
	}
}

func AddTenantLogo(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))
	// Load image
	formFile, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	// Make sure destination dir is created
	assetsDir := DOCKER_DIR + "app-deploy/" + tenantName
	err = os.MkdirAll(assetsDir, 0755)
	if err != nil && !strings.Contains(err.Error(), "already") {
		c.String(http.StatusInternalServerError, err.Error())
	}
	// Save image
	err = c.SaveUploadedFile(formFile, assetsDir+"/logo.png")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	c.String(http.StatusOK, "")
}

func RemoveTenant(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))

	// Stop and remove containers
	for _, str := range []string{"_webapp", "_api", "_db", "_doc"} {
		cmd := exec.Command("docker", "rm", "-v", "--force", strings.ToLower(tenantName)+str)
		cmd.Dir = DOCKER_DIR
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			c.IndentedJSON(http.StatusInternalServerError, stderr.String())
			return
		} else if str == "_db" {
			// Remove volume
			cmd = exec.Command("docker", "volume", "rm", strings.ToLower(tenantName)+str)
			cmd.Dir = DOCKER_DIR
			cmd.Stderr = &stderr
			if _, err := cmd.Output(); err != nil {
				fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
				c.IndentedJSON(http.StatusInternalServerError, stderr.String())
				return
			}
		}
	}

	// Remove assets
	os.RemoveAll(DOCKER_DIR + "app-deploy/" + tenantName)
	os.Remove(DOCKER_DIR + tenantName + ".env")

	// Update local file
	data, e := os.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []models.Tenant
	json.Unmarshal(data, &listTenants)
	for i, ten := range listTenants {
		if ten.Name == tenantName {
			listTenants = append(listTenants[:i], listTenants[i+1:]...)
		}
	}
	data, _ = json.MarshalIndent(listTenants, "", "  ")
	_ = os.WriteFile("tenants.json", data, 0755)
	c.IndentedJSON(http.StatusOK, "all good")
}

func UpdateTenant(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))
	println(tenantName)

	// Call BindJSON to bind the received JSON
	var newTenant models.Tenant
	if err := c.BindJSON(&newTenant); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	listTenants := getTenantsFromJSON()
	for i, tenant := range listTenants {
		if strings.ToLower(tenant.Name) == tenantName {
			// Docker compose stop
			println("Docker stop current tenant")
			args := []string{"compose", "-p", tenantName, "stop"}
			cmd := exec.Command("docker", args...)
			cmd.Dir = DOCKER_DIR
			if err := streamExecuteCmd(cmd, c); err != nil {
				errStr := "Error running docker: " + err.Error()
				println(errStr)
				c.IndentedJSON(http.StatusInternalServerError, errStr)
				return
			}
			println("Finished with docker")

			// Docker compose up
			if err := dockerCreateTenant(newTenant, c); err != "" {
				c.IndentedJSON(http.StatusInternalServerError, err)
				// Try to recreate previous config
				err = dockerCreateTenant(tenant, c)
				if err != "" {
					println("Error recovering:" + err)
				}
				return
			}

			listTenants[i] = newTenant
			println(listTenants)
			data, _ := json.MarshalIndent(listTenants, "", "  ")
			_ = os.WriteFile("tenants.json", data, 0755)
			break
		}
	}

	c.String(http.StatusOK, "")
}

func BackupTenantDB(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))
	t := time.Now()

	// Call BindJSON to bind the received JSON
	var backupInfo models.Backup
	if err := c.BindJSON(&backupInfo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	println("Docker backup current tenant")
	args := []string{"exec", tenantName + "_db", "sh", "-c",
		"exec mongodump --username ogree" + tenantName + "Admin --password " + backupInfo.DBPassword + " -d ogree" + tenantName + " --archive"}
	cmd := exec.Command("docker", args...)
	cmd.Dir = DOCKER_DIR
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	outfile, err := os.Create(tenantName + "_db_" + t.Format("2006-01-02T150405") + ".archive")
	if err != nil {
		println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusInternalServerError, stderr.String())
		return
	}

	dir, _ := os.Getwd()
	println("Finished with docker")
	if backupInfo.IsDownload {
		c.File(outfile.Name())
	} else {
		c.String(http.StatusOK, "Backup file created as "+outfile.Name()+" at "+dir)
	}
}

func StopStartTentant(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))
	path := strings.Split(c.FullPath(), "/")
	command := path[len(path)-1]
	println(command)
	println(tenantName)
	// Docker compose stop/start
	println("Docker current tenant")
	args := []string{"compose", "-p", tenantName, command}
	cmd := exec.Command("docker", args...)
	cmd.Dir = DOCKER_DIR
	if err := streamExecuteCmd(cmd, c); err != nil {
		errStr := "Error running docker: " + err.Error()
		println(errStr)
		c.IndentedJSON(http.StatusInternalServerError, errStr)
		return
	}
	println("Finished with docker")
	c.String(http.StatusOK, "")
}
