package models

import (
	"bufio"
	l "cli/logger"
	"cli/readline"
	"net"
)

//This section under a separate goroutine constantly
//monitors for messages on a port specified by the .env file
//and prints these messages to the Readline terminal
//This is meant for Unity interactivity
func ListenForUnity(rl *readline.Instance, addr string) error {
	//addr := "0.0.0.0:" + getListenerPort(rl)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		println("LISTEN ERROR: ", err.Error())
		l.GetListenErrorLogger().Println(err.Error())
		return nil
	}
	l.GetListenInfoLogger().Println("Listening server started")

	for {
		cn, err := ln.Accept()
		if err == nil {

			reply, err1 := bufio.NewReader(cn).ReadBytes('\n')
			if err1 == nil {
				msg := []byte("Received from Unity: ")

				msg = append(msg, reply...)

				rl.Write(msg)

			} else {
				l.GetListenErrorLogger().Println(err1.Error())
			}

		} else {
			l.GetListenErrorLogger().Println(err.Error())
		}
		cn.Close()
	}

}
