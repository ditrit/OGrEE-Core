package models

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisconnectInReceiveLoop(t *testing.T) {
	acceptAndCloseServer(t)

	connection := &Ogree3DConnection{}
	err := connection.Connect("localhost:3000", 5*time.Second)
	require.Nil(t, err)
	require.True(t, connection.IsConnected())

	waitReceiveLoop(connection)

	assert.False(t, connection.IsConnected())
	err = connection.Send(nil, 0)
	assert.ErrorContains(t, err, "not connected to OGrEE-3D")
}

func TestDisconnectInSend(t *testing.T) {
	serverFinishedWG := acceptAndCloseServer(t)

	connection := &Ogree3DConnection{}
	err := connection.Connect("localhost:3000", 5*time.Second)
	require.Nil(t, err)
	require.True(t, connection.IsConnected())

	serverFinishedWG.Wait()

	err = connection.Send(nil, 0)
	require.ErrorContains(t, err, "error contacting OGrEE-3D")
	require.False(t, connection.IsConnected())

	waitReceiveLoop(connection)
	assert.False(t, connection.IsConnected())
}

func acceptAndCloseServer(t *testing.T) *sync.WaitGroup {
	ln, err := net.Listen("tcp", ":3000")
	require.Nil(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		acceptAndClose(ln)
	}()

	return &wg
}

func acceptAndClose(ln net.Listener) {
	conn, _ := ln.Accept()
	conn.Close()
	ln.Close()
}

func waitReceiveLoop(connection *Ogree3DConnection) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		connection.ReceiveLoop()
	}()
	wg.Wait()
}
