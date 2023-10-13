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
	pathutil "path"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

func PWD() string {
	println(State.CurrPath)
	return State.CurrPath
}

func PostObj(ent int, entity string, data map[string]any) error {
	resp, err := API.Request("POST", "/api/"+entity+"s", data, http.StatusCreated)
	if err != nil {
		return err
	}

	if ent != TAG && IsInObjForUnity(entity) {
		entInt := EntityStrToInt(entity)
		Ogree3D.InformOptional("PostObj", entInt, map[string]any{"type": "create", "data": resp.Body["data"]})
	}

	return nil
}

func ObjectUrl(path string, depth int) (string, error) {
	prefix, id, err := models.SplitPath(path)
	if err != nil {
		return "", err
	}
	var baseUrl string
	switch prefix {
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
	case models.DomainsPath:
		baseUrl = "/api/domains"
	default:
		return "", fmt.Errorf("invalid object path")
	}
	baseUrl += "/" + id
	params := url.Values{}
	if depth > 0 {
		baseUrl += "/all"
		params.Add("limit", strconv.Itoa(depth))
	}
	parsedUrl, _ := url.Parse(baseUrl)
	parsedUrl.RawQuery = params.Encode()
	return parsedUrl.String(), nil
}

func ObjectUrlGeneric(path string, depth int, filters map[string]string) (string, error) {
	params := url.Values{}
	prefix, id, err := models.SplitPath(path)
	if err != nil {
		return "", err
	}
	switch prefix {
	case models.StayPath:
		params.Add("namespace", "physical.stray")
		params.Add("id", id)
	case models.PhysicalPath:
		params.Add("namespace", "physical.hierarchy")
		params.Add("id", id)
	case models.ObjectTemplatesPath:
		params.Add("namespace", "logical.objtemplate")
		params.Add("slug", id)
	case models.RoomTemplatesPath:
		params.Add("namespace", "logical.roomtemplate")
		params.Add("slug", id)
	case models.BuildingTemplatesPath:
		params.Add("namespace", "logical.bldgtemplate")
		params.Add("slug", id)
	case models.TagsPath:
		params.Add("namespace", "logical.tag")
		params.Add("slug", id)
	case models.GroupsPath:
		params.Add("namespace", "logical")
		params.Add("category", "group")
		params.Add("id", id)
	case models.DomainsPath:
		params.Add("namespace", "organisational")
		params.Add("id", id)
	default:
		return "", fmt.Errorf("invalid object path")
	}
	if depth > 0 {
		params.Add("limit", strconv.Itoa(depth))
	}
	for key, value := range filters {
		params.Add(key, value)
	}
	url, _ := url.Parse("/api/objects")
	url.RawQuery = params.Encode()
	return strings.ReplaceAll(url.String(), "%2A", "*"), nil
}

func PollObject(path string) (map[string]any, error) {
	return C.PollObjectWithChildren(path, 0)
}

func lsObjectsWithoutFilters(path string) ([]map[string]any, error) {
	n, err := Tree(path, 1)
	if err != nil {
		return nil, err
	}
	objects := []map[string]any{}
	for _, child := range n.Children {
		if child.Obj != nil {
			if strings.HasPrefix(path, "/Logical/Groups") {
				child.Obj["name"] = strings.ReplaceAll(child.Obj["id"].(string), ".", "/")
			}
			objects = append(objects, child.Obj)
		} else {
			objects = append(objects, map[string]any{"name": child.Name})
		}
	}
	return objects, nil
}

func lsObjectsWithFilters(path string, filters map[string]string) ([]map[string]any, error) {
	url, err := ObjectUrlGeneric(path+"/*", 0, filters)
	if err != nil {
		return nil, fmt.Errorf("cannot use filters at this location")
	}
	resp, err := API.Request("GET", url, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	objectsAny := resp.Body["data"].([]any)
	objects := []map[string]any{}
	for _, objAny := range objectsAny {
		obj, ok := objAny.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid response from API on GET %s", url)
		}
		objects = append(objects, obj)
	}
	return objects, nil
}

