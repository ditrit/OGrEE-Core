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

type pathNode struct {
	path node
}

func (n pathNode) getStr() (string, error) {
	val, err := n.path.execute()
	if err != nil {
		return "", err
	}
	p, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("path should be a string")
	}
	if p == "" {
		p = "."
	}
	if p == "_" {
		return "_", nil
	}
	var output_words []string
	if p[0] != '/' {
		output_words = strings.Split(cmd.State.CurrPath, "/")[1:]
	} else {
		p = p[1:]
	}
	input_words := strings.Split(p, "/")
	for _, word := range input_words {
		if word == "." {
			continue
		} else if word == ".." {
			if len(output_words) > 0 {
				output_words = output_words[:len(output_words)-1]
			}
		} else {
			output_words = append(output_words, word)
		}
	}
	if len(output_words) > 0 {
		if output_words[0] == "P" {
			output_words[0] = "Physical"
		} else if output_words[0] == "L" {
			output_words[0] = "Logical"
		}
	}
	return "/" + strings.Join(output_words, "/"), nil
}

func (n pathNode) execute() (interface{}, error) {
	return n.getStr()
}

type formatStringNode struct {
	str       string
	varsDeref []symbolReferenceNode
}

func (n *formatStringNode) getStr() (string, error) {
	vals := []any{}
	for _, varDeref := range n.varsDeref {
		val, err := varDeref.execute()
		if err != nil {
			return "", err
		}
		vals = append(vals, val)
	}
	return fmt.Sprintf(n.str, vals...), nil
}

func (n *formatStringNode) execute() (interface{}, error) {
	return n.getStr()
}
