package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for true {
		//scanner := bufio.NewScanner(os.Stdin)
		//lex := NewLexer(bufio.NewReader(os.Stdin))
		fmt.Printf("OGRE-$: ")
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		lex := NewLexer(strings.NewReader(line))
		//lex := NewLexer(strings.NewReader(scanner.Text()))
		e := yyParse(lex)
		println("Return Code: ", e)
	}
}
