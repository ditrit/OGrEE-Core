package main

import (
	"bytes"
	cmd "cli/controllers"
	l "cli/logger"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
)

var dynamicMap = make(map[string]int)
var dynamicSymbolTable = make(map[int]interface{})
var funcTable = make(map[string]interface{})
var dCatchPtr interface{}
var dCatchNodePtr interface{}
var varCtr = 1 //Started at 1 because unset could cause var data loss

type node interface {
	execute() interface{}
}

type array interface {
	node
	getLength() int
}

type postObjNode struct {
	entity string
	data   map[string]interface{}
}

func (n *postObjNode) execute() interface{} {
	v := cmd.PostObj(cmd.EntityStrToInt(n.entity), n.entity, n.data)
	return &jsonObjNode{v}
}

type easyPostNode struct {
	entity string
	path   string
}

func (n *easyPostNode) execute() interface{} {
	var data map[string]interface{}
	/*x, e := ioutil.ReadFile(n.path)
	if e != nil {
		println("Error while opening file! " + e.Error())
		return nil
	}
	json.Unmarshal(x, &data)*/
	data = fileToJSON(n.path)
	if data == nil {
		return nil
	}
	v := cmd.PostObj(cmd.EntityStrToInt(n.entity), n.entity, data)
	return &jsonObjNode{v}
}

type helpNode struct {
	entry string
}

func (n *helpNode) execute() interface{} {
	cmd.Help(n.entry)
	return nil
}

type focusNode struct {
	path string
}

func (n *focusNode) execute() interface{} {
	cmd.FocusUI(n.path)
	return nil
}

type cdNode struct {
	path string
}

func (n *cdNode) execute() interface{} {
	v := cmd.CD(n.path)
	return &strNode{v}
}

type lsNode struct {
	path string
}

func (n *lsNode) execute() interface{} {
	v := cmd.LS(n.path)
	return &jsonObjArrNode{len(v), v}
}

type loadNode struct {
	path string
}

func (n *loadNode) execute() interface{} {
	v := cmd.LoadFile(n.path)
	return &strNode{v}
}

type loadTemplateNode struct {
	path string
}

func (n *loadTemplateNode) execute() interface{} {
	data := fileToJSON(n.path)
	if data == nil {
		return nil
	}
	cmd.LoadTemplate(data, n.path)
	return &strNode{n.path}
}

type printNode struct {
	args []interface{}
}

func (n *printNode) execute() interface{} {
	res := []interface{}{}
	for i := range n.args {
		res = append(res, n.args[i].(node).execute())
	}
	v := cmd.Print(res)
	return &strNode{v}
}

type deleteObjNode struct {
	path string
}

func (n *deleteObjNode) execute() interface{} {
	v := cmd.DeleteObj(n.path)
	return &boolNode{v}
}

type deleteSelectionNode struct{}

func (n *deleteSelectionNode) execute() interface{} {
	return cmd.DeleteSelection()
}

type isEntityDrawableNode struct {
	path string
}

func (n *isEntityDrawableNode) execute() interface{} {
	return cmd.IsEntityDrawable(n.path)
}

type isAttrDrawableNode struct {
	objInf string
	factor node
}

func (n *isAttrDrawableNode) execute() interface{} {
	attrInf := n.factor.execute()
	if _, ok := attrInf.(string); !ok {
		println("Attribute operand is invalid")
		l.GetInfoLogger().Println("Attribute operand is invalid")
		return nil
	}
	return cmd.IsAttrDrawable(n.objInf, attrInf.(string), nil, false)
}

type getObjectNode struct {
	path string
}

func (n *getObjectNode) execute() interface{} {
	v, _ := cmd.GetObject(n.path, false)
	return &jsonObjNode{v}
}

type searchObjectsNode struct {
	objType string
	resMap  map[string]interface{}
}

func (n *searchObjectsNode) execute() interface{} {
	v := cmd.SearchObjects(n.objType, n.resMap)
	return &jsonObjArrNode{len(v), v}
}

type recursiveUpdateObjNode struct {
	arg0 interface{}
	arg1 interface{}
	arg2 interface{}
}

