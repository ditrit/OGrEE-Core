package docker

import (
	"fmt"
	"kube-admin/models"
	sshcmd "kube-admin/services/ssh"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Add a binary of this same backend in another server
func CreateNewBackend(c *gin.Context) {
	var newServer models.BackendServer
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
		pKey, err := os.ReadFile(newServer.Pkey)
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

	//Create .env file for distant copy
	if e := createEnvFile(newServer.DstPath + "/"); e != "" {
		c.String(http.StatusInternalServerError, e)
		return
	}

	sshcmd.SSHRunCmd("mkdir -p "+newServer.DstPath+"/docker", conn, true)
	sshcmd.SSHRunCmd("mkdir -p "+newServer.DstPath+"/backend-assets", conn, true)
	sshcmd.SSHRunCmd("mkdir -p "+newServer.DstPath+"/flutter-assets", conn, true)

	sshcmd.SSHCopyFile("ogree_app_backend", newServer.DstPath+"/ogree_app_backend", conn)
	sshcmd.SSHCopyFile("backend-assets/docker-env-template.txt", newServer.DstPath+"/backend-assets/docker-env-template.txt", conn)
	sshcmd.SSHCopyFile("backend-assets/template.service", newServer.DstPath+"/backend-assets/template.service", conn)
	sshcmd.SSHCopyFile("flutter-assets/flutter-env-template.txt", newServer.DstPath+"/flutter-assets/flutter-env-template.txt", conn)
	sshcmd.SSHCopyFile("flutter-assets/logo.png", newServer.DstPath+"/flutter-assets/logo.png", conn)
	sshcmd.SSHCopyFile(".envcopy", newServer.DstPath+"/.env", conn)
	sshcmd.SSHCopyFile(DOCKER_DIR+"docker-compose.yml", newServer.DstPath+"/docker/docker-compose.yml", conn)
	sshcmd.SSHCopyFile(DEPLOY_DIR+"createdb.js", newServer.DstPath+"/createdb.js", conn)
	sshcmd.SSHCopyFile(DOCKER_DIR+"init.sh", newServer.DstPath+"/docker/init.sh", conn)
	if newServer.AtStartup {
		// Create service file and send it to server
		file, _ := os.Create("ogree_app_backend.service")
		err = servertmplt.Execute(file, newServer)
		if err != nil {
			fmt.Println("Error creating service file: " + err.Error())
		}
		file.Close()
		sshcmd.SSHCopyFile("ogree_app_backend.service", "/etc/systemd/system/ogree_app_backend.service", conn)
		sshcmd.SSHRunCmd("systemctl enable ogree_app_backend.service", conn, true)
	}

	sshcmd.SSHRunCmd("chmod +x "+newServer.DstPath+"/ogree_app_backend", conn, true)
	sshcmd.SSHRunCmd("cd "+newServer.DstPath+" && nohup "+newServer.DstPath+"/ogree_app_backend -port "+newServer.RunPort+" > "+newServer.DstPath+"/ogree_backend.out", conn, false)

	c.String(http.StatusOK, "all good")
}

func createEnvFile(dir string) string {
	input, err := os.ReadFile(".env")
	if err != nil {
		return err.Error()
	}

	lines := strings.Split(string(input), "\n")

	replaced := false
	for i, line := range lines {
		if strings.Contains(line, "DEPLOY_DIR") {
			lines[i] = "DEPLOY_DIR=" + dir
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, "DEPLOY_DIR="+dir)
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(".envcopy", []byte(output), 0644)
	if err != nil {
		return err.Error()
	}

	return ""
}
