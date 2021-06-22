package cmd

import (
	"cli/utils"
	"encoding/json"
	"io/ioutil"
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

func PostObj(entity string, data map[string]interface{}) {

	resp, e := utils.Send("POST",
		"https://ogree.chibois.net/api/user/"+entity+"s", data)

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

func DeleteObj(entity string, data map[string]interface{}) {
	resp, e := utils.Send("DELETE",
		"https://ogree.chibois.net/api/user/"+entity+"s/"+
			string(data["id"].(string)), nil)
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
