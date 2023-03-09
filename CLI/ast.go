package main

import (
	"bytes"
	cmd "cli/controllers"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	for i := range a.statements {
		if a.statements[i] != nil {
			_, err := a.statements[i].execute()
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

type funcDefNode struct {
	name string
	body node
}

func (n *funcDefNode) execute() (interface{}, error) {
	funcTable[n.name] = n.body
	if cmd.State.DebugLvl >= 3 {
		println("New function ", n.name)
	}
	return nil, nil
}

type funcCallNode struct {
	name string
}

func (n *funcCallNode) execute() (interface{}, error) {
	val, ok := funcTable[n.name]
	if !ok {
		return nil, fmt.Errorf("undefined function %s", n.name)
	}
	body, ok := val.(node)
	if !ok {
		return nil, fmt.Errorf("variable %s does not contain a function", n.name)
	}
	return body.execute()
}

// At this time arrays are all []floats
type arrNode struct {
	nodes []node
}

func (n *arrNode) execute() (interface{}, error) {
	var r []float64
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
		return nil, fmt.Errorf("Undefined variable %s", n.variable)
	}
	arr, ok := val.([]float64)
	if !ok {
		return nil, fmt.Errorf("Variable %s does not contain an array.", n.variable)
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
	u    node
}

func (n *getUNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	uAny, err := n.u.execute()
	if err != nil {
		return nil, err
	}
	u, ok := uAny.(int)
	if !ok {
		return nil, fmt.Errorf("u should be an integer")
	}
	cmd.GetByAttr(path, u)
	return nil, nil
}

type getSlotNode struct {
	path node
	slot node
}

func (n *getSlotNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Path should be a string")
	}
	slot, err := n.slot.execute()
	if err != nil {
		return nil, err
	}
	cmd.GetByAttr(path, slot)
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
		return nil, fmt.Errorf("path should be a string")
	}

	//Usually functions from 'controller' pkg are called
	//But in this case we are calling a function from 'main' pkg
	return nil, LoadFile(path)
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
		return nil, fmt.Errorf("path should be a string")
	}
	data := fileToJSON(path)
	if data == nil {
		return nil, fmt.Errorf("cannot read json file : %s", path)
	}
	return path, cmd.LoadTemplate(data, path)
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
	path node
	attr string
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
	return cmd.IsAttrDrawable(path, n.attr, nil, false), nil
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
		return nil, fmt.Errorf("Cannot find object at path %s", path)
	}
	return v, nil
}

type selectObjectNode struct {
	path node
}

func (n *selectObjectNode) execute() (interface{}, error) {
	var selection []string
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("Object path should be a string")
	}
	if path != "" {
		selection = []string{path}
	}

	cmd.CD(path)
	return cmd.SetClipBoard(selection)
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

// TODO: Need to restore recursive updates or to remove it
// entirely
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

func setRoomAreas(path string, values []any) (map[string]any, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("2 values (reserved, technical) expected to set room areas")
	}
	areas := map[string]any{"reserved": values[0], "technical": values[1]}
	attributes, e := parseAreas(areas)
	if e != nil {
		return nil, e
	}
	return cmd.UpdateObj(path, "", "", attributes, false)
}

func setLabel(path string, values []any, hasSharpe bool) (map[string]any, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("only 1 value expected")
	}
	value, ok := values[0].(string)
	if !ok {
		return nil, fmt.Errorf("value should be a string")
	}
	return nil, cmd.InteractObject(path, "label", value, hasSharpe)
}

func setLabelFont(path string, values []any) (map[string]any, error) {
	msg := "The font can only be bold or italic" +
		" or be in the form of color@[colorValue]." +
		"\n\nFor more information please refer to: " +
		"\nhttps://github.com/ditrit/OGrEE-3D/wiki/CLI-langage#interact-with-objects"

	switch len(values) {
	case 1:
		if values[0] != "bold" && values[0] != "italic" {
			return nil, fmt.Errorf(msg)
		}
		return nil, cmd.InteractObject(path, "labelFont", values[0], false)
	case 2:
		if values[0] != "color" {
			return nil, fmt.Errorf(msg)
		}
		c, ok := AssertColor(values[1])
		if !ok {
			return nil, fmt.Errorf("please provide a valid 6 length hex value for the color")
		}
		return nil, cmd.InteractObject(path, "labelFont", "color@"+c, false)
	default:
		return nil, fmt.Errorf(msg)
	}
}

