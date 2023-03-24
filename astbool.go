package main

import "fmt"

type boolNode interface {
	getBool() (bool, error)
	execute() (interface{}, error)
}

type boolLeaf struct {
	val bool
}

func (l boolLeaf) getBool() (bool, error) {
	return l.val, nil
}

func (l boolLeaf) execute() (interface{}, error) {
	return l.val, nil
}

type equalityNode struct {
	op    string
	left  node
	right node
}

func (n *equalityNode) getBool() (bool, error) {
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

func (n *equalityNode) execute() (interface{}, error) {
	return n.getBool()
}

type comparatorNode struct {
	op    string
	left  node
	right node
}

func (n *comparatorNode) getBool() (bool, error) {
	left, err := n.left.execute()
	if err != nil {
		return false, err
	}
	leftNum, err := getFloat(left)
	if err != nil {
		return false, fmt.Errorf("left expression should return a number")
	}
	right, err := n.right.execute()
	if err != nil {
		return false, err
	}
	rightNum, err := getFloat(right)
	if err != nil {
		return false, fmt.Errorf("right expression should return a number")
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

func (n *comparatorNode) execute() (interface{}, error) {
	return n.getBool()
}

type logicalNode struct {
	op    string
	left  node
	right node
}

func (n *logicalNode) getBool() (bool, error) {
	left, err := n.left.execute()
	if err != nil {
		return false, err
	}
	leftBool, ok := left.(bool)
	if !ok {
		return false, fmt.Errorf("Left expression should return a boolean")
	}
	right, err := n.right.execute()
	if err != nil {
		return false, err
	}
	rightBool, ok := right.(bool)
	if !ok {
		return false, fmt.Errorf("Right expression should return a boolean")
	}
	switch n.op {
	case "||":
		return leftBool || rightBool, nil
	case "&&":
		return leftBool && rightBool, nil
	}
	return false, fmt.Errorf("Invalid logical operator : " + n.op)
}

func (n *logicalNode) execute() (interface{}, error) {
	return n.getBool()
}

type negateBoolNode struct {
	expr node
}

func (n *negateBoolNode) getBool() (bool, error) {
	val, err := n.expr.execute()
	if err != nil {
		return false, err
	}
	b, ok := val.(bool)
	if !ok {
		return false, fmt.Errorf("Expression should return a boolean to be negated")
	}
	return !b, nil
}

func (n *negateBoolNode) execute() (interface{}, error) {
	return n.getBool()
}
