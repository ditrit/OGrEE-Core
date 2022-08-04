package main

import (
	cmd "cli/controllers"
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

type PathMode int64

const (
	STD PathMode = iota
	PHYSICAL
	STRAY_DEV
)

type pathNode struct {
	path node
	mode PathMode
}

func (n pathNode) getStr() (string, error) {
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
	// ignore starting dot
	if path[0] == '.' {
		path = path[1:]
	}
	// split between /, then between dots
	words := strings.Split(path, "/")
	words = append(words[:len(words)-1], strings.Split(words[len(words)-1], ".")...)
	// if it starts with a /
	if words[0] == "" {
		words = words[1:]
	} else {
		switch n.mode {
		case STD:
			words = append(strings.Split(cmd.State.CurrPath, "/")[1:], words...)
		case PHYSICAL:
			words = append([]string{"Physical"}, words...)
		case STRAY_DEV:
			words = append([]string{"Physical", "Stray", "Devices"}, words...)
		}
	}
	r := "/" + strings.Join(words, "/")
	return r, nil
}

func (n pathNode) execute() (interface{}, error) {
	return n.getStr()
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
