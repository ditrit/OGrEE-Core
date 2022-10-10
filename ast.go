package main

import (
	"bytes"
	cmd "cli/controllers"
	"fmt"
	"strings"
)

var dynamicSymbolTable = make(map[string]interface{})
var funcTable = make(map[string]interface{})

func GetFuncTable() map[string]interface{} {
	return funcTable
}

func GetDynamicSymbolTable() map[string]interface{} {
	return dynamicSymbolTable
}

type node interface {
	execute() (interface{}, error)
}

type valueNode struct {
	val interface{}
}

func (n *valueNode) execute() (interface{}, error) {
	return n.val, nil
}

type ast struct {
	statements []node
}

func (a *ast) execute() (interface{}, error) {
	for i, _ := range a.statements {
		_, err := a.statements[i].execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type funcDefNode struct {
	name string
	body node
}

func (n *funcDefNode) execute() (interface{}, error) {
	dynamicSymbolTable[n.name] = n.body
	if cmd.State.DebugLvl >= 3 {
		println("New function ", n.name)
	}
	return nil, nil
}

type funcCallNode struct {
	name string
}

func (n *funcCallNode) execute() (interface{}, error) {
	val, ok := dynamicSymbolTable[n.name]
	if !ok {
		return nil, fmt.Errorf("undefined function ", n.name)
	}
	body, ok := val.(node)
	if !ok {
		return nil, fmt.Errorf("variable %s does not contain a function", n.name)
	}
	return body.execute()
}

type arrNode struct {
	nodes []node
}

func (n *arrNode) execute() (interface{}, error) {
	var r []interface{}
	for i := range n.nodes {
		v, err := n.nodes[i].execute()
		if err != nil {
			return nil, err
		}
		val, err := getFloat(v)
		if err != nil {
			return nil, fmt.Errorf("Array should contain floats")
		}
		r = append(r, val)
	}
	return r, nil
}

type lenNode struct {
	variable string
}

func (n *lenNode) execute() (interface{}, error) {
	val, ok := dynamicSymbolTable[n.variable]
	if !ok {
		return nil, fmt.Errorf("Undefined variable ", n.variable)
	}
	arr, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Variable ", n.variable, " does not contain an array.")
	}
	return len(arr), nil
}

type postObjNode struct {
	entity string
	data   map[string]interface{}
}

func (n *postObjNode) execute() (interface{}, error) {
	return cmd.PostObj(cmd.EntityStrToInt(n.entity), n.entity, n.data)
}

type easyPostNode struct {
	entity string
	path   node
}

func (n *easyPostNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}

	data := make(map[string]interface{})
	/*x, e := ioutil.ReadFile(n.path)
	if e != nil {
		println("Error while opening file! " + e.Error())
		return nil
	}
	json.Unmarshal(x, &data)*/
	data = fileToJSON(path)
	if data == nil {
		return nil, fmt.Errorf("Cannot read json file.")
	}
	return cmd.PostObj(cmd.EntityStrToInt(n.entity), n.entity, data)
}

type helpNode struct {
	entry string
}

func (n *helpNode) execute() (interface{}, error) {
	cmd.Help(n.entry)
	return nil, nil
}

type focusNode struct {
	path node
}

func (n *focusNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.FocusUI(path)
	return nil, nil
}

type cdNode struct {
	path node
}

func (n *cdNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.CD(path), nil
}

type lsNode struct {
	path node
}

func (n *lsNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.LS(path), nil
}

type lsAttrNode struct {
	path node
	attr string
}

func (n *lsAttrNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.LSATTR(path, n.attr)
	return nil, nil
}

type getUNode struct {
	path node
	u    interface{}
}

func (n *getUNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.GetByAttr(path, n.u)
	return nil, nil
}

type loadNode struct {
	path node
}