func addRoomSeparator(path string, values []any) (map[string]any, error) {
	if len(values) != 3 {
		return nil, fmt.Errorf("3 values (startPos, endPos, type) expected to add a separator")
	}
	startPos, ok := values[0].([]float64)
	if !ok || len(startPos) != 2 {
		return nil, fmt.Errorf("startPos should be a vector2")
	}
	endPos, ok := values[1].([]float64)
	if !ok || len(startPos) != 2 {
		return nil, fmt.Errorf("endPos should be a vector2")
	}
	sepType, ok := values[2].(string)
	if !ok {
		return nil, fmt.Errorf("type of separator should \"wireframe\" or \"plain\"")
	}
	sepType = strings.ToLower(sepType)
	if sepType != "wireframe" && sepType != "plain" {
		return nil, fmt.Errorf("type of separator should \"wireframe\" or \"plain\"")
	}
	nextSep := map[string]any{"startPosXYm": startPos, "endPosXYm": endPos, "type": sepType}
	obj, _ := cmd.GetObject(path, true)
	if obj == nil {
		return nil, fmt.Errorf("cannot find object")
	}
	attr := obj["attributes"].(map[string]any)
	var sepArray []any
	separators := attr["separators"]
	if IsInfArr(separators) {
		sepArray = separators.([]any)
		sepArray = append(sepArray, nextSep)
		sepArrStr, _ := json.Marshal(&sepArray)
		attr["separators"] = string(sepArrStr)
	} else {
		var sepStr string
		nextSepStr, _ := json.Marshal(nextSep)
		if IsString(separators) && separators != "" && separators != "[]" {
			sepStr = separators.(string)
			size := len(sepStr)
			sepStr = sepStr[:size-1] + "," + string(nextSepStr) + "]"
		} else {
			sepStr = "[" + string(nextSepStr) + "]"
		}
		attr["separators"] = sepStr
	}
	return cmd.UpdateObj(path, "", "", attr, false)
}

func addRoomPillar(path string, values []any) (map[string]any, error) {
	centerXY, ok := values[0].([]float64)
	if !ok || len(centerXY) != 2 {
		return nil, fmt.Errorf("centerXY should be a vector2")
	}
	sizeXY, ok := values[1].([]float64)
	if !ok || len(sizeXY) != 2 {
		return nil, fmt.Errorf("sizeXY should be a vector2")
	}
	rotation, err := getFloat(values[2])
	if err != nil {
		return nil, fmt.Errorf("rotation should be a number")
	}
	obj, _ := cmd.GetObject(path, true)
	if obj == nil {
		return nil, fmt.Errorf("cannot find object")
	}
	var pillarArray []any
	attr := obj["attributes"].(map[string]any)
	pillars := attr["pillars"]

	if IsInfArr(pillars) {
		pillarArray = pillars.([]any)
		pillarArray = append(pillarArray, map[string]any{
			"centerXY": centerXY, "sizeXY": sizeXY, "rotation": rotation})

		pillarArrStr, _ := json.Marshal(&pillarArray)
		attr["pillars"] = string(pillarArrStr)
	} else {
		var pillStr string
		nextPill := map[string]any{
			"centerXY": centerXY, "sizeXY": sizeXY, "rotation": rotation}

		nextPillStr, _ := json.Marshal(nextPill)
		if IsString(pillars) && pillars != "" && pillars != "[]" {
			pillStr = pillars.(string)
			size := len(pillStr)
			pillStr = pillStr[:size-1] + "," + string(nextPillStr) + "]"
		} else {
			pillStr = "[" + string(nextPillStr) + "]"
		}
		attr["pillars"] = pillStr
	}
	return attr, nil
}

type updateObjNode struct {
	path      node
	attr      string
	values    []node
	hasSharpe bool
}

