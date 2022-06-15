package models

import (
	"bufio"
	l "cli/logger"
	"cli/readline"
	"net"
	"os"
	"strings"
)

func getListenerPort(rl *readline.Instance) string {
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

	rl.Write([]byte("Falling back to default Listening Port\n"))
	l.ListenerInfoLogger.Println("Falling back to default Listening Port")
	//InfoLogger.Println("Falling back to Listening Port")
	return "5501"
}

//This section under a separate goroutine constantly
//monitors for messages on a port specified by the .env file
//and prints these messages to the Readline terminal
//This is meant for Unity interactivity
func ListenForUnity(rl *readline.Instance) error {
	addr := "localhost:" + getListenerPort(rl)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		println("LISTEN ERROR: ", err.Error())
		l.ListenerErrorLogger.Println(err.Error())
		return nil
	}
	l.ListenerInfoLogger.Println("Listening server started")

	for {
		cn, err := ln.Accept()
		if err == nil {

			reply, err1 := bufio.NewReader(cn).ReadBytes('\n')
			if err1 == nil {
				msg := []byte("Received from Unity: ")

				msg = append(msg, reply...)

				rl.Write(msg)

			} else {
				l.ListenerErrorLogger.Println(err1.Error())
			}

		} else {
			l.ListenerErrorLogger.Println(err.Error())
		}
		cn.Close()
	}

}
