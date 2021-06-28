package controllers

import (
	"cli/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func Execute() {
	println("Congratulations, nobody cares")
	return
}

func PWD() {
	println(State.CurrPath)
}

func Disp(x map[string]interface{}) {
	/*for i, k := range x {
		println("We got: ", i, " and ", k)
	}*/

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

func PostObj(entity string, data map[string]interface{}) {

	resp, e := models.Send("POST",
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
	resp, e := models.Send("DELETE",
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

	resp, e := models.Send("GET", URL, nil)
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

		resp, e := models.Send("PUT", URL, data)
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

func LS(x string) {
	switch x {
	case "", ".":
		DispAtLevel(&State.TreeHierarchy, *StrToStack(State.CurrPath))
	default:
		DispAtLevel(&State.TreeHierarchy, *StrToStack(x))
	}
}

func DispTree1() {
	//DispTree1()
}

func CD(x string) {
	if x == ".." {
		//strings.Split(State.CurrPath, "/")
		//strings.
		//State.CurrPath =
	}
	State.CurrPath += "/" + x
}

func Help() {
	fmt.Printf(`A Shell interface to the API and your datacenter visualisation solution`)
}