func (n *loadNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	err = cmd.LoadFile(path, InterpretLine)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type loadTemplateNode struct {
	path node
}

func (n *loadTemplateNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	data := fileToJSON(path)
	if data == nil {
		return nil, fmt.Errorf("Cannot read json file : %s", path)
	}
	cmd.LoadTemplate(data, path)
	return path, nil
}

type printNode struct {
	expr node
}

func (n *printNode) execute() (interface{}, error) {
	val, err := n.expr.execute()
	if err != nil {
		return nil, err
	}
	return cmd.Print([]interface{}{val}), nil
}

type deleteObjNode struct {
	path node
}

func (n *deleteObjNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.DeleteObj(path), nil
}

type deleteSelectionNode struct{}

func (n *deleteSelectionNode) execute() (interface{}, error) {
	return cmd.DeleteSelection(), nil
}

type isEntityDrawableNode struct {
	path node
}

func (n *isEntityDrawableNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.IsEntityDrawable(path), nil
}

type isAttrDrawableNode struct {
	path   node
	factor node
}

func (n *isAttrDrawableNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	attrInf, err := n.factor.execute()
	if err != nil {
		return nil, err
	}
	if _, ok := attrInf.(string); !ok {
		return nil, fmt.Errorf("Attribute operand is invalid")
	}
	return cmd.IsAttrDrawable(path, attrInf.(string), nil, false), nil
}

type getObjectNode struct {
	path node
}

func (n *getObjectNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	v, _ := cmd.GetObject(path, false)
	if v == nil {
		return nil, fmt.Errorf("Cannot find object at path ", path)
	}
	return v, nil
}

type selectObjectNode struct {
	path node
}

func (n *selectObjectNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	v, _ := cmd.GetObject(path, false)
	if v == nil {
		return nil, fmt.Errorf("Cannot find object at path ", path)
	}
	cmd.CD(path)
	cmd.SetClipBoard([]string{path})
	return v, nil
}

type searchObjectsNode struct {
	objType string
	nodeMap map[string]interface{}
}

func (n *searchObjectsNode) execute() (interface{}, error) {
	valMap, err := evalMapNodes(n.nodeMap)
	if err != nil {
		return nil, err
	}
	resMap, err := resMap(valMap, n.objType, false)
	if err != nil {
		return nil, err
	}
	v := cmd.SearchObjects(n.objType, resMap)
	return v, nil
}

type recursiveUpdateObjNode struct {
	arg0 interface{}
	arg1 interface{}
	arg2 interface{}
}

func (n *recursiveUpdateObjNode) execute() (interface{}, error) {
	//Old code was removed since
	//it broke the OCLI syntax easy update
	if _, ok := n.arg2.(bool); ok {
		//Weird edge case
		//to solve issue with:
		// for i in $(ls) do $i[attr]="string"

		//n.arg0 = referenceToNode
		//n.arg1 = attributeString, (used as an index)
		//n.arg2 = someValue (usually a string)
		nodeVal, err := n.arg0.(node).execute()
		if err != nil {
			return nil, err
		}
		objMap := nodeVal.(map[string]interface{})

		if checkIfObjectNode(objMap) == true {
			val, err := n.arg2.(node).execute()
			if err != nil {
				return nil, err
			}
			updateArgs := map[string]interface{}{n.arg1.(string): val}
			id := objMap["id"].(string)
			entity := objMap["category"].(string)
			cmd.RecursivePatch("", id, entity, updateArgs)
		}

	} else {
		if n.arg2.(string) == "recursive" {
			cmd.RecursivePatch(n.arg0.(string), "", "", n.arg1.(map[string]interface{}))
		}
	}
	return nil, nil
}

type updateObjNode struct {
	path       node
	attributes map[string]interface{}
}

func (n *updateObjNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	attributes, err := evalMapNodes(n.attributes)
	if err != nil {
		return nil, err
	}
	if path == "_" {
		return nil, cmd.UpdateSelection(attributes)
	}
	return cmd.UpdateObj(path, "", "", attributes, false)
}

type specialUpdateNode struct {
	path     node
	variable string
	first    node
	second   node
}

func (n *specialUpdateNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	first, err := n.first.execute()
	if err != nil {
		return nil, err
	}
	second, err := n.second.execute()
	if err != nil {
		return nil, err
	}
	if n.variable == "areas" {
		areas := map[string]interface{}{"reserved": first, "technical": second}
		attributes, _ := parseAreas(areas)
		return cmd.UpdateObj(path, "", "", attributes, false)
	} else if n.variable == "separator" {
		obj, _ := cmd.GetObject(path, true)
		if obj == nil {
			return nil, fmt.Errorf("cannot find object")
		}
		attr := obj["attributes"].(map[string]interface{})
		var sepArray []interface{}
		separators, ok := attr["separators"]
		if ok {
			sepArray = separators.([]interface{})
		}
		sepArray = append(sepArray, map[string]interface{}{"startPosXYm": first, "endPosXYm": second})
		attr["separators"] = sepArray
		return cmd.UpdateObj(path, "", "", attr, false)
	} else {
		return nil, fmt.Errorf("Invalid special update")
	}
}

type easyUpdateNode struct {
	nodePath     string
	jsonPath     string
	deleteAndPut bool
}

func (n *easyUpdateNode) execute() (interface{}, error) {
	data := make(map[string]interface{})
	data = fileToJSON(n.jsonPath)
	if data == nil {
		return nil, fmt.Errorf("Cannot open json file")
	}
	return cmd.UpdateObj(n.nodePath, "", "", data, n.deleteAndPut)
}

type lsObjNode struct {
	path   node
	entity int
}

func (n *lsObjNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.LSOBJECT(path, n.entity), nil
}

type treeNode struct {
	path  node
	depth int
}

func (n *treeNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	cmd.Tree(path, n.depth)
	return nil, nil
}

type drawNode struct {
	path  node
	depth int
}

func (n *drawNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return nil, cmd.Draw(path, n.depth)
}

type lsogNode struct{}

func (n *lsogNode) execute() (interface{}, error) {
	cmd.LSOG()
	return nil, nil
}

type lsenterpriseNode struct{}

func (n *lsenterpriseNode) execute() (interface{}, error) {
	cmd.LSEnterprise()
	return nil, nil
}

type exitNode struct{}

func (n *exitNode) execute() (interface{}, error) {
	cmd.Exit()
	return nil, nil
}

type clrNode struct{}

func (n *clrNode) execute() (interface{}, error) {
	cmd.Clear()
	return nil, nil
}

type envNode struct{}

func (n *envNode) execute() (interface{}, error) {
	cmd.Env(dynamicSymbolTable, funcTable)
	return nil, nil
}

type selectNode struct{}

func (n *selectNode) execute() (interface{}, error) {
	return cmd.ShowClipBoard(), nil
}

type pwdNode struct{}

func (n *pwdNode) execute() (interface{}, error) {
	return cmd.PWD(), nil
}

type grepNode struct{}

func (n *grepNode) execute() (interface{}, error) {
	return nil, nil
}

type selectChildrenNode struct {
	paths []node
}

func (n *selectChildrenNode) execute() (interface{}, error) {
	var paths []string
	for i := range n.paths {
		v, err := n.paths[i].execute()
		if err != nil {
			return nil, err
		}
		path, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("")
		}
		paths = append(paths, path)
	}
	v, err := cmd.SetClipBoard(paths)
	if err != nil {
		return nil, err
	}
	println("Selection made!")
	return v, nil
}

