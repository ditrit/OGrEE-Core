package main

import (
	"bytes"
	"cli/config"
	cmd "cli/controllers"
	"cli/models"
	"cli/utils"
	"cli/views"
	"encoding/json"
	"errors"
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
	cmd.State.DynamicSymbolTable = make(map[string]interface{})
	cmd.State.FuncTable = make(map[string]interface{})
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
	cmd.State.FuncTable[n.name] = n.body
	if cmd.State.DebugLvl >= 3 {
		println("New function ", n.name)
	}
	return nil, nil
}

type funcCallNode struct {
	name string
}

func (n *funcCallNode) execute() (interface{}, error) {
	val, ok := cmd.State.FuncTable[n.name]
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
	val, ok := cmd.State.DynamicSymbolTable[n.variable]
	if !ok {
		return nil, fmt.Errorf("Undefined variable %s", n.variable)
	}
	arr, err := utils.ValToVec(val, -1, "Variable "+n.variable)
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
	return nil, cmd.FocusUI(path)
}

type cdNode struct {
	path node
}

func (n *cdNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	return nil, cmd.C.CD(path)
}

type lsNode struct {
	path      *pathNode
	filters   map[string]node
	sortAttr  string
	recursive recursiveArgs
	attrList  []string
}

func (n *lsNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}

	filters, err := filtersToMapString(n.filters)
	if err != nil {
		return nil, err
	}

	pathEntered, err := n.path.Path()
	if err != nil {
		return nil, err
	}

	recursive, err := n.recursive.toParams(pathEntered)
	if err != nil {
		return nil, err
	}

	objects, err := cmd.C.Ls(path, filters, recursive)
	if err != nil {
		return nil, err
	}

	var relativePath *views.RelativePathArgs
	if n.recursive.isRecursive {
		relativePath = &views.RelativePathArgs{
			FromPath: path,
		}
	}

	if n.attrList == nil {
		n.attrList = []string{}
	}

	var toPrint string
	if len(n.attrList) == 0 {
		toPrint, err = views.Ls(objects, n.sortAttr, relativePath)
	} else {
		toPrint, err = views.LsWithFormat(objects, n.sortAttr, relativePath, n.attrList)
	}

	if err == nil {
		fmt.Print(toPrint)
	}

	return nil, err

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

	return nil, cmd.GetByAttr(path, u)
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

	return nil, cmd.GetByAttr(path, slot)
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
	return path, cmd.C.LoadTemplate(data)
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
	paths, err := cmd.C.DeleteObj(path)
	if err != nil {
		return nil, err
	}
	if len(paths) > 0 {
		fmt.Println("Objects deleted :")
		for _, path := range paths {
			fmt.Println(path)
		}
	} else {
		fmt.Println("Nothing got deleted")
	}
	return nil, nil
}

type deleteSelectionNode struct{}

func (n *deleteSelectionNode) execute() (interface{}, error) {
	var errBuilder strings.Builder
	deleted := 0
	if cmd.State.ClipBoard != nil {
		for _, obj := range cmd.State.ClipBoard {
			_, err := cmd.C.DeleteObj(obj)
			if err != nil {
				errBuilder.WriteString(fmt.Sprintf("    %s: %s\n", obj, err.Error()))
			} else {
				deleted += 1
			}
		}
	}
	println(fmt.Sprintf("%d objects deleted", deleted))
	notDeleted := len(cmd.State.ClipBoard) - deleted
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
	path      *pathNode
	filters   map[string]node
	recursive recursiveArgs
}

func (n *getObjectNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}

	filters, err := filtersToMapString(n.filters)
	if err != nil {
		return nil, err
	}

	pathEntered, err := n.path.Path()
	if err != nil {
		return nil, err
	}

	recursive, err := n.recursive.toParams(pathEntered)
	if err != nil {
		return nil, err
	}

	objs, _, err := cmd.C.GetObjectsWildcard(path, filters, recursive)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(path, "*") && !models.PathIsLayer(path) && len(objs) == 0 {
		return nil, errors.New("object not found")
	}

	objs, err = views.SortObjects(objs, "")
	if err != nil {
		return nil, err
	}

	for _, obj := range objs {
		views.Object(path, obj)
	}

	return objs, nil
}

