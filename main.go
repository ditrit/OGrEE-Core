package main

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
import (
	"strings"

	"github.com/chzyer/readline"
)

func DeleteMeWhenYouCan(str *string) {
	lex := NewLexer(strings.NewReader(*str))
	e := yyParse(lex)
	println("Return Code: ", e)
	return
}

func main() {
	rl, err := readline.New("OGRE3D:$> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		DeleteMeWhenYouCan(&line)
	}
}