type updateSelectNode struct {
	data map[string]interface{}
}

func (n *updateSelectNode) execute() (interface{}, error) {
	cmd.UpdateSelection(n.data)
	return nil, nil
}

type unsetVarNode struct {
	option string
	name   string
}

func (n *unsetVarNode) execute() (interface{}, error) {
	switch n.option {
	case "-f":
		funcTable[n.name] = nil
	case "-v":
		dynamicSymbolTable[n.name] = nil
	default:
		return nil, fmt.Errorf("unset option needed (-v or -f)")
	}
	return nil, nil
}

type unsetAttrNode struct {
	path node
}

func (n *unsetAttrNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	arr := strings.Split(path, ":")
	if len(arr) != 2 {
		msg := "You must specify the attribute to delete with a colon!\n" +
			"(ie. $> unset path/to/object:attributex). \n" +
			"Please refer to the language reference help for more details" +
			"\n($> man unset)"
		return nil, fmt.Errorf(msg)
	}
	path = arr[0]
	data := map[string]interface{}{arr[1]: nil}

	return cmd.UpdateObj(path, "", "", data, true)

}

type setEnvNode struct {
	arg  string
	expr node
}

func (n *setEnvNode) execute() (interface{}, error) {
	val, err := n.expr.execute()
	if err != nil {
		return nil, err
	}
	cmd.SetEnv(n.arg, val)
	return nil, nil
}

