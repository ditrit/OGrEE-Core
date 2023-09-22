package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type backendServer struct {
	Host      string `json:"host" binding:"required"`
	User      string `json:"user" binding:"required"`
	Password  string `json:"password"`
	Pkey      string `json:"pkey"`
	PkeyPass  string `json:"pkeypass"`
	DstPath   string `json:"dstpath" binding:"required"`
	RunPort   string `json:"runport" binding:"required"`
	AtStartup bool   `json:"startup"`
}

// Add a binary of this same backend in another server
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

	//Create .env file for distant copy
	if e := createEnvFile(newServer.DstPath + "/"); e != "" {
		c.String(http.StatusInternalServerError, e)
		return
	}

	SSHRunCmd("mkdir -p "+newServer.DstPath+"/docker", conn, true)
	SSHRunCmd("mkdir -p "+newServer.DstPath+"/backend-assets", conn, true)
	SSHRunCmd("mkdir -p "+newServer.DstPath+"/flutter-assets", conn, true)

	SSHCopyFile("ogree_app_backend", newServer.DstPath+"/ogree_app_backend", conn)
	SSHCopyFile("backend-assets/docker-env-template.txt", newServer.DstPath+"/backend-assets/docker-env-template.txt", conn)
	SSHCopyFile("backend-assets/template.service", newServer.DstPath+"/backend-assets/template.service", conn)
	SSHCopyFile("flutter-assets/flutter-env-template.txt", newServer.DstPath+"/flutter-assets/flutter-env-template.txt", conn)
	SSHCopyFile("flutter-assets/logo.png", newServer.DstPath+"/flutter-assets/logo.png", conn)
	SSHCopyFile(".envcopy", newServer.DstPath+"/.env", conn)
	SSHCopyFile(DOCKER_DIR+"docker-compose.yml", newServer.DstPath+"/docker/docker-compose.yml", conn)
	SSHCopyFile(DEPLOY_DIR+"createdb.js", newServer.DstPath+"/createdb.js", conn)
	SSHCopyFile(DOCKER_DIR+"init.sh", newServer.DstPath+"/docker/init.sh", conn)
	if newServer.AtStartup {
		// Create service file and send it to server
		file, _ := os.Create("ogree_app_backend.service")
		err = servertmplt.Execute(file, newServer)
		if err != nil {
			fmt.Println("Error creating service file: " + err.Error())
		}
		file.Close()
		SSHCopyFile("ogree_app_backend.service", "/etc/systemd/system/ogree_app_backend.service", conn)
		SSHRunCmd("systemctl enable ogree_app_backend.service", conn, true)
	}

	SSHRunCmd("chmod +x "+newServer.DstPath+"/ogree_app_backend", conn, true)
	SSHRunCmd("cd "+newServer.DstPath+" && nohup "+newServer.DstPath+"/ogree_app_backend -port "+newServer.RunPort+" > "+newServer.DstPath+"/ogree_backend.out", conn, false)

	c.String(http.StatusOK, "all good")
}

func SSHCopyFile(srcPath, dstPath string, client *ssh.Client) error {
	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		println(err.Error())
		return err
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		println(err.Error())
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := sftp.Create(dstPath)
	if err != nil {
		println(err.Error())
		return err
	}
	defer dstFile.Close()

	// write to file
	if _, err := dstFile.ReadFrom(srcFile); err != nil {
		println(err.Error())
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

func createEnvFile(dir string) string {
	input, err := ioutil.ReadFile(".env")
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
	err = ioutil.WriteFile(".envcopy", []byte(output), 0644)
	if err != nil {
		return err.Error()
	}

	return ""
}
