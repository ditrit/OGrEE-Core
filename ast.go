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
)

type controllerAST struct{}

type node interface {
	execute() interface{}
}

type commonNode struct {
	nodeType int
	fun      interface{}
	val      string
	args     interface{}
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
	symbol   interface{}
}

func (s *symbolReferenceNode) execute() interface{} {
	if ref, ok := s.symbol.(string); ok {
		val := dynamicSymbolTable[dynamicMap[ref]]
		return val
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