type hierarchyNode struct {
	path  node
	depth int
}

func (n *hierarchyNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return cmd.GetHierarchy(path, n.depth, false), nil

}

type getOCAttrNode struct {
	path       node
	ent        int
	attributes map[string]interface{}
}

func (n *getOCAttrNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	attributes, err := evalMapNodes(n.attributes)
	if err != nil {
		return nil, err
	}
	err = cmd.GetOCLIAtrributes(path, n.ent, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createRackNode struct {
	path  node
	attrs [3]node
}

func (n *createRackNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	var vals [3]interface{}
	for i := 0; i < 3; i++ {
		vals[i], err = n.attrs[i].execute()
		if err != nil {
			return nil, err
		}
	}
	attr := make(map[string]interface{})
	if checkIfTemplate(vals[1]) == false {
		attr["size"] = vals[1]
	} else {
		attr["template"] = vals[1]
	}
	attr["posXY"] = vals[0]
	attr["orientation"] = vals[2]
	attributes := map[string]interface{}{"attributes": attr}
	err = cmd.GetOCLIAtrributes(path, cmd.RACK, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createDeviceNode struct {
	path  node
	attrs [3]node
}

func (n *createDeviceNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	var vals [3]interface{}
	for i := 0; i < 3; i++ {
		if n.attrs[i] != nil {
			vals[i], err = n.attrs[i].execute()
			if err != nil {
				return nil, err
			}
		}
	}
	attr := map[string]interface{}{"posU/slot": vals[0]}
	if checkIfTemplate(vals[1]) == false {
		attr["sizeU"] = vals[1]
	} else {
		attr["template"] = vals[1]
	}
	if n.attrs[2] != nil {
		attr["orientation"] = vals[2]
	}
	attributes := map[string]interface{}{"attributes": attr}
	err = cmd.GetOCLIAtrributes(path, cmd.DEVICE, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createGroupNode struct {
	path  node
	paths []node
}

func (n *createGroupNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	var paths []string
	for i := range n.paths {
		v, err := n.paths[i].execute()
		if err != nil {
			return nil, err
		}
		path, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("")
		}
		paths = append(paths, path)
	}
	attributes := map[string]interface{}{"racks": paths}
	err = cmd.GetOCLIAtrributes(path, cmd.GROUP, attributes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type uiDelayNode struct {
	time node
}

func (n *uiDelayNode) execute() (interface{}, error) {
	val, err := n.time.execute()
	if err != nil {
		return nil, err
	}
	time, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("delay should be a float")
	}
	cmd.UIDelay(time)
	return nil, nil
}

type uiToggleNode struct {
	feature string
	enable  node
}

func (n *uiToggleNode) execute() (interface{}, error) {
	val, err := n.enable.execute()
	if err != nil {
		return nil, err
	}
	enable, ok := val.(bool)
	if !ok {
		return nil, fmt.Errorf("feature %s expects a boolean", n.feature)
	}
	cmd.UIToggle(n.feature, enable)
	return nil, nil
}

type uiHighlightNode struct {
	path node
}

func (n *uiHighlightNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	return nil, cmd.UIHighlight(path)
}

type cameraMoveNode struct {
	command  string
	position node
	rotation node
}

func (n *cameraMoveNode) execute() (interface{}, error) {
	posVal, err := n.position.execute()
	if err != nil {
		return nil, err
	}
	position, ok := posVal.([]float64)
	if !ok {
		return nil, fmt.Errorf("OGREE: Error, command args are invalid\nPlease provide a vector3 and a vector2")
	}
	rotVal, err := n.rotation.execute()
	if err != nil {
		return nil, err
	}
	rotation, ok := rotVal.([]float64)
	if !ok {
		return nil, fmt.Errorf("OGREE: Error, command args are invalid\nPlease provide a vector3 and a vector2")
	}
	cmd.CameraMove(n.command, position, rotation)
	return nil, nil
}

type cameraWaitNode struct {
	time node
}

func (n *cameraWaitNode) execute() (interface{}, error) {
	val, err := n.time.execute()
	if err != nil {
		return nil, err
	}
	time, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("delay should be a float")
	}
	cmd.CameraWait(time)
	return nil, nil
}

type linkObjectNode struct {
	paths []interface{}
}

func (n *linkObjectNode) execute() (interface{}, error) {
	if len(n.paths) == 3 {
		newVal, err := n.paths[2].(node).execute()
		if err != nil {
			return nil, err
		}
		n.paths[2] = newVal
	}
	cmd.LinkObject(n.paths)
	return nil, nil
}

type unlinkObjectNode struct {
	paths []interface{}
}

func (n *unlinkObjectNode) execute() (interface{}, error) {
	cmd.UnlinkObject(n.paths)
	return nil, nil
}

type symbolReferenceNode struct {
	va string
}

func (s *symbolReferenceNode) execute() (interface{}, error) {
	val, ok := dynamicSymbolTable[s.va]
	if !ok {
		return nil, fmt.Errorf("Undefined variable ", s.va)
	}
	switch v := val.(type) {
	case string, int, bool, float64, float32, map[int]interface{}:
		if cmd.State.DebugLvl >= 3 {
			println("So You want the value: ", v)
		}
	}
	return val, nil
}

type arrayReferenceNode struct {
	variable string
	idx      node
}

func (n *arrayReferenceNode) execute() (interface{}, error) {
	v, ok := dynamicSymbolTable[n.variable]
	if !ok {
		return nil, fmt.Errorf("Undefined variable ", n.variable)
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("You can only index an array.")
	}
	idx, err := n.idx.execute()
	if err != nil {
		return nil, err
	}
	i, ok := idx.(int)
	if !ok {
		return nil, fmt.Errorf("Index should be an integer.")
	}
	if i < 0 || i >= len(arr) {
		return nil, fmt.Errorf("Index out of range error!\n Array Length Of: ",
			len(arr), "\nBut desired index at: ", i)
	}
	return arr[i], nil
}

type assignNode struct {
	variable string
	val      node
}

func (a *assignNode) execute() (interface{}, error) {
	val, err := a.val.execute()
	if err != nil {
		return nil, err
	}
	switch v := val.(type) {
	case bool, int, float64, string, []interface{}, map[string]interface{}:
		dynamicSymbolTable[a.variable] = v
		if cmd.State.DebugLvl >= 3 {
			println("You want to assign", a.variable, "with value of", v)
		}
		return nil, nil
	}
	return nil, fmt.Errorf("Invalid type to assign variable ", a.variable)
}

//Checks the map and sees if it is an object type
func checkIfObjectNode(x map[string]interface{}) bool {
	if idInf, ok := x["id"]; ok {
		if id, ok := idInf.(string); ok {
			if len(id) == 24 {
				if catInf, ok := x["category"]; ok {
					if _, ok := catInf.(string); ok {
						return true
					}
				}

				if slugInf, ok := x["slug"]; ok {
					if _, ok := slugInf.(string); ok {
						return true
					}
				}
			}
		}
	}
	return false
}

//Hack function for the [room]:areas=[r1,r2,r3,r4]@[t1,t2,t3,t4]
//command
func parseAreas(areas map[string]interface{}) (map[string]interface{}, error) {
	var reservedStr string
	var techStr string
	if reserved, ok := areas["reserved"].([]interface{}); ok {
		if tech, ok := areas["technical"].([]interface{}); ok {
			if len(reserved) == 4 && len(tech) == 4 {
				var r [4]*bytes.Buffer
				var t [4]*bytes.Buffer
				for i := 3; i >= 0; i-- {
					r[i] = bytes.NewBufferString("")
					fmt.Fprintf(r[i], "%v", reserved[i])
					t[i] = bytes.NewBufferString("")
					fmt.Fprintf(t[i], "%v", tech[i])
				}
				reservedStr = "{\"left\":" + r[3].String() + ",\"right\":" + r[2].String() + ",\"top\":" + r[0].String() + ",\"bottom\":" + r[1].String() + "}"
				techStr = "{\"left\":" + t[3].String() + ",\"right\":" + t[2].String() + ",\"top\":" + t[0].String() + ",\"bottom\":" + t[1].String() + "}"
				areas["reserved"] = reservedStr
				areas["technical"] = techStr
			}
		}
	}
	return areas, nil
}
