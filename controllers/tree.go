package controllers

//This file describes the 'tree' command
//since it a has more complex algorithm

import (
	"cli/models"
	"fmt"
	"strconv"
	"strings"
)

// New Tree funcs here
func StrayWalk(root **Node, prefix string, depth int) {

	if depth > 0 {
		for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
			node := i.Value.(*Node)

			if i.Next() == nil {
				fmt.Println(prefix+"└──", node.Name)
				StrayWalk(&node, prefix+"    ", depth-1)
			} else {
				fmt.Println(prefix+("├──"), node.Name)
				StrayWalk(&node, prefix+"│   ", depth-1)
			}
		}

		if (*root).Nodes.Len() == 0 && depth > 0 {
			switch (*root).Name {
			case "Device":
				//Get Stray Devices and print them
				StrayAndDomain("stray-devices", prefix, depth)
			case "Sensor":
				//Get Stray Sensors and print them
				r, e := models.Send("GET",
					State.APIURL+"/api/stray-sensors", GetKey(), nil)
				resp := ParseResponse(r, e, "fetch objects")
				if resp != nil {
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				}
			default: //Error, execution should not reach here

			}
			return
		}
	}

}

func RootWalk(root **Node, path string, depth int) {
	org := FindNodeInTree(root, StrToStack("/Organisation"), true)
	fmt.Println("├──" + "Organisation")
	OrganisationWalk(org, "│   ", depth-1)

	logical := FindNodeInTree(root, StrToStack("/Logical"), true)
	fmt.Println("├──" + "Logical")
	LogicalWalk(logical, "│   ", depth-1)

	phys := FindNodeInTree(root, StrToStack("/Physical"), true)
	fmt.Println("└──" + "Physical")
	PhysicalWalk(phys, "    ", path, depth-1)
}

func LogicalWalk(root **Node, prefix string, depth int) {

	if root != nil {
		if depth >= 0 {
			if (*root).Nodes.Len() == 0 {
				switch (*root).Name {
				case "ObjectTemplates":
					//Get All Obj Templates and print them
					r, e := models.Send("GET",
						State.APIURL+"/api/obj-templates", GetKey(), nil)
					resp := ParseResponse(r, e, "fetching objects")
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				case "RoomTemplates":
					//Get All Room Templates and print them
					r, e := models.Send("GET",
						State.APIURL+"/api/room-templates", GetKey(), nil)
					resp := ParseResponse(r, e, "fetching objects")
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				case "BldgTemplates":
					//Get All Bldg Templates and print them
					r, e := models.Send("GET",
						State.APIURL+"/api/bldg-templates", GetKey(), nil)
					resp := ParseResponse(r, e, "fetching objects")
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				case "Groups":
					//Get All Groups and print them
					r, e := models.Send("GET",
						State.APIURL+"/api/groups", GetKey(), nil)
					resp := ParseResponse(r, e, "fetching objects")
					RemoteGetAllWalk(resp["data"].(map[string]interface{}), prefix)
				default: //Error case, execution should not reach here

				}
				return
			}

			for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
				if i.Next() == nil {
					fmt.Println(prefix+"└──", (i.Value.(*Node).Name))
					value := i.Value.(*Node)
					LogicalWalk(&(value), prefix+"    ", depth-1)

				} else {
					fmt.Println(prefix+("├──"), (i.Value.(*Node).Name))
					value := i.Value.(*Node)
					LogicalWalk(&(value), prefix+"│   ", depth-1)

				}
			}
		}
	}

}

func OrganisationWalk(root **Node, prefix string, depth int) {

	if root != nil {
		if depth >= 0 {
			if (*root).Nodes.Len() == 0 {
				switch (*root).Name {
				case "Domain":
					StrayAndDomain("domains", prefix, depth)
				case "Enterprise":
					//Most likely same as Domain case
					//TODO Will have to update this section later on
				}
			}

			for i := (*root).Nodes.Front(); i != nil; i = i.Next() {
				if i.Next() == nil {
					fmt.Println(prefix+"└──", (i.Value.(*Node).Name))
					value := i.Value.(*Node)
					OrganisationWalk(&(value), prefix+"    ", depth-1)

				} else {
					fmt.Println(prefix+("├──"), (i.Value.(*Node).Name))
					value := i.Value.(*Node)
					OrganisationWalk(&(value), prefix+"│   ", depth-1)

				}
			}
		}
	}

}

