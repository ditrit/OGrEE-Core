package main

//Since readline hasn't been updated since 2018
//it may be worth switching to peterh/liner
//https://stackoverflow.com/
// questions/33025599/move-the-cursor-in-a-c-program

// Adding TAB completion support
//https://thoughtbot.com/blog/tab-completion-in-gnu-readline
import (
	c "cli/controllers"
	"strings"

	"github.com/chzyer/readline"
)

func InterpretLine(str *string) {
	lex := NewLexer(strings.NewReader(*str))
	e := yyParse(lex)
	println("\nReturn Code: ", e)
	return
}

func main() {

	user, _ := c.Login()

	user = (strings.Split(user, "@"))[0]
	rl, err := readline.New(user + "@" + "OGRE3D:$> ")
	if err != nil {
		panic(err)
	}

	defer rl.Close()
	c.InitState()
	c.AddHistory(rl)
	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}
		InterpretLine(&line)
		c.UpdateSessionState(&line)
	}
}