type selectObjectNode struct {
	path node
}

func (n *selectObjectNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}

	selection, err := cmd.C.Select(path)
	if err != nil {
		return nil, err
	}

	if len(selection) == 0 {
		fmt.Println("Selection is now empty")
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
	return cmd.C.UpdateObj(path, map[string]any{"attributes": attributes})
}

func setLabel(path string, values []any, hasSharpe bool) (map[string]any, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("only 1 value expected")
	}
	value, err := utils.ValToString(values[0], "value")
	if err != nil {
		return nil, err
	}
	return nil, cmd.C.InteractObject(path, "label", value, hasSharpe)
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
		return nil, cmd.C.InteractObject(path, "labelFont", values[0], false)
	case 2:
		if values[0] != "color" {
			return nil, fmt.Errorf(msg)
		}
		c, ok := utils.ValToColor(values[1])
		if !ok {
			return nil, fmt.Errorf("please provide a valid 6 length hex value for the color")
		}
		return nil, cmd.C.InteractObject(path, "labelFont", "color@"+c, false)
	default:
		return nil, fmt.Errorf(msg)
	}
}

func setLabelBackground(path string, values []any) (map[string]any, error) {
	if len(values) != 1 {
		return nil, fmt.Errorf("only 1 value expected")
	}
	c, ok := utils.ValToColor(values[0])
	if !ok {
		return nil, fmt.Errorf("please provide a valid 6 length hex value for the color")
	}
	return nil, cmd.C.InteractObject(path, "labelBackground", c, false)
}

func addToStringMap[T any](stringMap string, key string, val T) (string, bool) {
	m := map[string]T{}
	if stringMap != "" {
		json.Unmarshal([]byte(stringMap), &m)
	}
	_, keyExist := m[key]
	m[key] = val
	mBytes, _ := json.Marshal(m)
	return string(mBytes), keyExist
}

func removeFromStringMap[T any](stringMap string, key string) (string, bool) {
	m := map[string]T{}
	if stringMap != "" {
		json.Unmarshal([]byte(stringMap), &m)
	}
	_, ok := m[key]
	if !ok {
		return stringMap, false
	}
	delete(m, key)
	mBytes, _ := json.Marshal(m)
	return string(mBytes), true
}

type Separator struct {
	StartPos []float64 `json:"startPosXYm"`
	EndPos   []float64 `json:"endPosXYm"`
	Type     string    `json:"type"`
}

func addRoomSeparator(path string, values []any) (map[string]any, error) {
	if len(values) != 4 {
		return nil, fmt.Errorf("4 values (name, startPos, endPos, type) expected to add a separator")
	}
	name, err := utils.ValToString(values[0], "name")
	if err != nil {
		return nil, err
	}
	startPos, err := utils.ValToVec(values[1], 2, "startPos")
	if err != nil {
		return nil, err
	}
	endPos, err := utils.ValToVec(values[2], 2, "endPos")
	if err != nil {
		return nil, err
	}
	sepType, err := utils.ValToString(values[3], "separator type")
	if err != nil {
		return nil, err
	}
	obj, err := cmd.C.GetObject(path)
	if err != nil {
		return nil, err
	}
	attr := obj["attributes"].(map[string]any)
	separators, _ := attr["separators"].(string)
	newSeparator := Separator{startPos, endPos, sepType}
	var keyExist bool
	attr["separators"], keyExist = addToStringMap[Separator](separators, name, newSeparator)
	obj, err = cmd.C.UpdateObj(path, map[string]any{"attributes": attr})
	if err != nil {
		return nil, err
	}
	if keyExist {
		fmt.Printf("Separator %s replaced\n", name)
	}
	return obj, nil
}

type Pillar struct {
	CenterXY []float64 `json:"centerXY"`
	SizeXY   []float64 `json:"sizeXY"`
	Rotation float64   `json:"rotation"`
}

