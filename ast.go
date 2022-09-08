package main

import (
	"bytes"
	c "cli/controllers"
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

const (
	COMMON = iota
	NUM
	FLOAT
	BOOL
	STR
	BOOLOP
	ARITHMETIC
	COMPARATOR
	REFERENCE
	IF
	FOR
	WHILE
	ASSIGN
	BLOCK
	FUNC
	ELIF
	ARRAY
	OBJND
	JSONND
)

type node interface {
	execute() interface{}
	getType() int
}

type array interface {
	node
	getLength() int
}

type commonNode struct {
	nodeType int
	fun      interface{}
	val      string
	args     []interface{}
}

func (c *commonNode) execute() interface{} {
	switch c.val {
	case "PostObj":
		if f, ok := c.fun.(func(int, string, map[string]interface{}) map[string]interface{}); ok {
			v := f(c.args[0].(int),
				c.args[1].(string), c.args[2].(map[string]interface{}))
			return &jsonObjNode{JSONND, v}
		}

	case "EasyPost":
		if f, ok := c.fun.(func(int, string, map[string]interface{}) map[string]interface{}); ok {
			var data map[string]interface{}
			/*x, e := ioutil.ReadFile(c.args[2].(string))
			if e != nil {
				println("Error while opening file! " + e.Error())
				return nil
			}
			json.Unmarshal(x, &data)*/
			data = fileToJSON(c.args[2].(string))
			if data == nil {
				return nil
			}

			v := f(c.args[0].(int), c.args[1].(string), data)

			return &jsonObjNode{JSONND, v}
		}

	case "LSATTR":
		if f, ok := c.fun.(func(string, string)); ok {
			f(c.args[0].(string), c.args[1].(string))
		}

	case "Help", "Focus":
		if f, ok := c.fun.(func(string)); ok {
			f(c.args[0].(string))
		}

	case "CD", "Load":
		//Script Load
		if f, ok := c.fun.(func(string) string); ok {
			v := f(c.args[0].(string))
			return &strNode{STR, v}
		}

		//Template Load
		if f, ok := c.fun.(func(map[string]interface{}, string)); ok {
			data := fileToJSON(c.args[0].(string))
			if data == nil {
				return nil
			}
			f(data, c.args[0].(string))
			return &strNode{STR, c.args[0].(string)}
		}

	case "Print":
		if f, ok := c.fun.(func([]interface{}) string); ok {
			res := []interface{}{}
			for i := range c.args {
				res = append(res, c.args[i].(node).execute())
			}
			v := f(res)
			return &strNode{STR, v}
		}

	case "LS":
		if f, ok := c.fun.(func(string) []map[string]interface{}); ok {
			v := f(c.args[0].(string))
			return &jsonObjArrNode{JSONND, len(v), v}

		}

	case "DeleteObj":
		if f, ok := c.fun.(func(string) bool); ok {
			v := f(c.args[0].(string))
			return &boolNode{BOOL, v}
		}

	case "DeleteSelection":
		if f, ok := c.fun.(func() bool); ok {
			return f() //returns bool
		}

	case "IsEntityDrawable":
		if f, ok := c.fun.(func(string) bool); ok {
			if arg, ok := c.args[0].(string); ok {
				t := f(arg)
				return t
			}
			//Error if reached here
		}

	case "IsAttrDrawable":
		if f, ok := c.fun.(func(string, string, map[string]interface{}, bool) bool); ok {
			objInf := c.args[0]
			attrInf := c.args[1].(node).execute()
			if _, ok := objInf.(string); !ok {
				println("Object operand is invalid")
				l.GetInfoLogger().Println("Object operand is invalid")
				return nil
			}
			if _, ok := attrInf.(string); !ok {
				println("Attribute operand is invalid")
				l.GetInfoLogger().Println("Attribute operand is invalid")
				return nil
			}
			return f(objInf.(string), attrInf.(string), nil, false)
		}

	case "GetObject":
		if f, ok := c.fun.(func(string, bool) (map[string]interface{}, string)); ok {
			v, _ := f(c.args[0].(string), false)
			return &jsonObjNode{COMMON, v}
		}

	case "SearchObjects":
		if f, ok := c.fun.(func(string, map[string]interface{}) []map[string]interface{}); ok {
			v := f(c.args[0].(string), c.args[1].(map[string]interface{}))
			return &jsonObjArrNode{COMMON, len(v), v}
		}

	case "RecursiveUpdateObj":
		if f, ok := c.fun.(func(string, string, string, map[string]interface{})); ok {
			//Old code was removed since
			//it broke the OCLI syntax easy update
			x := len(c.args)
			if _, ok := c.args[x-1].(bool); ok {
				var objMap map[string]interface{}
				//Weird edge case
				//to solve issue with:
				// for i in $(ls) do $i[attr]="string"

				//c.args[0] = referenceToNode
				//c.args[1] = attributeString, (used as an index)
				//c.args[2] = someValue (usually a string)
				mp := c.args[0]
				objMap = mp.(node).execute().(map[string]interface{})

				if checkIfObjectNode(objMap) == true {
					updateArgs := map[string]interface{}{c.args[1].(string): c.args[2].(node).execute()}
					id := objMap["id"].(string)
					entity := objMap["category"].(string)
					f("", id, entity, updateArgs)
				}

			} else {
				if c.args[2].(string) == "recursive" {
					f(c.args[0].(string), "", "", c.args[1].(map[string]interface{}))

				}

			}
			//return &jsonObjNode{COMMON, v}
		}

	case "UpdateObj":
		var v map[string]interface{}
		if f, ok := c.fun.(func(string, string, string, map[string]interface{}, bool) map[string]interface{}); ok {
			//Old code was removed since
			//it broke the OCLI syntax easy update
			x := len(c.args)
			if _, ok := c.args[x-1].(bool); ok {
				var objMap map[string]interface{}
				//Weird edge case
				//to solve issue with:
				// for i in $(ls) do $i[attr]="string"

				//c.args[0] = referenceToNode
				//c.args[1] = attributeString, (used as an index)
				//c.args[2] = someValue (usually a string)
				mp := c.args[0]
				objMap = mp.(node).execute().(map[string]interface{})

				if checkIfObjectNode(objMap) == true {
					updateArgs := map[string]interface{}{c.args[1].(string): c.args[2].(node).execute()}
					id := objMap["id"].(string)
					entity := objMap["category"].(string)
					v = f("", id, entity, updateArgs, false)
				}

			} else {
				if rawArr, ok := c.args[1].(map[string]interface{}); ok {
					if areas, ok := rawArr["areas"].(map[string]interface{}); ok {
						c.args[1] = parseAreas(areas)
					}
				}
				v = f(c.args[0].(string), "", "", c.args[1].(map[string]interface{}), false)
			}
			return &jsonObjNode{COMMON, v}
		}

	case "EasyUpdate":
		if f, ok := c.fun.(func(string, map[string]interface{}, bool) map[string]interface{}); ok {
			var data map[string]interface{}
			//var op string
			//0 -> path to node
			//1 -> path to json
			//2 -> put or patch
			data = fileToJSON(c.args[1].(string))
			if data == nil {
				return nil
			}

			v := f(c.args[0].(string), data, c.args[2].(bool))

			return &jsonObjNode{JSONND, v}
		}
	case "LSOBJ":
		if f, ok := c.fun.(func(string, int) []map[string]interface{}); ok {
			v := f(c.args[0].(string), c.args[1].(int))
			//return &objNdArrNode{COMMON, len(v), v}
			return &jsonObjArrNode{JSONND, len(v), v}
		}

	case "GetU":
		if f, ok := c.fun.(func(string, interface{})); ok {
			f(c.args[0].(string), c.args[1])
		}

	case "Tree", "Draw":
		if f, ok := c.fun.(func(string, int)); ok {
			f(c.args[0].(string), c.args[1].(int))
		}

	case "LSOG", "Exit", "CLR", "LSEnterprise":
		if f, ok := c.fun.(func()); ok {
			f()
		}

	case "Env":
		if f, ok := c.fun.(func(map[string]interface{},
			map[string]interface{})); ok {
			x := map[string]interface{}{}
			for i := range dynamicMap {
				x[i] = dynamicSymbolTable[dynamicMap[i]]
			}
			f(x, GetFuncTable())
		}

	case "select":
		if f, ok := c.fun.(func() []string); ok {
			v := f()
			return &strArrNode{COMMON, len(v), v}
		}

	case "PWD":
		if f, ok := c.fun.(func() string); ok {
			v := f()
			return &strNode{STR, v}
		}

	case "setCB":
		if f, ok := c.fun.(func(*[]string) []string); ok {
			v := f(c.args[0].(*[]string))
			return &strArrNode{COMMON, len(v), v}
		}

	case "UpdateSelect":
		if f, ok := c.fun.(func(map[string]interface{})); ok {
			f(c.args[0].(map[string]interface{}))
		}

	case "Unset":
		if f, ok := c.fun.(func(string, string, interface{}, interface{})); ok {
			if len(c.args) > 2 {
				//Means we are deleting an attribute
				//of an object
				/*idx := c.args[2].(node).execute()
				println("IDX WAS:", idx)
				return*/
				f(c.args[0].(string), c.args[1].(string),
					c.args[2].(node), c.args[3])
			} else {
				f(c.args[0].(string), c.args[1].(string), nil, nil)
			}

		}

	case "SetEnv":
		if f, ok := c.fun.(func(string, interface{})); ok {
			arg := c.args[0].(string)
			val := c.args[1].(node).execute()
			f(arg, val)
		}

	case "Hierarchy":
		if f, ok := c.fun.(func(string, int, bool) []map[string]interface{}); ok {
			path := c.args[0].(string)
			depth := c.args[1].(int)
			f(path, depth, false)
		}

	case "GetOCAttr":
		if f, ok := c.fun.(func(string, int,
			map[string]interface{})); ok {

			//Since the attributes is a map of nodes
			//execute and receive the values
			attributes := c.args[2].(map[string]interface{})
			attributes = evalMapNodes(attributes)

			f(c.args[0].(string),
				c.args[1].(int),
				c.args[2].(map[string]interface{}))
		}

	case "HandleUnity":
		if f, ok := c.fun.(func(map[string]interface{})); ok {
			data := map[string]interface{}{}
			data["command"] = c.args[1]
			if len(c.args) == 4 {

				firstArr := c.args[2].([]map[int]interface{})
				secondArr := c.args[3].([]map[int]interface{})

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

				if c.args[1].(string) == "wait" && c.args[0].(string) == "camera" {
					data["position"] = map[string]float64{"x": 0, "y": 0, "z": 0}

					if y, ok := c.args[2].([]map[int]interface{}); ok {
						data["rotation"] = map[string]interface{}{"x": 999,
							"y": y[0][0].(node).execute()}
					} else {
						data["rotation"] = map[string]interface{}{"x": 999,
							"y": c.args[2]}
					}

				} else {
					if _, ok := c.args[2].([]map[int]interface{}); ok {
						data["data"] = c.args[2].([]map[int]interface{})[0][0].(node).execute()
					} else {
						data["data"] = c.args[2]
					}

				}

			}
			fullJson := map[string]interface{}{
				"type": c.args[0].(string),
				"data": data,
			}
			f(fullJson)
		}

	case "LinkObject":
		if f, ok := c.fun.(func([]interface{})); ok {
			if len(c.args) == 3 {
				c.args[2] = c.args[2].(node).execute()
			}
			f(c.args)
		}

	case "UnlinkObject":
		if f, ok := c.fun.(func([]interface{})); ok {
			f(c.args)
		}
	}

	return nil
}

func (c *commonNode) getType() int {
	return c.nodeType
}

type arrNode struct {
	nodeType int
	len      int
	val      []map[int]interface{}
}

func (a *arrNode) execute() interface{} {
	return a.val
}

func (a *arrNode) getLength() int {
	return a.len
}

func (a *arrNode) getType() int {
	return a.nodeType
}

type objNdNode struct {
	nodeType int
	val      *cmd.Node
}

func (n *objNdNode) execute() interface{} {
	return n.val
}

func (n *objNdNode) getType() int {
	return n.nodeType
}

type objNdArrNode struct {
	nodeType int
	len      int
	val      []*cmd.Node
}

func (o *objNdArrNode) execute() interface{} {
	return o.val
}

func (o *objNdArrNode) getLength() int {
	return o.len
}

func (o *objNdArrNode) getType() int {
	return o.nodeType
}

type jsonObjNode struct {
	nodeType int
	val      map[string]interface{}
}

func (j *jsonObjNode) execute() interface{} {
	return j.val
}

func (j *jsonObjNode) getType() int {
	return j.nodeType
}

type jsonObjArrNode struct {
	nodeType int
	len      int
	val      []map[string]interface{}
}

func (j *jsonObjArrNode) execute() interface{} {
	return j.val
}

func (j *jsonObjArrNode) getLength() int {
	return j.len
}

func (j *jsonObjArrNode) getType() int {
	return j.nodeType
}

type numNode struct {
	nodeType int
	val      int
}

func (n *numNode) execute() interface{} {
	return n.val
}

func (n *numNode) getType() int {
	return n.nodeType
}

type floatNode struct {
	nodeType int
	val      float64
}

func (f *floatNode) execute() interface{} {
	return f.val
}

func (f *floatNode) getType() int {
	return f.nodeType
}

type strNode struct {
	nodeType int
	val      string
}

func (s *strNode) execute() interface{} {
	return s.val
}

func (s *strNode) getType() int {
	return s.nodeType
}

type strArrNode struct {
	nodeType int
	len      int
	val      []string
}

func (s *strArrNode) execute() interface{} {
	return s.val
}

func (s *strArrNode) getLength() int {
	return s.len
}

func (s *strArrNode) getType() int {
	return s.nodeType
}

type boolNode struct {
	nodeType int
	val      bool
}

func (b *boolNode) execute() interface{} {
	return b.val
}

func (b *boolNode) getType() int {
	return b.nodeType
}

type boolOpNode struct {
	nodeType int
	op       string
	operand  interface{}
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

func (b *boolOpNode) getType() int {
	return b.nodeType
}

type arithNode struct {
	nodeType int
	op       interface{}
	left     interface{}
	right    interface{}
}

func (a *arithNode) getType() int {
	return a.nodeType
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
	nodeType int
	op       interface{}
	left     interface{}
	right    interface{}
}

func (c *comparatorNode) getType() int {
	return c.nodeType
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
	nodeType int
	val      interface{}
	offset   interface{} //Used to index into arrays and node types
	key      interface{} //Used to index in []map[string] types
}

func (s *symbolReferenceNode) getType() int {
	return s.nodeType
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
						if cmd.State.DebugLvl > 0 {
							println("Index out of range error!")
							println("Array Length Of: ", len(x))
							println("But desired index at: ", i)
						}

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
							if cmd.State.DebugLvl > 0 {
								println("Index out of range error!")
								println("Array Length Of: ",
									len(val.([]map[string]interface{})))
								println("But desired index at: ", o)
							}

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
	nodeType int
	arg      interface{}
	val      interface{}
}

func (a *assignNode) getType() int {
	return a.nodeType
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
				if arithNdInf.getType() == ARITHMETIC {
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
	nodeType   int
	condition  interface{}
	ifBranch   interface{}
	elseBranch interface{}
	elif       interface{}
}

func (i *ifNode) getType() int {
	return i.nodeType
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
	nodeType int
	cond     interface{}
	taken    interface{}
}

func (e *elifNode) getType() int {
	return e.nodeType
}

func (e *elifNode) execute() interface{} {
	if e.cond.(node).execute().(bool) == true {
		e.taken.(node).execute()
		return true
	}
	return false
}

type forNode struct {
	nodeType    int
	init        interface{}
	condition   interface{}
	incrementor interface{}
	body        interface{}
}

func (f *forNode) getType() int {
	return f.nodeType
}

func (f *forNode) execute() interface{} {
	f.init.(node).execute()
	for ; f.condition.(node).execute().(bool); f.incrementor.(node).execute() {
		f.body.(node).execute()
	}
	return nil
}

type rangeNode struct {
	nodeType  int
	init      interface{}
	container interface{}
	body      interface{}
}

func (r *rangeNode) getType() int {
	return r.nodeType
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
	nodeType int
	//val       interface{}
	condition interface{}
	body      interface{}
}

func (w *whileNode) getType() int {
	return w.nodeType
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
	nodeType   int
	statements []node
}

func (a *ast) getType() int {
	return a.nodeType
}

func (a *ast) execute() interface{} {
	for i, _ := range a.statements {
		if a.statements[i] != nil {
			a.statements[i].execute()
		}

		if a.statements[i] == nil {
			if cmd.State.DebugLvl > 0 {
				fmt.Printf("\nOGREE: Unrecognised command!\n")
			}

			l.GetWarningLogger().Println("Unrecognised Command")
			if cmd.State.ScriptCalled == true {
				println("Line: ", cmd.State.LineNumber)
			}
		}

	}

	return nil
}

type funcNode struct {
	nodeType int
	block    interface{}
}

func (f *funcNode) getType() int {
	return f.nodeType
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
		//funcTable[name] = nil
		delete(funcTable, name)
	case "-v":
		v := dynamicMap[name]
		//dynamicSymbolTable[v] = nil
		delete(dynamicMap, name)
		delete(dynamicSymbolTable, v)
	default:
		//This section is for deleting an attribute
		myArg := ""
		identifier := ref.(*symbolReferenceNode)
		idx := dynamicMap[identifier.val.(string)] //Get the idx
		if idx < 0 {
			l.GetWarningLogger().Println("Object to update not found")
			if c.State.DebugLvl > 0 {
				println("Object to update not found")
			}

			return
		}

		if _, ok := dynamicSymbolTable[idx]; !ok {
			if cmd.State.DebugLvl > 1 {
				println("Requested Object to update not found")
			}
			msg := "Object not found in dynamicSymbolTable while deleting attr"
			l.GetErrorLogger().Println(msg)

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

func GetFuncTable() map[string]interface{} {
	return funcTable
}

func GetDynamicMap() map[string]int {
	return dynamicMap
}

func GetDynamicSymbolTable() map[int]interface{} {
	return dynamicSymbolTable
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
		return &numNode{NUM, x.(int)}
	case float64:
		return &floatNode{FLOAT, x.(float64)}
	case string:
		return &strNode{STR, x.(string)}
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
