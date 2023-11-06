package controllers

import (
	"cli/commands"
	l "cli/logger"
	"cli/models"
	"cli/readline"
	"cli/utils"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

func PWD() string {
	println(State.CurrPath)
	return State.CurrPath
}

func UnfoldPath(path string) ([]string, error) {
	if strings.Contains(path, "*") {
		_, subpaths, err := C.GetObjectsWildcard(path)
		return subpaths, err
	}
	if path == "_" {
		return State.ClipBoard, nil
	}
	return []string{path}, nil
}

func (controller Controller) ObjectUrl(pathStr string, depth int) (string, error) {
	path, err := controller.SplitPath(pathStr)
	if err != nil {
		return "", err
	}

	var baseUrl string
	switch path.Prefix {
	case models.StayPath:
		baseUrl = "/api/stray-objects"
	case models.PhysicalPath:
		baseUrl = "/api/hierarchy-objects"
	case models.ObjectTemplatesPath:
		baseUrl = "/api/obj-templates"
	case models.RoomTemplatesPath:
		baseUrl = "/api/room-templates"
	case models.BuildingTemplatesPath:
		baseUrl = "/api/bldg-templates"
	case models.GroupsPath:
		baseUrl = "/api/groups"
	case models.TagsPath:
		baseUrl = "/api/tags"
	case models.LayersPath:
		baseUrl = LayersURL
	case models.DomainsPath:
		baseUrl = "/api/domains"
	default:
		return "", fmt.Errorf("invalid object path")
	}
	baseUrl += "/" + path.ObjectID
	params := url.Values{}
	if depth > 0 {
		baseUrl += "/all"
		params.Add("limit", strconv.Itoa(depth))
	}
	parsedUrl, _ := url.Parse(baseUrl)
	parsedUrl.RawQuery = params.Encode()
	return parsedUrl.String(), nil
}

func (controller Controller) ObjectUrlGeneric(pathStr string, depth int, filters map[string]string) (string, error) {
	params := url.Values{}
	path, err := controller.SplitPath(pathStr)
	if err != nil {
		return "", err
	}

	if filters == nil {
		filters = map[string]string{}
	}

	if path.Layer != nil {
		path.Layer.ApplyFilters(filters)
	}

	switch path.Prefix {
	case models.StayPath:
		params.Add("namespace", "physical.stray")
		params.Add("id", path.ObjectID)
	case models.PhysicalPath:
		params.Add("namespace", "physical.hierarchy")
		params.Add("id", path.ObjectID)
	case models.ObjectTemplatesPath:
		params.Add("namespace", "logical.objtemplate")
		params.Add("slug", path.ObjectID)
	case models.RoomTemplatesPath:
		params.Add("namespace", "logical.roomtemplate")
		params.Add("slug", path.ObjectID)
	case models.BuildingTemplatesPath:
		params.Add("namespace", "logical.bldgtemplate")
		params.Add("slug", path.ObjectID)
	case models.TagsPath:
		params.Add("namespace", "logical.tag")
		params.Add("slug", path.ObjectID)
	case models.LayersPath:
		params.Add("namespace", "logical.layer")
		params.Add("slug", path.ObjectID)
	case models.GroupsPath:
		params.Add("namespace", "logical")
		params.Add("category", "group")
		params.Add("id", path.ObjectID)
	case models.DomainsPath:
		params.Add("namespace", "organisational")
		params.Add("id", path.ObjectID)
	default:
		return "", fmt.Errorf("invalid object path")
	}
	if depth > 0 {
		params.Add("limit", strconv.Itoa(depth))
	}

	for key, value := range filters {
		params.Set(key, value)
	}

	url, _ := url.Parse("/api/objects")
	url.RawQuery = params.Encode()
	return strings.ReplaceAll(url.String(), "%2A", "*"), nil
}

func PollObject(path string) (map[string]any, error) {
	return C.PollObjectWithChildren(path, 0)
}

func GetSlot(rack map[string]any, location string) (map[string]any, error) {
	templateAny, ok := rack["attributes"].(map[string]any)["template"]
	if !ok {
		return nil, nil
	}
	template := templateAny.(string)
	if template == "" {
		return nil, nil
	}
	resp, err := API.Request("GET", "/api/obj-templates/"+template, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	slots, ok := resp.Body["data"].(map[string]any)["slots"]
	if !ok {
		return nil, nil
	}
	for _, slotAny := range slots.([]any) {
		slot := slotAny.(map[string]any)
		if slot["location"] == location {
			return slot, nil
		}
	}
	return nil, fmt.Errorf("the slot %s does not exist", location)
}

func UnsetAttribute(path string, attr string) error {
	obj, err := C.GetObject(path)
	if err != nil {
		return err
	}
	delete(obj, "id")
	delete(obj, "lastUpdated")
	delete(obj, "createdDate")
	attributes, hasAttributes := obj["attributes"].(map[string]any)
	if !hasAttributes {
		return fmt.Errorf("object has no attributes")
	}
	delete(attributes, attr)
	url, err := C.ObjectUrl(path, 0)
	if err != nil {
		return err
	}
	_, err = API.Request("PUT", url, obj, http.StatusOK)
	return err
}

// Specific update for deleting elements in an array of an obj
func UnsetInObj(Path, attr string, idx int) (map[string]interface{}, error) {
	var arr []interface{}

	//Check for valid idx
	if idx < 0 {
		return nil,
			fmt.Errorf("Index out of bounds. Please provide an index greater than 0")
	}

	//Get the object
	obj, err := C.GetObject(Path)
	if err != nil {
		return nil, err
	}

	//Check if attribute exists in object
	existing, nested := AttrIsInObj(obj, attr)
	if !existing {
		if State.DebugLvl > ERROR {
			l.GetErrorLogger().Println("Attribute :" + attr + " was not found")
		}
		return nil, fmt.Errorf("Attribute :" + attr + " was not found")
	}

	//Check if attribute is an array
	if nested {
		objAttributes := obj["attributes"].(map[string]interface{})
		if _, ok := objAttributes[attr].([]interface{}); !ok {
			if State.DebugLvl > ERROR {
				println("Attribute is not an array")
			}
			return nil, fmt.Errorf("Attribute is not an array")

		}
		arr = objAttributes[attr].([]interface{})

	} else {
		if _, ok := obj[attr].([]interface{}); !ok {
			if State.DebugLvl > ERROR {
				l.GetErrorLogger().Println("Attribute :" + attr + " was not found")
			}
			return nil, fmt.Errorf("Attribute :" + attr + " was not found")
		}
		arr = obj[attr].([]interface{})
	}

	//Ensure that we can delete elt in array
	if len(arr) == 0 {
		if State.DebugLvl > ERROR {
			println("Cannot delete anymore elements")
		}
		return nil, fmt.Errorf("Cannot delete anymore elements")
	}

	//Perform delete
	if idx >= len(arr) {
		idx = len(arr) - 1
	}
	arr = slices.Delete(arr, idx, idx+1)

	//Save back into obj
	if nested {
		obj["attributes"].(map[string]interface{})[attr] = arr
	} else {
		obj[attr] = arr
	}

	entity := obj["category"].(string)
	URL, err := C.ObjectUrl(Path, 0)
	if err != nil {
		return nil, err
	}

	resp, err := API.Request("PUT", URL, obj, http.StatusOK)
	if err != nil {
		return nil, err
	}

	message := map[string]interface{}{
		"type": "modify", "data": resp.Body["data"]}

	//Update and inform unity
	if models.IsPhysical(Path) && IsInObjForUnity(entity) {
		entInt := models.EntityStrToInt(entity)
		Ogree3D.InformOptional("UpdateObj", entInt, message)
	}

	return nil, nil
}

func Clear() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Printf("\033[2J\033[H")
	}
}