func (n *updateObjNode) execute() (interface{}, error) {
	path, err := AssertString(&n.path, "Object path")
	if err != nil {
		return nil, err
	}
	values := []any{}
	for _, valueNode := range n.values {
		val, err := valueNode.execute()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	if path == "_" {
		if len(values) != 1 {
			return nil, fmt.Errorf("only one value is expected when updating selection")
		}
		return nil, cmd.UpdateSelection(map[string]any{n.attr: values[0]})
	}
	boolInteractVals := []string{"content", "alpha", "tilesName", "tilesColor", "U", "slots", "localCS"}
	if AssertInStringValues(n.attr, boolInteractVals) {
		if !IsBool(values[0]) {
			return nil, fmt.Errorf("boolean value expected")
		}
		return nil, cmd.InteractObject(path, n.attr, values[0], n.hasSharpe)
	}
	switch n.attr {
	case "areas":
		return setRoomAreas(path, values)
	case "label":
		return setLabel(path, values, n.hasSharpe)
	case "labelFont":
		return setLabelFont(path, values)
	case "separator":
		return addRoomSeparator(path, values)
	case "pillar":
		return addRoomPillar(path, values)
	}
	return cmd.UpdateObj(path, "", "", map[string]any{n.attr: values[0]}, false)
}

type lsObjNode struct {
	path      node
	entity    int
	recursive bool
	sort      string
	attrList  []string
	format    string
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
	var objects []any
	if n.recursive {
		objects = cmd.LSOBJECTRecursive(path, n.entity)
	} else {
		objects = cmd.LSOBJECT(path, n.entity)
	}
	if n.sort != "" {
		objects = cmd.SortObjects(&objects, n.sort).GetData()
	}
	if n.attrList != nil {
		if n.format != "" {
			cmd.DispfWithAttrs(n.format, &objects, &n.attrList)
		} else {
			cmd.DispWithAttrs(&objects, &n.attrList)
		}
	} else {
		if n.sort != "" {
			//We want to display the attribute used for sorting
			attrList := append(n.attrList, n.sort)
			cmd.DispWithAttrs(&objects, &attrList)
		} else {
			for i := range objects {
				object, ok := objects[i].(map[string]interface{})
				if ok && object != nil && object["name"] != nil {
					println(object["name"].(string))
				}
			}
		}
	}
	return objects, nil
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
	force bool
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
	return nil, cmd.Draw(path, n.depth, n.force)
}

type undrawNode struct {
	path node
}

func (n *undrawNode) execute() (interface{}, error) {
	if n.path == nil {
		return nil, cmd.Undraw("")
	}
	path, e := AssertString(&(n.path), "Path")
	if e != nil {
		return nil, e
	}
	return nil, cmd.Undraw(path)
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
	paths, err := evalNodeArr[string](&n.paths, []string{})
	if err != nil {
		return nil, err
	}
	v, err := cmd.SetClipBoard(paths)
	if err != nil {
		return nil, err
	}
	if cmd.State.DebugLvl > cmd.NONE {
		println("Selection made!")
	}

	return v, nil
}

type updateSelectNode struct {
	data map[string]interface{}
}

func (n *updateSelectNode) execute() (interface{}, error) {
	cmd.UpdateSelection(n.data)
	return nil, nil
}

type unsetFuncNode struct {
	funcName string
}

func (n *unsetFuncNode) execute() (interface{}, error) {
	delete(funcTable, n.funcName)
	return nil, nil
}

type unsetVarNode struct {
	varName string
}

func (n *unsetVarNode) execute() (interface{}, error) {
	delete(dynamicSymbolTable, n.varName)
	return nil, nil
}

type unsetAttrNode struct {
	path  node
	attr  string
	index node
}

func (n *unsetAttrNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	if n.index != nil {
		idxAny, err := n.index.execute()
		if err != nil {
			return nil, err
		}
		idx, ok := idxAny.(int)
		if !ok {
			return nil, fmt.Errorf("index should be an integer")
		}
		return cmd.UnsetInObj(path, n.attr, idx)
	}
	return cmd.UpdateObj(path, "", "", map[string]any{n.attr: nil}, true)
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

type createTenantNode struct {
	path  node
	color node
}

func (n *createTenantNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	colorInf, err := n.color.execute()
	if err != nil {
		return nil, err
	}
	color, ok := AssertColor(colorInf)
	if !ok {
		return nil, fmt.Errorf("please provide a valid 6 length hex value for the color")
	}
	attributes := map[string]any{"color": color}
	err = cmd.GetOCLIAtrributes(path, cmd.TENANT, map[string]any{"attributes": attributes})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createSiteNode struct {
	path node
}

func (n *createSiteNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	err = cmd.GetOCLIAtrributes(path, cmd.SITE, map[string]any{})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createBuildingNode struct {
	path           node
	posXY          node
	rotation       node
	sizeOrTemplate node
}

func (n *createBuildingNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	posXYany, err := n.posXY.execute()
	if err != nil {
		return nil, err
	}
	posXY, ok := posXYany.([]float64)
	if !ok || len(posXY) != 2 {
		fmt.Printf("%v\n", posXYany)
		return nil, fmt.Errorf("posXY should be a vector2")
	}
	rotationAny, err := n.rotation.execute()
	if err != nil {
		return nil, err
	}
	rotation, err := getFloat(rotationAny)
	if err != nil {
		return nil, fmt.Errorf("rotation should be a number")
	}
	attributes := map[string]any{"posXY": posXY, "rotation": rotation}

	sizeOrTemplateAny, err := n.sizeOrTemplate.execute()
	if err != nil {
		return nil, err
	}
	template, ok := sizeOrTemplateAny.(string)
	if ok && checkIfTemplate(template, cmd.BLDG) {
		attributes["template"] = template
	} else {
		size, ok := sizeOrTemplateAny.([]float64)
		if !ok || len(size) != 3 {
			return nil, fmt.Errorf("vector3 (size) or template expected")
		}
		attributes["size"] = size
	}
	err = cmd.GetOCLIAtrributes(path, cmd.BLDG, map[string]any{"attributes": attributes})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createRoomNode struct {
	path            node
	posXY           node
	rotation        node
	size            node
	axisOrientation node
	floorUnit       node
	template        node
}

func (n *createRoomNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	posXYany, err := n.posXY.execute()
	if err != nil {
		return nil, err
	}
	posXY, ok := posXYany.([]float64)
	if !ok || len(posXY) != 2 {
		return nil, fmt.Errorf("posXY should be a vector2")
	}
	rotationAny, err := n.rotation.execute()
	if err != nil {
		return nil, err
	}
	rotation, err := getFloat(rotationAny)
	if err != nil {
		return nil, fmt.Errorf("rotation should be a number")
	}
	attributes := map[string]any{"posXY": posXY, "rotation": rotation}

	if n.template != nil {
		templateAny, err := n.template.execute()
		if err != nil {
			return nil, err
		}
		template, ok := templateAny.(string)
		if !ok || !checkIfTemplate(template, cmd.ROOM) {
			return nil, fmt.Errorf("invalid template")
		}
		attributes["template"] = template
	} else {
		sizeAny, err := n.size.execute()
		if err != nil {
			return nil, err
		}
		size, ok := sizeAny.([]float64)
		if !ok || len(size) != 3 {
			return nil, fmt.Errorf("size should be a vector3")
		}
		attributes["size"] = size
		axisOrientationAny, err := n.axisOrientation.execute()
		if err != nil {
			return nil, err
		}
		axisOrientation, ok := axisOrientationAny.(string)
		if !ok || (axisOrientation != "+x+y" && axisOrientation != "+x-y" &&
			axisOrientation != "-x-y" && axisOrientation != "-x+y") {
			return nil, fmt.Errorf("orientation should be +x+y, +x-y, -x-y or x+y")
		}
		attributes["axisOrientation"] = axisOrientation
	}
	if n.floorUnit != nil {
		floorUnitAny, err := n.floorUnit.execute()
		if err != nil {
			return nil, err
		}
		floorUnit, ok := floorUnitAny.(string)
		if !ok {
			return nil, fmt.Errorf("floorUnit should be a string")
		}
		attributes["floorUnit"] = floorUnit
	}
	err = cmd.GetOCLIAtrributes(path, cmd.ROOM, map[string]any{"attributes": attributes})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createRackNode struct {
	path           node
	pos            node
	sizeOrTemplate node
	orientation    node
}

func (n *createRackNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	posAny, err := n.pos.execute()
	if err != nil {
		return nil, err
	}
	pos, ok := posAny.([]float64)
	if !ok || (len(pos) != 2 && len(pos) != 3) {
		return nil, fmt.Errorf("position should be a vector2 or a vector3")
	}
	orientationAny, err := n.orientation.execute()
	if err != nil {
		return nil, err
	}
	orientation, ok := orientationAny.(string)
	if !ok || (orientation != "front" && orientation != "rear" && orientation != "left" && orientation != "right") {
		return nil, fmt.Errorf("orientation should be a front, rear, left or right")
	}
	attributes := map[string]any{"posXYZ": pos, "orientation": orientation}

	sizeOrTemplateAny, err := n.sizeOrTemplate.execute()
	if err != nil {
		return nil, err
	}
	template, ok := sizeOrTemplateAny.(string)
	if ok && checkIfTemplate(template, cmd.RACK) {
		attributes["template"] = template
	} else {
		size, ok := sizeOrTemplateAny.([]float64)
		if !ok || len(size) != 3 {
			return nil, fmt.Errorf("vector3 (size) or template expected")
		}
		attributes["size"] = size
	}
	err = cmd.GetOCLIAtrributes(path, cmd.RACK, map[string]any{"attributes": attributes})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createDeviceNode struct {
	path            node
	posUOrSlot      node
	sizeUOrTemplate node
	side            node
}

func (n *createDeviceNode) execute() (interface{}, error) {
	val, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	posUOrSlot, err := n.posUOrSlot.execute()
	if err != nil {
		return nil, err
	}
	attr := map[string]any{"posU/slot": posUOrSlot}

	sizeUOrTemplate, err := n.sizeUOrTemplate.execute()
	if err != nil {
		return nil, err
	}
	if !checkIfTemplate(sizeUOrTemplate, cmd.DEVICE) {
		attr["sizeU"] = sizeUOrTemplate
	} else {
		attr["template"] = sizeUOrTemplate
	}
	if n.side != nil {
		side, err := n.side.execute()
		if err != nil {
			return nil, err
		}
		attr["orientation"] = side
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
		return nil, fmt.Errorf("path should be a string")
	}
	var objs []string
	data := map[string]interface{}{}
	for i := range n.paths {
		v, err := n.paths[i].execute()
		if err != nil {
			return nil, err
		}
		obj, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("")
		}
		obj = filepath.Base(obj)
		objs = append(objs, obj)
	}

	data["attributes"] = map[string]interface{}{"content": objs}
	err = cmd.GetOCLIAtrributes(path, cmd.GROUP, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createCorridorNode struct {
	path      node
	leftRack  node
	rightRack node
	temp      node
}

func (n *createCorridorNode) execute() (interface{}, error) {
	path, err := AssertString(&n.path, "Path for corridor")
	if err != nil {
		return nil, err
	}

	leftRack, err2 := AssertString(&n.leftRack, "Path for left rack")
	if err2 != nil {
		return nil, err2
	}

	rightRack, err3 := AssertString(&n.rightRack, "Path for right rack")
	if err3 != nil {
		return nil, err3
	}

	temp, err4 := AssertString(&n.temp, "Temperature")
	if err4 != nil {
		return nil, err4
	}
	tempIsValid := AssertInStringValues(temp, []string{"warm", "cold"})
	if !tempIsValid {
		return nil,
			fmt.Errorf("temperature should be either 'warm' or 'cold'")
	}
	leftRack = filepath.Base(leftRack)
	rightRack = filepath.Base(rightRack)

	attributes := map[string]interface{}{
		"content": leftRack + "," + rightRack, "temperature": temp}

	data := map[string]interface{}{"attributes": attributes}

	err = cmd.GetOCLIAtrributes(path, cmd.CORIDOR, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createOrphanNode struct {
	path     node
	template node
	sensor   bool
}

func (n *createOrphanNode) execute() (interface{}, error) {
	pathVal, err := n.path.execute()
	if err != nil {
		return nil, err
	}
	path, ok := pathVal.(string)
	if !ok {
		return nil, fmt.Errorf("path should be a string")
	}
	var t int
	if n.sensor {
		t = cmd.STRAYSENSOR
	} else {
		t = cmd.STRAY_DEV
	}
	templateAny, err := n.template.execute()
	if err != nil {
		return nil, err
	}
	template, ok := templateAny.(string)
	if !ok || !checkIfTemplate(template, t) {
		return nil, fmt.Errorf("invalid template")
	}
	attributes := map[string]any{"template": template}

	err = cmd.GetOCLIAtrributes(path, t, map[string]any{"attributes": attributes})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type uiDelayNode struct {
	time float64
}

func (n *uiDelayNode) execute() (interface{}, error) {
	cmd.UIDelay(n.time)
	return nil, nil
}

type uiToggleNode struct {
	feature string
	enable  bool
}

func (n *uiToggleNode) execute() (interface{}, error) {
	cmd.UIToggle(n.feature, n.enable)
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

type uiClearCacheNode struct {
}

func (n *uiClearCacheNode) execute() (interface{}, error) {
	return nil, cmd.UIClearCache()
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
	if !ok || len(position) != 3 {
		return nil, fmt.Errorf("position (first argument) is invalid\nPlease provide a vector3")
	}
	rotVal, err := n.rotation.execute()
	if err != nil {
		return nil, err
	}
	rotation, ok := rotVal.([]float64)
	if !ok || len(rotation) != 2 {
		return nil, fmt.Errorf("rotation (second argument) is invalid\nPlease provide a vector2")
	}
	cmd.CameraMove(n.command, position, rotation)
	return nil, nil
}

type cameraWaitNode struct {
	time float64
}

func (n *cameraWaitNode) execute() (interface{}, error) {
	cmd.CameraWait(n.time)
	return nil, nil
}

type linkObjectNode struct {
	source      node
	destination node
	slot        node
}

func (n *linkObjectNode) execute() (interface{}, error) {
	var slot interface{}
	source, err := AssertString(&n.source, "Source Object Path")
	if err != nil {
		return nil, err
	}

	dest, err1 := AssertString(&n.destination, "Destination Object Path")
	if err1 != nil {
		return nil, err1
	}

	if n.slot != nil {
		s, e := n.slot.execute()
		if e != nil {
			return nil, e
		}
		slot = s
	} else {
		slot = nil
	}

	cmd.LinkObject(source, dest, slot)
	return nil, nil
}

type unlinkObjectNode struct {
	source      node
	destination node
}

func (n *unlinkObjectNode) execute() (interface{}, error) {
	destination := ""
	source, err := AssertString(&n.source, "Source Object Path")
	if err != nil {
		return nil, err
	}

	if n.destination != nil {
		var e error
		destination, e = AssertString(&n.destination, "Destination Object Path")
		if e != nil {
			return nil, e
		}
	}
	cmd.UnlinkObject(source, destination)
	return nil, nil
}

type symbolReferenceNode struct {
	va string
}

func (s *symbolReferenceNode) execute() (interface{}, error) {
	val, ok := dynamicSymbolTable[s.va]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", s.va)
	}
	switch v := val.(type) {
	case string, int, bool, float64, float32, map[int]interface{}:
		if cmd.State.DebugLvl >= 3 {
			println("So You want the value: ", v)
		}
	}
	return val, nil
}

type objReferenceNode struct {
	va    string
	index node
}

func (o *objReferenceNode) execute() (interface{}, error) {
	val, ok := dynamicSymbolTable[o.va]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", o.va)
	}
	if _, ok := val.(map[string]interface{}); !ok {
		return nil, fmt.Errorf(o.va + " Is not an indexable object")
	}
	object := val.(map[string]interface{})

	idx, e := o.index.execute()
	if e != nil {
		return nil, e
	}
	if _, ok := idx.(string); !ok {
		return nil, fmt.Errorf("The index must resolve to a string")
	}
	index := idx.(string)

	if mainAttr, ok := object[index]; ok {
		return mainAttr, nil
	} else {
		if attrInf, ok := object["attributes"]; ok {
			if attrDict, ok := attrInf.(map[string]interface{}); ok {
				if _, ok := attrDict[index]; ok {
					return attrDict[index], nil
				}
			}
		}
	}

	msg := "This object " + o.va + " cannot be indexed with " + index +
		". Please check the object you are referencing and try again"

	return nil, fmt.Errorf(msg)

}

type arrayReferenceNode struct {
	variable string
	idx      node
}

func (n *arrayReferenceNode) execute() (interface{}, error) {
	v, ok := dynamicSymbolTable[n.variable]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", n.variable)
	}
	arr, ok := v.([]float64)
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
		return nil, fmt.Errorf(
			"Index out of range\n"+
				"Array length : %d"+
				"But desired index at : %d",
			len(arr), i,
		)
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
	case bool, int, float64, string, []float64, map[string]interface{}:
		dynamicSymbolTable[a.variable] = v
		if cmd.State.DebugLvl >= 3 {
			println("You want to assign", a.variable, "with value of", v)
		}
		return nil, nil
	}
	return nil, fmt.Errorf("Invalid type to assign variable %s", a.variable)
}

// Checks the map and sees if it is an object type
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

// Hack function for the [room]:areas=[r1,r2,r3,r4]@[t1,t2,t3,t4]
// command
func parseAreas(areas map[string]interface{}) (map[string]interface{}, error) {
	var reservedStr string
	var techStr string

	if reserved, ok := areas["reserved"].([]float64); ok {
		if tech, ok := areas["technical"].([]float64); ok {
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
			} else {
				if len(reserved) != 4 && len(tech) == 4 {
					return nil, errorResponder("reserved", "4", false)
				} else if len(tech) != 4 && len(reserved) == 4 {
					return nil, errorResponder("technical", "4", false)
				} else { //Both invalid
					return nil, errorResponder("reserved and technical", "4", true)
				}
			}
		} else {
			return nil, errorResponder("technical", "4", false)
		}
	} else {
		return nil, errorResponder("reserved", "4", false)
	}
	return areas, nil
}