func (n *recursiveUpdateObjNode) execute() interface{} {
	//Old code was removed since
	//it broke the OCLI syntax easy update
	if _, ok := n.arg2.(bool); ok {
		var objMap map[string]interface{}
		//Weird edge case
		//to solve issue with:
		// for i in $(ls) do $i[attr]="string"

		//n.arg0 = referenceToNode
		//n.arg1 = attributeString, (used as an index)
		//n.arg2 = someValue (usually a string)
		objMap = n.arg0.(node).execute().(map[string]interface{})
		if checkIfObjectNode(objMap) == true {
			updateArgs := map[string]interface{}{n.arg1.(string): n.arg2.(node).execute()}
			id := objMap["id"].(string)
			entity := objMap["category"].(string)
			cmd.RecursivePatch("", id, entity, updateArgs)
		}

	} else {
		if n.arg2.(string) == "recursive" {
			cmd.RecursivePatch(n.arg0.(string), "", "", n.arg1.(map[string]interface{}))
		}
	}
	return nil
	//return &jsonObjNode{v}
}

type updateObjNode struct {
	args []interface{}
}

func (n *updateObjNode) execute() interface{} {
	var v map[string]interface{}
	//Old code was removed since
	//it broke the OCLI syntax easy update
	x := len(n.args)
	if _, ok := n.args[x-1].(bool); ok {
		var objMap map[string]interface{}
		//Weird edge case
		//to solve issue with:
		// for i in $(ls) do $i[attr]="string"

		//n.args[0] = referenceToNode
		//n.args[1] = attributeString, (used as an index)
		//n.args[2] = someValue (usually a string)
		mp := n.args[0]
		objMap = mp.(node).execute().(map[string]interface{})

		if checkIfObjectNode(objMap) == true {
			updateArgs := map[string]interface{}{n.args[1].(string): n.args[2].(node).execute()}
			id := objMap["id"].(string)
			entity := objMap["category"].(string)
			v = cmd.UpdateObj("", id, entity, updateArgs, false)
		}

	} else {
		if rawArr, ok := n.args[1].(map[string]interface{}); ok {
			if areas, ok := rawArr["areas"].(map[string]interface{}); ok {
				n.args[1] = parseAreas(areas)
			}
		}
		v = cmd.UpdateObj(n.args[0].(string), "", "", n.args[1].(map[string]interface{}), false)
	}
	return &jsonObjNode{v}
}

type easyUpdateNode struct {
	nodePath     string
	jsonPath     string
	deleteAndPut bool
}

func (n *easyUpdateNode) execute() interface{} {
	var data map[string]interface{}
	data = fileToJSON(n.jsonPath)
	if data == nil {
		return nil
	}
	v := cmd.UpdateObj(n.nodePath, "", "", data, n.deleteAndPut)
	return &jsonObjNode{v}
}

type lsObjNode struct {
	path   string
	entity int
}

func (n *lsObjNode) execute() interface{} {
	v := cmd.LSOBJECT(n.path, n.entity)
	//return &objNdArrNode{len(v), v}
	return &jsonObjArrNode{len(v), v}
}

type treeNode struct {
	path  string
	depth int
}

func (n *treeNode) execute() interface{} {
	cmd.Tree(n.path, n.depth)
	return nil
}

type drawNode struct {
	path  string
	depth int
}

func (n *drawNode) execute() interface{} {
	cmd.Draw(n.path, n.depth)
	return nil
}

type lsogNode struct{}

func (n *lsogNode) execute() interface{} {
	cmd.LSOG()
	return nil
}

type exitNode struct{}

func (n *exitNode) execute() interface{} {
	cmd.Exit()
	return nil
}

type clrNode struct{}

func (n *clrNode) execute() interface{} {
	cmd.Clear()
	return nil
}

type envNode struct{}

func (n *envNode) execute() interface{} {
	cmd.Env()
	return nil
}

type selectNode struct{}

func (n *selectNode) execute() interface{} {
	v := cmd.ShowClipBoard()
	return &strArrNode{len(v), v}
}

type pwdNode struct{}

func (n *pwdNode) execute() interface{} {
	v := cmd.PWD()
	return &strNode{v}
}

type grepNode struct{}

func (n *grepNode) execute() interface{} {
	return nil
}

type setCBNode struct {
	cb *[]string
}

func (n *setCBNode) execute() interface{} {
	v := cmd.SetClipBoard(n.cb)
	return &strArrNode{len(v), v}
}

type updateSelectNode struct {
	data map[string]interface{}
}

func (n *updateSelectNode) execute() interface{} {
	cmd.UpdateSelection(n.data)
	return nil
}

type unsetNode struct {
	x     string
	name  string
	ref   interface{}
	value interface{}
}

