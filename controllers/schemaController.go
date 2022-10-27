package controllers

//Auxillary functions used to verify the API responses are valid according
//to the specification and other response parsing functions

import (
	l "cli/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func ParseResponse(resp *http.Response, e error, purpose string) map[string]interface{} {
	ans := map[string]interface{}{}

	if e != nil {
		l.GetWarningLogger().Println("Error while sending "+purpose+" to server: ", e)
		if State.DebugLvl > 0 {
			println("There was an error!")
		}
		return nil
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		if State.DebugLvl > 0 {
			println("Error: " + err.Error() + " Now Exiting")
		}

		l.GetErrorLogger().Println("Error while trying to read server response: ", err)
		if purpose == "POST" || purpose == "Search" {
			os.Exit(-1)
		}
		return nil
	}
	json.Unmarshal(bodyBytes, &ans)
	return ans
}

//Checks the map as x["data"].(map[string]interface{})["objects"]
func GetRawObjectsLength(x map[string]interface{}) (int, error) {
	if objs := GetRawObjects(x); objs != nil {
		return len(objs), nil
	}
	return -1, fmt.Errorf("Response did not meet schema spec")
}

func GetRawObjects(x map[string]interface{}) []interface{} {
	if x != nil {
		if dataInf, ok := x["data"]; ok {
			if data, ok := dataInf.(map[string]interface{}); ok {
				if objInf, ok := data["objects"]; ok {
					if objects, ok := objInf.([]interface{}); ok {
						return objects
					}
				}
			}
		}
	}
	return nil
}