func PhysicalWalk(root **Node, prefix, path string, depth int) {
	arr := strings.Split(path, "/")
	if len(arr) == 3 {
		//println("DEBUG ENTERED")
		if arr[2] == "Stray" {
			fmt.Println(prefix + "├──Device")
			if depth >= 1 {
				//Get and Print Stray Devices
				StrayAndDomain("stray-devices", prefix+"│   ", depth)

				//Get and Print Stray Sensors
				fmt.Println(prefix + "└──Sensor")
				r1, e1 := models.Send("GET",
					State.APIURL+"/api/stray-sensors", GetKey(), nil)
				resp1 := ParseResponse(r1, e1, "fetch objects")

				if resp1 != nil {
					RemoteGetAllWalk(resp1["data"].(map[string]interface{}),
						prefix+"    ")
				}
			} else { //Extra else block for correct printing
				fmt.Println(prefix + "└──Sensor")
			}

		} else { //Interacting with Tenants
			ObjectAndHierarchWalk(path, prefix, depth)

		}
	}
	if len(arr) == 2 { //Means path== "/Physical"

		var resp map[string]interface{}
		if arr[1] == "Physical" { //Means path== "/Physical"

			//Need to check num tenants before passing the prefix
			//Get and Print Tenants Block

			r, e := models.Send("GET",
				State.APIURL+"/api/tenants", GetKey(), nil)
			resp = ParseResponse(r, e, "fetch objects")
			strayNode := FindNodeInTree(&State.TreeHierarchy,
				StrToStack("/Physical/Stray"), true)

			if length, _ := GetRawObjectsLength(resp); length > 0 {
				fmt.Println(prefix + "├──" + " Stray")
				StrayWalk(strayNode, prefix+"│   ", depth)
			} else {
				fmt.Println(prefix + "└──" + " Stray")
				StrayWalk(strayNode, prefix+"   ", depth)
			}

			if resp != nil {
				if depth == 0 {
					if _, ok := resp["data"]; ok {
						RemoteGetAllWalk(resp["data"].(map[string]interface{}),
							prefix)
					}
					return
				}

			}

			if depth > 0 {
				if _, ok := resp["data"]; ok {
					tenants := GetRawObjects(resp)

					size := len(tenants)
					for idx, tInf := range tenants {
						tenant := tInf.(map[string]interface{})
						ID := tenant["id"].(string)
						depthStr := strconv.Itoa(depth)

						var subPrefix string
						var currPrefix string
						if idx == size-1 {
							subPrefix = prefix + "    "
							currPrefix = prefix + "└──"
						} else {
							subPrefix = prefix + "│   "
							currPrefix = prefix + "├──"
						}

						fmt.Println(currPrefix + tenant["name"].(string))

						//Get Hierarchy for each tenant and walk
						r, e := models.Send("GET",
							State.APIURL+"/api/tenants/"+ID+"/all?limit="+depthStr, GetKey(), nil)
						resp := ParseResponse(r, e, "fetch objects")
						if resp != nil {
							RemoteHierarchyWalk(resp["data"].(map[string]interface{}),
								subPrefix, depth)
						}

					}
				}

			}
		} else { //Means path == "/"

			if depth >= 0 {

				strayNode := FindNodeInTree(&State.TreeHierarchy,
					StrToStack("/Physical/Stray"), true)

				//Get and Print Tenants Block
				r, e := models.Send("GET",
					State.APIURL+"/api/tenants", GetKey(), nil)
				resp = ParseResponse(r, e, "fetch objects")

				//Need to check num tenants before passing the prefix
				if length, _ := GetRawObjectsLength(resp); length > 0 {
					fmt.Println(prefix + "├──" + " Stray")
					StrayWalk(strayNode, prefix+"│   ", depth)
				} else {
					fmt.Println(prefix + "└──" + " Stray")
					StrayWalk(strayNode, prefix+"   ", depth)
				}
				if resp != nil {
					if depth == 0 {
						if _, ok := resp["data"]; ok {
							RemoteGetAllWalk(resp["data"].(map[string]interface{}),
								prefix)
						}
						return
					}

				}

				//If hierarchy happens to be greater than 1
				if depth > 0 && resp != nil {
					if tenants := GetRawObjects(resp); tenants != nil {
						size := len(tenants)
						for idx, tInf := range tenants {
							tenant := tInf.(map[string]interface{})
							ID := tenant["id"].(string)
							depthStr := strconv.Itoa(depth)

							//Get Hierarchy for each tenant and walk
							r, e := models.Send("GET",
								State.APIURL+"/api/tenants/"+ID+"/all?limit="+depthStr, GetKey(), nil)
							resp := ParseResponse(r, e, "fetch objects")

							var subPrefix string
							var currPrefix string
							if idx == size-1 {
								subPrefix = prefix + "    "
								currPrefix = prefix + "└──"
							} else {
								subPrefix = prefix + "│   "
								currPrefix = prefix + "├──"
							}

							fmt.Println(currPrefix + tenant["name"].(string))
							if resp != nil {
								RemoteHierarchyWalk(resp["data"].(map[string]interface{}),
									subPrefix, depth)
							}
						}
					}
				}

			}
		}

	}

	if len(arr) > 3 { //Could still be Stray not sure yet
		if arr[2] == "Stray" && len(arr) <= 4 {
			StrayAndDomain("stray-devices", prefix, depth)
		} else {
			//Get Object hierarchy and walk
			ObjectAndHierarchWalk(path, prefix, depth)
		}
	}
}

