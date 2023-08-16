package main

import (
	"bytes"
	"cli/config"
	c "cli/controllers"
	cmd "cli/controllers"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

func InitVars(variables []config.Vardef) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cannot parse config variables")
		}
	}()
	c.State.DynamicSymbolTable = make(map[string]interface{})
	c.State.FuncTable = make(map[string]interface{})
	for _, v := range variables {
		var varNode node
		switch val := v.Value.(type) {
		case string:
			p := newParser("\"" + val + "\"")
			varNode = p.parseExpr("")
		case int64:
			varNode = &valueNode{int(val)}
		default:
			varNode = &valueNode{val}
		}
		n := &assignNode{
			variable: v.Name,
			val:      varNode,
		}
		if _, err = n.execute(); err != nil {
			return err
		}
	}
	return err
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
	c.State.FuncTable[n.name] = n.body
	if cmd.State.DebugLvl >= 3 {
		println("New function ", n.name)
	}
	return nil, nil
}

type funcCallNode struct {
	name string
}

func (n *funcCallNode) execute() (interface{}, error) {
	val, ok := c.State.FuncTable[n.name]
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
		v, err := nodeToFloat(n.nodes[i], "array element")
		if err != nil {
			return nil, err
		}
		r = append(r, v)
	}
	return r, nil
}

type lenNode struct {
	variable string
}

func (n *lenNode) execute() (interface{}, error) {
	val, ok := c.State.DynamicSymbolTable[n.variable]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", n.variable)
	}
	arr, err := valToVec(val, -1, "Variable "+n.variable)
	if err != nil {
		return nil, err
	}
	return len(arr), nil
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	cmd.FocusUI(path)
	return nil, nil
}

type cdNode struct {
	path node
}

func (n *cdNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	return nil, cmd.CD(path)
}

type lsNode struct {
	path node
}

func (n *lsNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	items, err := cmd.Ls(path)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		println(item)
	}
	return nil, nil
}

type lsAttrNode struct {
	path node
	attr string
}

func (n *lsAttrNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	cmd.LSATTR(path, n.attr)
	return nil, nil
}

type getUNode struct {
	path node
	u    node
}

func (n *getUNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	u, err := nodeToInt(n.u, "u")
	if err != nil {
		return nil, err
	}
	if u < 0 {
		return nil, fmt.Errorf("The U value must be positive")
	}
	cmd.GetByAttr(path, u)
	return nil, nil
}

type getSlotNode struct {
	path node
	slot node
}

func (n *getSlotNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	//Usually functions from 'controller' pkg are called
	//But in this case we are calling a function from 'main' pkg
	return nil, LoadFile(path)
}

type loadTemplateNode struct {
	path node
}

func (n *loadTemplateNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
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
	fmt.Printf("%v\n", val)
	return nil, nil
}

type deleteObjNode struct {
	path node
}

func (n *deleteObjNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	return nil, cmd.DeleteObj(path)
}

type deleteSelectionNode struct{}

func (n *deleteSelectionNode) execute() (interface{}, error) {
	var errBuilder strings.Builder
	deleted := 0
	if c.State.ClipBoard != nil {
		for _, obj := range c.State.ClipBoard {
			err := c.DeleteObj(obj)
			if err != nil {
				errBuilder.WriteString(fmt.Sprintf("    %s: %s\n", obj, err.Error()))
			} else {
				deleted += 1
			}
		}
	}
	println(fmt.Sprintf("%d objects deleted", deleted))
	notDeleted := len(c.State.ClipBoard) - deleted
	if notDeleted > 0 {
		fmt.Printf("%d objects could not be deleted :\n%s", notDeleted, errBuilder.String())
	}
	return nil, nil
}

type isEntityDrawableNode struct {
	path node
}

func (n *isEntityDrawableNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	drawable, err := cmd.IsEntityDrawable(path)
	if err != nil {
		return nil, err
	}
	println(drawable)
	return drawable, nil
}

type isAttrDrawableNode struct {
	path node
	attr string
}