func LSOG() error {
	fmt.Println("********************************************")
	fmt.Println("OGREE Shell Information")
	fmt.Println("********************************************")

	fmt.Println("USER EMAIL:", State.User.Email)
	fmt.Println("API URL:", State.APIURL+"/api/")
	fmt.Println("OGrEE-3D URL:", Ogree3D.URL())
	fmt.Println("OGrEE-3D connected: ", Ogree3D.IsConnected())
	fmt.Println("BUILD DATE:", BuildTime)
	fmt.Println("BUILD TREE:", BuildTree)
	fmt.Println("BUILD HASH:", BuildHash)
	fmt.Println("COMMIT DATE: ", GitCommitDate)
	fmt.Println("CONFIG FILE PATH: ", State.ConfigPath)
	fmt.Println("LOG PATH:", "./log.txt")
	fmt.Println("HISTORY FILE PATH:", State.HistoryFilePath)
	fmt.Println("DEBUG LEVEL: ", State.DebugLvl)

	fmt.Printf("\n\n")
	fmt.Println("********************************************")
	fmt.Println("API Information")
	fmt.Println("********************************************")

	//Get API Information here
	resp, err := API.Request("GET", "/api/version", nil, http.StatusOK)
	if err != nil {
		return err
	}
	apiInfo, ok := resp.Body["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid response from API on GET /api/version")
	}
	fmt.Println("BUILD DATE:", apiInfo["BuildDate"])
	fmt.Println("BUILD TREE:", apiInfo["BuildTree"])
	fmt.Println("BUILD HASH:", apiInfo["BuildHash"])
	fmt.Println("COMMIT DATE: ", apiInfo["CommitDate"])
	fmt.Println("CUSTOMER: ", apiInfo["Customer"])
	return nil
}

