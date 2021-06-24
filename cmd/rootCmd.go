package cmd

import (
	"cli/models"
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
	models.Exit()
	runtime.Goexit()
}

func PWD() {
	println(models.State.CurrPath)
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

func GetObjQ(entity string, data map[string]interface{}) {
	URL := "https://ogree.chibois.net/api/user/" + entity + "s?"

	for i, k := range data {
		if i == "attributes" {
			for j, _ := range k.(map[string]string) {
				URL = URL + "&" + j + "=" + data[i].(map[string]string)[j]
			}
		} else {
			URL = URL + "&" + i + "=" + string(data[i].(string))
		}
	}

	println("Here is URL: ", URL)

	resp, e := utils.Send("GET", URL, nil)
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
}

func UpdateObj(entity string, data map[string]interface{}) {
	println("OK. Attempting to update...")
	if data["id"] != nil {
		URL := "https://ogree.chibois.net/api/user/" + entity + "s/" +
			string(data["id"].(string))

		resp, e := utils.Send("PUT", URL, data)
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
	} else {
		println("Error! Please enter ID of Object to be updated")
	}

}

func LS() {
	models.DispAtLevel(&models.State.TreeHierarchy,
		*models.StrToStack(models.State.CurrPath))
}

//This version prints out comments
func DispTree() {
	models.DispTree()
}

func DispTree1() {
	//models.DispTree1()
}

func CD(x string) {
	if x == ".." {
		//strings.Split(models.State.CurrPath, "/")
		//strings.
		//models.State.CurrPath =
	}
	models.State.CurrPath += "/" + x
}