func (n *unsetNode) execute() interface{} {
	UnsetUtil(n.x, n.name, n.ref, n.value)
	return nil
}

type setEnvNode struct {
	arg  string
	expr node
}

func (n *setEnvNode) execute() interface{} {
	val := n.expr.execute()
	cmd.SetEnv(n.arg, val)
	return nil
}

type hierarchyNode struct {
	path  string
	depth int
}

func (n *hierarchyNode) execute() interface{} {
	cmd.GetHierarchy(n.path, n.depth, false)
	return nil
}

type getOCAttrNode struct {
	path       string
	ent        int
	attributes map[string]interface{}
}

func (n *getOCAttrNode) execute() interface{} {
	//Since the attributes is a map of nodes
	//execute and receive the values
	evalMapNodes(n.attributes)
	cmd.GetOCLIAtrributes(n.path, n.ent, n.attributes)
	return nil
}

type handleUnityNode struct {
	args []interface{}
}

func (n *handleUnityNode) execute() interface{} {
	data := map[string]interface{}{}
	data["command"] = n.args[1].(string)
	if len(n.args) == 4 {

		firstArr := n.args[2].([]map[int]interface{})
		secondArr := n.args[3].([]map[int]interface{})

		if len(firstArr) != 3 || len(secondArr) != 2 {
			println("OGREE: Error, command args are invalid")
			print("Please provide a vector3 and a vector2")
			return nil
		}

		pos := map[string]interface{}{"x": firstArr[0][0].(node).execute(),
			"y": firstArr[1][0].(node).execute(), "z": firstArr[2][0].(node).execute(),
		}

		rot := map[string]interface{}{"x": secondArr[0][0].(node).execute(),
			"y": secondArr[1][0].(node).execute(),
		}

		data["position"] = pos
		data["rotation"] = rot

	} else {
		if n.args[1].(string) == "wait" && n.args[0].(string) == "camera" {
			data["position"] = map[string]float64{"x": 0, "y": 0, "z": 0}

			if y, ok := n.args[2].([]map[int]interface{}); ok {
				data["rotation"] = map[string]interface{}{"x": 999,
					"y": y[0][0].(node).execute()}
			} else {
				data["rotation"] = map[string]interface{}{"x": 999,
					"y": n.args[2]}
			}

		} else {
			if _, ok := n.args[2].([]map[int]interface{}); ok {
				data["data"] = n.args[2].([]map[int]interface{})[0][0].(node).execute()
			} else {
				data["data"] = n.args[2]
			}

		}
	}
	fullJson := map[string]interface{}{
		"type": n.args[0].(string),
		"data": data,
	}
	cmd.HandleUI(fullJson)
	return nil
}

type linkObjectNode struct {
	paths []interface{}
}

func (n *linkObjectNode) execute() interface{} {
	if len(n.paths) == 3 {
		n.paths[2] = n.paths[2].(node).execute()
	}
	cmd.LinkObject(n.paths)
	return nil
}

type unlinkObjectNode struct {
	paths []interface{}
}

func (n *unlinkObjectNode) execute() interface{} {
	cmd.UnlinkObject(n.paths)
	return nil
}

type commonNode struct {
	fun  interface{}
	val  string
	args []interface{}
}

type arrNode struct {
	len int
	val []map[int]interface{}
}

func (a *arrNode) execute() interface{} {
	return a.val
}

func (a *arrNode) getLength() int {
	return a.len
}

type objNdNode struct {
	val *cmd.Node
}

func (n *objNdNode) execute() interface{} {
	return n.val
}

type objNdArrNode struct {
	len int
	val []*cmd.Node
}

func (o *objNdArrNode) execute() interface{} {
	return o.val
}

func (o *objNdArrNode) getLength() int {
	return o.len
}

type jsonObjNode struct {
	val map[string]interface{}
}

func (j *jsonObjNode) execute() interface{} {
	return j.val
}

type jsonObjArrNode struct {
	len int
	val []map[string]interface{}
}

func (j *jsonObjArrNode) execute() interface{} {
	return j.val
}

func (j *jsonObjArrNode) getLength() int {
	return j.len
}

type numNode struct {
	val int
}

func (n *numNode) execute() interface{} {
	return n.val
}

type floatNode struct {
	val float64
}

func (f *floatNode) execute() interface{} {
	return f.val
}

type strNode struct {
	val string
}

func (s *strNode) execute() interface{} {
	return s.val
}

type strArrNode struct {
	len int
	val []string
}

func (s *strArrNode) execute() interface{} {
	return s.val
}

