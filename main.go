package main

//https://www.reddit.com/r/golang/comments/
// kpyuw6/how_to_stop_reading_from_osstdin_with/

//https://github.com/gdamore/tcell
import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/MarinX/keylogger"
)

func KeyLoggerInit() *(keylogger.KeyLogger) {
	kb := keylogger.FindKeyboardDevice()

	if len(kb) <= 0 {
		println("Sorry, Keyboard wasn't found!")
		return nil
	}

	println("Found a keyboard at", kb)
	// init keylogger with keyboard
	k, err := keylogger.New(kb)
	if err != nil {
		println("Error!")
		return nil
	}
	return k
}

func DeleteMeWhenYouCan() {
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	lex := NewLexer(strings.NewReader(line))
	//lex := NewLexer(strings.NewReader(scanner.Text()))
	e := yyParse(lex)
	println("Return Code: ", e)
	return
}

func tput(args ...string) error {
	cmd := exec.Command("tput", args...)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func main() {
	for true {
		//scanner := bufio.NewScanner(os.Stdin)
		//lex := NewLexer(bufio.NewReader(os.Stdin))
		fmt.Printf("OGRE-$: ")
		k := KeyLoggerInit()
		defer k.Close()

		events := k.Read()

		for e := range events {
			switch e.Type {
			// EvKey is used to describe state changes of keyboards,
			// buttons, or other key-like devices.
			// check the input_event.go for more events
			case keylogger.EvKey:

				// if the state of key is pressed
				if e.KeyPress() && e.KeyString() == "Left" {
					//println("[event] press key ", e.KeyString())
					//println("Value: ", e.Value)
					println("\033[1D")
				} else if e.KeyPress() && e.KeyString() == "Right" {
					println("[event] press key ", e.KeyString())
					println("Value: ", e.Value)
				} else if e.KeyPress() && e.KeyString() == "Up" {
					println("[event] press key ", e.KeyString())
					println("Value: ", e.Value)
				} else if e.KeyPress() && e.KeyString() == "Down" {
					println("[event] press key ", e.KeyString())
					println("Value: ", e.Value)
				} else if e.KeyPress() && e.KeyString() == "ENTER" {
					//line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
					var buf string
					fmt.Scanln("%s", &buf)
					lex := NewLexer(strings.NewReader(string(buf)))
					//lex := NewLexer(strings.NewReader(scanner.Text()))
					e := yyParse(lex)
					println("Return Code: ", e)
				} else {
					//k.Close()
					//DeleteMeWhenYouCan()
				}

				// if the state of key is released
				/*if e.KeyRelease() {
					println("[event] release key ", e.KeyString())
				}*/

				break
			default:
				//k.Close()

				//DeleteMeWhenYouCan()
			}

		}
		//DeleteMeWhenYouCan()
	}
}
