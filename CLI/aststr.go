package main

import (
	c "cli/controllers"
	"fmt"
	"path"
	"strings"
)

type pathNode struct {
	path node
}

func (n pathNode) execute() (interface{}, error) {
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
	if p == "-" {
		return c.State.PrevPath, nil
	}
	var output_words []string
	if p[0] != '/' {
		output_words = strings.Split(c.State.CurrPath, "/")[1:]
		if len(output_words) == 1 && output_words[0] == "" {
			output_words = output_words[0:0]
		}
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
		} else if output_words[0] == "O" {
			output_words[0] = "Organisation"
		}
	}
	return path.Clean("/" + strings.Join(output_words, "/")), nil
}

type formatStringNode struct {
	str  node
	vals []node
}

func (n *formatStringNode) execute() (interface{}, error) {
	str, err := nodeToString(n.str, "string")
	if err != nil {
		return "", err
	}
	vals := []any{}
	for _, val := range n.vals {
		v, err := val.execute()
		if err != nil {
			return "", err
		}
		vals = append(vals, v)
	}
	return fmt.Sprintf(str, vals...), nil
}
