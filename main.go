package main

import (
	"cli/interpreter"
	"os"
)

func main() {
	lex := interpreter.NewLexer(os.Stdin)
	e := interpreter.yyParse(lex)
	println("Return Code: ", e)
}
