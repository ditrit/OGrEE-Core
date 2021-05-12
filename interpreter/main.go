package main

import (
	"os"
)

//import ("./interpreter")

func main() {
	lex := NewLexer(os.Stdin)
	e := yyParse(lex)
	println("Return Code: ", e)
}
