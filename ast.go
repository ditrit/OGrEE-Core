package main

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

type commonNode struct {
	nodeType int
	val      string
}

type numNode struct {
	nodeType int
	val      int
}

type boolNode struct {
	nodeType int
	val      bool
}

type arithNode struct {
	nodeType int
	op       interface{}
	left     interface{}
	right    interface{}
}

type comparatorNode struct {
	nodeType int
	op       interface{}
	left     interface{}
	right    interface{}
}

type symbolReferenceNode struct {
	nodeType int
	val      interface{}
	symbol   interface{}
}

type ifNode struct {
	nodeType   int
	val        interface{}
	ifBranch   interface{}
	elseBranch interface{}
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

type ast struct {
	nodeType int
	val      interface{}
	left     interface{}
	right    interface{}
}