func (s *strArrNode) getLength() int {
	return s.len
}

type boolNode struct {
	val bool
}

func (b *boolNode) execute() interface{} {
	return b.val
}

type boolOpNode struct {
	op      string
	operand interface{}
}

func (b *boolOpNode) execute() interface{} {
	if b.op == "!" {
		//return !operand.(node).execute().(bool)
		if val, ok := b.operand.(node); ok {
			if v, ok := val.execute().(bool); ok {
				return v
			}
		}
	}
	return nil
}

type arithNode struct {
	op    interface{}
	left  interface{}
	right interface{}
}

func (a *arithNode) execute() interface{} {
	if v, ok := a.op.(string); ok {
		lv := (a.left.(node).execute())
		if cmd.State.DebugLvl >= 3 {
			println("Left:", lv)
		}
		rv := (a.right.(node).execute())
		if cmd.State.DebugLvl >= 3 {
			println("Right: ", rv)
		}

		switch v {
		case "+":
			if checkTypesAreSame(lv, rv) == true {
				switch lv.(type) {
				case int:
					return lv.(int) + rv.(int)
				case float64:
					return lv.(float64) + rv.(float64)
				case float32:
					return lv.(float64) + rv.(float64)
				case string:
					return lv.(string) + rv.(string)
				}
			} else if checkTypeAreNumeric(lv, rv) == true {
				if _, ok := lv.(float64); ok {
					return lv.(float64) + float64(rv.(int))
				} else {
					return rv.(float64) + float64(lv.(int))
				}
			} else { //we have string and numeric type
				//this code occurs when assigning and not
				//when using + while printing

				switch lv.(type) {
				case int:
					return strconv.Itoa(lv.(int)) + rv.(string)
				case float64:
					return strconv.FormatFloat(lv.(float64), 'f', -1, 64) + rv.(string)
				}

				switch rv.(type) {
				case int:
					return lv.(string) + strconv.Itoa(rv.(int))
				case float64:
					return lv.(string) + strconv.FormatFloat(rv.(float64), 'f', -1, 64)
				}
			}
			//Otherwise the types are incompatible so return nil
			//TODO:see if team would want to have bool support
			return nil

		case "-":
			if checkTypesAreSame(lv, rv) == true {
				switch lv.(type) {
				case int:
					return lv.(int) - rv.(int)
				case float64:
					return lv.(float64) - rv.(float64)
				case float32:
					return lv.(float64) - rv.(float64)
				}
			} else if checkTypeAreNumeric(lv, rv) == true {
				if _, ok := lv.(float64); ok {
					return lv.(float64) - float64(rv.(int))
				} else {
					return float64(lv.(int)) - rv.(float64)
				}
			}

		case "*":
			if checkTypesAreSame(lv, rv) == true {
				switch lv.(type) {
				case int:
					return lv.(int) * rv.(int)
				case float64:
					return lv.(float64) * rv.(float64)
				case float32:
					return lv.(float64) * rv.(float64)
				}
			} else if checkTypeAreNumeric(lv, rv) == true {
				if _, ok := lv.(float64); ok {
					return lv.(float64) * float64(rv.(int))
				} else {
					return float64(lv.(int)) * rv.(float64)
				}
			}

		case "%":
			if checkTypesAreSame(lv, rv) == true {
				switch lv.(type) {
				case int:
					return lv.(int) % rv.(int)
				case float64:
					return int(lv.(float64)) % int(rv.(float64))
				case float32:
					return int(lv.(float32)) % int(rv.(float32))
				}
			} else if checkTypeAreNumeric(lv, rv) == true {
				if _, ok := lv.(float64); ok {
					return int(lv.(float64)) % rv.(int)
				} else {
					return lv.(int) % int(rv.(float64))
				}
			}

		case "/":
			if checkTypesAreSame(lv, rv) == true {
				switch lv.(type) {
				case int:
					return lv.(int) / rv.(int)
				case float64:
					return lv.(float64) / rv.(float64)
				case float32:
					return lv.(float64) / rv.(float64)
				}
			} else if checkTypeAreNumeric(lv, rv) == true {
				if _, ok := lv.(float64); ok {
					return lv.(float64) / float64(rv.(int))
				} else {
					return float64(lv.(int)) / rv.(float64)
				}
			}
		}
	}
	l.GetWarningLogger().Println("Invalid arithmetic operation attempted")
	return nil
}