func filterObjectsWithoutAttr(objects []map[string]any, attr string) []map[string]any {
	remainingObjects := []map[string]any{}
	for _, obj := range objects {
		_, hasAttr := utils.ObjectAttr(obj, attr)
		if hasAttr {
			remainingObjects = append(remainingObjects, obj)
		}
	}
	return remainingObjects
}

func objectsAreSortable(objects []map[string]any, attr string) bool {
	for i := 1; i < len(objects); i++ {
		val0, _ := utils.ObjectAttr(objects[0], attr)
		vali, _ := utils.ObjectAttr(objects[i], attr)
		_, comparable := utils.CompareVals(val0, vali)
		if !comparable {
			return false
		}
	}
	return true
}

func Ls(path string, filters map[string]string, sortAttr string) ([]map[string]any, error) {
	var objects []map[string]any
	var err error
	if len(filters) == 0 {
		objects, err = lsObjectsWithoutFilters(path)
		if err != nil {
			return nil, err
		}
	} else {
		objects, err = lsObjectsWithFilters(path, filters)
		if err != nil {
			return nil, err
		}
	}
	if sortAttr != "" {
		objects = filterObjectsWithoutAttr(objects, sortAttr)
		if !objectsAreSortable(objects, sortAttr) {
			return nil, fmt.Errorf("objects cannot be sorted according to this attribute")
		}
	}
	less := func(i, j int) bool {
		if sortAttr != "" {
			vali, _ := utils.ObjectAttr(objects[i], sortAttr)
			valj, _ := utils.ObjectAttr(objects[j], sortAttr)
			res, _ := utils.CompareVals(vali, valj)
			return res
		}
		return utils.NameOrSlug(objects[i]) < utils.NameOrSlug(objects[j])
	}
	sort.Slice(objects, less)
	return objects, nil
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
	url, err := ObjectUrl(path, 0)
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
	URL, err := ObjectUrl(Path, 0)
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
	if models.IsHierarchical(Path) && IsInObjForUnity(entity) {
		entInt := EntityStrToInt(entity)
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
		fmt.Println(EntityToString(k))
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
	for name, _ := range userFuncs {
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

func CD(path string) error {
	if State.DebugLvl >= 3 {
		println("THE PATH: ", path)
	}
	_, err := Tree(path, 0)
	if err != nil {
		return err
	}
	State.PrevPath = State.CurrPath
	State.CurrPath = path
	return nil
}

func Help(entry string) {
	var path string
	entry = strings.TrimSpace(entry)
	switch entry {
	case "ls", "pwd", "print", "cd", "tree", "get", "clear",
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

func CreateObject(path string, ent int, data map[string]interface{}) error {
	var attr map[string]interface{}
	var parent map[string]interface{}

	ogPath := path
	path = pathutil.Dir(path)
	name := pathutil.Base(ogPath)
	if name == "." || name == "" {
		l.GetWarningLogger().Println("Invalid path name provided for OCLI object creation")
		return fmt.Errorf("Invalid path name provided for OCLI object creation")
	}

	data["name"] = name
	data["category"] = EntityToString(ent)
	data["description"] = []interface{}{}

	//Retrieve Parent
	if ent != SITE && ent != STRAY_DEV && ent != STRAYSENSOR {
		var err error
		parent, err = PollObject(path)
		if err != nil {
			return err
		}
		if parent == nil && (ent != DOMAIN || path != "/Organisation/Domain") {
			return fmt.Errorf("parent not found")
		}
	}

	if ent != DOMAIN {
		if parent != nil {
			data["domain"] = parent["domain"]
		} else {
			data["domain"] = State.Customer
		}
	}

	var err error
	switch ent {
	case DOMAIN:
		if parent != nil {
			data["parentId"] = parent["id"]
		} else {
			data["parentId"] = ""
		}

	case SITE:
		//Default values
		//data["parentId"] = parent["id"]
		data["attributes"] = map[string]interface{}{}

	case BLDG:
		attr = data["attributes"].(map[string]interface{})

		//Check for template
		if _, ok := attr["template"]; ok {
			err := GetOCLIAtrributesTemplateHelper(attr, data, BLDG)
			if err != nil {
				return err
			}
		} else {
			//Serialise size and posXY manually instead
			if _, ok := attr["size"].(string); ok {
				attr["size"] = serialiseAttr(attr, "size")
			} else {
				attr["size"] = serialiseAttr2(attr, "size")
			}

			//Since template was not provided, set it empty
			attr["template"] = ""
		}

		if attr["size"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid size value for creating building")
				return fmt.Errorf("Invalid size attribute provided." +
					" \nIt must be an array/list/vector with 3 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		if _, ok := attr["posXY"].(string); ok {
			attr["posXY"] = serialiseAttr(attr, "posXY")
		} else {
			attr["posXY"] = serialiseAttr2(attr, "posXY")
		}

		if attr["posXY"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXY value for creating building")
				return fmt.Errorf("Invalid posXY attribute provided." +
					" \nIt must be an array/list/vector with 2 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		//Check rotation
		if _, ok := attr["rotation"].(float64); ok {
			attr["rotation"] =
				strconv.FormatFloat(attr["rotation"].(float64), 'f', -1, 64)
		}

		attr["posXYUnit"] = "m"
		attr["sizeUnit"] = "m"
		attr["heightUnit"] = "m"
		//attr["height"] = 0 //Should be set from parser by default
		data["parentId"] = parent["id"]

	case ROOM:
		attr = data["attributes"].(map[string]interface{})

		baseAttrs := map[string]interface{}{
			"floorUnit": "t",
			"posXYUnit": "m", "sizeUnit": "m",
			"heightUnit": "m"}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		//NOTE this function also assigns value for "size" attribute
		err := GetOCLIAtrributesTemplateHelper(attr, data, ent)
		if err != nil {
			return err
		}

		if _, ok := attr["posXY"].(string); ok {
			attr["posXY"] = serialiseAttr(attr, "posXY")
		} else {
			attr["posXY"] = serialiseAttr2(attr, "posXY")
		}

		if attr["posXY"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXY value for creating room")
				return fmt.Errorf("Invalid posXY attribute provided." +
					" \nIt must be an array/list/vector with 2 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		//Check rotation
		if _, ok := attr["rotation"].(float64); ok {
			attr["rotation"] =
				strconv.FormatFloat(attr["rotation"].(float64), 'f', -1, 64)
		}

		if attr["size"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid size value for creating room")
				return fmt.Errorf("Invalid size attribute provided." +
					" \nIt must be an array/list/vector with 3 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		data["parentId"] = parent["id"]
		data["attributes"] = attr
		if State.DebugLvl >= 3 {
			println("DEBUG VIEW THE JSON")
			Disp(data)
		}

	case RACK, CORRIDOR:
		attr = data["attributes"].(map[string]interface{})
		//Save rotation because it gets overwritten by
		//GetOCLIAtrributesTemplateHelper()
		rotation := attr["rotation"].([]float64)

		baseAttrs := map[string]interface{}{
			"sizeUnit":   "cm",
			"heightUnit": "U",
		}
		if ent == CORRIDOR {
			baseAttrs["heightUnit"] = "cm"
		}

		MergeMaps(attr, baseAttrs, false)

		//If user provided templates, get the JSON
		//and parse into templates
		err := GetOCLIAtrributesTemplateHelper(attr, data, ent)
		if err != nil {
			return err
		}

		if attr["size"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid size value for creating rack")
				return fmt.Errorf("Invalid size attribute/template provided." +
					" \nThe size must be an array/list/vector with " +
					"3 elements." + "\n\nIf you have provided a" +
					" template, please check that you are referring to " +
					"an existing template" +
					"\n\nFor more information " +
					"please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		//Serialise posXY if given
		if _, ok := attr["posXYZ"].(string); ok {
			attr["posXYZ"] = serialiseAttr(attr, "posXYZ")
		} else {
			attr["posXYZ"] = serialiseAttr2(attr, "posXYZ")
		}

		//Restore the rotation overwritten
		//by the helper func
		attr["rotation"] = fmt.Sprintf("{\"x\":%v, \"y\":%v, \"z\":%v}", rotation[0], rotation[1], rotation[2])

		if attr["posXYZ"] == "" {
			if State.DebugLvl > 0 {
				l.GetErrorLogger().Println(
					"User gave invalid posXYZ value for creating rack")
				return fmt.Errorf("Invalid posXYZ attribute provided." +
					" \nIt must be an array/list/vector with 2 or 3 elements." +
					" Please refer to the wiki or manual reference" +
					" for more details on how to create objects " +
					"using this syntax")
			}
			return nil
		}

		data["parentId"] = parent["id"]
		data["attributes"] = attr

	case DEVICE:
		attr = data["attributes"].(map[string]interface{})

		//Special routine to perform on device
		//based on if the parent has a "slot" attribute

		//First check if attr has only posU & sizeU
		//reject if true while also converting sizeU to string if numeric
		//if len(attr) == 2 {
		if sizeU, ok := attr["sizeU"]; ok {
			sizeUValid := checkNumeric(attr["sizeU"])

			if _, ok := attr["template"]; !ok && sizeUValid == false {
				l.GetWarningLogger().Println("Invalid template / sizeU parameter provided for device ")
				return fmt.Errorf("Please provide a valid device template or sizeU")
			}

			//Convert block
			//And Set height
			if _, ok := sizeU.(int); ok {
				attr["sizeU"] = strconv.Itoa(sizeU.(int))
				attr["height"] = strconv.FormatFloat(
					(float64(sizeU.(int)) * 44.5), 'G', -1, 64)
			} else if _, ok := sizeU.(float64); ok {
				attr["sizeU"] = strconv.FormatFloat(sizeU.(float64), 'G', -1, 64)
				attr["height"] = strconv.FormatFloat(sizeU.(float64)*44.5, 'G', -1, 64)
			}
			//End of convert block
			if _, ok := attr["slot"]; ok {
				l.GetWarningLogger().Println("Invalid device syntax encountered")
				return fmt.Errorf("Invalid device syntax: If you have provided a template, it was not found")
			}
		}
		//}

		var slot map[string]any
		//Process the posU/slot attribute
		if x, ok := attr["posU/slot"]; ok {
			delete(attr, "posU/slot")
			if _, err := strconv.Atoi(x.(string)); err == nil {
				attr["posU"] = x
				attr["slot"] = ""
			} else {
				attr["slot"] = x
			}
			slot, err = GetSlot(parent, x.(string))
			if err != nil {
				return err
			}
		}

		//If user provided templates, get the JSON
		//and parse into templates
		if _, ok := attr["template"]; ok {
			err := GetOCLIAtrributesTemplateHelper(attr, data, DEVICE)
			if err != nil {
				return err
			}
		} else {
			attr["template"] = ""
			if slot != nil {
				size := slot["elemSize"].([]any)
				attr["size"] = fmt.Sprintf(
					"{\"x\":%f, \"y\":%f}", size[0].(float64)/10., size[1].(float64)/10.)
			} else {
				if parAttr, ok := parent["attributes"].(map[string]interface{}); ok {
					if rackSize, ok := parAttr["size"]; ok {
						attr["size"] = rackSize
					}
				}
			}
		}
		//End of device special routine

		baseAttrs := map[string]interface{}{
			"orientation": "front",
			"sizeUnit":    "mm",
			"heightUnit":  "mm",
		}

		MergeMaps(attr, baseAttrs, false)

		data["parentId"] = parent["id"]
		data["attributes"] = attr

	case GROUP:
		//name, category, domain, pid
		data["parentId"] = parent["id"]
		attr := data["attributes"].(map[string]interface{})

		groups := strings.Join(attr["content"].([]string), ",")
		attr["content"] = groups

	case STRAYSENSOR:
		attr = data["attributes"].(map[string]interface{})
		if _, ok := attr["template"]; ok {
			//GetOCLIAtrributesTemplateHelper(attr, data, DEVICE)
			tmpl, err := fetchTemplate(attr["template"].(string), STRAYSENSOR)
			if err != nil {
				return err
			}
			MergeMaps(attr, tmpl, true)
		} else {
			attr["template"] = ""
		}

	case STRAY_DEV:
		attr = data["attributes"].(map[string]interface{})
		if _, ok := attr["template"]; ok {
			err := GetOCLIAtrributesTemplateHelper(attr, data, DEVICE)
			if err != nil {
				return err
			}
		} else {
			attr["template"] = ""
		}

	default:
		//Execution should not reach here!
		return fmt.Errorf("Invalid Object Specified!")
	}

	//Stringify the attributes if not already
	if _, ok := data["attributes"]; ok {
		if attributes, ok := data["attributes"].(map[string]interface{}); ok {
			for i := range attributes {
				attributes[i] = Stringify(attributes[i])
			}
		}
	}

	//Because we already stored the string conversion in category
	//we can do the conversion for templates here
	data["category"] = strings.Replace(data["category"].(string), "_", "-", 1)

	err = PostObj(ent, data["category"].(string), data)
	if err != nil {
		return err
	}
	return nil
}

// If user provided templates, get the JSON
// and parse into templates
func GetOCLIAtrributesTemplateHelper(attr, data map[string]interface{}, ent int) error {
	//Inner func declaration used for importing
	//data from templates
	attrSerialiser := func(someVal interface{}, idx string, ent int) string {
		if x, ok := someVal.(int); ok {
			if ent == DEVICE || ent == ROOM || ent == BLDG {
				return strconv.Itoa(x)
			}
			return strconv.Itoa(x / 10)
		} else if x, ok := someVal.(float64); ok {
			if ent == DEVICE || ent == ROOM || ent == BLDG {
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
			if ent == ROOM {
				tInt = ROOMTMPL
			} else if ent == BLDG {
				tInt = BLDGTMPL
			} else {
				tInt = OBJTMPL
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

				if ent == DEVICE {
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

				} else if ent == ROOM {
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

				} else if ent == BLDG {
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
					if ent != BLDG {
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
		if ent != CORRIDOR {
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
		category := EntityStrToInt(obj["category"].(string))
		if models.IsNonHierarchical(path) || category == SITE || category == BLDG || category == ROOM {
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
		return CD(path)
	} else {
		fmt.Println("Focus is now empty")
	}

	return nil
}

func LinkObject(source string, destination string, posUOrSlot string) error {
	sourceUrl, err := ObjectUrl(source, 0)
	if err != nil {
		return err
	}
	_, destId, err := models.SplitPath(destination)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(sourceUrl, "/api/stray-objects/") {
		return fmt.Errorf("only stray objects can be linked")
	}
	payload := map[string]any{"parentId": destId}
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
	sourceUrl, err := ObjectUrl(path, 0)
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

func CreateTag(slug, color string) error {
	jsonData := map[string]any{
		"slug":        slug,
		"description": slug, // the description is initially set with the value of the slug
		"color":       color,
	}

	_, err := API.Request("POST", "/api/tags", jsonData, http.StatusCreated)
	if err != nil {
		return err
	}

	return nil
}

func SetClipBoard(x []string) ([]string, error) {
	State.ClipBoard = x
	var data map[string]interface{}

	if len(x) == 0 { //This means deselect
		data = map[string]interface{}{"type": "select", "data": "[]"}
		err := Ogree3D.InformOptional("SetClipBoard", -1, data)
		if err != nil {
			return nil, fmt.Errorf("cannot reset clipboard : %s", err.Error())
		}
	} else {
		//Verify paths
		arr := []string{}
		for _, val := range x {
			obj, err := C.GetObject(val)
			if err != nil {
				return nil, err
			}
			id, ok := obj["id"].(string)
			if ok {
				arr = append(arr, id)
			}
		}
		serialArr := "[\"" + strings.Join(arr, "\",\"") + "\"]"
		data = map[string]interface{}{"type": "select", "data": serialArr}
		err := Ogree3D.InformOptional("SetClipBoard", -1, data)
		if err != nil {
			return nil, fmt.Errorf("cannot set clipboard : %s", err.Error())
		}
	}
	return State.ClipBoard, nil
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
	if objType == ROOMTMPL {
		url = "/api/room_templates/"
	} else if objType == BLDGTMPL {
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
