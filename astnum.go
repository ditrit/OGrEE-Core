package main

import (
	cmd "cli/controllers"
	l "cli/logger"
	"fmt"
	"strconv"
)

type numNode interface {
	getNum() (float64, error)
	execute() (interface{}, error)
}

type floatNode interface {
	getFloat() (float64, error)
	execute() (interface{}, error)
}

type floatLeaf struct {
	val float64
}

func (l floatLeaf) getFloat() (float64, error) {
	return l.val, nil
}
func (l floatLeaf) execute() (interface{}, error) {
	return l.val, nil
}

func (l floatLeaf) getNum() (float64, error) {
	return l.val, nil
}

type intNode interface {
	getInt() (int, error)
}

type intLeaf struct {
	val int
}

func (l intLeaf) getInt() (int, error) {
	return l.val, nil
}
func (l intLeaf) execute() (interface{}, error) {
	return l.val, nil
}
func (l intLeaf) getNum() (float64, error) {
	return float64(l.val), nil
}

type arithNode struct {
	op    string
	left  node
	right node
}

func (a *arithNode) execute() (interface{}, error) {
	lv, err := a.left.execute()
	if err != nil {
		return nil, err
	}
	if cmd.State.DebugLvl >= 3 {
		println("Left:", lv)
	}
	rv, err := a.right.execute()
	if err != nil {
		return nil, err
	}
	if cmd.State.DebugLvl >= 3 {
		println("Right: ", rv)
	}
	switch a.op {
	case "+":
		if checkTypesAreSame(lv, rv) == true {
			switch lv.(type) {
			case int:
				return lv.(int) + rv.(int), nil
			case float64:
				return lv.(float64) + rv.(float64), nil
			case float32:
				return lv.(float64) + rv.(float64), nil
			case string:
				return lv.(string) + rv.(string), nil
			}
		} else if checkTypeAreNumeric(lv, rv) == true {
			if _, ok := lv.(float64); ok {
				return lv.(float64) + float64(rv.(int)), nil
			} else {
				return rv.(float64) + float64(lv.(int)), nil
			}
		} else { //we have string and numeric type
			//this code occurs when assigning and not
			//when using + while printing

			switch lv.(type) {
			case int:
				return strconv.Itoa(lv.(int)) + rv.(string), nil
			case float64:
				return strconv.FormatFloat(lv.(float64), 'f', -1, 64) + rv.(string), nil
			}

			switch rv.(type) {
			case int:
				return lv.(string) + strconv.Itoa(rv.(int)), nil
			case float64:
				return lv.(string) + strconv.FormatFloat(rv.(float64), 'f', -1, 64), nil
			}
		}
		//Otherwise the types are incompatible so return nil
		//TODO:see if team would want to have bool support
		return nil, fmt.Errorf("Incompatible types")

	case "-":
		if checkTypesAreSame(lv, rv) == true {
			switch lv.(type) {
			case int:
				return lv.(int) - rv.(int), nil
			case float64:
				return lv.(float64) - rv.(float64), nil
			case float32:
				return lv.(float64) - rv.(float64), nil
			}
		} else if checkTypeAreNumeric(lv, rv) == true {
			if _, ok := lv.(float64); ok {
				return lv.(float64) - float64(rv.(int)), nil
			} else {
				return float64(lv.(int)) - rv.(float64), nil
			}
		}

	case "*":
		if checkTypesAreSame(lv, rv) == true {
			switch lv.(type) {
			case int:
				return lv.(int) * rv.(int), nil
			case float64:
				return lv.(float64) * rv.(float64), nil
			case float32:
				return lv.(float64) * rv.(float64), nil
			}
		} else if checkTypeAreNumeric(lv, rv) == true {
			if _, ok := lv.(float64); ok {
				return lv.(float64) * float64(rv.(int)), nil
			} else {
				return float64(lv.(int)) * rv.(float64), nil
			}
		}
	case "%":
		if checkTypesAreSame(lv, rv) == true {
			switch lv.(type) {
			case int:
				return lv.(int) % rv.(int), nil
			case float64:
				return int(lv.(float64)) % int(rv.(float64)), nil
			case float32:
				return int(lv.(float32)) % int(rv.(float32)), nil
			}
		} else if checkTypeAreNumeric(lv, rv) == true {
			if _, ok := lv.(float64); ok {
				return int(lv.(float64)) % rv.(int), nil
			} else {
				return lv.(int) % int(rv.(float64)), nil
			}
		}
	case "/":
		if checkTypesAreSame(lv, rv) == true {
			switch lv.(type) {
			case int:
				return lv.(int) / rv.(int), nil
			case float64:
				return lv.(float64) / rv.(float64), nil
			case float32:
				return lv.(float64) / rv.(float64), nil
			}
		} else if checkTypeAreNumeric(lv, rv) == true {
			if _, ok := lv.(float64); ok {
				return lv.(float64) / float64(rv.(int)), nil
			} else {
				return float64(lv.(int)) / rv.(float64), nil
			}
		}
	}
	l.GetWarningLogger().Println("Invalid arithmetic operation attempted")
	return nil, fmt.Errorf("Invalid arithmetic operation attempted")
}
