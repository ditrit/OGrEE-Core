package readline

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type AutoCompleter interface {
	// Readline will pass the whole line and current offset to it
	// Completer need to pass all the candidates, and how long they shared the same characters in line
	// Example:
	//   [go, git, git-shell, grep]
	//   Do("g", 1) => ["o", "it", "it-shell", "rep"], 1
	//   Do("gi", 2) => ["t", "t-shell"], 2
	//   Do("git", 3) => ["", "-shell"], 3
	Do(line []rune, pos int) (newLine [][]rune, length int)
	DoDynamicFS(line []rune, pos int) (newLine [][]rune, length int, fs bool)
	ReturnFileSysMode() bool
}

type TabCompleter struct{}

func (t *TabCompleter) Do([]rune, int) ([][]rune, int) {
	return [][]rune{[]rune("\t")}, 0
}

func (t *TabCompleter) DoDynamicFS([]rune, int) ([][]rune, int, bool) {
	return [][]rune{[]rune("\t")}, 0, false
}

func (t *TabCompleter) ReturnFileSysMode() bool {
	println("HITTING THE TAB COMPLETER")
	return false
}

type opCompleter struct {
	w     io.Writer
	op    *Operation
	width int

	inCompleteMode  bool
	inSelectMode    bool
	candidate       [][]rune
	candidateSource []rune
	candidateOff    int
	candidateChoise int
	candidateColNum int
}

func newOpCompleter(w io.Writer, op *Operation, width int) *opCompleter {
	return &opCompleter{
		w:     w,
		op:    op,
		width: width,
	}
}

func (o *opCompleter) doSelect() {
	if len(o.candidate) == 1 {
		//Append slash to end of completion
		//word if FileSysMode is enabled
		//if o.op.IsInFileSystemMode() {
		//	o.op.buf.WriteRunes(append(
		//		o.candidate[0][:len(o.candidate[0])-1], '/'))
		//} else {
		o.op.buf.WriteRunes(o.candidate[0])
		//}

		o.ExitCompleteMode(false)
		return
	}
	o.nextCandidate(1)
	o.CompleteRefresh()
}

func (o *opCompleter) nextCandidate(i int) {
	o.candidateChoise += i
	o.candidateChoise = o.candidateChoise % len(o.candidate)
	if o.candidateChoise < 0 {
		o.candidateChoise = len(o.candidate) + o.candidateChoise
	}
}

func (o *opCompleter) OnComplete() bool {
	if o.width == 0 {
		return false
	}
	if o.IsInCompleteSelectMode() {
		o.doSelect()
		return true
	}

	buf := o.op.buf   //This might be golden
	rs := buf.Runes() //This might be golden, indeed rs is the buffer

	//println("THIS IS SUPPOSED TO REMOVE THE PRECEDING SLASH", MyTrimPrecedingSlash((string(rs))))
	//
	//rs = []rune(" " + MyTrimPrecedingSlash((string(rs))))
	//buf.idx = len(rs)

	//println("RS HAS:", string(rs)) <--prints out whatever in buffer

	if o.IsInCompleteMode() && o.candidateSource != nil && runes.Equal(rs, o.candidateSource) {
		//println("WE ENTERED THIS BLOCK")
		//This block pertains to the select menu
		//and control wont come here the first time tab
		//is pressed
		o.EnterCompleteSelectMode()
		o.doSelect()
		return true
	}

	o.ExitCompleteSelectMode()
	o.candidateSource = rs
	//println("I THOUGHT WE FIXED THE RS BUT HERE IT IS: ", string(rs))
	//newLines, offset := o.op.cfg.AutoComplete.Do(rs, buf.idx)
	//println("LENGTH OF AUTOComplete.DO: ", len(newLines))
	newLines, offset, fs := o.op.cfg.AutoComplete.DoDynamicFS(rs, buf.idx)
	//println("ARE WE IN FS MODE? WHICH PREFIX ARE WE USING?", fs)

	if len(newLines) == 0 {
		//println("CLASSIC HAVENT PLAYED IT YET")
		//This block executes when there are no matches
		o.ExitCompleteMode(false)
		o.op.Refresh()
		return true
	}

	// only Aggregate candidates in non-complete mode
	//This section looks like it prints out matches
	if !o.IsInCompleteMode() {
		//This part gets executed when TAB is pressed
		//AND WHEN THERE IS ONLY 1 MATCHING STRING
		if len(newLines) == 1 {
			//Append a slash to completing word
			//if FileSysMode is enabled
			if fs == true {
				buf.WriteRunes(append(
					newLines[0][:len(newLines[0])-1], '/'))
			} else {
				buf.WriteRunes(newLines[0])
			}

			o.ExitCompleteMode(false)
			return true
		}

		//println("OUTPUTTING ALL MATCHES")
		same, size := runes.Aggregate(newLines)
		//println("WHAT THE HECK IS SAME?", string(same))
		if size > 0 {
			buf.WriteRunes(same)
			o.ExitCompleteMode(false)
			return true
		}
	}

	//Sets variables of operation struct then outputs matches
	o.EnterCompleteMode(offset, newLines)
	return true
}

func (o *opCompleter) IsInCompleteSelectMode() bool {
	return o.inSelectMode
}

func (o *opCompleter) IsInCompleteMode() bool {
	return o.inCompleteMode
}