func LSEnterprise() error {
	resp, err := API.Request("GET", "/api/stats", nil, http.StatusOK)
	if err != nil {
		return err
	}
	DisplayObject(resp.Body)
	return nil
}

// Displays environment variable values
// and user defined variables and funcs
func Env(userVars, userFuncs map[string]interface{}) {
	fmt.Println("Filter: ", State.FilterDisplay)
	fmt.Println()
	fmt.Println("Objects Unity shall be informed of upon update:")
	for _, k := range State.ObjsForUnity {
		fmt.Println(k)
	}
	fmt.Println()
	fmt.Println("Objects Unity shall draw:")
	for _, k := range State.DrawableObjs {
		fmt.Println(models.EntityToString(k))
	}

	fmt.Println()
	fmt.Println("Currently defined user variables:")
	for name, k := range userVars {
		if k != nil {
			fmt.Println("Name:", name, "  Value: ", k)
		}

	}

	fmt.Println()
	fmt.Println("Currently defined user functions:")
	for name := range userFuncs {
		fmt.Println("Name:", name)
	}
}

func GetByAttr(path string, u interface{}) error {
	obj, err := C.GetObjectWithChildren(path, 1)
	if err != nil {
		return err
	}
	cat := obj["category"].(string)
	if cat != "rack" {
		return fmt.Errorf("command may only be performed on rack objects")
	}
	children := obj["children"].([]any)
	devices := infArrToMapStrinfArr(children)
	switch u.(type) {
	case int:
		for i := range devices {
			if attr, ok := devices[i]["attributes"].(map[string]interface{}); ok {
				uStr := strconv.Itoa(u.(int))
				if attr["height"] == uStr {
					DisplayObject(devices[i])
					return nil //What if the user placed multiple devices at same height?
				}
			}
		}
		if State.DebugLvl > NONE {
			println("The 'U' you provided does not correspond to any device in this rack")
		}
	default: //String
		for i := range devices {
			if attr, ok := devices[i]["attributes"].(map[string]interface{}); ok {
				if attr["slot"] == u.(string) {
					DisplayObject(devices[i])
					return nil //What if the user placed multiple devices at same slot?
				}
			}
		}
		if State.DebugLvl > NONE {
			println("The slot you provided does not correspond to any device in this rack")
		}
	}
	return nil
}

func Help(entry string) {
	var path string
	entry = strings.TrimSpace(entry)
	switch entry {
	case "ls", "pwd", "print", "printf", "cd", "tree", "get", "clear",
		"lsog", "grep", "for", "while", "if", "env",
		"cmds", "var", "unset", "selection", commands.Connect3D, "camera", "ui", "hc", "drawable",
		"link", "unlink", "draw", "getu", "getslot", "undraw",
		"lsenterprise":
		path = "./other/man/" + entry + ".md"

	case ">":
		path = "./other/man/focus.md"

	case "+":
		path = "./other/man/plus.md"

	case "=":
		path = "./other/man/equal.md"

	case "-":
		path = "./other/man/minus.md"

	case ".template":
		path = "./other/man/template.md"

	case ".cmds":
		path = "./other/man/cmds.md"

	case ".var":
		path = "./other/man/var.md"

	case "lsobj", "lsten", "lssite", "lsbldg", "lsroom", "lsrack",
		"lsdev", "lsac", "lscorridor", "lspanel", "lssensor", "lscabinet":
		path = "./other/man/lsobj.md"

	default:
		path = "./other/man/default.md"
	}
	text, e := os.ReadFile(utils.ExeDir() + "/" + path)
	if e != nil {
		println("Manual Page not found!")
	} else {
		println(string(text))
	}

}

func DisplayObject(obj map[string]interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	if err := enc.Encode(obj); err != nil {
		log.Fatal(err)
	}
}

