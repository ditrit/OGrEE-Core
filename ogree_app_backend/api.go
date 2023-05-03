package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"ogree_app_backend/auth"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var tmplt *template.Template

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}
	// hashedPassword, _ := bcrypt.GenerateFromPassword(
	// 	[]byte("password"), bcrypt.DefaultCost)
	// println(string(hashedPassword))
	tmplt = template.Must(template.ParseFiles("docker-env-template.txt"))
}

func main() {
	port := flag.Int("port", 8082, "an int")
	flag.Parse()
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"X-Requested-With", "Content-Type", "Authorization", "Origin", "Accept"}
	router.Use(cors.New(corsConfig))

	router.POST("/api/login", login) // public endpoint

	router.Use(auth.JwtAuthMiddleware()) // protected
	router.GET("/api/tenants", getTenants)
	router.GET("/api/tenants/:name", getTenantDockerInfo)
	router.DELETE("/api/tenants/:name", removeTenant)
	router.POST("/api/tenants", addTenant)
	router.GET("/api/containers/:name", getContainerLogs)
	router.POST("/api/servers", createNewBackend)

	router.Run(":" + strconv.Itoa(*port))

}

type tenant struct {
	Name             string `json:"name" binding:"required"`
	CustomerPassword string `json:"customerPassword"`
	ApiUrl           string `json:"apiUrl"`
	WebUrl           string `json:"webUrl"`
	ApiPort          string `json:"apiPort"`
	WebPort          string `json:"webPort"`
	HasWeb           bool   `json:"hasWeb"`
	HasCli           bool   `json:"hasCli"`
}

type container struct {
	Name       string `json:"Names"`
	RunningFor string `json:"RunningFor"`
	State      string `json:"State"`
	Image      string `json:"Image"`
	Size       string `json:"Size"`
	Ports      string `json:"Ports"`
}

type user struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Token    string `json:"token"`
}

type backendServer struct {
	Host     string `json:"host" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password"`
	Pkey     string `json:"pkey"`
	PkeyPass string `json:"pkeypass"`
	DstPath  string `json:"dstpath" binding:"required"`
	RunPort  string `json:"runport" binding:"required"`
}

func getTenants(c *gin.Context) {
	data, e := ioutil.ReadFile("tenants.json")
	if e != nil {
		if strings.Contains(e.Error(), "no such file") || strings.Contains(e.Error(), "cannot find") {
			var file, e = os.Create("tenants.json")
			if e != nil {
				panic(e.Error())
			} else {
				file.WriteString("[]")
				file.Sync()
				defer file.Close()
				response := make(map[string][]tenant)
				response["tenants"] = []tenant{}
				c.IndentedJSON(http.StatusOK, response)
				return
			}
		} else {
			panic(e.Error())
		}
	}
	var listTenants []tenant
	json.Unmarshal(data, &listTenants)
	fmt.Println(listTenants)
	response := make(map[string][]tenant)
	response["tenants"] = listTenants
	c.IndentedJSON(http.StatusOK, response)
}

func getTenantDockerInfo(c *gin.Context) {
	name := c.Param("name")
	println(name)
	cmd := exec.Command("docker", "ps", "--all", "--format", "\"{{json .}}\"")
	cmd.Dir = "docker/"
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if output, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusInternalServerError, stderr.String())
		return
	} else {
		response := []container{}
		s := bufio.NewScanner(bytes.NewReader(output))
		for s.Scan() {
			var dc container
			jsonOutput := s.Text()
			jsonOutput, _ = strings.CutPrefix(jsonOutput, "\"")
			jsonOutput, _ = strings.CutSuffix(jsonOutput, "\"")
			fmt.Println(jsonOutput)
			if err := json.Unmarshal([]byte(jsonOutput), &dc); err != nil {
				//handle error
				fmt.Println(err.Error())
			}
			fmt.Println(dc)
			if strings.Contains(dc.Name, name) {
				response = append(response, dc)
			}
		}
		if s.Err() != nil {
			// handle scan error
			fmt.Println(s.Err().Error())
		}

		c.IndentedJSON(http.StatusOK, response)
	}
}

func getContainerLogs(c *gin.Context) {
	name := c.Param("name")
	println(name)
	cmd := exec.Command("docker", "logs", name, "--tail", "1000")
	cmd.Dir = "docker/"
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

func addTenant(c *gin.Context) {
	data, e := ioutil.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []tenant
	json.Unmarshal(data, &listTenants)

	// Call BindJSON to bind the received JSON
	var newTenant tenant
	if err := c.BindJSON(&newTenant); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else {
		// Create .env file
		file, _ := os.Create("docker/.env")
		err = tmplt.Execute(file, newTenant)
		if err != nil {
			panic(err)
		}
		file.Close()

		// Docker compose up
		args := []string{"-p", strings.ToLower(newTenant.Name)}
		if newTenant.HasWeb {
			args = append(args, "--profile")
			args = append(args, "web")
		}
		if newTenant.HasCli {
			args = append(args, "--profile")
			args = append(args, "cli")
		}
		args = append(args, "up")
		args = append(args, "-d")
		cmd := exec.Command("docker-compose", args...)
		cmd.Dir = "docker/"
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			c.IndentedJSON(http.StatusInternalServerError, stderr.String())
			return
		}

		// Add to local json
		listTenants = append(listTenants, newTenant)
		data, _ := json.MarshalIndent(listTenants, "", "  ")
		_ = ioutil.WriteFile("tenants.json", data, 0644)
		// Create .env copy
		args = []string{"docker/.env", "docker/" + strings.ToLower(newTenant.Name) + ".env"}
		cmd = exec.Command("cp", args...)
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		}

		c.IndentedJSON(http.StatusOK, "all good")
	}

}