//When you already pressed TAB
//This block will indicate how the buttons behave
//in the TAB selection Menu
func (o *opCompleter) HandleCompleteSelect(r rune) bool {
	//println("WE HANDLING THE COMPLETE SELECT")
	next := true
	switch r {
	case CharEnter, CharCtrlJ:
		next = false
		//Append slash on completing word
		//if FileSysMode is enabled
		//if fs == true {
		//	outString := o.op.candidate[o.op.candidateChoise]
		//	outString = append(outString[:len(outString)-1], '/')
		//	o.op.buf.WriteRunes(outString)
		//} else {
		target := o.op.candidate[o.op.candidateChoise]
		o.op.buf.WriteRunes(target[:len(target)-1])
		//}

		o.ExitCompleteMode(false)
	case CharLineStart:
		num := o.candidateChoise % o.candidateColNum
		o.nextCandidate(-num)
	case CharLineEnd:
		num := o.candidateColNum - o.candidateChoise%o.candidateColNum - 1
		o.candidateChoise += num
		if o.candidateChoise >= len(o.candidate) {
			o.candidateChoise = len(o.candidate) - 1
		}
	case CharBackspace:
		o.ExitCompleteSelectMode()
		next = false
	case CharTab, CharForward:
		o.doSelect()
	case CharBell, CharInterrupt:
		o.ExitCompleteMode(true)
		next = false
	case CharNext:
		tmpChoise := o.candidateChoise + o.candidateColNum
		if tmpChoise >= o.getMatrixSize() {
			tmpChoise -= o.getMatrixSize()
		} else if tmpChoise >= len(o.candidate) {
			tmpChoise += o.candidateColNum
			tmpChoise -= o.getMatrixSize()
		}
		o.candidateChoise = tmpChoise
	case CharBackward:
		o.nextCandidate(-1)
	case CharPrev:
		tmpChoise := o.candidateChoise - o.candidateColNum
		if tmpChoise < 0 {
			tmpChoise += o.getMatrixSize()
			if tmpChoise >= len(o.candidate) {
				tmpChoise -= o.candidateColNum
			}
		}
		o.candidateChoise = tmpChoise
	default:
		next = false
		o.ExitCompleteSelectMode()
	}
	if next {
		//This just moves the highlighted option
		o.CompleteRefresh()
		return true
	}
	return false
}

func (o *opCompleter) getMatrixSize() int {
	line := len(o.candidate) / o.candidateColNum
	if len(o.candidate)%o.candidateColNum != 0 {
		line++
	}
	return line * o.candidateColNum
}

func (o *opCompleter) OnWidthChange(newWidth int) {
	o.width = newWidth
}

func (o *opCompleter) CompleteRefresh() {
	if !o.inCompleteMode {
		return
	}
	lineCnt := o.op.buf.CursorLineCount()
	colWidth := 0
	for _, c := range o.candidate {
		w := runes.WidthAll(c)
		if w > colWidth {
			colWidth = w
		}
	}
	colWidth += o.candidateOff + 1
	same := o.op.buf.RuneSlice(-o.candidateOff)

	// -1 to avoid reach the end of line
	width := o.width - 1
	colNum := width / colWidth
	if colNum != 0 {
		colWidth += (width - (colWidth * colNum)) / colNum
	}

	o.candidateColNum = colNum
	buf := bufio.NewWriter(o.w)
	buf.Write(bytes.Repeat([]byte("\n"), lineCnt))

	colIdx := 0
	lines := 1
	buf.WriteString("\033[J")
	for idx, c := range o.candidate {
		inSelect := idx == o.candidateChoise && o.IsInCompleteSelectMode()
		if inSelect {
			buf.WriteString("\033[30;47m")
		}
		buf.WriteString(string(same))
		buf.WriteString(string(c))
		buf.Write(bytes.Repeat([]byte(" "), colWidth-runes.WidthAll(c)-runes.WidthAll(same)))

		if inSelect {
			buf.WriteString("\033[0m")
		}

		colIdx++
		if colIdx == colNum {
			buf.WriteString("\n")
			lines++
			colIdx = 0
		}
	}

	// move back
	fmt.Fprintf(buf, "\033[%dA\r", lineCnt-1+lines)
	fmt.Fprintf(buf, "\033[%dC", o.op.buf.idx+o.op.buf.PromptLen())
	buf.Flush()
}

func (o *opCompleter) aggCandidate(candidate [][]rune) int {
	offset := 0
	for i := 0; i < len(candidate[0]); i++ {
		for j := 0; j < len(candidate)-1; j++ {
			if i > len(candidate[j]) {
				goto aggregate
			}
			if candidate[j][i] != candidate[j+1][i] {
				goto aggregate
			}
		}
		offset = i
	}
aggregate:
	return offset
}

func (o *opCompleter) EnterCompleteSelectMode() {
	o.inSelectMode = true
	o.candidateChoise = -1
	o.CompleteRefresh()
}

func (o *opCompleter) EnterCompleteMode(offset int, candidate [][]rune) {
	o.inCompleteMode = true
	o.candidate = candidate
	o.candidateOff = offset
	o.CompleteRefresh()
}

func (o *opCompleter) ExitCompleteSelectMode() {
	o.inSelectMode = false
	o.candidate = nil
	o.candidateChoise = -1
	o.candidateOff = -1
	o.candidateSource = nil
}

func (o *opCompleter) ExitCompleteMode(revent bool) {
	o.inCompleteMode = false
	o.ExitCompleteSelectMode()
}
