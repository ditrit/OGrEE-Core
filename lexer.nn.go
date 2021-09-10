package main

import (
	cmd "cli/controllers"
	"os"
	"strconv"
)
import (
	"bufio"
	"io"
	"strings"
)

type frame struct {
	i            int
	s            string
	line, column int
}
type Lexer struct {
	// The lexer runs in its own goroutine, and communicates via channel 'ch'.
	ch      chan frame
	ch_stop chan bool
	// We record the level of nesting because the action could return, and a
	// subsequent call expects to pick up where it left off. In other words,
	// we're simulating a coroutine.
	// TODO: Support a channel-based variant that compatible with Go's yacc.
	stack []frame
	stale bool

	// The 'l' and 'c' fields were added for
	// https://github.com/wagerlabs/docker/blob/65694e801a7b80930961d70c69cba9f2465459be/buildfile.nex
	// Since then, I introduced the built-in Line() and Column() functions.
	l, c int

	parseResult interface{}

	// The following line makes it easy for scripts to insert fields in the
	// generated code.
	// [NEX_END_OF_LEXER_STRUCT]
}

// NewLexerWithInit creates a new Lexer object, runs the given callback on it,
// then returns it.
func NewLexerWithInit(in io.Reader, initFun func(*Lexer)) *Lexer {
	yylex := new(Lexer)
	if initFun != nil {
		initFun(yylex)
	}
	yylex.ch = make(chan frame)
	yylex.ch_stop = make(chan bool, 1)
	var scan func(in *bufio.Reader, ch chan frame, ch_stop chan bool, family []dfa, line, column int)
	scan = func(in *bufio.Reader, ch chan frame, ch_stop chan bool, family []dfa, line, column int) {
		// Index of DFA and length of highest-precedence match so far.
		matchi, matchn := 0, -1
		var buf []rune
		n := 0
		checkAccept := func(i int, st int) bool {
			// Higher precedence match? DFAs are run in parallel, so matchn is at most len(buf), hence we may omit the length equality check.
			if family[i].acc[st] && (matchn < n || matchi > i) {
				matchi, matchn = i, n
				return true
			}
			return false
		}
		var state [][2]int
		for i := 0; i < len(family); i++ {
			mark := make([]bool, len(family[i].startf))
			// Every DFA starts at state 0.
			st := 0
			for {
				state = append(state, [2]int{i, st})
				mark[st] = true
				// As we're at the start of input, follow all ^ transitions and append to our list of start states.
				st = family[i].startf[st]
				if -1 == st || mark[st] {
					break
				}
				// We only check for a match after at least one transition.
				checkAccept(i, st)
			}
		}
		atEOF := false
		stopped := false
		for {
			if n == len(buf) && !atEOF {
				r, _, err := in.ReadRune()
				switch err {
				case io.EOF:
					atEOF = true
				case nil:
					buf = append(buf, r)
				default:
					panic(err)
				}
			}
			if !atEOF {
				r := buf[n]
				n++
				var nextState [][2]int
				for _, x := range state {
					x[1] = family[x[0]].f[x[1]](r)
					if -1 == x[1] {
						continue
					}
					nextState = append(nextState, x)
					checkAccept(x[0], x[1])
				}
				state = nextState
			} else {
			dollar: // Handle $.
				for _, x := range state {
					mark := make([]bool, len(family[x[0]].endf))
					for {
						mark[x[1]] = true
						x[1] = family[x[0]].endf[x[1]]
						if -1 == x[1] || mark[x[1]] {
							break
						}
						if checkAccept(x[0], x[1]) {
							// Unlike before, we can break off the search. Now that we're at the end, there's no need to maintain the state of each DFA.
							break dollar
						}
					}
				}
				state = nil
			}

			if state == nil {
				lcUpdate := func(r rune) {
					if r == '\n' {
						line++
						column = 0
					} else {
						column++
					}
				}
				// All DFAs stuck. Return last match if it exists, otherwise advance by one rune and restart all DFAs.
				if matchn == -1 {
					if len(buf) == 0 { // This can only happen at the end of input.
						break
					}
					lcUpdate(buf[0])
					buf = buf[1:]
				} else {
					text := string(buf[:matchn])
					buf = buf[matchn:]
					matchn = -1
					select {
					case ch <- frame{matchi, text, line, column}:
						{
						}
					case stopped = <-ch_stop:
						{
						}
					}
					if stopped {
						break
					}
					if len(family[matchi].nest) > 0 {
						scan(bufio.NewReader(strings.NewReader(text)), ch, ch_stop, family[matchi].nest, line, column)
					}
					if atEOF {
						break
					}
					for _, r := range text {
						lcUpdate(r)
					}
				}
				n = 0
				for i := 0; i < len(family); i++ {
					state = append(state, [2]int{i, 0})
				}
			}
		}
		ch <- frame{-1, "", line, column}
	}
	go scan(bufio.NewReader(in), yylex.ch, yylex.ch_stop, dfas, 0, 0)
	return yylex
}