func removeTenant(c *gin.Context) {
	tenantName := c.Param("name")

	for _, str := range []string{"_cli", "_webapp", "_api", "_db"} {
		cmd := exec.Command("docker", "rm", "--force", strings.ToLower(tenantName)+str)
		cmd.Dir = "docker/"
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			c.IndentedJSON(http.StatusInternalServerError, stderr.String())
			return
		}
	}

	// Update local file
	data, e := ioutil.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []tenant
	json.Unmarshal(data, &listTenants)
	for i, ten := range listTenants {
		if ten.Name == tenantName {
			listTenants = append(listTenants[:i], listTenants[i+1:]...)
		}
	}
	data, _ = json.MarshalIndent(listTenants, "", "  ")
	_ = ioutil.WriteFile("tenants.json", data, 0644)
	c.IndentedJSON(http.StatusOK, "all good")
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

func createNewBackend(c *gin.Context) {
	var newServer backendServer
	if err := c.BindJSON(&newServer); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	var err error
	var signer ssh.Signer
	var homeDir string
	sshAuthMethod := []ssh.AuthMethod{}

	if newServer.Password != "" {
		println("password")
		sshAuthMethod = append(sshAuthMethod, ssh.Password(newServer.Password))
	} else {
		pKey, err := ioutil.ReadFile(newServer.Pkey)
		if err != nil {
			fmt.Println("Failed to read ssh_host_key")
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		if newServer.PkeyPass != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(pKey, []byte(newServer.PkeyPass))
		} else {
			signer, err = ssh.ParsePrivateKey(pKey)
		}
		if err != nil {
			fmt.Println(err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		sshAuthMethod = append(sshAuthMethod, ssh.PublicKeys(signer))
	}

	homeDir, err = os.UserHomeDir()
	if err != nil {
		fmt.Println(err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	var hostkeyCallback ssh.HostKeyCallback
	hostkeyCallback, err = knownhosts.New(homeDir + "/.ssh/known_hosts")
	if err != nil {
		fmt.Println(err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	conf := &ssh.ClientConfig{
		User:            newServer.User,
		HostKeyCallback: hostkeyCallback,
		Auth:            sshAuthMethod,
	}

	var conn *ssh.Client

	conn, err = ssh.Dial("tcp", newServer.Host, conf)
	if err != nil {
		fmt.Println(err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer conn.Close()

	SSHRunCmd("mkdir -p "+newServer.DstPath+"/docker", conn, true)

	SSHCopyFile("ogree_app_backend_linux", newServer.DstPath+"/ogree_app_backend", conn)
	SSHCopyFile("docker-env-template.txt", newServer.DstPath+"/docker-env-template.txt", conn)
	SSHCopyFile(".env", newServer.DstPath+"/.env", conn)
	SSHCopyFile("docker/docker-compose.yml", newServer.DstPath+"/docker/docker-compose.yml", conn)
	SSHCopyFile("docker/addCustomer.js", newServer.DstPath+"/docker/addCustomer.js", conn)
	SSHCopyFile("docker/addCustomer.sh", newServer.DstPath+"/docker/addCustomer.sh", conn)
	SSHCopyFile("docker/dbft.js", newServer.DstPath+"/docker/dbft.js", conn)
	SSHCopyFile("docker/init.sh", newServer.DstPath+"/docker/init.sh", conn)

	SSHRunCmd("chmod +x "+newServer.DstPath+"/ogree_app_backend", conn, true)
	SSHRunCmd("cd "+newServer.DstPath+" && nohup "+newServer.DstPath+"/ogree_app_backend -port "+newServer.RunPort+" > "+newServer.DstPath+"/ogree_backend.out", conn, false)

	c.String(http.StatusOK, "all good")
}

func SSHCopyFile(srcPath, dstPath string, client *ssh.Client) error {
	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := sftp.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// write to file
	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		return err
	}
	return nil
}

func SSHRunCmd(cmd string, client *ssh.Client, wait bool) {
	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
	}
	defer session.Close()

	var buff bytes.Buffer
	session.Stdout = &buff
	// var buff2 bytes.Buffer
	// session.Stderr = &buff2
	if !wait {
		println(cmd)
		// if err := session.Run(cmd); err != nil {
		// 	fmt.Println(err.Error())
		// 	fmt.Println(buff.String())
		// 	fmt.Println(buff2.String())
		// }
		session.Start(cmd)
		time.Sleep(2 * time.Second)
		session.Close()
		println("out")

	} else {
		if err := session.Run(cmd); err != nil {
			fmt.Println(err)
			fmt.Println(buff.String())
		}
	}
}