type comparatorNode struct {
	op    interface{}
	left  interface{}
	right interface{}
}

func (c *comparatorNode) execute() interface{} {
	if op, ok := c.op.(string); ok {
		switch op {
		case "<":
			lvint, lokint := c.left.(node).execute().(int)
			rvint, rokint := c.right.(node).execute().(int)

			if lokint && rokint {
				return lvint < rvint
			}

			lvf64, lokf64 := c.left.(node).execute().(float64)
			rvf64, rokf64 := c.right.(node).execute().(float64)

			if lokf64 && rokf64 {
				return lvf64 < rvf64
			}

			return nil
		case "<=":
			lvint, lokint := c.left.(node).execute().(int)
			rvint, rokint := c.right.(node).execute().(int)

			if lokint && rokint {
				return lvint <= rvint
			}

			lvf64, lokf64 := c.left.(node).execute().(float64)
			rvf64, rokf64 := c.right.(node).execute().(float64)

			if lokf64 && rokf64 {
				return lvf64 <= rvf64
			}

			return nil
		case "==":
			left := c.left.(node).execute()
			right := c.right.(node).execute()
			if checkTypesAreSame(left, right) == true {
				return left == right
			}
			return nil
		case "!=":
			left := c.left.(node).execute()
			right := c.right.(node).execute()
			if checkTypesAreSame(left, right) == true {
				return left != right
			}
			return nil
		case ">":
			lvint, lokint := c.left.(node).execute().(int)
			rvint, rokint := c.right.(node).execute().(int)

			if lokint && rokint {
				return lvint > rvint
			}

			lvf64, lokf64 := c.left.(node).execute().(float64)
			rvf64, rokf64 := c.right.(node).execute().(float64)

			if lokf64 && rokf64 {
				return lvf64 > rvf64
			}

			return nil
		case ">=":
			lvint, lokint := c.left.(node).execute().(int)
			rvint, rokint := c.right.(node).execute().(int)

			if lokint && rokint {
				return lvint >= rvint
			}

			lvf64, lokf64 := c.left.(node).execute().(float64)
			rvf64, rokf64 := c.right.(node).execute().(float64)

			if lokf64 && rokf64 {
				return lvf64 >= rvf64
			}

			return nil
		}
	}
	return nil
}

type symbolReferenceNode struct {
	val    interface{}
	offset interface{} //Used to index into arrays and node types
	key    interface{} //Used to index in []map[string] types
}

func (s *symbolReferenceNode) execute() interface{} {
	if ref, ok := s.val.(string); ok {
		idx, ok := dynamicMap[ref]
		if ok {
			val, ok := dynamicSymbolTable[idx]
			if ok {
				switch val.(type) {
				case string:
					x := val.(string)
					if cmd.State.DebugLvl >= 3 {
						println("So You want the value: ", x)
					}
				case int:
					x := val.(int)
					if cmd.State.DebugLvl >= 3 {
						println("So You want the value: ", x)
					}
				case bool:
					x := val.(bool)
					if cmd.State.DebugLvl >= 3 {
						println("So You want the value: ", x)
					}
				case float64, float32:
					x := val.(float64)
					if cmd.State.DebugLvl >= 3 {
						println("So You want the value: ", x)
					}
				case []map[int]interface{}:
					x := val.([]map[int]interface{})
					if cmd.State.DebugLvl >= 3 {
						println("Referring to Array")
					}
					i := s.offset.(node).execute().(int)
					if i >= len(x) {
						println("Index out of range error!")
						println("Array Length Of: ", len(x))
						println("But desired index at: ", i)
						l.GetWarningLogger().Println("Index out of range error!")
						return nil
					}
					//A bad implementation to implement len
					if i == -1 {
						val = len(x)
					} else {
						q := ((x[i][0]).(node).execute())
						switch q.(type) {
						case bool:
							if cmd.State.DebugLvl >= 3 {
								println("So you want the value: ", q.(bool))
							}
						case int:
							if cmd.State.DebugLvl >= 3 {
								println("So you want the value: ", q.(int))
							}
						case float64:
							if cmd.State.DebugLvl >= 3 {
								println("So you want the value: ", q.(float64))
							}
						case string:
							if cmd.State.DebugLvl >= 3 {
								println("So you want the value: ", q.(string))
							}
						}

						val = q
					}

				case []map[string]interface{}:
					if o, ok := s.offset.(node).execute().(int); ok {
						if o >= len(val.([]map[string]interface{})) {
							println("Index out of range error!")
							println("Array Length Of: ",
								len(val.([]map[string]interface{})))
							println("But desired index at: ", o)
							l.GetWarningLogger().Println("Index out of range error!")
							return nil
						}
						x := val.([]map[string]interface{})[o]
						if s.key != nil {
							if i, ok := s.key.(node).execute().(string); ok {
								val = x[i]
							} else {
								val = x
							}
						} else {
							val = x
						}

					}
					//println("This is a mapSTRINFArr but maybe it should be anode")

				case map[string]interface{}:
					if o, ok := s.offset.(node).execute().(string); ok {
						switch o {
						case "id", "name", "category", "parentID",
							"description", "domain", "parentid", "parentId":
							val = val.(map[string]interface{})[o]

						default:
							val = val.(map[string]interface{})["attributes"].(map[string]interface{})[o]
						}

					} else if idx, ok := s.offset.(node).execute().(int); ok {
						if idx != -1 {
							val = val.(map[string]interface{})["name"]
						} else {
							val = val.(map[string]interface{})
						}

					}
				case []string:
					val = val.([]string)[s.offset.(node).execute().(int)]

				case (*commonNode):
					val = val.(node).execute()
				}
				return val
			}
		}
	}
	return nil
}

