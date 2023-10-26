USAGE: printf [format], [arg1], ..., [argn]
Where [format] is an expression returning a string, that can include format specifiers (subsequences beginning with %).
Each [arg] is an expression returning a value that will replace the corresponding format specifier.
For more details about the format specifiers, see https://pkg.go.dev/fmt.

EXAMPLE
    printf "41 + 1 equals %03d", 42 // prints "41 + 1 equals 042"