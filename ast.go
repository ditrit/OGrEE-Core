package main

import (
	cmd "cli/controllers"
	"cli/readline"
	"encoding/json"
	"io/ioutil"
	"reflect"
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
			x, e := ioutil.ReadFile(c.args[2].(string))
			if e != nil {
				println("Error while opening file! " + e.Error())
				return nil
			}
			json.Unmarshal(x, &data)
			v := f(c.args[0].(int), c.args[1].(string), data)

			return &jsonObjNode{JSONND, v}
		}

	case "Help":
		if f, ok := c.fun.(func(string)); ok {
			f(c.args[0].(string))
		}

	case "CD", "Load":
		if f, ok := c.fun.(func(string) string); ok {
			v := f(c.args[0].(string))
			return &strNode{STR, v}
		}

	case "Print":
		if f, ok := c.fun.(func(...interface{}) string); ok {
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

	case "GetObject":
		if f, ok := c.fun.(func(string, bool) map[string]interface{}); ok {
			v := f(c.args[0].(string), false)
			return &jsonObjNode{COMMON, v}
		}

	case "SearchObjects":
		if f, ok := c.fun.(func(string, map[string]interface{}) []map[string]interface{}); ok {
			v := f(c.args[0].(string), c.args[1].(map[string]interface{}))
			return &jsonObjArrNode{COMMON, len(v), v}
		}

	case "UpdateObj":
		var v map[string]interface{}
		if f, ok := c.fun.(func(string, map[string]interface{}, bool) map[string]interface{}); ok {
			//Old code was removed since
			//it broke the OCLI syntax easy update
			x := len(c.args)
			if _, ok := c.args[x-1].(bool); ok {
				//Weird edge case
				//to solve issue with:
				// for i in $(ls) do $i[attr]="string"

				//c.args[0] = referenceToNode
				//c.args[1] = attributeString, (used as an index)
				//c.args[2] = someValue (usually a string)
				mp := c.args[0]
				nd := getNodeFromMapInf(mp.(node).execute().(map[string]interface{}))
				updateArgs := map[string]interface{}{c.args[1].(string): c.args[2].(node).execute()}

				v = f(nd.Path, updateArgs, false)
			} else {
				v = f(c.args[0].(string), c.args[1].(map[string]interface{}), false)
			}
			return &jsonObjNode{COMMON, v}
		}

	case "EasyUpdate":
		if f, ok := c.fun.(func(string, string, map[string]interface{}) map[string]interface{}); ok {
			var data map[string]interface{}
			var op string
			//0 -> path to node
			//1 -> path to json
			//2 -> put or patch
			x, e := ioutil.ReadFile(c.args[1].(string))
			if e != nil {
				println("Error while opening file! " + e.Error())
				return nil
			}
			json.Unmarshal(x, &data)

			if c.args[2].(bool) == true {
				op = "PATCH"
			} else {
				op = "PUT"
			}

			v := f(c.args[0].(string), op, data)

			return &jsonObjNode{JSONND, v}
		}
	case "LSOBJ":
		if f, ok := c.fun.(func(string, int) []map[string]interface{}); ok {
			v := f(c.args[0].(string), c.args[1].(int))
			//return &objNdArrNode{COMMON, len(v), v}
			return &jsonObjArrNode{JSONND, len(v), v}
		}

	case "Tree":
		if f, ok := c.fun.(func(string, int)); ok {
			f(c.args[0].(string), c.args[1].(int))
		}

	case "LSOG", "Exit":
		if f, ok := c.fun.(func()); ok {
			f()
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
			}
			f(c.args[0].(string), c.args[1].(string),
				c.args[2].(node), c.args[3])
		}

	case "GetOCAttr":
		if f, ok := c.fun.(func(*cmd.Stack, int,
			map[string]interface{}, *readline.Instance)); ok {
			f(c.args[0].(*cmd.Stack),
				c.args[1].(int),
				c.args[2].(map[string]interface{}),
				c.args[3].(*readline.Instance))
		}
	}
	return nil
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

type objNdNode struct {
	nodeType int
	val      *cmd.Node
}

