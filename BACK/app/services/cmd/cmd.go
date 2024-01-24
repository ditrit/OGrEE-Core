package cmd


import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)
func ExecCommand(command string)(string,error){
	
	args := strings.Split(command," ")
	cmd := exec.Command(args[0],args[1:]...)
	
	var stderr ,stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		
		return "", errors.New("fail running: " + stderr.String())
	}
	cmd.Wait()
	return stdout.String(),nil
}