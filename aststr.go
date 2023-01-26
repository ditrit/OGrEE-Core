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
	if p == "_" {
		return "_", nil
	}
	if p == "." {
		return cmd.State.CurrPath, nil
	}
	var output_words []string
	if p[0] != '/' {
		switch n.mode {
		case STD:
			output_words = append(strings.Split(cmd.State.CurrPath, "/")[1:], output_words...)
		case PHYSICAL:
			output_words = append([]string{"Physical"}, output_words...)
		case STRAY_DEV:
			output_words = append([]string{"Physical", "Stray", "Devices"}, output_words...)
		}
	}
	// split between /, then between dots
	input_words := strings.Split(p, "/")
	if input_words[len(input_words)-1] != ".." && input_words[len(input_words)-1] != "." {
		input_words = append(input_words[:len(input_words)-1], strings.Split(input_words[len(input_words)-1], ".")...)
	}

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
	r := "/" + strings.Join(output_words, "/")
	return path.Clean(r), nil
}

func (n pathNode) execute() (interface{}, error) {
	return n.getStr()
}

type concatNode struct {
	nodes []node
}

func (n *concatNode) getStr() (string, error) {
	//var r string
	r := ""
	for i := range n.nodes {
		valAny, err := n.nodes[i].execute()
		if err != nil {
			return "", err
		}
		r += fmt.Sprintf("%v", valAny)
		/*switch val := valAny.(type) {
		case string:
			r = r + val
		case int:
			r = r + strconv.Itoa(val)
		case float64:
			if val == float64(int(val)) {
				r = r + strconv.Itoa((int(val)))
			} else {
				return "", fmt.Errorf("expression should return a string or an int, but returned a float (concatenation expr %d)", i+1)
			}
		default:
			return "", fmt.Errorf("expression should return a string or an int (concatenation expr %d)", i+1)
		}*/
	}
	return r, nil
}

func (n *concatNode) execute() (interface{}, error) {
	return n.getStr()
}
