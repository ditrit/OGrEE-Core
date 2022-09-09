package models

import (
	"bufio"
	"bytes"
	"cli/readline"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

var conn net.Conn
var connected bool = false

func ConnectToUnity(addr string, timeOut time.Duration) error {
	var dialErr error
	conn, dialErr = net.DialTimeout("tcp", addr, timeOut)
	if dialErr != nil {
		return fmt.Errorf("Unity Client (" + addr + ") unreachable\n" + dialErr.Error())
	}
	connected = true
	return nil
}

//This section under a separate goroutine constantly
//monitors for messages on a port specified by the .env file
//and prints these messages to the Readline terminal
//This is meant for Unity interactivity
func ReceiveLoop(rl *readline.Instance, addr string, shellConnected *bool) {
	reader := bufio.NewReader(conn)
	var err error
	for {
		var size int32
		err = binary.Read(reader, binary.LittleEndian, &size)
		if err != nil {
			break
		}
		msgBuffer := make([]byte, size)
		_, err = io.ReadFull(reader, msgBuffer)
		if err != nil {
			break
		}
		msg := string(msgBuffer)
		toPrint := "Received from Unity : " + msg + "\n"
		rl.Write([]byte(toPrint))
	}
	connected = false
	*shellConnected = false
	conn.Close()
	println("Disconnected from server")
}

//Function to communicate with Unity
func ContactUnity(data map[string]interface{}, debug int) error {
	if !connected {
		return fmt.Errorf("not connected to Unity")
	}
	dataJSON, _ := json.Marshal(data)
	if debug >= 4 {
		println("DEBUG OUTGOING JSON")
		Disp(data)
	}
	sizeBytesBuff := new(bytes.Buffer)
	sizeConvErr := binary.Write(sizeBytesBuff, binary.LittleEndian, int32(len(dataJSON)))
	if sizeConvErr != nil {
		return fmt.Errorf("error converting size to binary : %s", sizeConvErr.Error())
	}
	_, writeSizeErr := conn.Write(sizeBytesBuff.Bytes())
	if writeSizeErr != nil {
		return fmt.Errorf("error contacting Unity : %s", writeSizeErr.Error())
	}
	_, writeJsonErr := conn.Write(dataJSON)
	if writeJsonErr != nil {
		return fmt.Errorf("error contacting Unity: %s", writeJsonErr.Error())
	}
	return nil
}