type assignNode struct {
	arg interface{}
	val interface{}
}

func (a *assignNode) execute() interface{} {
	var idx int
	var id string
	if identifier, ok := a.arg.(*symbolReferenceNode); ok {
		idx = dynamicMap[identifier.val.(string)] //Get the idx
		id = identifier.val.(string)
	} else {
		idx = varCtr
		dynamicMap[a.arg.(string)] = idx
		varCtr += 1
		id = a.arg.(string)
	}

	if a.val != nil {
		var v interface{}
		if _, ok := a.val.(*commonNode); !ok {
			if _, ok := a.val.(node); !ok {
				v = a.val
			} else {
				if fn, ok := a.val.(*funcNode); ok {
					funcTable[a.arg.(string)] = fn

				} else {
					v = a.val.(node).execute() //Obtain val, execute block to get value
					// if it is not a common node
					/*if id == "_internalRes" {
						println("You need to check v here")
						q := a.val.(array).getLength()
						v = v.([]string)[q-1]
					}*/
				}

			}

		} else {
			v = a.val.(node)
		}

		if _, e := dynamicSymbolTable[idx].([]map[int]interface{}); e == true {
			//Modifying contents are user's desired Array
			arrayIdx := a.arg.(*symbolReferenceNode).offset.(node).execute().(int)

			//Bug fix for: $x[0] = $x[0] + 5
			if arithNdInf, ok := a.val.(node); ok {
				if _, ok := arithNdInf.(*arithNode); ok {
					dynamicSymbolTable[idx].([]map[int]interface{})[arrayIdx][0] = resolveArithNode(v)
				} else {
					dynamicSymbolTable[idx].([]map[int]interface{})[arrayIdx][0] = a.val
				}
			} else {
				dynamicSymbolTable[idx].([]map[int]interface{})[arrayIdx][0] = a.val
			}

		} else if mp, e := dynamicSymbolTable[idx].(map[string]interface{}); e == true {
			locIdx := a.arg.(*symbolReferenceNode).offset.(node).execute()
			switch locIdx.(type) {
			case string:
				//Check if map is an object
				if checkIfObjectNode(mp) == true {
					//No one should be updating templates at this time
					id := mp["id"].(string)
					cat := mp["category"].(string)
					cmd.UpdateObj("", id, cat,
						map[string]interface{}{locIdx.(string): v}, false)

				}

			case int:
				if locIdx.(int) > 0 {
					if cmd.State.DebugLvl >= 3 {
						println("I think should assign here")
					}

				}

				dynamicSymbolTable[idx] = v
			case nil:
				//Potential place for update
				//attribute @ a.arg.offset.val
				//new value @ a.val

				if checkIfObjectNode(mp) {
					id := mp["id"].(string)
					category := mp["category"].(string)
					attr := a.arg.(*symbolReferenceNode).offset.(*symbolReferenceNode).val.(string)
					updateArg := map[string]interface{}{attr: v}
					cmd.UpdateObj("", id, category, updateArg, false)
				}

			}

		} else {
			dynamicSymbolTable[idx] = v //Assign val into DStable
		}
		if cmd.State.DebugLvl >= 3 {
			switch v.(type) {
			case string:
				x := v.(string)
				println("You want to assign", id, "with value of", x)
			case int:
				x := v.(int)
				println("You want to assign", id, "with value of", x)
			case bool:
				x := v.(bool)
				println("You want to assign", id, "with value of", x)
			case float64, float32:
				x := v.(float64)
				println("You want to assign", id, "with value of", x)
			}
		}
	}

	return nil
}

