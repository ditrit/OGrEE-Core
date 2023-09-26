package controllers

import (
	"cli/models"
	"cli/readline"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	rl, _ := readline.New("")
	State.Terminal = &rl
}

func TestConnect3DReturnsErrorIfProvidedURLIsInvalid(t *testing.T) {
	err := Connect3D("not.valid")
	assert.ErrorContains(t, err, "OGrEE-3D URL is not valid: not.valid")
	assert.False(t, models.Ogree3D.IsConnected())
	assert.NotEqual(t, State.Ogree3DURL, "not.valid")
}

func TestConnect3DDoesNotConnectIfOgree3DIsUnreachable(t *testing.T) {
	err := Connect3D("localhost:3000")
	assert.Nil(t, err)
	assert.False(t, models.Ogree3D.IsConnected())
	assert.Equal(t, State.Ogree3DURL, "localhost:3000")
}

func TestConnect3DConnectsToProvidedURL(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	assert.Nil(t, err)
	assert.True(t, models.Ogree3D.IsConnected())
	assert.Equal(t, State.Ogree3DURL, "localhost:3000")
}

func TestConnect3DConnectsToStateOgreeURLIfNotProvidedURL(t *testing.T) {
	fakeOgree3D(t, "3000")

	State.SetOgree3DURL("localhost:3000")
	err := Connect3D("")
	assert.Nil(t, err)
	assert.True(t, models.Ogree3D.IsConnected())
	assert.Equal(t, State.Ogree3DURL, "localhost:3000")
}

func TestConnect3DReturnsErrorIfAlreadyConnectedAndNotUrlProvided(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, models.Ogree3D.IsConnected())

	err = Connect3D("")
	assert.ErrorContains(t, err, "already connected to OGrEE-3D url: localhost:3000")
	assert.True(t, models.Ogree3D.IsConnected())
}

func TestConnect3DReturnsErrorIfAlreadyConnectedAndSameUrlProvided(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, models.Ogree3D.IsConnected())

	err = Connect3D("localhost:3000")
	assert.ErrorContains(t, err, "already connected to OGrEE-3D url: localhost:3000")
	assert.True(t, models.Ogree3D.IsConnected())
}

func TestConnect3DTriesToConnectIfAlreadyConnectedAndDifferentUrlProvided(t *testing.T) {
	wg := fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, models.Ogree3D.IsConnected())

	err = Connect3D("localhost:5000")
	assert.Nil(t, err)
	assert.False(t, models.Ogree3D.IsConnected())
	assert.Equal(t, State.Ogree3DURL, "localhost:5000")

	wg.Wait()
}

func TestConnect3DConnectsIfAlreadyConnectedAndDifferentUrlProvidedIsReachable(t *testing.T) {
	wg := fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, models.Ogree3D.IsConnected())

	fakeOgree3D(t, "5000")

	err = Connect3D("localhost:5000")
	assert.Nil(t, err)
	assert.True(t, models.Ogree3D.IsConnected())
	assert.Equal(t, State.Ogree3DURL, "localhost:5000")

	wg.Wait()
}

func fakeOgree3D(t *testing.T, port string) *sync.WaitGroup {
	ln, err := net.Listen("tcp", ":"+port)
	require.Nil(t, err)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		conn, _ := ln.Accept()
		buf := make([]byte, 1)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				break
			}
		}
		conn.Close()
		ln.Close()
		wg.Done()
	}()

	t.Cleanup(func() {
		ln.Close()
		if models.Ogree3D.IsConnected() {
			models.Ogree3D.Disconnect()
		}
	})

	return &wg
}
