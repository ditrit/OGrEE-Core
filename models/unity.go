package models

import (
	"bufio"
	"cli/readline"
	"fmt"
	"net"
	"os"
	"strings"
)

func getListenerPort() string {
	file, err := os.Open("./.resources/.env")
	defer file.Close()
	if err == nil {
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanWords) // use scanwords
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "listenPORT=") {
				return scanner.Text()[11:]
			}
		}
	}

	fmt.Println("Falling back to default Listening Port")
	//InfoLogger.Println("Falling back to Listening Port")
	return "5501"
}

//This section under a separate goroutine constantly
//monitors for messages on a port specified by the .env file
//and prints these messages to the Readline terminal
//This is meant for Unity interactivity
func ListenForUnity(rl *readline.Instance) error {
	addr := "localhost:" + getListenerPort()
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		println("LISTEN ERROR: ", err.Error())
		return nil
	}

	for {
		cn, err := ln.Accept()
		if err == nil {

			reply, err1 := bufio.NewReader(cn).ReadBytes('\n')
			if err1 == nil {
				msg := []byte("Received from Unity: ")

				msg = append(msg, reply...)

				rl.Write(msg)

			}

		}
		cn.Close()
	}

}