func (n *objNdNode) execute() interface{} {
	return n.val
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

type jsonObjNode struct {
	nodeType int
	val      map[string]interface{}
}

func (j *jsonObjNode) execute() interface{} {
	return j.val
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

type numNode struct {
	nodeType int
	val      int
}

func (n *numNode) execute() interface{} {
	return n.val
}

type strNode struct {
	nodeType int
	val      string
}

func (s *strNode) execute() interface{} {
	return s.val
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

type boolNode struct {
	nodeType int
	val      bool
}

func (b *boolNode) execute() interface{} {
	return b.val
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

type arithNode struct {
	nodeType int
	op       interface{}
	left     interface{}
	right    interface{}
}

func (a *arithNode) execute() interface{} {
	if v, ok := a.op.(string); ok {
		switch v {
		case "+":
			lv, lok := (a.left.(node).execute()).(int)
			if cmd.State.DebugLvl >= 3 {
				println("Left:", lv)
			}
			rv, rok := (a.right.(node).execute()).(int)
			if cmd.State.DebugLvl >= 3 {
				println("Right: ", rv)
			}
			if lok && rok {
				//println("Adding", lv, rv)
				return lv + rv
			}
			return nil

		case "-":
			lv, lok := (a.left.(node).execute()).(int)
			rv, rok := (a.right.(node).execute()).(int)
			if lok && rok {
				return lv - rv
			}
			return nil

		case "*":
			lv, lok := (a.left.(node).execute()).(int)
			rv, rok := (a.right.(node).execute()).(int)
			if lok && rok {
				return lv * rv
			}
			return nil

		case "%":
			lv, lok := (a.left.(node).execute()).(int)
			rv, rok := (a.right.(node).execute()).(int)
			if lok && rok {
				return lv % rv
			}
			return nil

		case "/":
			lv, lok := (a.left.(node).execute()).(int)
			rv, rok := (a.right.(node).execute()).(int)
			if lok && rok {
				return lv / rv
			}
			return nil
		}
	}
	return nil
}

type comparatorNode struct {
	nodeType int
	op       interface{}
	left     interface{}
	right    interface{}
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
					x := dCatchPtr.(float64)
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
						cmd.WarningLogger.Println("Index out of range error!")
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
							cmd.WarningLogger.Println("Index out of range error!")
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
				v = a.val.(node).execute() //Obtain val, execute block to get value
				// if it is not a common node
				/*if id == "_internalRes" {
					println("You need to check v here")
					q := a.val.(array).getLength()
					v = v.([]string)[q-1]
				}*/
			}

		} else {
			v = a.val.(node)
		}

		if _, e := dynamicSymbolTable[idx].([]map[int]interface{}); e == true {
			//Modifying contents are user's desired Array
			arrayIdx := a.arg.(*symbolReferenceNode).offset.(node).execute().(int)
			dynamicSymbolTable[idx].([]map[int]interface{})[arrayIdx][0] = a.val

		} else if mp, e := dynamicSymbolTable[idx].(map[string]interface{}); e == true {
			locIdx := a.arg.(*symbolReferenceNode).offset.(node).execute()
			switch locIdx.(type) {
			case string:
				mp[locIdx.(string)] = v //Assign val into map[str]inf{} (node) type
				//Some kind of update needs to be done

				//Update if the map was a node
				if nd := getNodeFromMapInf(mp); nd != nil {
					data := map[string]interface{}{locIdx.(string): v}
					cmd.UpdateObj(nd.Path, data, false)
				}

			case int:
				if locIdx.(int) > 0 {
					if cmd.State.DebugLvl >= 3 {
						println("I think should assign here")
					}

				}
				//println("I think I should do nothing here")
				dynamicSymbolTable[idx] = v
			case nil:
				//Potential place for update
				//attribute @ a.arg.offset.val
				//new value @ a.val
				//the Node is mp
				nd := getNodeFromMapInf(mp)
				attr := a.arg.(*symbolReferenceNode).offset.(*symbolReferenceNode).val.(string)
				updateArg := map[string]interface{}{attr: v}
				cmd.UpdateObj(nd.Path, updateArg, false)
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

func (a *ast) execute() interface{} {
	for i, _ := range a.statements {
		if a.statements[i] != nil {
			a.statements[i].execute()
		}

	}

	return nil
}

type funcNode struct {
	nodeType int
	block    interface{}
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
		mp := dynamicSymbolTable[idx].(map[string]interface{})
		nd := getNodeFromMapInf(mp)

		myArg = ref.(*symbolReferenceNode).offset.(node).execute().(string)
		if v, ok := dynamicMap[myArg]; ok {
			myArg = dynamicSymbolTable[v].(string)
		}

		cmd.UpdateObj(nd.Path, map[string]interface{}{myArg: nil}, true)

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
	return xOK == yOK
}

//Gets node from Tree Hierarchy using a map[string]interface
func getNodeFromMapInf(x map[string]interface{}) *cmd.Node {
	pid, _ := x["parentId"].(string)
	id, _ := x["id"].(string)

	//There is no stable hard code way
	//to check if obj is Group or
	//Obj/Room Template
	if pid == "" {
		roomTmplOk := cmd.FindNodeByIDP(&cmd.State.TreeHierarchy, id, "2")
		if roomTmplOk != nil {
			return roomTmplOk
		}

		objTmplOk := cmd.FindNodeByIDP(&cmd.State.TreeHierarchy, id, "1")
		if objTmplOk != nil {
			return objTmplOk
		}

		groupOk := cmd.FindNodeByIDP(&cmd.State.TreeHierarchy, id, "3")
		if groupOk != nil {
			return groupOk
		}

		//Not found!
		return nil
	}

	return cmd.FindNodeByIDP(&cmd.State.TreeHierarchy, id, pid)
}
