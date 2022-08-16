package main

import (
	cmd "cli/controllers"
	"fmt"
	"path"
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
	p, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("Path should be a string")
	}
	if p == "." || p == ".." {
		return p, nil
	}
	if p == "_" {
		return "_", nil
	}
	// ignore starting dot
	if p[0] == '.' {
		p = p[1:]
	}
	// split between /, then between dots
	words := strings.Split(p, "/")
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
	return path.Clean(r), nil
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