func addRoomPillar(path string, values []any) (map[string]any, error) {
	if len(values) != 4 {
		return nil, fmt.Errorf("4 values (name, centerXY, sizeXY, rotation) expected to add a pillar")
	}
	name, err := utils.ValToString(values[0], "name")
	if err != nil {
		return nil, err
	}
	centerXY, err := utils.ValToVec(values[1], 2, "centerXY")
	if err != nil {
		return nil, err
	}
	sizeXY, err := utils.ValToVec(values[2], 2, "sizeXY")
	if err != nil {
		return nil, err
	}
	rotation, err := utils.ValToFloat(values[3], "rotation")
	if err != nil {
		return nil, err
	}
	obj, err := cmd.C.GetObject(path)
	if err != nil {
		return nil, err
	}
	attr := obj["attributes"].(map[string]any)
	pillars, _ := attr["pillars"].(string)
	newPillar := Pillar{centerXY, sizeXY, rotation}
	var keyExist bool
	attr["pillars"], keyExist = addToStringMap[Pillar](pillars, name, newPillar)
	obj, err = cmd.C.UpdateObj(path, map[string]any{"attributes": attr})
	if err != nil {
		return nil, err
	}
	if keyExist {
		fmt.Printf("Pillar %s replaced\n", name)
	}
	return obj, nil
}

