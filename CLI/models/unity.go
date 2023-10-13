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
	"sync/atomic"
	"time"
)

type Ogree3DConnection struct {
	conn        net.Conn
	isConnected atomic.Bool
}

func (connection *Ogree3DConnection) IsConnected() bool {
	return connection.isConnected.Load()
}

// Connect with OGrEE-3D by a tcp socket
func (connection *Ogree3DConnection) Connect(addr string, timeOut time.Duration) error {
	var dialErr error

	connection.conn, dialErr = net.DialTimeout("tcp", addr, timeOut)
	if dialErr != nil {
		return fmt.Errorf("OGrEE-3D (" + addr + ") unreachable\n" + dialErr.Error())
	}

	connection.isConnected.Store(true)

	return nil
}

func (connection *Ogree3DConnection) Disconnect() {
	if connection.conn != nil {
		connection.isConnected.Store(false)
		connection.conn.Close()
	}
}

// This section under a separate goroutine constantly
// monitors for messages on a port specified by the .env file
// and prints these messages to the Readline terminal
// This is meant for OGrEE-3D interactivity
func (connection *Ogree3DConnection) ReceiveLoop(terminal *readline.Instance) {
	if !connection.IsConnected() {
		return
	}

	reader := bufio.NewReader(connection.conn)
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
		toPrint := "Received from OGrEE-3D: " + msg + "\n"
		terminal.Write([]byte(toPrint))
	}

	// for loop has been exited, there is an error in the connection.
	connection.Disconnect()

	terminal.Write([]byte("Disconnected from OGrEE-3D\n"))
}

// Function to communicate with OGrEE-3D
func (connection *Ogree3DConnection) Send(data map[string]interface{}, debug int) error {
	if !connection.IsConnected() {
		return fmt.Errorf("not connected to OGrEE-3D")
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data : %s", err.Error())
	}

	if debug >= 4 {
		println("DEBUG OUTGOING JSON")
		println("JSON: ", string(dataJSON))
	}

	sizeBytesBuff := new(bytes.Buffer)
	sizeConvErr := binary.Write(sizeBytesBuff, binary.LittleEndian, int32(len(dataJSON)))
	if sizeConvErr != nil {
		return fmt.Errorf("error converting size to binary : %s", sizeConvErr.Error())
	}

	_, writeSizeErr := connection.conn.Write(sizeBytesBuff.Bytes())
	if writeSizeErr != nil {
		return fmt.Errorf("error contacting OGrEE-3D: %s", writeSizeErr.Error())
	}

	_, writeJsonErr := connection.conn.Write(dataJSON)
	if writeJsonErr != nil {
		return fmt.Errorf("error contacting OGrEE-3D: %s", writeJsonErr.Error())
	}

	return nil
}
