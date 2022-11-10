package controllers

import "encoding/json"

// Display contents of []map[string]inf array
func DispMapArr(x []map[string]interface{}) {
	for idx := range x {
		println()
		println()
		println("OBJECT: ", idx)
		displayObject(x[idx])
		println()
	}
}

// displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}