// attribute must be "separator" or "pillar"
func deleteRoomPillarOrSeparator(path, attribute, name string) (map[string]any, error) {
	obj, err := cmd.C.GetObject(path)
	if err != nil {
		return nil, err
	}
	attributes := obj["attributes"].(map[string]any)
	stringMap, _ := attributes[attribute+"s"].(string)
	var ok bool
	switch attribute {
	case "pillar":
		stringMap, ok = removeFromStringMap[Pillar](stringMap, name)
	case "separator":
		stringMap, ok = removeFromStringMap[Separator](stringMap, name)
	default:
		return nil, errors.New("\"separator\" or \"pillar\" expected")
	}

	if !ok {
		return nil, fmt.Errorf("%s %s does not exist", attribute, name)
	}
	attributes[attribute+"s"] = stringMap
	return cmd.C.UpdateObj(path, map[string]any{"attributes": attributes})
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
	newDesc, err := utils.ValToString(values[0], "description")
	if err != nil {
		return nil, err
	}
	data := map[string]any{}
	if attr == "description" {
		data["description"] = newDesc
	} else {
		obj, err := cmd.C.GetObject(path)
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
	return cmd.C.UpdateObj(path, data)
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
	paths, err := cmd.C.UnfoldPath(path)
	if err != nil {
		return nil, err
	}
	for _, path := range paths {
		var err error
		if models.IsTag(path) {
			if n.attr == "slug" || n.attr == "color" || n.attr == "description" {
				_, err = cmd.C.UpdateObj(path, map[string]any{n.attr: values[0]})
			}
		} else if models.IsLayer(path) {
			err = cmd.C.UpdateLayer(path, n.attr, values[0])
		} else {
			switch n.attr {
			case "content", "alpha", "tilesName", "tilesColor", "U", "slots", "localCS":
				var boolVal bool
				boolVal, err = utils.ValToBool(values[0], n.attr)
				if err != nil {
					return nil, err
				}
				err = cmd.C.InteractObject(path, n.attr, boolVal, n.hasSharpe)
			case "areas":
				_, err = setRoomAreas(path, values)
			case "label":
				_, err = setLabel(path, values, n.hasSharpe)
			case "labelFont":
				_, err = setLabelFont(path, values)
			case "labelBackground":
				_, err = setLabelBackground(path, values)
			case "separators+":
				_, err = addRoomSeparator(path, values)
			case "pillars+":
				_, err = addRoomPillar(path, values)
			case "separators-":
				_, err = deleteRoomPillarOrSeparator(path, "separator", values[0].(string))
			case "pillars-":
				_, err = deleteRoomPillarOrSeparator(path, "pillar", values[0].(string))
			case "domain", "tags+", "tags-":
				_, err = cmd.C.UpdateObj(path, map[string]any{n.attr: values[0]})
			case "tags", "separators", "pillars":
				err = fmt.Errorf(
					"object's %[1]s can not be updated directly, please use %[1]s+= and %[1]s-=",
					n.attr,
				)
			default:
				if strings.HasPrefix(n.attr, "description") {
					_, err = updateDescription(path, n.attr, values)
				} else {
					_, err = updateAttributes(path, n.attr, values)
				}
			}
		}

		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func updateAttributes(path, attributeName string, values []any) (map[string]any, error) {
	var attributes map[string]any
	if attributeName == "slot" {
		slots := []string{}
		for _, value := range values {
			slots = append(slots, value.(string))
		}
		value := "[" + strings.Join(slots, ",") + "]"
		attributes = map[string]any{attributeName: value}
	} else {
		if len(values) > 1 {
			return nil, fmt.Errorf("attributes can only be assigned a single value")
		}
		attributes = map[string]any{attributeName: values[0]}
	}

	return cmd.C.UpdateObj(path, map[string]any{"attributes": attributes})
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
	root, err := cmd.C.Tree(path, n.depth)
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

	return nil, cmd.C.Draw(path, n.depth, n.force)
}

type undrawNode struct {
	path node
}

func (n *undrawNode) execute() (interface{}, error) {
	if n.path == nil {
		return nil, cmd.C.Undraw("")
	}

	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}

	return nil, cmd.C.Undraw(path)
}

type lsogNode struct{}

func (n *lsogNode) execute() (interface{}, error) {
	return nil, cmd.LSOG()
}

type lsenterpriseNode struct{}

func (n *lsenterpriseNode) execute() (interface{}, error) {
	return nil, cmd.LSEnterprise()
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
	cmd.Env(cmd.State.DynamicSymbolTable, cmd.State.FuncTable)
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

type selectChildrenNode struct {
	paths []node
}

func (n *selectChildrenNode) execute() (interface{}, error) {
	paths, err := evalNodeArr[string](&n.paths, []string{})
	if err != nil {
		return nil, err
	}
	v, err := cmd.C.SetClipBoard(paths)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		fmt.Println("Selection is now empty")

	} else if cmd.State.DebugLvl > cmd.NONE {
		fmt.Println("Selection made")
	}
	return v, nil
}

type unsetFuncNode struct {
	funcName string
}

func (n *unsetFuncNode) execute() (interface{}, error) {
	delete(cmd.State.FuncTable, n.funcName)
	return nil, nil
}

type unsetVarNode struct {
	varName string
}

func (n *unsetVarNode) execute() (interface{}, error) {
	delete(cmd.State.DynamicSymbolTable, n.varName)
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

	color, err := nodeToColorString(n.color)
	if err != nil {
		return nil, err
	}

	attributes := map[string]interface{}{"attributes": map[string]interface{}{"color": color}}

	return nil, cmd.C.CreateObject(path, models.DOMAIN, attributes)
}

type createSiteNode struct {
	path node
}

func (n *createSiteNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}

	return nil, cmd.C.CreateObject(path, models.SITE, map[string]any{})
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

	addSizeOrTemplate(n.sizeOrTemplate, attributes, models.BLDG)

	return nil, cmd.C.CreateObject(path, models.BLDG, map[string]any{"attributes": attributes})
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

		attributes["template"] = template
	} else {
		size, err := nodeToSize(n.size)
		if err != nil {
			return nil, err
		}
		attributes["size"] = size

		axisOrientation, err := nodeToString(n.axisOrientation, "orientation")
		if err != nil {
			return nil, err
		}
		attributes["axisOrientation"] = axisOrientation

		if n.floorUnit != nil {
			floorUnit, err := nodeToString(n.floorUnit, "floorUnit")
			if err != nil {
				return nil, err
			}
			attributes["floorUnit"] = floorUnit
		}
	}

	return nil, cmd.C.CreateObject(path, models.ROOM, map[string]any{"attributes": attributes})
}

