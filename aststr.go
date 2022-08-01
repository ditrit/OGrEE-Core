package main

import (
	"fmt"
	"strings"
)

type strLeaf struct {
	val string
}

func (l strLeaf) getStr() (string, error) {
	return l.val, nil
}
func (l strLeaf) execute() (interface{}, error) {
	return l.getStr()
}

type pathConvert struct {
	path node
}

func (n pathConvert) getStr() (string, error) {
	val, err := n.path.execute()
	if err != nil {
		return "", err
	}
	path, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("Path should be a string")
	}
	if path == "." || path == ".." {
		return path, nil
	}
	return strings.ReplaceAll(path, ".", "/"), nil
}
func (l pathConvert) execute() (interface{}, error) {
	return l.getStr()
}

type concatNode struct {
	nodes []node
}

func (n *concatNode) getStr() (string, error) {
	var r string
	for i := range n.nodes {
		v, err := n.nodes[i].execute()
		if err != nil {
			return "", err
		}
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("Expression should return a string (concatenation expr %d)", i)
		}
		r = r + s
	}
	return r, nil
}

func (n *concatNode) execute() (interface{}, error) {
	return n.getStr()
}
