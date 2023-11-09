package controllers

import (
	"cli/readline"
	"fmt"
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
	assert.False(t, Ogree3D.IsConnected())
	assert.NotEqual(t, Ogree3D.URL(), "not.valid")
}

func TestConnect3DDoesNotConnectIfOgree3DIsUnreachable(t *testing.T) {
	err := Connect3D("localhost:3000")
	assert.ErrorContains(t, err, "OGrEE-3D is not reachable caused by OGrEE-3D (localhost:3000) unreachable\ndial tcp 127.0.0.1:3000: connect: connection refused")
	assert.False(t, Ogree3D.IsConnected())
	assert.Equal(t, Ogree3D.URL(), "localhost:3000")
}

func TestConnect3DConnectsToProvidedURL(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	assert.Nil(t, err)
	assert.True(t, Ogree3D.IsConnected())
	assert.Equal(t, Ogree3D.URL(), "localhost:3000")
}

func TestConnect3DConnectsToStateOgreeURLIfNotProvidedURL(t *testing.T) {
	fakeOgree3D(t, "3000")

	Ogree3D.SetURL("localhost:3000")
	err := Connect3D("")
	assert.Nil(t, err)
	assert.True(t, Ogree3D.IsConnected())
	assert.Equal(t, Ogree3D.URL(), "localhost:3000")
}

func TestConnect3DReturnsErrorIfAlreadyConnectedAndNotUrlProvided(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, Ogree3D.IsConnected())

	err = Connect3D("")
	assert.ErrorContains(t, err, "already connected to OGrEE-3D url: localhost:3000")
	assert.True(t, Ogree3D.IsConnected())
}

func TestConnect3DReturnsErrorIfAlreadyConnectedAndSameUrlProvided(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, Ogree3D.IsConnected())

	err = Connect3D("localhost:3000")
	assert.ErrorContains(t, err, "already connected to OGrEE-3D url: localhost:3000")
	assert.True(t, Ogree3D.IsConnected())
}

func TestConnect3DTriesToConnectIfAlreadyConnectedAndDifferentUrlProvided(t *testing.T) {
	wg := fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, Ogree3D.IsConnected())

	err = Connect3D("localhost:5000")
	assert.ErrorContains(t, err, "OGrEE-3D is not reachable caused by OGrEE-3D (localhost:5000) unreachable\ndial tcp 127.0.0.1:5000: connect: connection refused")
	assert.False(t, Ogree3D.IsConnected())
	assert.Equal(t, Ogree3D.URL(), "localhost:5000")

	wg.Wait()
}

func TestConnect3DConnectsIfAlreadyConnectedAndDifferentUrlProvidedIsReachable(t *testing.T) {
	wg := fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, Ogree3D.IsConnected())

	fakeOgree3D(t, "5000")

	err = Connect3D("localhost:5000")
	assert.Nil(t, err)
	assert.True(t, Ogree3D.IsConnected())
	assert.Equal(t, Ogree3D.URL(), "localhost:5000")

	wg.Wait()
}

func TestInformOgree3DOptionalDoesNothingIfOgree3DNotConnected(t *testing.T) {
	require.False(t, Ogree3D.IsConnected())
	err := Ogree3D.InformOptional("Interact", -1, map[string]any{})
	assert.Nil(t, err)
}

func TestInformOgree3DOptionalSendDataWhenOgree3DIsConnected(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, Ogree3D.IsConnected())

	err = Ogree3D.InformOptional("Interact", -1, map[string]any{})
	assert.Nil(t, err)
}

func TestInformOgree3DFailsIfOgree3DNotReachable(t *testing.T) {
	require.False(t, Ogree3D.IsConnected())
	Ogree3D.SetURL("localhost:3000")
	err := Ogree3D.Inform("Interact", -1, map[string]any{})
	assert.ErrorContains(t, err, "OGrEE-3D is not reachable caused by OGrEE-3D (localhost:3000) unreachable\ndial tcp")
	assert.ErrorContains(t, err, "connect: connection refused")
}

func TestInformOgree3DEstablishConnectionIfOgree3DIsReachable(t *testing.T) {
	require.False(t, Ogree3D.IsConnected())

	Ogree3D.SetURL("localhost:3000")
	fakeOgree3D(t, "3000")

	err := Ogree3D.Inform("Interact", -1, map[string]any{})
	assert.Nil(t, err)
}

func TestInformOgree3DSendsDataIfEstablishConnectionWithOgree3DAlreadyEstablished(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, Ogree3D.IsConnected())

	err = Ogree3D.Inform("Interact", -1, map[string]any{})
	assert.Nil(t, err)
}

func fakeOgree3D(t *testing.T, port string) *sync.WaitGroup {
	ln, err := net.Listen("tcp", ":"+port)
	require.Nil(t, err)

	fmt.Println("Fake OGrEE-3D running on", ":"+port)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Fake OGrEE-3D could not accept connection")
			return
		}

		buf := make([]byte, 1)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				break
			}
		}
		fmt.Println("Fake OGrEE-3D connection broken")
		conn.Close()
		wg.Done()
	}()

	t.Cleanup(func() {
		ln.Close()
		if Ogree3D.IsConnected() {
			Ogree3D.Disconnect()
		}
	})

	return &wg
}