type dfa struct {
	acc          []bool           // Accepting states.
	f            []func(rune) int // Transitions.
	startf, endf []int            // Transitions at start and end of input.
	nest         []dfa
}

var dfas = []dfa{
	// [ \t]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 9:
				return 1
			case 32:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 9:
				return -1
			case 32:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// create
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return 1
			case 101:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 114:
				return 2
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return 3
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 4
			case 99:
				return -1
			case 101:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 114:
				return -1
			case 116:
				return 5
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return 6
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// gt
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 103:
				return 1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 103:
				return -1
			case 116:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 103:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// update
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			case 117:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 112:
				return 2
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 100:
				return 3
			case 101:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 4
			case 100:
				return -1
			case 101:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 112:
				return -1
			case 116:
				return 5
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 100:
				return -1
			case 101:
				return 6
			case 112:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// delete
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 100:
				return 1
			case 101:
				return -1
			case 108:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return 2
			case 108:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return 3
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return 4
			case 108:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 116:
				return 5
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return 6
			case 108:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// search
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 104:
				return -1
			case 114:
				return -1
			case 115:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return 2
			case 104:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 3
			case 99:
				return -1
			case 101:
				return -1
			case 104:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 104:
				return -1
			case 114:
				return 4
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return 5
			case 101:
				return -1
			case 104:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 104:
				return 6
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 104:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// \+
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 43:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 43:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// -
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 45:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// :
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 58:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 58:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// @
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 64:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 64:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \$
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 36:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 36:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// ;
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 59:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 59:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \[
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 91:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 91:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 93:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 93:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \(
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 40:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 40:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \)
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 41:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 41:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// ||
	{[]bool{true}, []func(rune) int{ // Transitions
		func(r rune) int {
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1}, []int{ /* End-of-input transitions */ -1}, nil},

	// &&
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 38:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 38:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 38:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// \!
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 33:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 33:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \*
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 42:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 42:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// >
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 62:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 62:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// <
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 60:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 60:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// false|true
	{[]bool{false, false, false, false, false, true, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return 1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 2
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 6
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return 3
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return 5
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return 7
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return 8
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return 9
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// if
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 102:
				return -1
			case 105:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 102:
				return 2
			case 105:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 102:
				return -1
			case 105:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// for
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 102:
				return 1
			case 111:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 102:
				return -1
			case 111:
				return 2
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 102:
				return -1
			case 111:
				return -1
			case 114:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 102:
				return -1
			case 111:
				return -1
			case 114:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// in
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 105:
				return 1
			case 110:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 110:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 110:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// while
	{[]bool{false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 119:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return 2
			case 105:
				return -1
			case 108:
				return -1
			case 119:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return -1
			case 105:
				return 3
			case 108:
				return -1
			case 119:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 4
			case 119:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 5
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 119:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 119:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1}, nil},

	// else
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return 1
			case 108:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return 2
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 4
			case 108:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// then
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return -1
			case 110:
				return -1
			case 116:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return 2
			case 110:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 3
			case 104:
				return -1
			case 110:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return -1
			case 110:
				return 4
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 104:
				return -1
			case 110:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// fi
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 102:
				return 1
			case 105:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 102:
				return -1
			case 105:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 102:
				return -1
			case 105:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// elif
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return 1
			case 102:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 102:
				return -1
			case 105:
				return -1
			case 108:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 102:
				return -1
			case 105:
				return 3
			case 108:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 102:
				return 4
			case 105:
				return -1
			case 108:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 102:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// done
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 100:
				return 1
			case 101:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 110:
				return -1
			case 111:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 110:
				return 3
			case 111:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return 4
			case 110:
				return -1
			case 111:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// do
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 100:
				return 1
			case 111:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 111:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 111:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// print
	{[]bool{false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 110:
				return -1
			case 112:
				return 1
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 110:
				return -1
			case 112:
				return -1
			case 114:
				return 2
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return 3
			case 110:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 110:
				return 4
			case 112:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 110:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 116:
				return 5
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 110:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1}, nil},

	// unset
	{[]bool{false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 110:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 110:
				return 2
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 110:
				return -1
			case 115:
				return 3
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 4
			case 110:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 110:
				return -1
			case 115:
				return -1
			case 116:
				return 5
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 110:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1}, nil},

	// \"
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 34:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 34:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// tn
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 110:
				return -1
			case 116:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 110:
				return 2
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 110:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// si
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 115:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return 2
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 105:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// bd
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 98:
				return 1
			case 100:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// ro
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 111:
				return -1
			case 114:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 111:
				return 2
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 111:
				return -1
			case 114:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// rk
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 107:
				return -1
			case 114:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 107:
				return 2
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 107:
				return -1
			case 114:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// dv
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 100:
				return 1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 118:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 118:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// sd
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 115:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return 2
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// sd1
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 100:
				return -1
			case 115:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 100:
				return 2
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return 3
			case 100:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 100:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// selection
	{[]bool{false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 115:
				return 1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return 2
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return 3
			case 110:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return 4
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return 5
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			case 116:
				return 6
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return -1
			case 105:
				return 7
			case 108:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 111:
				return 8
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return 9
			case 111:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// cmds
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 99:
				return 1
			case 100:
				return -1
			case 109:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			case 109:
				return 2
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return 3
			case 109:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			case 109:
				return -1
			case 115:
				return 4
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			case 109:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// template
	{[]bool{false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 112:
				return -1
			case 116:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return 2
			case 108:
				return -1
			case 109:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 109:
				return 3
			case 112:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 112:
				return 4
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 108:
				return 5
			case 109:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 6
			case 101:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 112:
				return -1
			case 116:
				return 7
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return 8
			case 108:
				return -1
			case 109:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 112:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// var
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 114:
				return -1
			case 118:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 2
			case 114:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 114:
				return 3
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 114:
				return -1
			case 118:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// {
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 123:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 123:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// }
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 125:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 125:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// ,
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 44:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 44:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \.
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 46:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 46:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// tenant
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 110:
				return -1
			case 116:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return 2
			case 110:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 110:
				return 3
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 4
			case 101:
				return -1
			case 110:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 110:
				return 5
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 110:
				return -1
			case 116:
				return 6
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 101:
				return -1
			case 110:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// site
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return 1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return 2
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 116:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 4
			case 105:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// bldg|building
	{[]bool{false, false, false, false, false, false, false, false, false, true, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 98:
				return 1
			case 100:
				return -1
			case 103:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 105:
				return -1
			case 108:
				return 2
			case 110:
				return -1
			case 117:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return 10
			case 103:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 105:
				return 4
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 105:
				return -1
			case 108:
				return 5
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return 6
			case 103:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 105:
				return 7
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return 8
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return 9
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return 11
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 117:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// room
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 109:
				return -1
			case 111:
				return -1
			case 114:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 109:
				return -1
			case 111:
				return 2
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 109:
				return -1
			case 111:
				return 3
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 109:
				return 4
			case 111:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 109:
				return -1
			case 111:
				return -1
			case 114:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// rack
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 107:
				return -1
			case 114:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 2
			case 99:
				return -1
			case 107:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return 3
			case 107:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 107:
				return 4
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 107:
				return -1
			case 114:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// device
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return 1
			case 101:
				return -1
			case 105:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 2
			case 105:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 118:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return 4
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return 5
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 6
			case 105:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 118:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// subdevice
	{[]bool{false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return 1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return 2
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return 3
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 4
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 5
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return 6
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return 7
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return 8
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 9
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// subdevice1
	{[]bool{false, false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return 1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return 2
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return 3
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 4
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 5
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return 6
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return 7
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return 8
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 9
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return 10
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 105:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// address|category|city|color|country|description|domain|gps|height|heightUnit|id|mainContact|mainEmail|mainPhone|model|name|nbFloors|orientation|parentId|posU|posXY|posXYUnit|posZ|posZUnit|reserved|reservedColor|serial|size|sizeU|sizeUnit|slot|technical|technicalColor|template|TOK|type|usableColor|vendor|zipcode
	{[]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false, false, false, false, false, true, false, false, false, true, false, false, false, false, false, false, true, false, false, false, false, false, true, false, false, false, false, true, false, false, false, false, true, false, true, true, false, false, true, false, false, false, true, false, false, false, false, false, false, true, false, false, false, false, true, false, false, false, true, false, true, false, false, false, true, true, false, false, false, true, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, false, false, false, false, false, true, false, false, false, true, false, false, false, false, false, true, true, false, false, false, false, true, false, false, false, true, false, true, false, false, false, false, false, true, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, true, false, true, false, true, false, false, false, false, false, true, false, false, false, false, false, true, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return 1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 2
			case 98:
				return -1
			case 99:
				return 3
			case 100:
				return 4
			case 101:
				return -1
			case 103:
				return 5
			case 104:
				return 6
			case 105:
				return 7
			case 108:
				return -1
			case 109:
				return 8
			case 110:
				return 9
			case 111:
				return 10
			case 112:
				return 11
			case 114:
				return 12
			case 115:
				return 13
			case 116:
				return 14
			case 117:
				return 15
			case 118:
				return 16
			case 121:
				return -1
			case 122:
				return 17
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return 205
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 199
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 180
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 181
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 182
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 165
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 166
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return 163
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 154
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 153
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 129
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 130
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 119
			case 98:
				return 120
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 109
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 88
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 89
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 76
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 61
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 62
			case 108:
				return 63
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 39
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return 40
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return 29
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 24
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 18
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return 19
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return 20
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 21
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 22
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 23
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 25
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 26
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 27
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 28
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 30
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return 31
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 32
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 33
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return 34
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 35
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 36
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 37
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 38
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return 43
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return 44
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return 41
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 42
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return 50
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return 45
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 46
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 47
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 48
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 49
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 51
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 52
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return 53
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 54
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 55
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return 56
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 57
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 58
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 59
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 60
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 72
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return 66
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 64
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 65
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 67
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return 68
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 69
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 70
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 71
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 73
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 74
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 75
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return 77
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 78
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 79
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return 80
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 81
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 82
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return 83
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 84
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 85
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 86
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 87
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 103
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return 90
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return 91
			case 88:
				return 92
			case 89:
				return -1
			case 90:
				return 93
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return 98
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return 94
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 95
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 96
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 97
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return 99
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 100
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 101
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 102
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 104
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 105
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 106
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return 107
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 108
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 110
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 111
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 112
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 113
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 114
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 115
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 116
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 117
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 118
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return 127
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return 121
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 122
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 123
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 124
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 125
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return 126
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 128
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 134
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 131
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 132
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 133
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 135
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return 136
			case 69:
				return 137
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return 138
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 147
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return 143
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return 139
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 140
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 141
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 142
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 144
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 145
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 146
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 148
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 149
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 150
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return 151
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 152
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 155
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return 156
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return 157
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 158
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return 159
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 160
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 161
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 162
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return 164
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return 171
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return 167
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return 168
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 169
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 170
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return 172
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 173
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 174
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return 175
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 176
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 177
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 178
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 179
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 193
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 191
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return 183
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return 184
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 189
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 185
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return 186
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 187
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return 188
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 190
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return 192
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 194
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return 195
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 196
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 197
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return 198
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return 200
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return 201
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return 202
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return 203
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return 204
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return 206
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 67:
				return -1
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 75:
				return -1
			case 79:
				return -1
			case 80:
				return -1
			case 84:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return -1
			case 90:
				return -1
			case 97:
				return -1
			case 98:
				return -1
			case 99:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			case 121:
				return -1
			case 122:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// [0-9]+
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch {
			case 48 <= r && r <= 57:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch {
			case 48 <= r && r <= 57:
				return 1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// lsten
	{[]bool{false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return 1
			case 110:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 115:
				return 2
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 115:
				return -1
			case 116:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 4
			case 108:
				return -1
			case 110:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return -1
			case 110:
				return 5
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1}, nil},

	// lssite
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return 1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 115:
				return 2
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 115:
				return 3
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return 4
			case 108:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 116:
				return 5
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 6
			case 105:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// lsbldg
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 108:
				return 1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 108:
				return -1
			case 115:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return 3
			case 100:
				return -1
			case 103:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 108:
				return 4
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return 5
			case 103:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return 6
			case 108:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 103:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// lsroom
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 108:
				return 1
			case 109:
				return -1
			case 111:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 109:
				return -1
			case 111:
				return -1
			case 114:
				return -1
			case 115:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 109:
				return -1
			case 111:
				return -1
			case 114:
				return 3
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 109:
				return -1
			case 111:
				return 4
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 109:
				return -1
			case 111:
				return 5
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 109:
				return 6
			case 111:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 109:
				return -1
			case 111:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// lsrack
	{[]bool{false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 107:
				return -1
			case 108:
				return 1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 107:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 107:
				return -1
			case 108:
				return -1
			case 114:
				return 3
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 4
			case 99:
				return -1
			case 107:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return 5
			case 107:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 107:
				return 6
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 107:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1}, nil},

	// lsdev
	{[]bool{false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return 1
			case 115:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return 2
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return 3
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return 4
			case 108:
				return -1
			case 115:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 118:
				return 5
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 118:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1}, nil},

	// lssubdev
	{[]bool{false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return 1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return 2
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return 3
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return 4
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return 5
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return 6
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return 7
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return 8
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// lssubdev1
	{[]bool{false, false, false, false, false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return 1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return 2
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return 3
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return 4
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return 5
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 100:
				return 6
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return 7
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return 8
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return 9
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 49:
				return -1
			case 98:
				return -1
			case 100:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 115:
				return -1
			case 117:
				return -1
			case 118:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

	// tree
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 114:
				return -1
			case 116:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 114:
				return 2
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 3
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 4
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 114:
				return -1
			case 116:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// lsog
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 103:
				return -1
			case 108:
				return 1
			case 111:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 103:
				return -1
			case 108:
				return -1
			case 111:
				return -1
			case 115:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 103:
				return -1
			case 108:
				return -1
			case 111:
				return 3
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 103:
				return 4
			case 108:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 103:
				return -1
			case 108:
				return -1
			case 111:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// cd
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 99:
				return 1
			case 100:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 99:
				return -1
			case 100:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// pwd
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 112:
				return 1
			case 119:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 112:
				return -1
			case 119:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return 3
			case 112:
				return -1
			case 119:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 100:
				return -1
			case 112:
				return -1
			case 119:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// clear
	{[]bool{false, false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return 1
			case 101:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 108:
				return 2
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return 3
			case 108:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 4
			case 99:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 114:
				return 5
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 99:
				return -1
			case 101:
				return -1
			case 108:
				return -1
			case 114:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1}, nil},

	// grep
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 103:
				return 1
			case 112:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 103:
				return -1
			case 112:
				return -1
			case 114:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 3
			case 103:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 103:
				return -1
			case 112:
				return 4
			case 114:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 103:
				return -1
			case 112:
				return -1
			case 114:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// ls
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 108:
				return 1
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 115:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 108:
				return -1
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// exit
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return 1
			case 105:
				return -1
			case 116:
				return -1
			case 120:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 116:
				return -1
			case 120:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return 3
			case 116:
				return -1
			case 120:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 116:
				return 4
			case 120:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 105:
				return -1
			case 116:
				return -1
			case 120:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

	// len
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return 1
			case 110:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return 2
			case 108:
				return -1
			case 110:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return -1
			case 110:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 108:
				return -1
			case 110:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// -l
	{[]bool{false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 45:
				return 1
			case 108:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 108:
				return 2
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 45:
				return -1
			case 108:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1}, nil},

	// [=]
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 61:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 61:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// \/
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 47:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 47:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},

	// man
	{[]bool{false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 109:
				return 1
			case 110:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 2
			case 109:
				return -1
			case 110:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 109:
				return -1
			case 110:
				return 3
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1}, nil},

	// [A-Za-z0-9_]+
	{[]bool{false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 95:
				return 1
			}
			switch {
			case 48 <= r && r <= 57:
				return 1
			case 65 <= r && r <= 90:
				return 1
			case 97 <= r && r <= 122:
				return 1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 95:
				return 1
			}
			switch {
			case 48 <= r && r <= 57:
				return 1
			case 65 <= r && r <= 90:
				return 1
			case 97 <= r && r <= 122:
				return 1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1}, []int{ /* End-of-input transitions */ -1, -1}, nil},
}

func NewLexer(in io.Reader) *Lexer {
	return NewLexerWithInit(in, nil)
}

func (yyLex *Lexer) Stop() {
	yyLex.ch_stop <- true
}

// Text returns the matched text.
func (yylex *Lexer) Text() string {
	return yylex.stack[len(yylex.stack)-1].s
}

// Line returns the current line number.
// The first line is 0.
func (yylex *Lexer) Line() int {
	if len(yylex.stack) == 0 {
		return 0
	}
	return yylex.stack[len(yylex.stack)-1].line
}

// Column returns the current column number.
// The first column is 0.
func (yylex *Lexer) Column() int {
	if len(yylex.stack) == 0 {
		return 0
	}
	return yylex.stack[len(yylex.stack)-1].column
}

func (yylex *Lexer) next(lvl int) int {
	if lvl == len(yylex.stack) {
		l, c := 0, 0
		if lvl > 0 {
			l, c = yylex.stack[lvl-1].line, yylex.stack[lvl-1].column
		}
		yylex.stack = append(yylex.stack, frame{0, "", l, c})
	}
	if lvl == len(yylex.stack)-1 {
		p := &yylex.stack[lvl]
		*p = <-yylex.ch
		yylex.stale = false
	} else {
		yylex.stale = true
	}
	return yylex.stack[lvl].i
}
func (yylex *Lexer) pop() {
	yylex.stack = yylex.stack[:len(yylex.stack)-1]
}
func (yylex Lexer) Error(e string) {
	panic(e)
}

// Lex runs the lexer. Always returns 0.
// When the -s option is given, this function is not generated;
// instead, the NN_FUN macro runs the lexer.
func (yylex *Lexer) Lex(lval *yySymType) int {
OUTER0:
	for {
		switch yylex.next(0) {
		case 0:
			{ /* Skip blanks and tabs. */
			}
		case 1:
			{
				printToks("TOK_CREATE")
				return TOK_CREATE
			}
		case 2:
			{
				printToks("TOK_GET")
				return TOK_GET
			}
		case 3:
			{
				printToks("TOK_UPDATE")
				return TOK_UPDATE
			}
		case 4:
			{
				printToks("TOK_DELETE")
				return TOK_DELETE
			}
		case 5:
			{
				printToks("TOK_SEARCH")
				return TOK_SEARCH
			}
		case 6:
			{
				printToks("TOK_PLUS")
				lval.s = yylex.Text()
				return TOK_PLUS
			}
		case 7:
			{
				printToks("TOK_OCDEL")
				lval.s = yylex.Text()
				return TOK_OCDEL
			}
		case 8:
			{
				printToks("TOK_COL")
				return TOK_COL
			}
		case 9:
			{
				printToks("TOK_ATTRSPEC")
				return TOK_ATTRSPEC
			}
		case 10:
			{
				printToks("TOK_DEREF")
				return TOK_DEREF
			}
		case 11:
			{
				printToks("TOK_SEMICOL")
				return TOK_SEMICOL
			}
		case 12:
			{
				printToks("TOK_LBLOCK")
				return TOK_LBLOCK
			}
		case 13:
			{
				printToks("TOK_RBLOCK")
				return TOK_RBLOCK
			}
		case 14:
			{
				printToks("TOK_LPAREN")
				return TOK_LPAREN
			}
		case 15:
			{
				printToks("TOK_RPAREN")
				return TOK_RPAREN
			}
		case 16:
			{
				printToks("TOK_OR")
				return TOK_OR
			}
		case 17:
			{
				printToks("TOK_AND")
				return TOK_AND
			}
		case 18:
			{
				printToks("TOK_NOT")
				return TOK_NOT
			}
		case 19:
			{
				printToks("TOK_MULT")
				return TOK_MULT
			}
		case 20:
			{
				printToks("TOK_GREATER")
				return TOK_GREATER
			}
		case 21:
			{
				printToks("TOK_LESS")
				return TOK_LESS
			}
		case 22:
			{
				printToks("TOK_BOOL")
				lval.s = yylex.Text()
				return TOK_BOOL
			}
		case 23:
			{
				printToks("TOK_IF")
				return TOK_IF
			}
		case 24:
			{
				printToks("TOK_FOR")
				return TOK_FOR
			}
		case 25:
			{
				printToks("TOK_IN")
				return TOK_IN
			}
		case 26:
			{
				printToks("TOK_WHILE")
				return TOK_WHILE
			}
		case 27:
			{
				printToks("TOK_ELSE")
				return TOK_ELSE
			}
		case 28:
			{
				printToks("TOK_THEN")
				return TOK_THEN
			}
		case 29:
			{
				printToks("TOK_FI")
				return TOK_FI
			}
		case 30:
			{
				printToks("TOK_ELIF")
				return TOK_ELIF
			}
		case 31:
			{
				printToks("TOK_DONE")
				return TOK_DONE
			}
		case 32:
			{
				printToks("TOK_DO")
				return TOK_DO
			}
		case 33:
			{
				printToks("TOK_PRNT")
				return TOK_PRNT
			}
		case 34:
			{
				printToks("TOK_UNSET")
				return TOK_UNSET
			}
		case 35:
			{
				printToks("TOK_QUOT")
				return TOK_QUOT
			}
		case 36:
			{
				printToks("TOK_OCTENANT")
				return TOK_OCTENANT
			}
		case 37:
			{
				printToks("TOK_OCSITE")
				return TOK_OCSITE
			}
		case 38:
			{
				printToks("TOK_OCBLDG")
				return TOK_OCBLDG
			}
		case 39:
			{
				printToks("TOK_OCROOM")
				return TOK_OCROOM
			}
		case 40:
			{
				printToks("TOK_OCRACK")
				return TOK_OCRACK
			}
		case 41:
			{
				printToks("TOK_OCDEV")
				return TOK_OCDEV
			}
		case 42:
			{
				printToks("TOK_OCSDEV")
				return TOK_OCSDEV
			}
		case 43:
			{
				printToks("TOK_OCSDEV1")
				return TOK_OCSDEV1
			}
		case 44:
			{
				printToks("TOK_SELECT")
				return TOK_SELECT
			}
		case 45:
			{
				printToks("TOK_CMDS")
				return TOK_CMDS
			}
		case 46:
			{
				printToks("TOK_TEMPLATE")
				return TOK_TEMPLATE
			}
		case 47:
			{
				printToks("TOK_VAR")
				return TOK_VAR
			}
		case 48:
			{
				printToks("TOK_LBRAC")
				return TOK_LBRAC
			}
		case 49:
			{
				printToks("TOK_RBRAC")
				return TOK_RBRAC
			}
		case 50:
			{
				printToks("TOK_COMMA")
				return TOK_COMMA
			}
		case 51:
			{
				printToks("TOK_DOT")
				return TOK_DOT
			}
		case 52:
			{
				printToks("TOK_TENANT")
				lval.s = yylex.Text()
				return TOK_TENANT
			}
		case 53:
			{
				printToks("TOK_SITE")
				lval.s = yylex.Text()
				return TOK_SITE
			}
		case 54:
			{
				printToks("TOK_BLDG")
				lval.s = yylex.Text()
				return TOK_BLDG
			}
		case 55:
			{
				printToks("TOK_ROOM")
				lval.s = yylex.Text()
				return TOK_ROOM
			}
		case 56:
			{
				printToks("TOK_RACK")
				lval.s = yylex.Text()
				return TOK_RACK
			}
		case 57:
			{
				printToks("TOK_DEVICE")
				lval.s = yylex.Text()
				return TOK_DEVICE
			}
		case 58:
			{
				printToks("TOK_SUBDEVICE")
				lval.s = yylex.Text()
				return TOK_SUBDEVICE
			}
		case 59:
			{
				printToks("TOK_SUBDEVICE1")
				lval.s = yylex.Text()
				return TOK_SUBDEVICE1
			}
		case 60:
			{
				printToks("TOK_ATTR")
				lval.s = yylex.Text()
				return TOK_ATTR
			}
		case 61:
			{
				printToks("TOK_NUM")
				lval.n = atoi(yylex.Text())
				return TOK_NUM
			}
		case 62:
			{
				printToks("TOK_LSTEN")
				return TOK_LSTEN
			}
		case 63:
			{
				printToks("TOK_LSSITE")
				return TOK_LSSITE
			}
		case 64:
			{
				printToks("TOK_LSBLDG")
				return TOK_LSBLDG
			}
		case 65:
			{
				printToks("TOK_LSROOM")
				return TOK_LSROOM
			}
		case 66:
			{
				printToks("TOK_LSRACK")
				return TOK_LSRACK
			}
		case 67:
			{
				printToks("TOK_LSDEV")
				return TOK_LSDEV
			}
		case 68:
			{
				printToks("TOK_LSSUBDEV")
				return TOK_LSSUBDEV
			}
		case 69:
			{
				printToks("TOK_LSSUBDEV1")
				return TOK_LSSUBDEV1
			}
		case 70:
			{
				printToks("TOK_TREE")
				return TOK_TREE
			}
		case 71:
			{
				printToks("TOK_LSOG")
				return TOK_LSOG
			}
		case 72:
			{
				printToks("TOK_CD")
				return TOK_CD
			}
		case 73:
			{
				printToks("TOK_PWD")
				return TOK_PWD
			}
		case 74:
			{
				printToks("TOK_CLR")
				return TOK_CLR
			}
		case 75:
			{
				printToks("TOK_GREP")
				return TOK_GREP
			}
		case 76:
			{
				printToks("TOK_LS")
				return TOK_LS
			}
		case 77:
			{
				printToks("TOK_EXIT")
				return TOK_EXIT
			}
		case 78:
			{
				printToks("TOK_LEN")
				return TOK_LEN
			}
		case 79:
			{
				printToks("TOK_CMDFLAG")
				return TOK_CMDFLAG
			}
		case 80:
			{
				printToks("TOK_EQUAL")
				return TOK_EQUAL
			}
		case 81:
			{
				printToks("TOK_SLASH")
				return TOK_SLASH
			}
		case 82:
			{
				printToks("TOK_DOC")
				return TOK_DOC
			}
		case 83:
			{
				printToks("TOK_WORD")
				lval.s = yylex.Text()
				printCapturedWord(lval.s)
				return TOK_WORD
			}
		default:
			break OUTER0
		}
		continue
	}
	yylex.pop()

	return 0
}

type TOKType int

func atoi(x string) int {
	v, e := strconv.Atoi(x)
	if e != nil {
		println("STRCONV ERROR!")
		return 0
	}
	return v
}

func printToks(tok string) {
	if cmd.State.DebugLvl >= 2 {
		println("We got ", tok)
	}
}

func printCapturedWord(x string) {
	if cmd.State.DebugLvl >= 2 {
		println("LVAL: ", x)
	}
}

func lexBegin() {
	//NN_FUN(NewLexer(os.Stdin))
	//yyParse(NewLexer(os.Stdin))

	lex := NewLexer(strings.NewReader(os.Args[1]))
	e := yyParse(lex)
	println("Return Code: ", e)
}
