package main

import (
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

	// false
	{[]bool{false, false, false, false, false, true}, []func(rune) int{ // Transitions
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
			case 115:
				return -1
			}
			return -1
		},
		func(r rune) int {
			switch r {
			case 97:
				return 2
			case 101:
				return -1
			case 102:
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
			case 97:
				return -1
			case 101:
				return -1
			case 102:
				return -1
			case 108:
				return 3
			case 115:
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
			case 115:
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
			case 115:
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
			case 115:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1}, nil},

	// true
	{[]bool{false, false, false, false, true}, []func(rune) int{ // Transitions
		func(r rune) int {
			switch r {
			case 101:
				return -1
			case 114:
				return -1
			case 116:
				return 1
			case 117:
				return -1
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
			case 117:
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
			case 117:
				return 3
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
			case 117:
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
			case 117:
				return -1
			}
			return -1
		},
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1}, nil},

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

	// address|category|city|color|country|description|domain|gps|height|heightUnit|id|mainContact|mainEmail|mainPhone|model|name|nbFloors|orientation|parentId|posU|posXY|posXYUnit|posZ|posZUnit|reserved|reservedColor|serial|size|sizeU|sizeUnit|slot|technical|technicalColor|template|token|type|usableColor|vendor|zipcode
	{[]bool{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false, false, false, false, false, true, false, false, false, false, true, false, false, true, false, false, false, false, false, false, true, false, false, false, false, false, true, false, false, false, false, true, false, false, false, false, true, false, true, true, false, false, true, false, false, false, true, false, false, false, false, false, false, true, false, false, false, false, true, false, false, false, true, false, true, false, false, false, true, true, false, false, false, true, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, true, false, true, false, false, false, false, true, false, false, false, false, false, false, false, false, true, false, false, false, true, false, false, false, false, false, true, true, false, false, false, false, true, false, false, false, true, false, true, false, false, false, false, false, true, false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, true, false, true, false, true, false, false, false, false, false, true, false, false, false, false, false, true}, []func(rune) int{ // Transitions
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
			case 80:
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
				return 1
			case 98:
				return -1
			case 99:
				return 2
			case 100:
				return 3
			case 101:
				return -1
			case 103:
				return 4
			case 104:
				return 5
			case 105:
				return 6
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return 7
			case 110:
				return 8
			case 111:
				return 9
			case 112:
				return 10
			case 114:
				return 11
			case 115:
				return 12
			case 116:
				return 13
			case 117:
				return 14
			case 118:
				return 15
			case 121:
				return -1
			case 122:
				return 16
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
			case 80:
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
				return 202
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 183
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
				return 184
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 185
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
			case 80:
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
				return 168
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 169
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
			case 80:
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
			case 107:
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
				return 166
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
			case 80:
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
				return 157
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 156
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 132
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 133
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
			case 80:
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
				return 122
			case 98:
				return 123
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 112
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
			case 80:
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
				return 91
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 92
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
			case 80:
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
				return 79
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 64
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return 65
			case 107:
				return -1
			case 108:
				return 66
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
			case 80:
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
				return 38
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 39
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
			case 80:
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
			case 107:
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
				return 28
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 17
			case 107:
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
			case 80:
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
			case 107:
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
				return 18
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
			case 80:
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
				return 19
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 20
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
			case 80:
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
				return 21
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 22
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 24
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
			case 80:
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
				return 25
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 26
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
			case 80:
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
			case 107:
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
				return 27
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 29
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
			case 107:
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
			case 80:
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
				return 30
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 31
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
			case 80:
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
				return 32
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
				return 33
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 34
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 35
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 36
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
			case 80:
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
			case 107:
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
				return 37
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 46
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return 47
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
			case 80:
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
			case 107:
				return 43
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 44
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 45
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 53
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
				return 48
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 49
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
			case 80:
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
				return 50
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 51
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
			case 80:
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
				return 52
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 54
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
			case 80:
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
				return 55
			case 107:
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
			case 80:
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
				return 56
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
			case 107:
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
			case 80:
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
				return 57
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 59
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 60
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 61
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 62
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
			case 80:
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
			case 107:
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
				return 63
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 75
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
			case 80:
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
			case 107:
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
				return 69
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 67
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
			case 80:
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
			case 107:
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
				return 68
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 70
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
				return -1
			case 85:
				return 71
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 72
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 74
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 76
			case 107:
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
			case 80:
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
				return 77
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 78
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 80
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 82
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
			case 80:
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
			case 107:
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
				return 83
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
			case 80:
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
				return 84
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 85
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
				return 86
			case 69:
				return -1
			case 70:
				return -1
			case 73:
				return -1
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 87
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 88
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 90
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 106
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
			case 80:
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
			case 107:
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
				return 93
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
			case 80:
				return -1
			case 85:
				return 94
			case 88:
				return 95
			case 89:
				return -1
			case 90:
				return 96
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
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
				return -1
			case 85:
				return -1
			case 88:
				return -1
			case 89:
				return 101
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
			case 107:
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
			case 80:
				return -1
			case 85:
				return 97
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 98
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
			case 80:
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
				return 99
			case 107:
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
			case 80:
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
			case 107:
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
				return 100
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
			case 80:
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
			case 107:
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
			case 80:
				return -1
			case 85:
				return 102
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 103
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
			case 80:
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
				return 104
			case 107:
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
			case 80:
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
			case 107:
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
				return 105
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 107
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 108
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
			case 80:
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
			case 107:
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
				return 109
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
				return 110
			case 80:
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
			case 107:
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
			case 80:
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
				return 111
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 113
			case 107:
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
			case 80:
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
				return 114
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 115
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
			case 80:
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
			case 107:
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
				return 116
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
			case 80:
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
				return 117
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 118
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
			case 80:
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
				return 119
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 120
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 121
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return 130
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
				return 124
			case 73:
				return -1
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 125
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 126
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 127
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
			case 80:
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
			case 107:
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
				return 128
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
			case 80:
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
			case 107:
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
				return 129
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 131
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 137
			case 107:
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
			case 80:
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
				return 134
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 135
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 136
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 138
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
				return 139
			case 69:
				return 140
			case 70:
				return -1
			case 73:
				return -1
			case 80:
				return 141
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 150
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return 146
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
			case 80:
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
				return 142
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 143
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 144
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
			case 80:
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
				return 145
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 147
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
			case 107:
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
			case 80:
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
				return 148
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 149
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 151
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 153
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
			case 107:
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
			case 80:
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
				return 154
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 155
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 158
			case 107:
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
			case 80:
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
				return 159
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 160
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
				return 161
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
			case 80:
				return -1
			case 85:
				return 162
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
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 163
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
			case 80:
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
				return 164
			case 107:
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
			case 80:
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
			case 107:
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
				return 165
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 167
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 174
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return 170
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
			case 80:
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
				return 171
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
			case 107:
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
			case 80:
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
				return 172
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 173
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 175
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 176
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 178
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
			case 80:
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
			case 107:
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
				return 179
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
			case 80:
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
				return 180
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 181
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 182
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 196
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
			case 80:
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
			case 107:
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
				return 194
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return 186
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
				return 187
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 192
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return 188
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
			case 80:
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
			case 107:
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
				return 189
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 191
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 193
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
			case 80:
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
			case 107:
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
			case 80:
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
			case 107:
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
				return 195
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 197
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
				return 198
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
				return -1
			case 108:
				return -1
			case 109:
				return -1
			case 110:
				return -1
			case 111:
				return 199
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
			case 80:
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
			case 107:
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
				return 200
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
			case 80:
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
			case 107:
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
				return 201
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
			case 80:
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
			case 107:
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
			case 80:
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
				return 203
			case 101:
				return -1
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
				return 204
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
			case 80:
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
				return 205
			case 103:
				return -1
			case 104:
				return -1
			case 105:
				return -1
			case 107:
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
			case 80:
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
			case 107:
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
				return 206
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
			case 80:
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
			case 107:
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
				return 207
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
			case 80:
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
			case 107:
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
	}, []int{ /* Start-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, []int{ /* End-of-input transitions */ -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil},

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
				println("We got TOKEN_CREATE")
				return TOKEN_CREATE
			}
		case 2:
			{
				println("We got TOKEN_GET")
				return TOKEN_GET
			}
		case 3:
			{
				println("We got TOKEN_UPDATE")
				return TOKEN_UPDATE
			}
		case 4:
			{
				println("We got TOKEN_DELETE")
				return TOKEN_DELETE
			}
		case 5:
			{
				println("We got TOKEN_SEARCH")
				return TOKEN_SEARCH
			}
		case 6:
			{
				println("We got TOKEN_PLUS")
				lval.s = yylex.Text()
				return TOKEN_PLUS
			}
		case 7:
			{
				println("We got TOKEN_OCDEL")
				lval.s = yylex.Text()
				return TOKEN_OCDEL
			}
		case 8:
			{
				println("We got TOKEN_OCPSPEC")
				return TOKEN_OCPSPEC
			}
		case 9:
			{
				println("We got TOKEN_ATTRSPEC")
				return TOKEN_ATTRSPEC
			}
		case 10:
			{
				println("We got TOKEN_DEREF")
				return TOKEN_DEREF
			}
		case 11:
			{
				println("We got TOKEN_SEMICOL")
				return TOKEN_SEMICOL
			}
		case 12:
			{
				println("We got TOKEN_BOOL")
				lval.s = yylex.Text()
				return TOKEN_BOOL
			}
		case 13:
			{
				println("We got TOKEN_BOOL")
				lval.s = yylex.Text()
				return TOKEN_BOOL
			}
		case 14:
			{
				println("We got TOKEN_OCTENANT")
				return TOKEN_OCTENANT
			}
		case 15:
			{
				println("We got TOKEN_OCSITE")
				return TOKEN_OCSITE
			}
		case 16:
			{
				println("We got TOKEN_OCBLDG")
				return TOKEN_OCBLDG
			}
		case 17:
			{
				println("We got TOKEN_OCROOM")
				return TOKEN_OCROOM
			}
		case 18:
			{
				println("We got TOKEN_OCRACK")
				return TOKEN_OCRACK
			}
		case 19:
			{
				println("We got TOKEN_OCDEV")
				return TOKEN_OCDEV
			}
		case 20:
			{
				println("We got TOKEN_OCSDEV")
				return TOKEN_OCSDEV
			}
		case 21:
			{
				println("We got TOKEN_OCSDEV1")
				return TOKEN_OCSDEV1
			}
		case 22:
			{
				println("We got TOKEN_SELECT")
				return TOKEN_SELECT
			}
		case 23:
			{
				println("We got TOKEN_CMDS")
				return TOKEN_CMDS
			}
		case 24:
			{
				println("We got TOKEN_TEMPLATE")
				return TOKEN_TEMPLATE
			}
		case 25:
			{
				println("We got TOKEN_VAR")
				return TOKEN_VAR
			}
		case 26:
			{
				println("We got TOKEN_LBRAC")
				return TOKEN_LBRAC
			}
		case 27:
			{
				println("We got TOKEN_RBRAC")
				return TOKEN_RBRAC
			}
		case 28:
			{
				println("We got TOKEN_COMMA")
				return TOKEN_COMMA
			}
		case 29:
			{
				println("We got TOKEN_DOT")
				return TOKEN_DOT
			}
		case 30:
			{
				println("We got TOKEN_TENANT")
				lval.s = yylex.Text()
				return TOKEN_TENANT
			}
		case 31:
			{
				println("We got TOKEN_SITE")
				lval.s = yylex.Text()
				return TOKEN_SITE
			}
		case 32:
			{
				println("We got TOKEN_BLDG")
				lval.s = yylex.Text()
				return TOKEN_BLDG
			}
		case 33:
			{
				println("We got TOKEN_ROOM")
				lval.s = yylex.Text()
				return TOKEN_ROOM
			}
		case 34:
			{
				println("We got TOKEN_RACK")
				lval.s = yylex.Text()
				return TOKEN_RACK
			}
		case 35:
			{
				println("We got TOKEN_DEVICE")
				lval.s = yylex.Text()
				return TOKEN_DEVICE
			}
		case 36:
			{
				println("We got TOKEN_SUBDEVICE")
				lval.s = yylex.Text()
				return TOKEN_SUBDEVICE
			}
		case 37:
			{
				println("We got TOKEN_SUBDEVICE1")
				lval.s = yylex.Text()
				return TOKEN_SUBDEVICE1
			}
		case 38:
			{
				println("We got TOKEN_ATTR")
				lval.s = yylex.Text()
				return TOKEN_ATTR
			}
		case 39:
			{
				println("We got TOKEN_NUM")
				lval.n = atoi(yylex.Text())
				return TOKEN_NUM
			}
		case 40:
			{
				println("We got TOKEN_LSTEN")
				return TOKEN_LSTEN
			}
		case 41:
			{
				println("We got TOKEN_LSSITE")
				return TOKEN_LSSITE
			}
		case 42:
			{
				println("We got TOKEN_LSBLDG")
				return TOKEN_LSBLDG
			}
		case 43:
			{
				println("We got TOKEN_LSROOM")
				return TOKEN_LSROOM
			}
		case 44:
			{
				println("We got TOKEN_LSRACK")
				return TOKEN_LSRACK
			}
		case 45:
			{
				println("We got TOKEN_LSDEV")
				return TOKEN_LSDEV
			}
		case 46:
			{
				println("We got TOKEN_LSSUBDEV")
				return TOKEN_LSSUBDEV
			}
		case 47:
			{
				println("We got TOKEN_LSSUBDEV1")
				return TOKEN_LSSUBDEV1
			}
		case 48:
			{
				println("We got TOKEN_TREE")
				return TOKEN_TREE
			}
		case 49:
			{
				println("We got TOKEN_LSOG")
				return TOKEN_LSOG
			}
		case 50:
			{
				println("We got TOKEN_CD")
				return TOKEN_CD
			}
		case 51:
			{
				println("We got TOKEN_PWD")
				return TOKEN_PWD
			}
		case 52:
			{
				println("We got TOKEN_CLR")
				return TOKEN_CLR
			}
		case 53:
			{
				println("We got TOKEN_GREP")
				return TOKEN_GREP
			}
		case 54:
			{
				println("We got TOKEN_LS")
				return TOKEN_LS
			}
		case 55:
			{
				println("We got TOKEN_EXIT")
				return TOKEN_EXIT
			}
		case 56:
			{
				println("We got TOKEN_CMDFLAG")
				return TOKEN_CMDFLAG
			}
		case 57:
			{
				println("We got TOKEN_EQUAL")
				return TOKEN_EQUAL
			}
		case 58:
			{
				println("We got TOKEN_SLASH")
				return TOKEN_SLASH
			}
		case 59:
			{
				println("We got TOKEN_DOC")
				return TOKEN_DOC
			}
		case 60:
			{
				println("We got TOKEN_WORD")
				lval.s = yylex.Text()
				println("LVAL: ", lval.s)
				return TOKEN_WORD
			}
		default:
			break OUTER0
		}
		continue
	}
	yylex.pop()

	return 0
}

type TokenType int

func atoi(x string) int {
	v, e := strconv.Atoi(x)
	if e != nil {
		println("STRCONV ERROR!")
		return 0
	}
	return v
}

func lexBegin() {
	//NN_FUN(NewLexer(os.Stdin))
	//yyParse(NewLexer(os.Stdin))

	lex := NewLexer(strings.NewReader(os.Args[1]))
	e := yyParse(lex)
	println("Return Code: ", e)
}
