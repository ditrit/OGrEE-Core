USAGE: print [value]
where [value] can be either 
* a string (in which dereferencing variables is possible, or produced by the format function)
* the evaluation of an expression via: eval [expr]
* the result of a command via $(([command]))

EXAMPLE
    print hello world // prints "hello world"
    print 41 + 1 // prints "41 + 1"
    print eval 41 + 1 // prints "42"
    print format("41 + 1 equals %d", 42) // prints "41 + 1 equals 42"