func (n *isAttrDrawableNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	drawable, err := cmd.IsAttrDrawable(path, n.attr)
	if err != nil {
		return nil, err
	}
	println(drawable)
	return drawable, nil
}

type getObjectNode struct {
	path node
}

func (n *getObjectNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	obj, err := cmd.GetObject(path)
	if err != nil {
		return nil, err
	}
	cmd.DisplayObject(obj)
	return obj, nil
}

type selectObjectNode struct {
	path node
}

func (n *selectObjectNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	var selection []string
	if path != "" {
		selection = []string{path}
		err = cmd.CD(path)
		if err != nil {
			return nil, err
		}
	}
	return cmd.SetClipBoard(selection)
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
	return cmd.UpdateObj(path, map[string]any{"attributes": attributes})
}

func setLabel(path string, values []any, hasSharpe bool) (map[string]any, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("only 1 value expected")
	}
	value, err := valToString(values[0], "value")
	if err != nil {
		return nil, err
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
		c, ok := valToColor(values[1])
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
	startPos, err := valToVec(values[0], 2, "startPos")
	if err != nil {
		return nil, err
	}
	endPos, err := valToVec(values[1], 2, "endPos")
	if err != nil {
		return nil, err
	}
	sepType, err := valToString(values[2], "separator type")
	if err != nil {
		return nil, err
	}
	nextSep := map[string]any{"startPosXYm": startPos, "endPosXYm": endPos, "type": sepType}
	obj, err := cmd.GetObject(path)
	if err != nil {
		return nil, err
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
	return cmd.UpdateObj(path, map[string]any{"attributes": attr})
}

func addRoomPillar(path string, values []any) (map[string]any, error) {
	if len(values) != 2 {
		return nil, fmt.Errorf("2 values (centerXY, sizeXY) expected to add a pillar")
	}
	centerXY, err := valToVec(values[0], 2, "centerXY")
	if err != nil {
		return nil, err
	}
	sizeXY, err := valToVec(values[0], 2, "sizeXY")
	if err != nil {
		return nil, err
	}
	rotation, err := valToFloat(values[2], "rotation")
	if err != nil {
		return nil, err
	}
	obj, err := cmd.GetObject(path)
	if err != nil {
		return nil, err
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

func parseDescriptionIdx(desc string) (int, error) {
	numStr := desc[len("description"):]
	num, e := strconv.Atoi(numStr)
	if e != nil {
		return -1, e
	}
	num -= 1
	if num < 0 {
		return -1, fmt.Errorf("description index should be at least 1")
	}
	return num, nil
}

func updateDescription(path string, attr string, values []any) (map[string]any, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("a single value is expected to update a description")
	}
	newDesc, err := valToString(values[0], "description")
	if err != nil {
		return nil, err
	}
	data := map[string]any{}
	if attr == "description" {
		data["description"] = []any{newDesc}
	} else {
		obj, err := cmd.GetObject(path)
		if err != nil {
			return nil, err
		}
		curDesc := obj["description"].([]any)
		idx, e := parseDescriptionIdx(attr)
		if e != nil {
			return nil, e
		}
		if idx > len(curDesc) {
			return nil, fmt.Errorf("description index out of range")
		} else if idx == len(curDesc) {
			curDesc = append(curDesc, newDesc)
		} else {
			curDesc[idx] = newDesc
		}
		data["description"] = curDesc
	}
	return cmd.UpdateObj(path, data)
}

type updateObjNode struct {
	path      node
	attr      string
	values    []node
	hasSharpe bool
}

func (n *updateObjNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
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
	var paths []string
	if path == "_" {
		paths = cmd.State.ClipBoard
	} else {
		paths = []string{path}
	}
	for _, path := range paths {
		switch n.attr {
		case "content", "alpha", "tilesName", "tilesColor", "U", "slots", "localCS":
			boolVal, err := valToBool(values[0], n.attr)
			if err != nil {
				return nil, err
			}
			return nil, cmd.InteractObject(path, n.attr, boolVal, n.hasSharpe)
		}
		var err error
		switch n.attr {
		case "areas":
			_, err = setRoomAreas(path, values)
		case "label":
			_, err = setLabel(path, values, n.hasSharpe)
		case "labelFont":
			_, err = setLabelFont(path, values)
		case "separator":
			_, err = addRoomSeparator(path, values)
		case "pillar":
			_, err = addRoomPillar(path, values)
		case "domain":
			_, err = cmd.UpdateObj(path, map[string]any{"domain": values[0]})
		default:
			if strings.HasPrefix(n.attr, "description") {
				_, err = updateDescription(path, n.attr, values)
			} else {
				attributes := map[string]any{n.attr: values[0]}
				_, err = cmd.UpdateObj(path, map[string]any{"attributes": attributes})
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type lsObjNode struct {
	path      node
	entity    int
	recursive bool
	sort      string
	attrList  []string
}

func (n *lsObjNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	var objects []any
	if n.recursive {
		objects, err = cmd.LSOBJECTRecursive(path, n.entity)
		if err != nil {
			return nil, err
		}
	} else {
		objects = cmd.LSOBJECT(path, n.entity)
	}
	if n.sort != "" {
		objects = cmd.SortObjects(objects, n.sort).GetData()
	}
	if n.attrList != nil {
		cmd.DispWithAttrs(objects, n.attrList)
	} else {
		if n.sort != "" {
			//We want to display the attribute used for sorting
			attrList := append(n.attrList, n.sort)
			cmd.DispWithAttrs(objects, attrList)
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	root, err := cmd.Tree(path, n.depth)
	if err != nil {
		return nil, err
	}
	fmt.Println(path)
	s := root.String(n.depth)
	if s != "" {
		fmt.Println(s)
	}
	return nil, nil
}

type drawNode struct {
	path  node
	depth int
	force bool
}

func (n *drawNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
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
	cmd.Env(c.State.DynamicSymbolTable, c.State.FuncTable)
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

type unsetFuncNode struct {
	funcName string
}

func (n *unsetFuncNode) execute() (interface{}, error) {
	delete(c.State.FuncTable, n.funcName)
	return nil, nil
}

type unsetVarNode struct {
	varName string
}

func (n *unsetVarNode) execute() (interface{}, error) {
	delete(c.State.DynamicSymbolTable, n.varName)
	return nil, nil
}

type unsetAttrNode struct {
	path  node
	attr  string
	index node
}

func (n *unsetAttrNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	if n.index != nil {
		idx, err := nodeToInt(n.index, "index")
		if err != nil {
			return nil, err
		}
		return cmd.UnsetInObj(path, n.attr, idx)
	}
	return nil, cmd.UnsetAttribute(path, n.attr)
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

type createDomainNode struct {
	path  node
	color node
}

func (n *createDomainNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	colorInf, err := n.color.execute()
	if err != nil {
		return nil, err
	}
	var color string
	var ok bool
	if color, ok = valToColor(colorInf); !ok {
		return nil, fmt.Errorf("Please provide a valid 6 digit Hex value for the color")
	}

	attributes := map[string]interface{}{"attributes": map[string]interface{}{"color": color}}
	err = cmd.CreateObject(path, cmd.DOMAIN, attributes)
	return nil, err
}

type createSiteNode struct {
	path node
}

func (n *createSiteNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	err = cmd.CreateObject(path, cmd.SITE, map[string]any{})
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	posXY, err := nodeToVec(n.posXY, 2, "posXY")
	if err != nil {
		return nil, err
	}
	rotation, err := nodeToFloat(n.rotation, "rotation")
	if err != nil {
		return nil, err
	}
	attributes := map[string]any{"posXY": posXY, "rotation": rotation}
	sizeOrTemplateAny, err := n.sizeOrTemplate.execute()
	if err != nil {
		return nil, err
	}
	size, ok := sizeOrTemplateAny.([]float64)
	if ok {
		attributes["size"] = size
	} else {
		template, ok := sizeOrTemplateAny.(string)
		if !ok {
			return nil, fmt.Errorf("vector3 (size) or string (template) expected")
		}
		if !checkIfTemplate(template, cmd.BLDG) {
			return nil, fmt.Errorf("template not found")
		}
		attributes["template"] = template
	}
	err = cmd.CreateObject(path, cmd.BLDG, map[string]any{"attributes": attributes})
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	posXY, err := nodeToVec(n.posXY, 2, "posXY")
	if err != nil {
		return nil, err
	}
	rotation, err := nodeToFloat(n.rotation, "rotation")
	if err != nil {
		return nil, err
	}
	attributes := map[string]any{"posXY": posXY, "rotation": rotation}
	if n.template != nil {
		template, err := nodeToString(n.template, "template")
		if err != nil {
			return nil, err
		}
		if !checkIfTemplate(template, cmd.ROOM) {
			return nil, fmt.Errorf("template not found")
		}
		attributes["template"] = template
	} else {
		size, err := nodeToVec(n.size, 3, "size")
		if err != nil {
			return nil, err
		}
		attributes["size"] = size
		axisOrientation, err := nodeToString(n.axisOrientation, "orientation")
		if err != nil {
			return nil, err
		}
		attributes["axisOrientation"] = axisOrientation
	}
	if n.floorUnit != nil {
		floorUnit, err := nodeToString(n.floorUnit, "floorUnit")
		if err != nil {
			return nil, err
		}
		attributes["floorUnit"] = floorUnit
	}
	err = cmd.CreateObject(path, cmd.ROOM, map[string]any{"attributes": attributes})
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	pos, err := nodeToVec(n.pos, -1, "position")
	if err != nil {
		return nil, err
	}
	if len(pos) != 2 && len(pos) != 3 {
		return nil, fmt.Errorf("position should be a vector2 or a vector3")
	}
	orientation, err := nodeToString(n.orientation, "orientation")
	if err != nil {
		return nil, err
	}
	attributes := map[string]any{"posXYZ": pos, "orientation": orientation}
	sizeOrTemplateAny, err := n.sizeOrTemplate.execute()
	if err != nil {
		return nil, err
	}
	size, ok := sizeOrTemplateAny.([]float64)
	if ok {
		attributes["size"] = size
	} else {
		template, ok := sizeOrTemplateAny.(string)
		if !ok {
			return nil, fmt.Errorf("vector3 (size) or string (template) expected")
		}
		if !checkIfTemplate(template, cmd.RACK) {
			return nil, fmt.Errorf("template not found")
		}
		attributes["template"] = template
	}
	err = cmd.CreateObject(path, cmd.RACK, map[string]any{"attributes": attributes})
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	posUOrSlot, err := nodeToString(n.posUOrSlot, "posU/slot")
	if err != nil {
		return nil, err
	}
	attr := map[string]any{"posU/slot": posUOrSlot}

	sizeUOrTemplate, err := n.sizeUOrTemplate.execute()
	if err != nil {
		return nil, err
	}
	sizeU, err := valToInt(sizeUOrTemplate, "sizeU")
	if err != nil {
		attr["sizeU"] = sizeU
	} else {
		template, ok := sizeUOrTemplate.(string)
		if !ok {
			return nil, fmt.Errorf("int (sizeU) or string (template) expected")
		}
		if !checkIfTemplate(template, cmd.DEVICE) {
			return nil, fmt.Errorf("template not found")
		}
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
	err = cmd.CreateObject(path, cmd.DEVICE, attributes)
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	var objs []string
	data := map[string]interface{}{}
	for i := range n.paths {
		obj, err := nodeToString(n.paths[i], "path")
		if err != nil {
			return nil, err
		}
		obj = filepath.Base(obj)
		objs = append(objs, obj)
	}
	data["attributes"] = map[string]interface{}{"content": objs}
	err = cmd.CreateObject(path, cmd.GROUP, data)
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
	path, err := nodeToString(n.path, "path for corridor")
	if err != nil {
		return nil, err
	}
	leftRack, err := nodeToString(n.leftRack, "path for left rack")
	if err != nil {
		return nil, err
	}
	rightRack, err := nodeToString(n.rightRack, "path for right rack")
	if err != nil {
		return nil, err
	}
	temp, err := nodeToString(n.temp, "temperature")
	if err != nil {
		return nil, err
	}
	leftRack = filepath.Base(leftRack)
	rightRack = filepath.Base(rightRack)
	attributes := map[string]any{
		"content":     leftRack + "," + rightRack,
		"temperature": temp,
	}
	data := map[string]interface{}{"attributes": attributes}
	err = cmd.CreateObject(path, cmd.CORRIDOR, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createOrphanNode struct {
	path     node
	template node
}

func (n *createOrphanNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	template, err := nodeToString(n.template, "template")
	if err != nil {
		return nil, err
	}
	if !checkIfTemplate(template, cmd.STRAY_DEV) {
		return nil, fmt.Errorf("template not found")
	}
	attributes := map[string]any{"template": template}
	err = cmd.CreateObject(path, cmd.STRAY_DEV, map[string]any{"attributes": attributes})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type createUserNode struct {
	email  node
	role   node
	domain node
}

func (n *createUserNode) execute() (interface{}, error) {
	email, err := nodeToString(n.email, "email")
	if err != nil {
		return nil, err
	}
	role, err := nodeToString(n.role, "role")
	if err != nil {
		return nil, err
	}
	domain, err := nodeToString(n.domain, "domain")
	if err != nil {
		return nil, err
	}
	err = cmd.CreateUser(email, role, domain)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type addRoleNode struct {
	email  node
	role   node
	domain node
}

func (n *addRoleNode) execute() (interface{}, error) {
	email, err := nodeToString(n.email, "email")
	if err != nil {
		return nil, err
	}
	role, err := nodeToString(n.role, "role")
	if err != nil {
		return nil, err
	}
	domain, err := nodeToString(n.domain, "domain")
	if err != nil {
		return nil, err
	}
	err = cmd.AddRole(email, role, domain)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type changePasswordNode struct{}

func (n *changePasswordNode) execute() (interface{}, error) {
	return nil, cmd.ChangePassword()
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
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
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
	position, err := nodeToVec(n.position, 3, "position")
	if err != nil {
		return nil, err
	}
	rotation, err := nodeToVec(n.rotation, 2, "rotation")
	if err != nil {
		return nil, err
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
	posUOrSlot  node
}

func (n *linkObjectNode) execute() (interface{}, error) {
	source, err := nodeToString(n.source, "source object path")
	if err != nil {
		return nil, err
	}
	dest, err := nodeToString(n.destination, "destination object path")
	if err != nil {
		return nil, err
	}
	var posUOrSlot string
	if n.posUOrSlot != nil {
		posUOrSlot, err = nodeToString(n.posUOrSlot, "posU/slot")
		if err != nil {
			return nil, err
		}
	}
	return nil, cmd.LinkObject(source, dest, posUOrSlot)
}

type unlinkObjectNode struct {
	source node
}

func (n *unlinkObjectNode) execute() (interface{}, error) {
	source, err := nodeToString(n.source, "source object path")
	if err != nil {
		return nil, err
	}
	return nil, cmd.UnlinkObject(source)
}

type symbolReferenceNode struct {
	va string
}

func (s *symbolReferenceNode) execute() (interface{}, error) {
	val, ok := c.State.DynamicSymbolTable[s.va]
	if !ok {
		return nil, fmt.Errorf("undefined variable %s", s.va)
	}
	return val, nil
}

type arrayReferenceNode struct {
	variable string
	idx      node
}

func (n *arrayReferenceNode) execute() (interface{}, error) {
	v, ok := c.State.DynamicSymbolTable[n.variable]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", n.variable)
	}
	arr, ok := v.([]float64)
	if !ok {
		return nil, fmt.Errorf("You can only index an array.")
	}
	i, err := nodeToInt(n.idx, "index")
	if err != nil {
		return nil, err
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
		c.State.DynamicSymbolTable[a.variable] = v
		if cmd.State.DebugLvl >= 3 {
			println("You want to assign", a.variable, "with value of", v)
		}
		return nil, nil
	}
	return nil, fmt.Errorf("Invalid type to assign variable %s", a.variable)
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
