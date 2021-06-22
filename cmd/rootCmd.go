package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
)

func Execute() {
	println("Congratulations, nobody cares")
	return
}

func Exit() {
	runtime.Goexit()
}

func PWD(currPath *string) {
	println(*currPath)
}

func Disp(x map[string]interface{}) {
	/*for i, k := range x {
		println("We got: ", i, " and ", k)
	}*/

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func PostObj(entity string, entMap map[string]interface{}) {
	client := &http.Client{}
	key := os.Getenv("apikey")

	entJson, _ := json.Marshal(entMap)
	req, _ := http.NewRequest("POST",
		"https://ogree.chibois.net/api/user/"+entity+"s",
		bytes.NewBuffer(entJson))

	req.Header.Set("Authorization", "Bearer "+key)
	resp, e := client.Do(req)
	println("Response Code: ", resp.Status)
	if e != nil {
		println("There was an error!")
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		println("Error: " + err.Error() + " Now Exiting")
		os.Exit(-1)
	}
	println(string(bodyBytes))
	return
}
