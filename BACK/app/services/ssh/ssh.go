package sshcmd

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

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
	if !wait {
		println(cmd)
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