// Function is an abstraction of a normal exit
func Exit() {
	//writeHistoryOnExit(&State.sessionBuffer)
	//runtime.Goexit()
	os.Exit(0)
}

// If user provided templates, get the JSON
// and parse into templates
func GetOCLIAtrributesTemplateHelper(attr, data map[string]interface{}, ent int) error {
	//Inner func declaration used for importing
	//data from templates
	attrSerialiser := func(someVal interface{}, idx string, ent int) string {
		if x, ok := someVal.(int); ok {
			if ent == models.DEVICE || ent == models.ROOM || ent == models.BLDG {
				return strconv.Itoa(x)
			}
			return strconv.Itoa(x / 10)
		} else if x, ok := someVal.(float64); ok {
			if ent == models.DEVICE || ent == models.ROOM || ent == models.BLDG {
				return strconv.FormatFloat(x, 'G', -1, 64)
			}
			return strconv.FormatFloat(x/10.0, 'G', -1, 64)
		} else {
			msg := "Warning: Invalid " + idx +
				" value detected in size." +
				" Resorting to default"
			println(msg)
			return "5"
		}
	}

	if q, ok := attr["template"]; ok {
		if qS, ok := q.(string); ok {
			//Determine the type of template
			tInt := 0
			if ent == models.ROOM {
				tInt = models.ROOMTMPL
			} else if ent == models.BLDG {
				tInt = models.BLDGTMPL
			} else {
				tInt = models.OBJTMPL
			} //End of determine block
			tmpl, err := fetchTemplate(qS, tInt)
			if err != nil {
				return err
			}

			//MergeMaps(attr, tmpl, true)
			key := determineStrKey(tmpl, []string{"sizeWDHmm", "sizeWDHm"})

			if sizeInf, ok := tmpl[key].([]interface{}); ok && len(sizeInf) == 3 {
				var xS, yS, zS string
				xS = attrSerialiser(sizeInf[0], "x", ent)
				yS = attrSerialiser(sizeInf[1], "y", ent)
				zS = attrSerialiser(sizeInf[2], "height", ent)

				attr["size"] = "{\"x\":" + xS + ", \"y\":" + yS + "}"
				attr["height"] = zS

				if ent == models.DEVICE {
					attr["sizeUnit"] = "mm"
					attr["heightUnit"] = "mm"
					if tmpx, ok := tmpl["attributes"]; ok {
						if x, ok := tmpx.(map[string]interface{}); ok {
							if tmp, ok := x["type"]; ok {
								if t, ok := tmp.(string); ok {
									if t == "chassis" || t == "server" {
										res := 0
										if val, ok := sizeInf[2].(float64); ok {
											res = int((val / 1000) / RACKUNIT)
										} else if val, ok := sizeInf[2].(int); ok {
											res = int((float64(val) / 1000) / RACKUNIT)
										} else {
											//Resort to default value
											msg := "Warning, invalid value provided for" +
												" sizeU. Defaulting to 5"
											println(msg)
											res = int((5 / 1000) / RACKUNIT)
										}
										attr["sizeU"] = strconv.Itoa(res)

									}
								}
							}
						}
					}

				} else if ent == models.ROOM {
					attr["sizeUnit"] = "m"
					attr["heightUnit"] = "m"

					//Copy additional Room specific attributes
					var tmp []byte
					CopyAttr(attr, tmpl, "technicalArea")
					if _, ok := attr["technicalArea"]; ok {
						//tmp, _ := json.Marshal(attr["technicalArea"])
						attr["technical"] = attr["technicalArea"]
						delete(attr, "technicalArea")
					}

					CopyAttr(attr, tmpl, "axisOrientation")

					CopyAttr(attr, tmpl, "reservedArea")
					if _, ok := attr["reservedArea"]; ok {
						//tmp, _ = json.Marshal(attr["reservedArea"])
						attr["reserved"] = attr["reservedArea"]
						delete(attr, "reservedArea")
					}

					parseReservedTech(attr)

					CopyAttr(attr, tmpl, "separators")
					if _, ok := attr["separators"]; ok {
						tmp, _ = json.Marshal(attr["separators"])
						attr["separators"] = string(tmp)
					}

					CopyAttr(attr, tmpl, "pillars")
					if _, ok := attr["pillars"]; ok {
						tmp, _ = json.Marshal(attr["pillars"])
						attr["pillars"] = string(tmp)
					}

					CopyAttr(attr, tmpl, "floorUnit")
					if _, ok := attr["floorUnit"]; ok {
						if floorUnit, ok := attr["floorUnit"].(string); ok {
							attr["floorUnit"] = floorUnit
						}
					}

					CopyAttr(attr, tmpl, "tiles")
					if _, ok := attr["tiles"]; ok {
						tmp, _ = json.Marshal(attr["tiles"])
						attr["tiles"] = string(tmp)
					}

					CopyAttr(attr, tmpl, "rows")
					if _, ok := attr["rows"]; ok {
						tmp, _ = json.Marshal(attr["rows"])
						attr["rows"] = string(tmp)
					}

					CopyAttr(attr, tmpl, "aisles")
					if _, ok := attr["aisles"]; ok {
						tmp, _ = json.Marshal(attr["aisles"])
						attr["aisles"] = string(tmp)
					}

					CopyAttr(attr, tmpl, "vertices")
					if _, ok := attr["vertices"]; ok {
						tmp, _ = json.Marshal(attr["vertices"])
						attr["vertices"] = string(tmp)
					}

					CopyAttr(attr, tmpl, "colors")
					if _, ok := attr["colors"]; ok {
						tmp, _ = json.Marshal(attr["colors"])
						attr["colors"] = string(tmp)
					}

					CopyAttr(attr, tmpl, "tileAngle")
					if _, ok := attr["tileAngle"]; ok {
						if tileAngle, ok := attr["tileAngle"].(int); ok {
							attr["tileAngle"] = strconv.Itoa(tileAngle)
						}

						if tileAngleF, ok := attr["tileAngle"].(float64); ok {
							tileAngleStr := strconv.FormatFloat(tileAngleF, 'f', -1, 64)
							attr["tileAngle"] = tileAngleStr
						}
					}

				} else if ent == models.BLDG {
					attr["sizeUnit"] = "m"
					attr["heightUnit"] = "m"

				} else {
					attr["sizeUnit"] = "cm"
					attr["heightUnit"] = "cm"
				}

				//Copy Description
				if _, ok := tmpl["description"]; ok {
					if descTable, ok := tmpl["description"].([]interface{}); ok {
						data["description"] = descTable
					} else {
						data["description"] = []interface{}{tmpl["description"]}
					}
				} else {
					data["description"] = []string{}
				}

				//fbxModel section
				if check := CopyAttr(attr, tmpl, "fbxModel"); !check {
					if ent != models.BLDG {
						attr["fbxModel"] = ""
					}

				}

				//Copy orientation if available
				CopyAttr(attr, tmpl, "orientation")

				//Merge attributes if available
				if tmplAttrsInf, ok := tmpl["attributes"]; ok {
					if tmplAttrs, ok := tmplAttrsInf.(map[string]interface{}); ok {
						MergeMaps(attr, tmplAttrs, false)
					}
				}
			} else {
				if State.DebugLvl > 1 {
					println("Warning, invalid size value in template.",
						"Default values will be assigned")
				}

			}

		} else {
			attr["template"] = ""
			if State.DebugLvl > 1 {
				println("Warning: template must be a string that",
					" refers to an existing imported template.",
					q, " will not be used")
			}

			l.GetWarningLogger().Println("Invalid data type used to invoke template")
		}

	} else {
		if ent != models.CORRIDOR {
			attr["template"] = ""
		}
		//Serialise size and posXY if given
		if _, ok := attr["size"].(string); ok {
			attr["size"] = serialiseAttr(attr, "size")
		} else {
			attr["size"] = serialiseAttr2(attr, "size")
		}
	}
	return nil
}

