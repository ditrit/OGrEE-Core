package parser

import (
	c "cli/controllers"
	"fmt"
)

type ifNode struct {
	condition  node
	ifBranch   node
	elseBranch node
}

func (n *ifNode) execute() (interface{}, error) {
	condition, err := nodeToBool(n.condition, "condition")
	if err != nil {
		return nil, err
	}
	if condition {
		_, err := n.ifBranch.execute()
		if err != nil {
			return nil, err
		}
	} else {
		if n.elseBranch != nil {
			_, err := n.elseBranch.execute()
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

type whileNode struct {
	condition node
	body      node
}

func (n *whileNode) execute() (interface{}, error) {
	for {
		condition, err := nodeToBool(n.condition, "condition")
		if err != nil {
			return nil, err
		}
		if !condition {
			break
		}
		_, err = n.body.execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type forNode struct {
	init        node
	condition   node
	incrementor node
	body        node
}

// ToDo: this expression is not possible to obtain. Add it to parser
func (n *forNode) execute() (interface{}, error) {
	_, err := n.init.execute()
	if err != nil {
		return nil, err
	}
	for {
		condition, err := nodeToBool(n.condition, "condition")
		if err != nil {
			return nil, err
		}
		if !condition {
			break
		}
		_, err = n.body.execute()
		if err != nil {
			return nil, err
		}
		_, err = n.incrementor.execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type forArrayNode struct {
	variable string
	arr      node
	body     node
}

// ToDo: this expression is not possible to obtain. Add it to parser
func (n *forArrayNode) execute() (interface{}, error) {
	val, err := n.arr.execute()
	if err != nil {
		return nil, err
	}
	arr, ok := val.([]interface{})
	if !ok {
		return nil, fmt.Errorf("only an array can be iterated")
	}
	for _, v := range arr {
		_, err := (&assignNode{n.variable, &valueNode{v}}).execute()
		if err != nil {
			return nil, err
		}
		_, err = n.body.execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type forRangeNode struct {
	variable string
	start    node
	end      node
	body     node
}

func (n *forRangeNode) execute() (interface{}, error) {
	start, err := nodeToInt(n.start, "start index")
	if err != nil {
		return nil, err
	}
	end, err := nodeToInt(n.end, "end index")
	if err != nil {
		return nil, err
	}
	if start > end {
		return nil, fmt.Errorf("start index should be lower than end index")
	}
	for i := start; i <= end; i++ {
		c.State.DynamicSymbolTable[n.variable] = i
		if n.body == nil {
			return nil, fmt.Errorf("loop body should not be empty")
		}
		_, err = n.body.execute()
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}
