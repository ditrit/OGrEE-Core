package controllers

//Auxillary functions for parsing and verifying
//that the API responses are valid according
//to the specification

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
			println("Error: " + err.Error())
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

// Checks the map as x["data"].(map[string]interface{})["objects"]
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

func LoadArrFromResp(resp map[string]interface{}, idx string) []interface{} {
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if objs, ok1 := data[idx].([]interface{}); ok1 {
			return objs
		}
	}
	return nil
}

func LoadObjectFromInf(x interface{}) (map[string]interface{}, bool) {
	object, ok := x.(map[string]interface{})
	return object, ok
}

// Convert []interface{} array to
// []map[string]interface{} array
func infArrToMapStrinfArr(x []interface{}) []map[string]interface{} {
	ans := []map[string]interface{}{}
	for i := range x {
		ans = append(ans, x[i].(map[string]interface{}))
	}
	return ans
}

// Utility function used by FetchJsonNodes
func strArrToMapStrInfArr(x []string) []map[string]interface{} {
	ans := []map[string]interface{}{}
	for i := range x {
		ans = append(ans, map[string]interface{}{"name": x[i]})
	}
	return ans
}