type createRackNode struct {
	path           node
	pos            node
	unit           node
	rotation       node
	sizeOrTemplate node
}

func (n *createRackNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}

	pos, err := nodeToPosXYZ(n.pos)
	if err != nil {
		return nil, err
	}

	unit, err := nodeToString(n.unit, "unit")
	if err != nil {
		return nil, err
	}

	rotation, err := nodeTo3dRotation(n.rotation)
	if err != nil {
		return nil, err
	}

	attributes := map[string]any{"posXYZ": pos, "posXYUnit": unit, "rotation": rotation}

	addSizeOrTemplate(n.sizeOrTemplate, attributes, models.RACK)

	return nil, cmd.C.CreateObject(path, models.RACK, map[string]any{"attributes": attributes})
}

type createGenericNode struct {
	path           node
	pos            node
	unit           node
	rotation       node
	sizeOrTemplate node
	shape          node
	getype         node
}

func (n *createGenericNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}

	pos, err := nodeToPosXYZ(n.pos)
	if err != nil {
		return nil, err
	}

	unit, err := nodeToString(n.unit, "unit")
	if err != nil {
		return nil, err
	}

	rotation, err := nodeTo3dRotation(n.rotation)
	if err != nil {
		return nil, err
	}

	attributes := map[string]any{"posXYZ": pos, "posXYUnit": unit, "rotation": rotation}

	if n.shape != nil {
		shape, err := nodeToString(n.shape, "shape")
		if err != nil {
			return nil, err
		}
		attributes["shape"] = shape

		getype, err := nodeToString(n.getype, "type")
		if err != nil {
			return nil, err
		}
		attributes["type"] = getype
	}

	addSizeOrTemplate(n.sizeOrTemplate, attributes, models.GENERIC)

	return nil, cmd.C.CreateObject(path, models.GENERIC, map[string]any{"attributes": attributes})
}

type createDeviceNode struct {
	path            node
	posUOrSlot      []node
	sizeUOrTemplate node
	invertOffset    bool
	side            node
}

func (n *createDeviceNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}
	posUOrSlot := []string{}
	for _, node := range n.posUOrSlot {
		str, err := nodeToString(node, "posU/slot")
		posUOrSlot = append(posUOrSlot, str)
		if err != nil {
			return nil, err
		}
	}

	attributes := map[string]any{"posU/slot": posUOrSlot}

	sizeU, err := nodeToInt(n.sizeUOrTemplate, "sizeU")
	if err == nil {
		attributes["sizeU"] = sizeU
	} else {
		template, err := nodeToString(n.sizeUOrTemplate, "template")
		if err != nil {
			if errors.Is(err, utils.ErrShouldBeAString) {
				return nil, errors.New("int (sizeU) or string (template) expected")
			}

			return nil, err
		}

		attributes["template"] = template
	}

	if n.invertOffset {
		attributes["invertOffset"] = "true"
	} else {
		attributes["invertOffset"] = "false"
	}

	if n.side != nil {
		side, err := n.side.execute()
		if err != nil {
			return nil, err
		}
		attributes["orientation"] = side
	}

	return nil, cmd.C.CreateObject(path, models.DEVICE, map[string]any{"attributes": attributes})
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

	return nil, cmd.C.CreateObject(path, models.GROUP, data)
}

type createTagNode struct {
	slug  node
	color node
}

func (n *createTagNode) execute() (interface{}, error) {
	slug, err := nodeToString(n.slug, "slug")
	if err != nil {
		return nil, err
	}

	color, err := nodeToColorString(n.color)
	if err != nil {
		return nil, err
	}

	return nil, cmd.CreateTag(slug, color)
}

type createLayerNode struct {
	slug          node
	applicability node
	filterName    string
	filterValue   node
}

