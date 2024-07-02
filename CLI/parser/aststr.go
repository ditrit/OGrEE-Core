package parser

import (
	c "cli/controllers"
	"fmt"
)

type pathNode struct {
	path            node
	acceptSelection bool
}

func (n pathNode) Path() (string, error) {
	return nodeToString(n.path, "path")
}

func (n pathNode) execute() (interface{}, error) {
	p, err := n.Path()
	if err != nil {
		return nil, err
	}

	return c.TranslatePath(p, n.acceptSelection), nil
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
	if str == "%v" && len(vals) == 1 {
		vec, isVec := vals[0].([]float64)
		if isVec {
			return vec, nil
		}
	}
	return fmt.Sprintf(str, vals...), nil
}
