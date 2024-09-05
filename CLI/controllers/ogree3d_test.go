package controllers_test

import (
	"cli/controllers"
	"cli/readline"
	test_utils "cli/test"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	rl, _ := readline.New("")
	controllers.State.Terminal = &rl
}

func TestConnect3DReturnsErrorIfProvidedURLIsInvalid(t *testing.T) {
	err := controllers.Connect3D("not.valid")
	assert.ErrorContains(t, err, "OGrEE-3D URL is not valid: not.valid")
	assert.False(t, controllers.Ogree3D.IsConnected())
	assert.NotEqual(t, controllers.Ogree3D.URL(), "not.valid")
}

func TestConnect3DDoesNotConnectIfOgree3DIsUnreachable(t *testing.T) {
	err := controllers.Connect3D("localhost:3000")
	assert.ErrorContains(t, err, "OGrEE-3D is not reachable caused by OGrEE-3D (localhost:3000) unreachable\ndial tcp")
	assert.ErrorContains(t, err, "connect: connection refused")
	assert.False(t, controllers.Ogree3D.IsConnected())
	assert.Equal(t, controllers.Ogree3D.URL(), "localhost:3000")
}

func TestConnect3DConnectsToProvidedURL(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := controllers.Connect3D("localhost:3000")
	assert.Nil(t, err)
	assert.True(t, controllers.Ogree3D.IsConnected())
	assert.Equal(t, controllers.Ogree3D.URL(), "localhost:3000")
}

func TestConnect3DConnectsToStateOgreeURLIfNotProvidedURL(t *testing.T) {
	fakeOgree3D(t, "3000")

	controllers.Ogree3D.SetURL("localhost:3000")
	err := controllers.Connect3D("")
	assert.Nil(t, err)
	assert.True(t, controllers.Ogree3D.IsConnected())
	assert.Equal(t, controllers.Ogree3D.URL(), "localhost:3000")
}

func TestConnect3DReturnsErrorIfAlreadyConnectedAndNotUrlProvided(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := controllers.Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, controllers.Ogree3D.IsConnected())

	err = controllers.Connect3D("")
	assert.ErrorContains(t, err, "already connected to OGrEE-3D url: localhost:3000")
	assert.True(t, controllers.Ogree3D.IsConnected())
}

func TestConnect3DReturnsErrorIfAlreadyConnectedAndSameUrlProvided(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := controllers.Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, controllers.Ogree3D.IsConnected())

	err = controllers.Connect3D("localhost:3000")
	assert.ErrorContains(t, err, "already connected to OGrEE-3D url: localhost:3000")
	assert.True(t, controllers.Ogree3D.IsConnected())
}

func TestConnect3DTriesToConnectIfAlreadyConnectedAndDifferentUrlProvided(t *testing.T) {
	wg := fakeOgree3D(t, "3000")

	err := controllers.Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, controllers.Ogree3D.IsConnected())

	err = controllers.Connect3D("localhost:5000")
	assert.ErrorContains(t, err, "OGrEE-3D is not reachable caused by OGrEE-3D (localhost:5000) unreachable\ndial tcp")
	assert.ErrorContains(t, err, "connect: connection refused")
	assert.False(t, controllers.Ogree3D.IsConnected())
	assert.Equal(t, controllers.Ogree3D.URL(), "localhost:5000")

	wg.Wait()
}

func TestConnect3DConnectsIfAlreadyConnectedAndDifferentUrlProvidedIsReachable(t *testing.T) {
	wg := fakeOgree3D(t, "3000")

	err := controllers.Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, controllers.Ogree3D.IsConnected())

	fakeOgree3D(t, "5000")

	err = controllers.Connect3D("localhost:5000")
	assert.Nil(t, err)
	assert.True(t, controllers.Ogree3D.IsConnected())
	assert.Equal(t, controllers.Ogree3D.URL(), "localhost:5000")

	wg.Wait()
}

func TestInformOgree3DOptionalDoesNothingIfOgree3DNotConnected(t *testing.T) {
	require.False(t, controllers.Ogree3D.IsConnected())
	err := controllers.Ogree3D.InformOptional("Interact", -1, map[string]any{})
	assert.Nil(t, err)
}

func TestInformOgree3DOptionalSendDataWhenOgree3DIsConnected(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := controllers.Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, controllers.Ogree3D.IsConnected())

	err = controllers.Ogree3D.InformOptional("Interact", -1, map[string]any{})
	assert.Nil(t, err)
}

func TestInformOgree3DFailsIfOgree3DNotReachable(t *testing.T) {
	require.False(t, controllers.Ogree3D.IsConnected())
	controllers.Ogree3D.SetURL("localhost:3000")
	err := controllers.Ogree3D.Inform("Interact", -1, map[string]any{})
	assert.ErrorContains(t, err, "OGrEE-3D is not reachable caused by OGrEE-3D (localhost:3000) unreachable\ndial tcp")
	assert.ErrorContains(t, err, "connect: connection refused")
}

func TestInformOgree3DEstablishConnectionIfOgree3DIsReachable(t *testing.T) {
	require.False(t, controllers.Ogree3D.IsConnected())

	controllers.Ogree3D.SetURL("localhost:3000")
	fakeOgree3D(t, "3000")

	err := controllers.Ogree3D.Inform("Interact", -1, map[string]any{})
	assert.Nil(t, err)
}

func TestInformOgree3DSendsDataIfEstablishConnectionWithOgree3DAlreadyEstablished(t *testing.T) {
	fakeOgree3D(t, "3000")

	err := controllers.Connect3D("localhost:3000")
	require.Nil(t, err)
	require.True(t, controllers.Ogree3D.IsConnected())

	err = controllers.Ogree3D.Inform("Interact", -1, map[string]any{})
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
		if controllers.Ogree3D.IsConnected() {
			controllers.Ogree3D.Disconnect()
		}
	})

	return &wg
}

// Test GenerateFilteredJson
func TestGenerateFilteredJson(t *testing.T) {
	controllers.State.DrawableJsons = test_utils.GetTestDrawableJson()

	object := map[string]any{
		"name":        "rack",
		"parentId":    "site.building.room",
		"category":    "rack",
		"description": "",
		"domain":      "domain",
		"attributes": map[string]any{
			"color": "aaaaaa",
		},
	}

	filteredObject := controllers.GenerateFilteredJson(object)

	assert.Contains(t, filteredObject, "name")
	assert.Contains(t, filteredObject, "parentId")
	assert.Contains(t, filteredObject, "category")
	assert.Contains(t, filteredObject, "domain")
	assert.NotContains(t, filteredObject, "description")
	assert.Contains(t, filteredObject, "attributes")
	assert.Contains(t, filteredObject["attributes"], "color")
}