func (n *createLayerNode) execute() (interface{}, error) {
	slug, err := nodeToString(n.slug, "slug")
	if err != nil {
		return nil, err
	}

	applicability, err := nodeToString(n.applicability, models.LayerApplicability)
	if err != nil {
		return nil, err
	}

	filterValue, err := nodeToString(n.filterValue, "filterValue")
	if err != nil {
		return nil, err
	}

	return nil, cmd.C.CreateLayer(slug, applicability, n.filterName, filterValue)
}

type createCorridorNode struct {
	path     node
	pos      node
	unit     node
	rotation node
	size     node
	temp     node
}

func (n *createCorridorNode) execute() (interface{}, error) {
	path, err := nodeToString(n.path, "path")
	if err != nil {
		return nil, err
	}

	pos, err := nodeToPosXYZ(n.pos)
	if err != nil {
		return nil, err
	}

	unit, err := nodeToString(n.unit, "unit")
	if err != nil {
		return nil, err
	}

	rotation, err := nodeTo3dRotation(n.rotation)
	if err != nil {
		return nil, err
	}

	size, err := nodeToSize(n.size)
	if err != nil {
		return nil, err
	}

	temp, err := nodeToString(n.temp, "temperature")
	if err != nil {
		return nil, err
	}
	attributes := map[string]any{"posXYZ": pos, "posXYUnit": unit, "rotation": rotation, "size": size, "temperature": temp}

	return nil, cmd.C.CreateObject(path, models.CORRIDOR, map[string]any{"attributes": attributes})
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

	attributes := map[string]any{"template": template}

	return nil, cmd.C.CreateObject(path, models.STRAY_DEV, map[string]any{"attributes": attributes})
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

	return nil, cmd.CreateUser(email, role, domain)
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

	return nil, cmd.AddRole(email, role, domain)
}

type changePasswordNode struct{}

func (n *changePasswordNode) execute() (interface{}, error) {
	return nil, cmd.ChangePassword()
}

type connect3DNode struct {
	url string
}

func (n *connect3DNode) execute() (interface{}, error) {
	return nil, cmd.Connect3D(n.url)
}

type disconnect3DNode struct{}

func (n *disconnect3DNode) execute() (interface{}, error) {
	cmd.Disconnect3D()
	return nil, nil
}

type uiDelayNode struct {
	time float64
}

func (n *uiDelayNode) execute() (interface{}, error) {
	return nil, cmd.UIDelay(n.time)
}

type uiToggleNode struct {
	feature string
	enable  bool
}

func (n *uiToggleNode) execute() (interface{}, error) {
	return nil, cmd.UIToggle(n.feature, n.enable)
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

	return nil, cmd.CameraMove(n.command, position, rotation)
}

type cameraWaitNode struct {
	time float64
}

func (n *cameraWaitNode) execute() (interface{}, error) {
	return nil, cmd.CameraWait(n.time)
}

type linkObjectNode struct {
	source      node
	destination node
	slot        []node
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

	var slot []string
	if n.slot != nil {
		slot = []string{}
		for _, node := range n.slot {
			str, err := nodeToString(node, "posU/slot")
			slot = append(slot, str)
			if err != nil {
				return nil, err
			}
		}
	}

	return nil, cmd.LinkObject(source, dest, slot)
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
	val, ok := cmd.State.DynamicSymbolTable[s.va]
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
	v, ok := cmd.State.DynamicSymbolTable[n.variable]
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
		cmd.State.DynamicSymbolTable[a.variable] = v
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
				// [front/top, back/bottom, right, left]
				reservedStr = "[" + r[0].String() + ", " + r[1].String() + ", " + r[2].String() + ", " + r[3].String() + "]"
				techStr = "[" + t[0].String() + ", " + t[1].String() + ", " + t[2].String() + ", " + t[3].String() + "]"
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

type cpNode struct {
	source node
	dest   node
}

func (n *cpNode) execute() (interface{}, error) {
	source, err := nodeToString(n.source, "source")
	if err != nil {
		return nil, err
	}

	dest, err := nodeToString(n.dest, "dest")
	if err != nil {
		return nil, err
	}

	return nil, cmd.C.Cp(source, dest)
}
