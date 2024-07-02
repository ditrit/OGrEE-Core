package parser

import (
	"fmt"
)

type equalityNode struct {
	op    string
	left  node
	right node
}

func (n *equalityNode) execute() (any, error) {
	left, err := n.left.execute()
	if err != nil {
		return false, err
	}
	right, err := n.right.execute()
	if err != nil {
		return false, err
	}
	switch n.op {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	}
	return false, fmt.Errorf("Invalid equality node operator : " + n.op)
}

type comparatorNode struct {
	op    string
	left  node
	right node
}

func (n *comparatorNode) execute() (any, error) {
	leftNum, err := nodeToFloat(n.left, "left expression")
	if err != nil {
		return nil, err
	}
	rightNum, err := nodeToFloat(n.right, "right expression")
	if err != nil {
		return nil, err
	}
	switch n.op {
	case "<":
		return leftNum < rightNum, nil
	case "<=":
		return leftNum <= rightNum, nil
	case ">":
		return leftNum > rightNum, nil
	case ">=":
		return leftNum >= rightNum, nil
	}
	return false, fmt.Errorf("Invalid comparison operator : " + n.op)
}

type logicalNode struct {
	op    string
	left  node
	right node
}

func (n *logicalNode) execute() (any, error) {
	leftBool, err := nodeToBool(n.left, "left operand")
	if err != nil {
		return false, err
	}
	rightBool, err := nodeToBool(n.right, "right operand")
	if err != nil {
		return false, err
	}
	switch n.op {
	case "||":
		return leftBool || rightBool, nil
	case "&&":
		return leftBool && rightBool, nil
	}
	return false, fmt.Errorf("Invalid logical operator : " + n.op)
}

type negateBoolNode struct {
	expr node
}

func (n *negateBoolNode) execute() (any, error) {
	b, err := nodeToBool(n.expr, "expression")
	if err != nil {
		return false, err
	}
	return !b, nil
}
