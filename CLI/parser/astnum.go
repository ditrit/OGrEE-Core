package parser

import (
	"fmt"
)

type arithNode struct {
	op    string
	left  node
	right node
}

func (a *arithNode) execute() (interface{}, error) {
	lv, err := nodeToNum(a.left, "left operand")
	if err != nil {
		return nil, err
	}
	rv, err := nodeToNum(a.right, "right operand")
	if err != nil {
		return nil, err
	}
	leftIntVal, leftInt := lv.(int)
	rightIntVal, rightInt := rv.(int)
	leftFloatVal, _ := lv.(float64)
	rightFloatVal, _ := rv.(float64)
	if leftInt && rightInt {
		switch a.op {
		case "+":
			return leftIntVal + rightIntVal, nil
		case "-":
			return leftIntVal - rightIntVal, nil
		case "*":
			return leftIntVal * rightIntVal, nil
		case "/":
			if rightIntVal == 0 {
				return nil, fmt.Errorf("cannot divide by 0")
			}
			return float64(leftIntVal) / float64(rightIntVal), nil
		case "\\":
			if rightIntVal == 0 {
				return nil, fmt.Errorf("cannot divide by 0")
			}
			return leftIntVal / rightIntVal, nil
		case "%":
			return leftIntVal % rightIntVal, nil
		default:
			return nil, fmt.Errorf("invalid operator for integer operands")
		}
	}
	if leftInt {
		leftFloatVal = float64(leftIntVal)
	}
	if rightInt {
		rightFloatVal = float64(rightIntVal)
	}
	switch a.op {
	case "+":
		return leftFloatVal + rightFloatVal, nil
	case "-":
		return leftFloatVal - rightFloatVal, nil
	case "*":
		return leftFloatVal * rightFloatVal, nil
	case "/":
		if rightFloatVal == 0. {
			return nil, fmt.Errorf("cannot divide by 0")
		}
		return leftFloatVal / rightFloatVal, nil
	default:
		return nil, fmt.Errorf("invalid operator for float operands")
	}
}

type negateNode struct {
	val node
}

func (n *negateNode) execute() (interface{}, error) {
	v, err := nodeToNum(n.val, "expression")
	if err != nil {
		return nil, err
	}
	intVal, isInt := v.(int)
	if isInt {
		return -intVal, nil
	}
	floatVal, isFloat := v.(float64)
	if isFloat {
		return -floatVal, nil
	}
	return nil, fmt.Errorf("cannot negate non numeric value")
}