func Connect3D(url string) error {
	return Ogree3D.Connect(url, *State.Terminal)
}

func UIDelay(time float64) error {
	subdata := map[string]interface{}{"command": "delay", "data": time}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return Ogree3D.Inform("HandleUI", -1, data)
}

func UIToggle(feature string, enable bool) error {
	subdata := map[string]interface{}{"command": feature, "data": enable}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return Ogree3D.Inform("HandleUI", -1, data)
}

func UIHighlight(path string) error {
	obj, err := C.GetObject(path)
	if err != nil {
		return err
	}

	subdata := map[string]interface{}{"command": "highlight", "data": obj["id"]}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return Ogree3D.Inform("HandleUI", -1, data)
}

func UIClearCache() error {
	subdata := map[string]interface{}{"command": "clearcache", "data": ""}
	data := map[string]interface{}{"type": "ui", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return Ogree3D.Inform("HandleUI", -1, data)
}

func CameraMove(command string, position []float64, rotation []float64) error {
	subdata := map[string]interface{}{"command": command}
	subdata["position"] = map[string]interface{}{"x": position[0], "y": position[1], "z": position[2]}
	subdata["rotation"] = map[string]interface{}{"x": rotation[0], "y": rotation[1]}
	data := map[string]interface{}{"type": "camera", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return Ogree3D.Inform("HandleUI", -1, data)
}

func CameraWait(time float64) error {
	subdata := map[string]interface{}{"command": "wait"}
	subdata["position"] = map[string]interface{}{"x": 0, "y": 0, "z": 0}
	subdata["rotation"] = map[string]interface{}{"x": 999, "y": time}
	data := map[string]interface{}{"type": "camera", "data": subdata}
	if State.DebugLvl > WARNING {
		Disp(data)
	}

	return Ogree3D.Inform("HandleUI", -1, data)
}

func FocusUI(path string) error {
	var id string
	if path != "" {
		obj, err := C.GetObject(path)
		if err != nil {
			return err
		}
		category := models.EntityStrToInt(obj["category"].(string))
		if !models.IsPhysical(path) || category == models.SITE || category == models.BLDG || category == models.ROOM {
			msg := "You cannot focus on this object. Note you cannot" +
				" focus on Sites, Buildings and Rooms. " +
				"For more information please refer to the help doc  (man >)"
			return fmt.Errorf(msg)
		}
		id = obj["id"].(string)
	} else {
		id = ""
	}

	data := map[string]interface{}{"type": "focus", "data": id}
	err := Ogree3D.Inform("FocusUI", -1, data)
	if err != nil {
		return err
	}

	if path != "" {
		return C.CD(path)
	} else {
		fmt.Println("Focus is now empty")
	}

	return nil
}

func LinkObject(source string, destination string, posUOrSlot string) error {
	sourceUrl, err := C.ObjectUrl(source, 0)
	if err != nil {
		return err
	}
	destPath, err := C.SplitPath(destination)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(sourceUrl, "/api/stray-objects/") {
		return fmt.Errorf("only stray objects can be linked")
	}
	payload := map[string]any{"parentId": destPath.ObjectID}
	if posUOrSlot != "" {
		payload["slot"] = posUOrSlot
	}
	_, err = API.Request("PATCH", sourceUrl+"/link", payload, http.StatusOK)
	if err != nil {
		return err
	}
	return nil
}

func UnlinkObject(path string) error {
	sourceUrl, err := C.ObjectUrl(path, 0)
	if err != nil {
		return err
	}
	_, err = API.Request("PATCH", sourceUrl+"/unlink", nil, http.StatusOK)
	return err
}

func objectCounter(parent map[string]interface{}) int {
	count := 0
	if parent != nil {
		count += 1
		if _, ok := parent["children"]; ok {
			if arr, ok := parent["children"].([]interface{}); ok {
				for _, childInf := range arr {
					if child, ok := childInf.(map[string]interface{}); ok {
						count += objectCounter(child)
					}
				}
			}
			if arr, ok := parent["children"].([]map[string]interface{}); ok {
				for _, child := range arr {
					count += objectCounter(child)
				}
			}
		}
	}
	return count
}

// Unity UI will draw already existing objects
// by retrieving the hierarchy. 'force' bool is useful
// for scripting where the user can 'force' input if
// the num objects to draw surpasses threshold
func Draw(path string, depth int, force bool) error {
	obj, err := C.GetObjectWithChildren(path, depth)
	if err != nil {
		return err
	}

	count := objectCounter(obj)
	okToGo := true
	if count > State.DrawThreshold && !force {
		msg := "You are about to send " + strconv.Itoa(count) +
			" objects to the Unity 3D client. " +
			"Do you want to continue ? (y/n)\n"
		(*State.Terminal).Write([]byte(msg))
		(*State.Terminal).SetPrompt(">")
		ans, _ := (*State.Terminal).Readline()
		if ans != "y" && ans != "Y" {
			okToGo = false
		}
	} else if force {
		okToGo = true
	} else if !force && count > State.DrawThreshold {
		okToGo = false
	}
	if okToGo {
		data := map[string]interface{}{"type": "create", "data": obj}
		//0 to include the JSON filtration
		unityErr := Ogree3D.Inform("Draw", 0, data)
		if unityErr != nil {
			return unityErr
		}
	}
	return nil
}

func Undraw(x string) error {
	var id string
	if x == "" {
		id = ""
	} else {
		obj, err := C.GetObject(x)
		if err != nil {
			return err
		}
		var ok bool
		id, ok = obj["id"].(string)
		if !ok {
			return fmt.Errorf("this object has no id")
		}
	}

	data := map[string]interface{}{"type": "delete", "data": id}

	return Ogree3D.Inform("Undraw", 0, data)
}

func IsEntityDrawable(path string) (bool, error) {
	obj, err := C.GetObject(path)
	if err != nil {
		return false, err
	}
	if catInf, ok := obj["category"]; ok {
		if category, ok := catInf.(string); ok {
			return IsDrawableEntity(category), nil
		}
	}
	return false, nil
}

func IsCategoryAttrDrawable(category string, attr string) bool {
	templateJson := State.DrawableJsons[category]
	if templateJson == nil {
		return true
	}

	switch attr {
	case "id", "name", "category", "parentID",
		"description", "domain", "parentid", "parentId", "tags":
		if val, ok := templateJson[attr]; ok {
			if valBool, ok := val.(bool); ok {
				return valBool
			}
		}
		return false
	default:
		if tmp, ok := templateJson["attributes"]; ok {
			if attributes, ok := tmp.(map[string]interface{}); ok {
				if val, ok := attributes[attr]; ok {
					if valBool, ok := val.(bool); ok {
						return valBool
					}
				}
			}
		}
		return false
	}
}

func IsAttrDrawable(path string, attr string) (bool, error) {
	obj, err := C.GetObject(path)
	if err != nil {
		return false, err
	}
	category := obj["category"].(string)
	return IsCategoryAttrDrawable(category, attr), nil
}

func ShowClipBoard() []string {
	if State.ClipBoard != nil {
		for _, k := range State.ClipBoard {
			println(k)
		}
		return State.ClipBoard
	}
	return nil
}

func LoadTemplate(data map[string]interface{}, filePath string) error {
	var URL string
	if cat := data["category"]; cat == "room" {
		//Room template
		URL = "/api/room-templates"
	} else if cat == "bldg" || cat == "building" {
		//Bldg template
		URL = "/api/bldg-templates"
	} else if cat == "rack" || cat == "device" {
		// Obj template
		URL = "/api/obj-templates"
	} else {
		return fmt.Errorf("this template does not have a valid category. Please add a category attribute with a value of building or room or rack or device")
	}
	_, err := API.Request("POST", URL, data, http.StatusCreated)
	if err != nil {
		return err
	}
	return nil
}

func SetEnv(arg string, val interface{}) {
	switch arg {
	case "Filter":
		if _, ok := val.(bool); !ok {
			msg := "Can only assign bool values for " + arg + " Env Var"
			l.GetWarningLogger().Println(msg)
			if State.DebugLvl > 0 {
				println(msg)
			}
		} else {
			if arg == "Filter" {
				State.FilterDisplay = val.(bool)
			}

			println(arg + " Display Environment variable set")
		}

	default:
		println(arg + " is not an environment variable")
	}
}

// Utility functions
func determineStrKey(x map[string]interface{}, possible []string) string {
	for idx := range possible {
		if _, ok := x[possible[idx]]; ok {
			return possible[idx]
		}
	}
	return "" //The code should not reach this point!
}

// Function called by update node for interact commands (ie label, labelFont)
func InteractObject(path string, keyword string, val interface{}, fromAttr bool) error {
	//First retrieve the object
	obj, err := C.GetObject(path)
	if err != nil {
		return err
	}
	//Verify labelFont has valid values
	if fromAttr == true {
		//Check if the val refers to an attribute field in the object
		//this means to retrieve value from object
		if value, ok := val.(string); ok {

			innerMap := obj["attributes"].(map[string]interface{})

			if _, ok := obj[value]; ok {
				if value == "description" {

					desc := obj["description"].([]interface{})
					val = ""
					//Combine entire the description array into a string
					for i := 0; i < len(desc); i++ {
						if i == 0 {
							val = desc[i].(string)
						} else {
							val = val.(string) + "\n" + desc[i].(string)
						}

					}
				} else {
					val = obj[value]
				}

			} else if _, ok := innerMap[value]; ok {
				val = innerMap[value]
			} else {
				if strings.Contains(value, "description") == true {
					if desc, ok := obj["description"].([]interface{}); ok {
						if len(value) > 11 { //descriptionX format
							//split the number and description
							numStr := strings.Split(value, "description")[1]
							num, e := strconv.Atoi(numStr)
							if e != nil {
								return e
							}

							if num < 0 {
								return fmt.Errorf("Description index must be positive")
							}

							if num >= len(desc) {
								msg := "Description index is out of" +
									" range. The length for this object is: " +
									strconv.Itoa(len(desc))
								return fmt.Errorf(msg)
							}
							val = desc[num]

						} else {
							val = innerMap[value]
						}
					} //Otherwise the description is a string

				} else {
					msg := "The specified attribute does not exist" +
						" in the object. \nPlease view the object" +
						" (ie. $> get) and try again"
					return fmt.Errorf(msg)
				}

			}

		} else {
			return fmt.Errorf("The label value must be a string")
		}
	}

	data := map[string]interface{}{"id": obj["id"],
		"param": keyword, "value": val}
	ans := map[string]interface{}{"type": "interact", "data": data}

	//-1 since its not neccessary to check for filtering
	return Ogree3D.InformOptional("Interact", -1, ans)
}

// Helper function for GetOCLIAttr which retrieves
// template from server if available, this func mainly helps
// to keep code organised
func fetchTemplate(name string, objType int) (map[string]interface{}, error) {
	var url string
	if objType == models.ROOMTMPL {
		url = "/api/room_templates/"
	} else if objType == models.BLDGTMPL {
		url = "/api/bldg_templates/"
	} else {
		url = "/api/obj_templates/"
	}
	url += name
	resp, err := API.Request("GET", url, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	tmplInf, ok := resp.Body["data"]
	if !ok {
		return nil, fmt.Errorf("invalid response on GET %s", url)
	}
	tmpl, ok := tmplInf.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response on GET %s", url)
	}
	return tmpl, nil
}

func randPassword(n int) string {
	const passChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = passChars[rand.Intn(len(passChars))]
	}
	return string(b)
}

func CreateUser(email string, role string, domain string) error {
	password := randPassword(14)
	response, err := API.Request(
		"POST",
		"/api/users",
		map[string]any{
			"email":    email,
			"password": password,
			"roles": map[string]any{
				domain: role,
			},
		},
		http.StatusCreated,
	)
	if err != nil {
		return err
	}
	println(response.message)
	println("password:" + password)
	return nil
}

func AddRole(email string, role string, domain string) error {
	response, err := API.Request("GET", "/api/users", nil, http.StatusOK)
	if err != nil {
		return err
	}
	userList, userListOk := response.Body["data"].([]any)
	if !userListOk {
		return fmt.Errorf("response contains no user list")
	}
	userID := ""
	for _, user := range userList {
		userMap, ok := user.(map[string]any)
		if !ok {
			continue
		}
		userEmail, emailOk := userMap["email"].(string)
		id, idOk := userMap["_id"].(string)
		if emailOk && idOk && userEmail == email {
			userID = id
			break
		}
	}
	if userID == "" {
		return fmt.Errorf("user not found")
	}
	response, err = API.Request("PATCH", fmt.Sprintf("/api/users/%s", userID),
		map[string]any{
			"roles": map[string]any{
				domain: role,
			},
		},
		http.StatusOK,
	)
	if err != nil {
		return err
	}
	println(response.message)
	return nil
}

func ChangePassword() error {
	currentPassword, err := readline.Password("Current password: ")
	if err != nil {
		return err
	}
	newPassword, err := readline.Password("New password: ")
	if err != nil {
		return err
	}
	response, err := API.Request("POST", "/api/users/password/change",
		map[string]any{
			"currentPassword": string(currentPassword),
			"newPassword":     string(newPassword),
		},
		http.StatusOK,
	)
	if err != nil {
		return err
	}
	println(response.message)
	return nil
}

func (controller Controller) SplitPath(pathStr string) (models.Path, error) {
	for _, prefix := range models.PathPrefixes {
		if strings.HasPrefix(pathStr, string(prefix)) {
			id := pathStr[len(prefix):]
			id = strings.ReplaceAll(id, "/", ".")

			var layer *models.Layer
			var err error

			id, layer, err = controller.GetLayer(id)
			if err != nil {
				return models.Path{}, err
			}

			return models.Path{
				Prefix:   prefix,
				ObjectID: id,
				Layer:    layer,
			}, nil
		}
	}

	return models.Path{}, fmt.Errorf("invalid object path")
}