type ifNode struct {
	condition  interface{}
	ifBranch   interface{}
	elseBranch interface{}
	elif       interface{}
}

func (i *ifNode) execute() interface{} {
	if c, ok := i.condition.(node).execute().(bool); ok {
		if c == true {
			i.ifBranch.(node).execute()
		} else {
			//Check the array of Elif cases
			//println("Now checking the elifs......")
			if _, ok := i.elif.([]elifNode); ok {
				for idx := range i.elif.([]elifNode) {
					if i.elif.([]elifNode)[idx].execute() == true {
						return true
					}
				}
			}
			if i.elseBranch != nil {
				i.elseBranch.(node).execute()
			}

		}

	}
	return nil
}

type elifNode struct {
	cond  interface{}
	taken interface{}
}

func (e *elifNode) execute() interface{} {
	if e.cond.(node).execute().(bool) == true {
		e.taken.(node).execute()
		return true
	}
	return false
}

type forNode struct {
	init        interface{}
	condition   interface{}
	incrementor interface{}
	body        interface{}
}

func (f *forNode) execute() interface{} {
	f.init.(node).execute()
	for ; f.condition.(node).execute().(bool); f.incrementor.(node).execute() {
		f.body.(node).execute()
	}
	return nil
}

type rangeNode struct {
	init      interface{}
	container interface{}
	body      interface{}
}

func (r *rangeNode) execute() interface{} {
	r.init.(node).execute()
	data := r.container.(node).execute()
	var i int
	dynamicMap["_internalIdx"] = 0
	dynamicSymbolTable[0] = i

	switch data.(type) {
	case ([]string):

		for i := range data.([]string) {
			dynamicSymbolTable[0] = i
			r.body.(node).execute()
		}

	case ([]*cmd.Node):
		for i := range data.([]*cmd.Node) {
			dynamicSymbolTable[0] = i
			r.body.(node).execute()
		}

	case ([]map[string]interface{}):
		for i := range data.([]map[string]interface{}) {
			dynamicSymbolTable[0] = i
			r.body.(node).execute()
		}

	case ([]map[int]interface{}):
		for i := range data.([]map[int]interface{}) {
			dynamicSymbolTable[0] = i
			r.body.(node).execute()
		}
	}
	return nil
}

type whileNode struct {
	//val       interface{}
	condition interface{}
	body      interface{}
}

func (w *whileNode) execute() interface{} {
	if condNode, ok := w.condition.(node); ok {
		if _, cok := condNode.execute().(bool); cok {
			/*for val == true {
				w.body.(node).execute()
			}*/
			for condNode.execute().(bool) {
				w.body.(node).execute()
			}
		}
	}
	return nil
}

type ast struct {
	statements []node
}

func (a *ast) execute() interface{} {
	for i, _ := range a.statements {
		if a.statements[i] != nil {
			a.statements[i].execute()
		}

		if a.statements[i] == nil {
			fmt.Printf("\nOGREE: Unrecognised command!\n")
			l.GetWarningLogger().Println("Unrecognised Command")
			if cmd.State.ScriptCalled == true {
				println("Line: ", cmd.State.LineNumber)
			}
		}

	}

	return nil
}

type funcNode struct {
	block interface{}
}

func (f *funcNode) execute() interface{} {
	if f.block != nil {
		f.block.(*ast).execute()
	}
	return nil
}

//Helper Functions
func UnsetUtil(x, name string, ref, value interface{}) {
	switch x {
	case "-f":
		funcTable[name] = nil
	case "-v":
		v := dynamicMap[name]
		dynamicSymbolTable[v] = nil
	default:
		//This section is for deleting an attribute
		myArg := ""
		identifier := ref.(*symbolReferenceNode)
		idx := dynamicMap[identifier.val.(string)] //Get the idx
		if idx < 0 {
			l.GetWarningLogger().Println("Object to update not found")
			println("Object to update not found")
			return
		}

		if _, ok := dynamicSymbolTable[idx]; !ok {
			msg := "Object not found in dynamicSymbolTable while deleting attr"
			l.GetErrorLogger().Println(msg)
			println("Requested Object to update not found")
			return
		}
		mp := dynamicSymbolTable[idx].(map[string]interface{})

		myArg = ref.(*symbolReferenceNode).offset.(node).execute().(string)
		if v, ok := dynamicMap[myArg]; ok {
			myArg = dynamicSymbolTable[v].(string)
		}

		cmd.UpdateObj("", mp["id"].(string), mp["category"].(string), map[string]interface{}{myArg: nil}, true)

	}
}