// Helper function for TreeWalk commands, this will filter out
// objects that a have a ParentID
func Filter(root map[string]interface{}, depth int, ent string) {
	var arr []interface{}
	var replacement []interface{}
	if root == nil {
		return
	}

	if _, ok := root["objects"]; !ok {
		return
	}

	if _, ok := root["objects"].([]interface{}); !ok {
		return
	}
	arr = root["objects"].([]interface{})
	//length = len(arr)

	for _, m := range arr {
		if object, ok := m.(map[string]interface{}); ok {
			if object["parentId"] == nil {
				//Change m -> result of hierarchal API call
				ext := object["id"].(string) + "/all?limit=" + strconv.Itoa(depth)
				URL := State.APIURL + "/api/" + ent + "/" + ext
				r, _ := models.Send("GET", URL, GetKey(), nil)
				parsed := ParseResponse(r, nil, "Fetch "+ent)
				m = parsed["data"].(map[string]interface{})
				replacement = append(replacement, m)
				//Disp(m.(map[string]interface{}))
			}
		}
	}

	root["objects"] = replacement
}

func ObjectAndHierarchWalk(path, prefix string, depth int) {
	depthStr := strconv.Itoa(depth + 1)

	//Need to convert path to URL then append /all?limit=depthStr
	_, urls := CheckPathOnline(path)
	r, e := models.Send("GET", urls, GetKey(), nil)
	//WE need to get the Object in order for us to create
	//the correct GET /all?limit=depthStr URL
	//we get the object category and ID in the JSON response

	parsed := ParseResponse(r, e, "get object")
	if parsed != nil {

		obj := parsed["data"].(map[string]interface{})
		cat := obj["category"].(string)
		ID := obj["id"].(string)
		URL := State.APIURL + "/api/" +
			cat + "s/" + ID + "/all?limit=" + depthStr
		r1, e1 := models.Send("GET", URL, GetKey(), nil)
		parsedRoot := ParseResponse(r1, e1, "get object hierarchy")
		if parsedRoot != nil {
			if _, ok := parsedRoot["data"]; ok {
				RemoteHierarchyWalk(
					parsedRoot["data"].(map[string]interface{}),
					prefix, depth+1)
			}

		}
	}
}

// Gets all objects and filters out the objs with PID and adds
// the respective hierarchies of each object and walks them
// (meant for walking stray and domain objs)
func StrayAndDomain(ent, prefix string, depth int) {
	//Do the call, filter and perform remote
	//hierarchy walk
	//Get All Domains OR Stray Devices and print them
	r, e := models.Send("GET",
		State.APIURL+"/api/"+ent, GetKey(), nil)
	resp := ParseResponse(r, e, "fetching objects")
	if resp != nil {
		if _, ok := resp["data"]; ok {
			data := resp["data"].(map[string]interface{})
			Filter(data, depth, ent)

			if objects, ok := data["objects"]; ok {
				length := len(objects.([]interface{}))
				for i, obj := range objects.([]interface{}) {
					if m, ok := obj.(map[string]interface{}); ok {
						subname := m["name"].(string)

						if i == length-1 {
							fmt.Println(prefix+"└──", subname)
							RemoteHierarchyWalk(m, prefix+"    ", depth-1)
						} else {
							fmt.Println(prefix+("├──"), subname)
							RemoteHierarchyWalk(m, prefix+"│   ", depth-1)
						}
					}

				}
			}
		}

	}

}

func RemoteGetAllWalk(root map[string]interface{}, prefix string) {
	var arr []interface{}
	var length int
	if root == nil {
		return
	}

	if _, ok := root["objects"]; !ok {
		return
	}

	if _, ok := root["objects"].([]interface{}); !ok {
		return
	}
	arr = root["objects"].([]interface{})
	length = len(arr)

	for i, m := range arr {
		var subname string
		if n, ok := m.(map[string]interface{})["name"].(string); ok {
			subname = n
		} else {
			subname = m.(map[string]interface{})["slug"].(string)
		}

		if i == length-1 {
			fmt.Println(prefix+"└──", subname)

		} else {
			fmt.Println(prefix+("├──"), subname)
		}
	}
}

func RemoteHierarchyWalk(root map[string]interface{}, prefix string, depth int) {

	if depth == 0 || root == nil {
		return
	}
	if infants, ok := root["children"]; !ok || infants == nil {
		return
	}

	//name := root["name"].(string)
	//println(prefix + name)

	//or cast to []interface{}
	arr := root["children"].([]interface{})

	//or cast to []interface{}
	length := len(arr)

	//or cast to []interface{}
	for i, mInf := range arr {
		m := mInf.(map[string]interface{})
		subname := m["name"].(string)

		if i == length-1 {
			fmt.Println(prefix+"└──", subname)
			RemoteHierarchyWalk(m, prefix+"    ", depth-1)
		} else {
			fmt.Println(prefix+("├──"), subname)
			RemoteHierarchyWalk(m, prefix+"│   ", depth-1)
		}
	}
}
