package main

import cmd "cli/controllers"

var dynamicVarLimit = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
var dynamicMap = make(map[string]int)
var dynamicSymbolTable = make(map[int]interface{})
var dCatchPtr interface{}
var varCtr = 0

const (
	COMMON = iota
	NUM
	BOOL
	ARITHMETIC
	COMPARATOR
	REFERENCE
	IF
	FOR
	WHILE
	ASSIGN
)

type controllerAST struct{}

type node interface {
	execute() interface{}
}

type commonNode struct {
	nodeType int
	fun      interface{}
	val      string
	args     []interface{}
}

func (c *commonNode) execute() interface{} {
	switch c.nodeType {
	case 0:
		println("func called")
		if f, ok := c.fun.(func()); ok {
			f()
		}
		//c.fun.(func())()
	}

	switch c.val {
	case "PostObj":
		if f, ok := c.fun.(func(int, string, map[string]interface{})); ok {
			f(c.args[0].(int),
				c.args[1].(string), c.args[2].(map[string]interface{}))
		}

	case "GetObject", "DeleteObj", "CD", "LS", "Help", "Load":
		if f, ok := c.fun.(func(string)); ok {
			f(c.args[0].(string))
		}

	case "SearchObjects", "UpdateObj":
		if f, ok := c.fun.(func(string, map[string]interface{})); ok {
			f(c.args[0].(string), c.args[1].(map[string]interface{}))
		}

	case "Tree":
		if f, ok := c.fun.(func(string, int)); ok {
			f(c.args[0].(string), c.args[1].(int))
		}

	case "LSOG", "PWD", "select", "Exit":
		if f, ok := c.fun.(func()); ok {
			f()
		}

	case "ASSIGN":

	}
	return nil
}

type numNode struct {
	nodeType int
	val      int
}

func (n *numNode) execute() interface{} {
	return n.val
}

type boolNode struct {
	nodeType int
	val      bool
}

func (b *boolNode) execute() interface{} {
	return b.val
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
			rv, rok := (a.right.(node).execute()).(int)
			if lok && rok {
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
			lv, lok := c.left.(node).execute().(int)
			rv, rok := c.right.(node).execute().(int)
			if lok && rok {
				return lv < rv
			}
			return nil
		case "<=":
			lv, lok := c.left.(node).execute().(int)
			rv, rok := c.right.(node).execute().(int)
			if lok && rok {
				return lv <= rv
			}
			return nil
		case "==":
			lv, lok := c.left.(node).execute().(int)
			rv, rok := c.right.(node).execute().(int)
			if lok && rok {
				return lv == rv
			}
			return nil
		case ">":
			lv, lok := c.left.(node).execute().(int)
			rv, rok := c.right.(node).execute().(int)
			if lok && rok {
				return lv > rv
			}
			return nil
		case ">=":
			lv, lok := c.left.(node).execute().(int)
			rv, rok := c.right.(node).execute().(int)
			if lok && rok {
				return lv >= rv
			}
			return nil
		}
	}
	return nil
}

type symbolReferenceNode struct {
	nodeType int
	val      interface{}
}

func (s *symbolReferenceNode) execute() interface{} {
	if ref, ok := s.val.(string); ok {
		val := dynamicSymbolTable[dynamicMap[ref]]
		switch val.(type) {
		case string:
			x := val.(string)
			println("So You want the value: ", x)
		case int:
			x := val.(int)
			println("So You want the value: ", x)
		case bool:
			x := val.(bool)
			println("So You want the value: ", x)
		case float64, float32:
			x := dCatchPtr.(float64)
			println("So You want the value: ", x)
		}
		return val
	}
	return nil
}

type assignNode struct {
	nodeType int
	arg      interface{}
}

func (a *assignNode) execute() interface{} {
	dynamicMap[a.arg.(string)] = varCtr
	dynamicSymbolTable[varCtr] = dCatchPtr
	varCtr += 1
	switch dCatchPtr.(type) {
	case string:
		x := dCatchPtr.(string)
		println("You want to assign", a.arg.(string), "with value of", x)
	case int:
		x := dCatchPtr.(int)
		println("You want to assign", a.arg.(string), "with value of", x)
	case bool:
		x := dCatchPtr.(bool)
		println("You want to assign", a.arg.(string), "with value of", x)
	case float64, float32:
		x := dCatchPtr.(float64)
		println("You want to assign", a.arg.(string), "with value of", x)
	}
	return nil
}

type ifNode struct {
	nodeType   int
	condition  interface{}
	ifBranch   interface{}
	elseBranch interface{}
}

func (i *ifNode) execute() interface{} {
	if c, ok := i.condition.(node).execute().(bool); ok {
		if c == true {
			i.ifBranch.(node).execute()
		} else {
			i.elseBranch.(node).execute()
		}
	}
	return nil
}

type forNode struct {
	nodeType    int
	val         interface{}
	init        interface{}
	condition   interface{}
	incrementor interface{}
	body        interface{}
}

type whileNode struct {
	nodeType  int
	val       interface{}
	condition interface{}
	body      interface{}
}

func (w *whileNode) execute() interface{} {
	if condNode, ok := w.condition.(node); ok {
		if val, cok := condNode.execute().(bool); cok {
			for val == true {
				w.body.(node).execute()
			}
		}
	}
	return nil
}

type ast struct {
	nodeType int
	val      interface{}
	left     interface{}
	right    interface{}
}

func (c controllerAST) PWD() func() {
	return cmd.PWD
}