func checkTypesAreSame(x, y interface{}) bool {
	//println(reflect.TypeOf(x))
	return reflect.TypeOf(x) == reflect.TypeOf(y)
}

func checkTypeAreNumeric(x, y interface{}) bool {
	var xOK, yOK bool
	switch x.(type) {
	case int, float64, float32:
		xOK = true
	default:
		xOK = false
	}

	switch y.(type) {
	case int, float64, float32:
		yOK = true
	default:
		yOK = false
	}

	return xOK && yOK
}

//Generate nodes from the result of arithmetic node execution.
//This helps to maintain the definition of the array
//type in that all of its elements are nodes
func resolveArithNode(x interface{}) node {
	switch x.(type) {
	case int:
		return &numNode{x.(int)}
	case float64:
		return &floatNode{x.(float64)}
	case string:
		return &strNode{x.(string)}
	}
	return nil
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

//Open a file and return the JSON in the file
//Used by EasyPost, EasyUpdate and Load Template
func fileToJSON(path string) map[string]interface{} {
	data := map[string]interface{}{}
	x, e := ioutil.ReadFile(path)
	if e != nil {
		println("Error while opening file! " + e.Error())
		return nil
	}
	json.Unmarshal(x, &data)
	return data
}

//Executes nodes in the map and reassigns the keys
//to resulting values
func evalMapNodes(x map[string]interface{}) map[string]interface{} {
	for i := range x {

		if n, ok := x[i].(node); ok {
			tmpVal := n.(node).execute()

			//If the value is a raw array []map[int]interface{}
			//we have to cast it to []interface{}
			//this code is mainly used for when user specified
			//an array type
			if rawArr, ok := tmpVal.([]map[int]interface{}); ok {
				val := []interface{}{}
				for idx := range rawArr {
					val = append(val, rawArr[idx][0].(node).execute())
				}
				x[i] = val
			} else {
				x[i] = tmpVal
			}

		}

		//Recursively resolve values of map
		if sub, ok := x[i].(map[string]interface{}); ok {
			x[i] = evalMapNodes(sub)
		}

	}
	return x
}

//Hack function for the [room]:areas=[r1,r2,r3,r4]@[t1,t2,t3,t4]
//command
func parseAreas(x map[string]interface{}) map[string]interface{} {
	var reservedStr string
	var techStr string
	if reserved, ok := x["reserved"].([]map[int]interface{}); ok {
		if tech, ok := x["technical"].([]map[int]interface{}); ok {
			if len(reserved) == 4 && len(tech) == 4 {
				r4 := bytes.NewBufferString("")
				fmt.Fprintf(r4, "%v", reserved[3][0].(node).execute())
				r3 := bytes.NewBufferString("")
				fmt.Fprintf(r3, "%v", reserved[2][0].(node).execute())
				r2 := bytes.NewBufferString("")
				fmt.Fprintf(r2, "%v", reserved[1][0].(node).execute())
				r1 := bytes.NewBufferString("")
				fmt.Fprintf(r1, "%v", reserved[0][0].(node).execute())

				t4 := bytes.NewBufferString("")
				fmt.Fprintf(t4, "%v", tech[3][0].(node).execute())
				t3 := bytes.NewBufferString("")
				fmt.Fprintf(t3, "%v", tech[2][0].(node).execute())
				t2 := bytes.NewBufferString("")
				fmt.Fprintf(t2, "%v", tech[1][0].(node).execute())
				t1 := bytes.NewBufferString("")
				fmt.Fprintf(t1, "%v", tech[0][0].(node).execute())

				reservedStr = "{\"left\":" + r4.String() + ",\"right\":" + r3.String() + ",\"top\":" + r1.String() + ",\"bottom\":" + r2.String() + "}"
				techStr = "{\"left\":" + t4.String() + ",\"right\":" + t3.String() + ",\"top\":" + t1.String() + ",\"bottom\":" + t2.String() + "}"
				x["reserved"] = reservedStr
				x["technical"] = techStr
			}
		}
	}
	return x
}
