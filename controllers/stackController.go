package controllers

type (
	Stack struct {
		top    *node
		bottom *node
		length int
	}
	node struct {
		value interface{}
		prev  *node
	}
)

// Create a new stack
func New() *Stack {
	return &Stack{nil, nil, 0}
}

// Return the number of items in the stack
func (this *Stack) Len() int {
	return this.length
}

// View the top item on the stack
func (this *Stack) Peek() interface{} {
	if this.length == 0 {
		return nil
	}
	return this.top.value
}

func (this *Stack) PeekLast() interface{} {
	if this.length == 0 {
		return nil
	}
	return this.bottom.value
}

// Pop the top item of the stack and return it
func (this *Stack) Pop() interface{} {
	if this.length == 0 {
		return nil
	}

	n := this.top
	this.top = n.prev
	this.length--
	return n.value
}

//Reverse pop the stack
func (this *Stack) ReversePop() interface{} {
	if this.length == 0 {
		return nil
	}

	prev := this.top
	nxt := prev.prev

	//Single node case
	if nxt == nil {
		this.length--
		n := this.top
		this.top = nil //nxt
		return n.value
	}

	//Go to bottom of stack
	for nxt.prev != nil {
		nxt = nxt.prev
		prev = prev.prev
	}

	//At the bottom of the stack
	this.length--
	n := nxt
	prev.prev = nil
	return n.value

}

// Push a value onto the top of the stack
func (this *Stack) Push(value interface{}) {
	n := &node{value, this.top}
	if this.length == 0 {
		this.bottom = n
	}
	this.top = n
	this.length++
}
