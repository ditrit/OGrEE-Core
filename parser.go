package main

import (
	"regexp"
	"strconv"
	"strings"
)

type parseCommandFunc func(frame Frame) (node, Frame, *ParserError)

var commandDispatch map[string]parseCommandFunc
var createObjDispatch map[string]parseCommandFunc

var lsCommands = []string{"lsten", "lssite", "lsbldg", "lsroom", "lsrack", "lsdev", "lsac",
	"lspanel", "lscabinet", "lscorridor", "lssensor"}
var noArgsCommands map[string]node

var manCommands = []string{
	"get", "getu", "getslot",
	"+", "-", "=",
	".cmds", ".template", ".var",
	"ui", "camera",
	"link", "unlink",
	"lsten", "lssite", "lsbldg", "lsroom", "lsrack", "lsdev", "lsac",
	"lspanel", "lscabinet", "lscorridor", "lssensor", "lsenterprise",
	"drawable", "draw", "undraw",
	"tree", "lsog", "env", "cd", "pwd", "clear", "grep", "ls", "exit", "len", "man", "hc",
	"print", "unset", "selection",
}

func sliceContains(slice []string, s string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}
	return false
}

func indexOf(arr []string, val string) int {
	for pos, v := range arr {
		if v == val {
			return pos
		}
	}
	return -1
}

func regexMatch(regex string, str string) bool {
	reg, err := regexp.Compile(regex)
	if err != nil {
		panic("Regexp compilation error")
	}
	reg.Longest()
	return str == reg.FindString(str)
}

type ParserError struct {
	frames   []Frame
	messages []string
}

func buildColoredFrame(frame Frame) string {
	result := ""
	result += frame.buf[0:frame.start]
	result += "\033[31m"
	result += "|"
	result += frame.buf[frame.start:frame.end]
	result += "\033[0m"
	result += frame.buf[frame.end:]
	return result
}

func (err *ParserError) Error() string {
	errorString := ""
	for i := len(err.messages) - 1; i >= 0; i-- {
		frame := err.frames[i]
		errorString += buildColoredFrame(frame) + "\n"
		errorString += err.messages[i]
		if i > 0 {
			errorString += "\n"
		}
	}
	return errorString
}

func (err *ParserError) extend(frame Frame, message string) *ParserError {
	return &ParserError{append(err.frames, frame), append(err.messages, message)}
}

func (err *ParserError) extendMessage(message string) *ParserError {
	currentMessage := err.messages[len(err.messages)-1]
	err.messages[len(err.messages)-1] = message + " : " + currentMessage
	return err
}

func newParserError(frame Frame, message string) *ParserError {
	return &ParserError{[]Frame{frame}, []string{message}}
}

type Frame struct {
	buf   string
	start int
	end   int
}

func newFrame(buffer string) Frame {
	return Frame{buffer, 0, len(buffer)}
}

func (frame Frame) new(start int, end int) Frame {
	if start < frame.start || start > frame.end || end < frame.start || end > frame.end {
		panic("the subframe is not included in the topframe")
	}
	return Frame{frame.buf, start, end}
}

func (frame Frame) until(end int) Frame {
	return frame.new(frame.start, end)
}

func (frame Frame) from(start int) Frame {
	return frame.new(start, frame.end)
}

func (frame Frame) str() string {
	return frame.buf[frame.start:frame.end]
}

func (frame Frame) char(i int) byte {
	if i < frame.start || i >= frame.end {
		panic("index outside of frame bounds")
	}
	return frame.buf[i]
}

func (frame Frame) forward(offset int) Frame {
	if frame.start+offset > frame.end {
		panic("cannot go forward")
	}
	return Frame{frame.buf, frame.start + offset, frame.end}
}

func (frame Frame) first() byte {
	return frame.char(frame.start)
}

func (frame Frame) String() string {
	return buildColoredFrame(frame)
}

func lexerFromFrame(frame Frame) *lexer {
	return newLexer(frame.buf, frame.start, frame.end)
}

func skipWhiteSpaces(frame Frame) Frame {
	i := frame.start
	for i < frame.end && (frame.char(i) == ' ' || frame.char(i) == '\t') {
		i += 1
	}
	return frame.from(i)
}

func findNext(substring string, frame Frame) int {
	idx := strings.Index(frame.str(), substring)
	if idx != -1 {
		return frame.start + idx
	}
	return frame.end
}

func findNextQuote(frame Frame) int {
	return findNext("\"", frame)
}

func findClosing(frame Frame) int {
	openToClose := map[byte]byte{'(': ')', '{': '}', '[': ']'}
	open := frame.first()
	close, ok := openToClose[open]
	if !ok {
		panic("invalid opening character")
	}
	stackCount := 0
	inString := false
	for cursor := frame.start; cursor < frame.end; cursor++ {
		if inString {
			if frame.char(cursor) == '"' {
				inString = false
			}
			continue
		}
		if frame.char(cursor) == '"' {
			inString = true
		} else if frame.char(cursor) == open {
			stackCount++
		} else if frame.char(cursor) == close {
			stackCount--
		}
		if stackCount == 0 {
			return cursor
		}
	}
	return frame.end
}

func frameEnd(endChars string, frame Frame) bool {
	frame = skipWhiteSpaces(frame)
	return frame.start == frame.end || strings.Contains(endChars, string(frame.first()))
}

func commandEnd(frame Frame) bool {
	return frameEnd(";})", frame)
}

func exprEnd(frame Frame) bool {
	return frameEnd(";})@", frame)
}

func parseExact(word string, frame Frame) (bool, Frame) {
	if frame.start+len(word) <= frame.end && frame.until(frame.start+len(word)).str() == word {
		return true, frame.forward(len(word))
	}
	return false, frame
}

func isPrefix(prefix string, candidates []string) bool {
	for _, candidate := range candidates {
		if strings.HasPrefix(candidate, prefix) {
			return true
		}
	}
	return false
}

func parseKeyWord(candidates []string, frame Frame) (string, Frame) {
	commandEnd := frame.start
	for commandEnd < frame.end && isPrefix(frame.until(commandEnd+1).str(), candidates) {
		commandEnd++
	}
	longestPrefix := frame.until(commandEnd).str()
	if sliceContains(candidates, longestPrefix) {
		return longestPrefix, frame.from(commandEnd)
	}
	return "", frame
}

func parseWord(frame Frame) (string, Frame, *ParserError) {
	l := lexerFromFrame(frame)
	tok := l.nextToken(lexExpr)
	if tok.t != tokWord {
		return "", frame, newParserError(frame, "word expected")
	}
	return tok.str, frame.from(tok.end), nil
}

func parseSeparatedStuff(
	sep byte,
	frame Frame,
	parseStuff func(Frame) (any, Frame, *ParserError),
) ([]any, *ParserError) {
	items := []any{}

	for {
		var item any
		var err *ParserError
		item, frame, err = parseStuff(frame)
		if err != nil {
			return nil, err.extend(frame, "parsing item in list")
		}
		items = append(items, item)
		frame = skipWhiteSpaces(frame)
		if frame.start == frame.end {
			return items, nil
		}
		if frame.first() != sep {
			return nil, newParserError(frame, string(sep)+" expected")
		}
		frame = skipWhiteSpaces(frame.forward(1))
	}
}

func parseSeparatedWords(sep byte, frame Frame) ([]string, *ParserError) {
	parseFunc := func(frame Frame) (any, Frame, *ParserError) {
		return parseWord(frame)
	}
	wordsAny, err := parseSeparatedStuff(sep, frame, parseFunc)
	if err != nil {
		return nil, err.extend(frame, "parsing list of words")
	}
	words := []string{}
	for _, wordAny := range wordsAny {
		words = append(words, wordAny.(string))
	}
	return words, nil
}

func charIsNumber(char byte) bool {
	return char >= 48 && char <= 57
}

func parseInt(frame Frame) (int, Frame, *ParserError) {
	end := frame.start
	for end < frame.end && charIsNumber(frame.char(end)) {
		end++
	}
	if end == frame.start {
		return 0, frame, newParserError(frame, "integer expected")
	}
	intString := frame.until(end).str()
	val, err := strconv.Atoi(intString)
	if err != nil {
		panic("cannot convert " + intString + " to integer")
	}
	return val, frame.from(end), nil
}

func parseFloat(frame Frame) (float64, Frame, *ParserError) {
	end := frame.start
	dotseen := false
	for end < frame.end {
		if frame.char(end) == '.' {
			if dotseen {
				break
			}
			dotseen = true
		} else if !charIsNumber(frame.char(end)) {
			break
		}
		end++
	}
	if end == frame.start {
		return 0, frame, newParserError(frame, "float expected")
	}
	floatString := frame.until(end).str()
	val, err := strconv.ParseFloat(floatString, 64)
	if err != nil {
		panic("cannot convert " + floatString + " to float")
	}
	return val, frame.from(end), nil
}

func parseBool(frame Frame) (bool, Frame, *ParserError) {
	if frame.end-frame.start >= 4 && frame.until(frame.start+4).str() == "true" {
		return true, frame.forward(4), nil
	}
	if frame.end-frame.start >= 5 && frame.until(frame.start+5).str() == "false" {
		return false, frame.forward(5), nil
	}
	return false, frame, newParserError(frame, "bool expected")
}

func parseRawText(lexFunc stateFn, frame Frame) (node, Frame, *ParserError) {
	l := lexerFromFrame(frame)
	s := ""
	vars := []symbolReferenceNode{}
loop:
	for {
		tok := l.nextToken(lexFunc)
		switch tok.t {
		case tokText:
			s += tok.str
		case tokDeref:
			s += "%v"
			vars = append(vars, symbolReferenceNode{tok.val.(string)})
		case tokEOF:
			break loop
		default:
			return nil, frame, newParserError(frame, "unexpected token")
		}
	}
	frame = frame.from(l.tok.end)
	if len(vars) == 0 {
		return &strLeaf{s}, frame, nil
	}
	return &formatStringNode{s, vars}, frame, nil
}

func parsePath(frame Frame) (node, Frame, *ParserError) {
	frame = skipWhiteSpaces(frame)
	path, frame, err := parseRawText(lexPath, frame)
	if err != nil {
		return nil, frame, err.extend(frame, "parsing path")
	}
	return &pathNode{path}, skipWhiteSpaces(frame), nil
}

func parsePathGroup(frame Frame) ([]node, Frame, *ParserError) {
	var err *ParserError
	var path node

	ok, frame := parseExact("{", frame)
	if !ok {
		return nil, frame, newParserError(frame, "{ expected")
	}
	paths := []node{}
	for {
		frame = skipWhiteSpaces(frame)
		path, frame, err = parsePath(frame)
		if err != nil {
			return nil, frame, err
		}
		paths = append(paths, path)
		frame = skipWhiteSpaces(frame)
		ok, frame = parseExact("}", frame)
		if ok {
			break
		}
		ok, frame = parseExact(",", frame)
		if !ok {
			return nil, frame, newParserError(frame, "comma expected")
		}
	}
	return paths, frame, nil
}

func exprError(l *lexer, message string) *ParserError {
	frame := Frame{
		buf:   l.input,
		start: l.tok.start,
		end:   l.tok.end,
	}
	return newParserError(frame, message)
}

func parsePrimaryExpr(l *lexer) (node, *ParserError) {
	tok := l.tok
	l.nextToken(lexExpr)
	switch tok.t {
	case tokBool:
		return &boolLeaf{tok.val.(bool)}, nil
	case tokInt:
		return &intLeaf{tok.val.(int)}, nil
	case tokFloat:
		return &floatLeaf{tok.val.(float64)}, nil
	case tokString:
		n, _, err := parseRawText(lexQuotedString, newFrame(tok.val.(string)))
		if err != nil {
			return nil, exprError(l, "cannot parse string")
		}
		return n, nil
	case tokDeref:
		if l.tok.t != tokLeftBrac {
			return &symbolReferenceNode{tok.val.(string)}, nil
		}
		l.nextToken(lexExpr)
		index, err := parseExprFromLex(l)
		if err != nil {
			return nil, err
		}
		if l.tok.t != tokRightBrac {
			return nil, exprError(l, "square bracket opened but not closed")
		}
		l.nextToken(lexExpr)
		return &objReferenceNode{tok.val.(string), index}, nil
	case tokLeftParen:
		expr, err := parseExprFromLex(l)
		if err != nil {
			return nil, err
		}
		endTok := l.tok
		if endTok.t != tokRightParen {
			return nil, exprError(l, ") expected, got "+endTok.str)
		}
		l.nextToken(lexExpr)
		return expr, nil
	case tokLeftBrac:
		exprList := []node{}
		if l.tok.t == tokRightBrac {
			l.nextToken(lexExpr)
			return &arrNode{exprList}, nil
		}
		for {
			expr, err := parseExprFromLex(l)
			if err != nil {
				return nil, err
			}
			exprList = append(exprList, expr)
			if l.tok.t == tokRightBrac {
				l.nextToken(lexExpr)
				return &arrNode{exprList}, nil
			}
			if l.tok.t == tokComma {
				l.nextToken(lexExpr)
				continue
			}
			return nil, exprError(l, "] or comma expected")
		}
	}
	return nil, exprError(l, "unexpected token : "+tok.str)
}

func parseUnaryExpr(l *lexer) (node, *ParserError) {
	switch l.tok.t {
	case tokAdd:
		l.nextToken(lexExpr)
		return parseUnaryExpr(l)
	case tokSub:
		l.nextToken(lexExpr)
		x, err := parseUnaryExpr(l)
		if err != nil {
			return nil, err
		}
		return &negateNode{x}, nil
	case tokNot:
		l.nextToken(lexExpr)
		x, err := parseUnaryExpr(l)
		if err != nil {
			return nil, err
		}
		return &negateBoolNode{x}, nil
	}
	return parsePrimaryExpr(l)
}

func parseBinaryExpr(l *lexer, leftOperand node, precedence int) (node, *ParserError) {
	var err *ParserError
	if leftOperand == nil {
		leftOperand, err = parseUnaryExpr(l)
		if err != nil {
			return nil, err
		}
	}
	for {
		operator := l.tok
		operatorPrecedence := operator.precedence()
		if operatorPrecedence < precedence {
			return leftOperand, nil
		}
		l.nextToken(lexExpr)
		rightOperand, err := parseBinaryExpr(l, nil, operatorPrecedence+1)
		if err != nil {
			return nil, err
		}
		switch operator.t {
		case tokAdd, tokSub, tokMul, tokDiv, tokMod:
			leftOperand = &arithNode{operator.str, leftOperand, rightOperand}
		case tokOr, tokAnd:
			leftOperand = &logicalNode{operator.str, leftOperand, rightOperand}
		case tokEq, tokNeq:
			leftOperand = &equalityNode{operator.str, leftOperand, rightOperand}
		case tokLeq, tokGeq, tokGtr, tokLss:
			leftOperand = &comparatorNode{operator.str, leftOperand, rightOperand}
		}
	}
}

func parseExprFromLex(l *lexer) (node, *ParserError) {
	return parseBinaryExpr(l, nil, 1)
}

func parseExpr(frame Frame) (node, Frame, *ParserError) {
	l := lexerFromFrame(frame)
	l.nextToken(lexExpr)
	expr, err := parseExprFromLex(l)
	if err != nil {
		return nil, frame.from(l.tok.start), err
	}
	if expr == nil {
		return nil, frame.from(l.tok.start), newParserError(frame, "expression expected")
	}
	return expr, frame.from(l.tok.start), nil
}

func parseAssign(frame Frame) (string, Frame, *ParserError) {
	eqIdx := findNext("=", frame)
	if eqIdx == frame.end {
		return "", frame, newParserError(frame, "= expected")
	}
	varName, frame, err := parseWord(frame)
	if err != nil {
		return "", frame, err.extendMessage("parsing word on the left of =")
	}
	frame = skipWhiteSpaces(frame)
	if frame.first() != '=' {
		return "", frame, newParserError(skipWhiteSpaces(frame), "= expected")
	}
	return varName, frame.forward(1), nil
}

func parseIndexing(frame Frame) (node, Frame, *ParserError) {
	frame = skipWhiteSpaces(frame)
	ok, frame := parseExact("[", frame)
	if !ok {
		return nil, frame, newParserError(frame, "[ expected")
	}
	index, frame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, err.extend(frame, "parsing indexing")
	}
	ok, frame = parseExact("]", frame)
	if !ok {
		return nil, frame, newParserError(frame, "] expected")
	}
	return index, frame, nil
}

func parseArgValue(frame Frame) (string, Frame, *ParserError) {
	if commandEnd(frame) {
		return "", frame, newParserError(frame, "argument value expected")
	}
	if frame.first() == '(' {
		close := findClosing(frame)
		if close == frame.end {
			return "", frame, newParserError(frame, "( opened but never closed")
		}
		return frame.until(close + 1).str(), frame.from(close + 1), nil
	} else if frame.first() == '"' {
		endQuote := findNextQuote(frame)
		if endQuote == frame.end {
			return "", frame, newParserError(frame, "\" opened but never closed")
		}
		return frame.until(endQuote).str(), frame.from(endQuote + 1), nil
	}
	endValue := findNext(" ", frame)
	return frame.until(endValue).str(), skipWhiteSpaces(frame.from(endValue)), nil
}

func parseSingleArg(allowedArgs []string, allowedFlags []string, frame Frame) (
	string, string, Frame, *ParserError,
) {
	topFrame := frame
	frame = skipWhiteSpaces(frame.forward(1))
	arg, frame, err := parseWord(frame)
	if err != nil {
		return "", "", frame, err.extendMessage("parsing arg name").
			extend(topFrame, "parsing argument")
	}
	frame = skipWhiteSpaces(frame)
	var value string
	if sliceContains(allowedArgs, arg) {
		value, frame, err = parseArgValue(frame)
		if err != nil {
			return "", "", frame, err.extendMessage("pasing arg value").
				extend(topFrame, "parsing argument")
		}
	} else if sliceContains(allowedFlags, arg) {
		value = ""
	} else {
		panic("unexpected argument")
	}
	return arg, value, skipWhiteSpaces(frame), nil
}

func parseArgs(allowedArgs []string, allowedFlags []string, frame Frame) (
	map[string]string, Frame, *ParserError,
) {
	args := map[string]string{}
	frame = skipWhiteSpaces(frame)
	for frame.start < frame.end && frame.first() == '-' {
		arg, value, newFrame, err := parseSingleArg(allowedArgs, allowedFlags, frame)
		if err != nil {
			return nil, frame, err
		}
		args[arg] = value
		frame = newFrame
	}
	return args, frame, nil
}

func parseLsObj(lsIdx int, frame Frame) (node, Frame, *ParserError) {
	args, frame, err := parseArgs([]string{"s", "f"}, []string{"r"}, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing lsobj arguments")
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("pasing lsobj path")
	}
	_, recursive := args["r"]
	sort := args["s"]

	//msg := "Please provide a quote enclosed string for '-f' with arguments separated by ':'. Or provide an argument with printf formatting (ie -f (\"%d\",arg1))"

	var attrList []string
	var format string
	if formatArg, ok := args["f"]; ok {
		if regexMatch(`\(\s*".*"\s*,.+\)`, formatArg) {
			formatFrame := Frame{formatArg, 1, len(formatArg)}
			startFormat := findNextQuote(formatFrame)
			endFormat := findNextQuote(formatFrame.from(startFormat + 1))
			format = formatArg[startFormat+1 : endFormat]
			cursor := findNext(",", formatFrame.from(endFormat)) + 1
			attrList, err = parseSeparatedWords(',', formatFrame.new(cursor, len(formatArg)-1))
			if err != nil {
				return nil, frame, err.extendMessage("parsing lsobj format")
			}
		} else {
			formatFrame := newFrame(formatArg)
			attrList, err = parseSeparatedWords(':', formatFrame)
			if err != nil {
				return nil, frame, err.extendMessage("parsing lsobj format")
			}
		}
	}
	return &lsObjNode{path, lsIdx, recursive, sort, attrList, format}, frame, nil
}

func parseLs(frame Frame) (node, Frame, *ParserError) {
	args, frame, err := parseArgs([]string{"s", "f"}, []string{"r"}, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing ls arguments")
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing ls path")
	}
	if attr, ok := args["s"]; ok {
		return &lsAttrNode{path, attr}, frame, nil
	}
	return &lsNode{path}, frame, nil
}

func parseGet(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing get path")
	}
	return &getObjectNode{path}, frame, nil
}

func parseGetU(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing getu path")
	}
	u, frame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing getu depth")
	}
	return &getUNode{path, u}, frame, nil
}

func parseGetSlot(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing getslot path")
	}
	slotName, frame, err := parseStringExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing getslot slot name")
	}
	return &getSlotNode{path, slotName}, frame, nil
}

func parseUndraw(frame Frame) (node, Frame, *ParserError) {
	if commandEnd(frame) {
		return &undrawNode{nil}, frame, nil
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing undraw path")
	}
	return &undrawNode{path}, frame, nil
}

func parseDraw(frame Frame) (node, Frame, *ParserError) {
	args, frame, err := parseArgs([]string{}, []string{"f"}, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing draw arguments")
	}
	_, force := args["f"]
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing draw path")
	}
	depth := 0
	if !commandEnd(frame) {
		depth, frame, err = parseInt(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing draw depth")
		}
	}
	return &drawNode{path, depth, force}, frame, nil
}

func parseDrawable(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing drawable path")
	}
	if commandEnd(frame) {
		return &isEntityDrawableNode{path}, frame, nil
	}
	attrName, _, err := parseWord(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing drawable attribute name")
	}
	return &isAttrDrawableNode{path, attrName}, frame, nil
}

func parseHc(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing hc path")
	}
	if commandEnd(frame) {
		return &hierarchyNode{path, 1}, frame, nil
	}
	depth, frame, err := parseInt(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing hc depth")
	}
	return &hierarchyNode{path, depth}, frame, nil
}

func parseUnset(frame Frame) (node, Frame, *ParserError) {
	args, frame, err := parseArgs([]string{"f", "v"}, []string{}, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing unset arguments")
	}
	if len(args) == 0 {
		path, frame, err := parsePath(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing unset path")
		}
		ok, frame := parseExact(":", frame)
		if !ok {
			return nil, frame, newParserError(frame, ": expected")
		}
		attr, frame, err := parseWord(frame)
		if err != nil {
			return nil, frame, err.extend(frame, "parsing attribute name")
		}
		index, frame, _ := parseIndexing(frame)
		return &unsetAttrNode{path, attr, index}, frame, nil
	}
	if funcName, ok := args["f"]; ok {
		return &unsetFuncNode{funcName}, frame, nil
	}
	if varName, ok := args["v"]; ok {
		return &unsetVarNode{varName}, frame, nil
	}
	panic("unexpected argument while parsing unset command")
}

func parseEnv(frame Frame) (node, Frame, *ParserError) {
	if commandEnd(frame) {
		return &envNode{}, frame, nil
	}
	arg, valueFrame, err := parseAssign(frame)
	if err != nil {
		return nil, frame, err
	}
	value, frame, err := parseStringExpr(valueFrame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing env variable value")
	}
	return &setEnvNode{arg, value}, frame, nil
}

func parseDelete(frame Frame) (node, Frame, *ParserError) {
	deleteSelection, frame := parseExact("selection", frame)
	if deleteSelection {
		return &deleteSelectionNode{}, frame, nil
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing deletion path")
	}
	return &deleteObjNode{path}, frame, nil
}

func parseEqual(frame Frame) (node, Frame, *ParserError) {
	if frame.first() == '{' {
		paths, frame, err := parsePathGroup(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing selection paths")
		}
		return &selectChildrenNode{paths}, frame, nil
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing selection path")
	}
	return &selectObjectNode{path}, frame, nil
}

func parseVar(frame Frame) (node, Frame, *ParserError) {
	topFrame := frame
	varName, frame, err := parseAssign(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing variable assignment")
	}
	frame = skipWhiteSpaces(frame)
	commandExpr, frame := parseExact("$(", frame)
	if commandExpr {
		value, frame, err := parseCommand(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing variable value (command expression)")
		}
		closed, frame := parseExact(")", frame)
		if !closed {
			return nil, frame, newParserError(frame, "$( opened but never closed").
				extend(topFrame, "parsing variable assignment")
		}
		return &assignNode{varName, value}, frame, nil
	}
	value, frame, err := parseStringExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing variable value")
	}
	return &assignNode{varName, value}, frame, nil
}

func parseLoad(frame Frame) (node, Frame, *ParserError) {
	filePath, frame, err := parseStringExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing file path")
	}
	return &loadNode{filePath}, frame, nil
}

func parseTemplate(frame Frame) (node, Frame, *ParserError) {
	filePath, frame, err := parseStringExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing file path")
	}
	return &loadTemplateNode{filePath}, frame, nil
}

func parseLen(frame Frame) (node, Frame, *ParserError) {
	varName, frame, err := parseWord(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing variable name")
	}
	return &lenNode{varName}, frame, nil
}

func parseLink(frame Frame) (node, Frame, *ParserError) {
	sourcePath, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing source path (physical)")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return nil, frame, newParserError(frame, "@ expected")
	}
	destPath, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing destination path (physical)")
	}
	ok, frame = parseExact("@", frame)
	if ok {
		slot, frame, err := parseStringExpr(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing slot name")
		}
		return &linkObjectNode{sourcePath, destPath, slot}, frame, nil
	}
	return &linkObjectNode{sourcePath, destPath, nil}, frame, nil
}

func parseUnlink(frame Frame) (node, Frame, *ParserError) {
	sourcePath, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing source path (physical)")
	}
	ok, frame := parseExact("@", frame)
	if ok {
		destPath, frame, err := parsePath(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing destination path (physical)")
		}
		return &unlinkObjectNode{sourcePath, destPath}, frame, nil
	}
	return &unlinkObjectNode{sourcePath, nil}, frame, nil
}

func parsePrint(frame Frame) (node, Frame, *ParserError) {
	str, frame, err := parseStringExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing message to print")
	}
	return &printNode{str}, frame, nil
}

func parseMan(frame Frame) (node, Frame, *ParserError) {
	if commandEnd(frame) {
		return &helpNode{""}, frame, nil
	}
	endCommandName := findNext(" ", frame)
	commandName := frame.until(endCommandName).str()
	if !sliceContains(manCommands, commandName) {
		return nil, frame, newParserError(frame, "unknown command")
	}
	return &helpNode{commandName}, frame, nil
}

func parseCd(frame Frame) (node, Frame, *ParserError) {
	if commandEnd(frame) {
		return &cdNode{strLeaf{"/"}}, frame, nil
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing path")
	}
	return &cdNode{path}, frame, nil
}

func parseTree(frame Frame) (node, Frame, *ParserError) {
	if commandEnd(frame) {
		return &treeNode{&pathNode{&strLeaf{"."}}, 0}, frame, nil
	}
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing tree path")
	}
	if commandEnd(frame) {
		return &treeNode{path, 0}, frame, nil
	}
	u, frame, err := parseInt(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing tree depth")
	}
	return &treeNode{path, u}, frame, nil
}

func parseUi(frame Frame) (node, Frame, *ParserError) {
	key, valueFrame, err := parseAssign(frame)
	if err != nil {
		return nil, frame, err
	}
	if key == "delay" {
		delay, frame, err := parseFloat(valueFrame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing ui delay")
		}
		return &uiDelayNode{delay}, frame, nil
	}
	if key == "debug" || key == "infos" || key == "wireframe" {
		val, frame, err := parseBool(valueFrame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing ui toggle " + key)
		}
		return &uiToggleNode{key, val}, frame, nil
	}
	if key == "highlight" || key == "hl" {
		path, frame, err := parsePath(valueFrame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing ui highlight")
		}
		return &uiHighlightNode{path}, frame, nil
	}
	return nil, frame, newParserError(frame, "unknown ui command")
}

func parseCamera(frame Frame) (node, Frame, *ParserError) {
	key, frame, err := parseAssign(frame)
	if err != nil {
		return nil, frame, err
	}
	if key == "move" || key == "translate" {
		position, frame, err := parseExpr(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing position vector")
		}
		ok, frame := parseExact("@", frame)
		if !ok {
			return nil, frame, newParserError(frame, "@ expected")
		}
		rotation, frame, err := parseExpr(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing rotation vector")
		}
		return &cameraMoveNode{key, position, rotation}, frame, nil
	}
	if key == "wait" {
		time, frame, err := parseFloat(frame)
		if err != nil {
			return nil, frame, err.extendMessage("parsing waiting time")
		}
		return &cameraWaitNode{time}, frame, nil
	}
	return nil, frame, newParserError(frame, "unknown ui command")
}

func parseFocus(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing path")
	}
	return &focusNode{path}, frame, nil
}

func parseWhile(frame Frame) (node, Frame, *ParserError) {
	condition, frame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing condition")
	}
	ok, frame := parseExact("{", frame)
	if !ok {
		return nil, frame, newParserError(frame, "{ expected")
	}
	body, frame, err := parseCommand(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing while body")
	}
	ok, frame = parseExact("}", skipWhiteSpaces(frame))
	if !ok {
		return nil, frame, newParserError(frame, "} expected")
	}
	return &whileNode{condition, body}, frame, nil
}

func parseFor(frame Frame) (node, Frame, *ParserError) {
	varName, frame, err := parseWord(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing for loop variable")
	}
	ok, frame := parseExact("in", skipWhiteSpaces(frame))
	if !ok {
		return nil, frame, newParserError(frame, "\"in\" expected")
	}
	start, frame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing for loop start index")
	}
	ok, frame = parseExact("..", frame)
	if !ok {
		return nil, frame, newParserError(frame, ".. expected")
	}
	end, frame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing for loop end index")
	}
	ok, frame = parseExact("{", frame)
	if !ok {
		return nil, frame, newParserError(frame, "{ expected")
	}
	body, frame, err := parseCommand(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing for loop body")
	}
	ok, frame = parseExact("}", skipWhiteSpaces(frame))
	if !ok {
		return nil, frame, newParserError(frame, "} expected")
	}
	return &forRangeNode{varName, start, end, body}, frame, nil
}

func parseIf(frame Frame) (node, Frame, *ParserError) {
	condition, frame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, err
	}
	ok, frame := parseExact("{", frame)
	if !ok {
		return nil, frame, newParserError(frame, "{ expected")
	}
	body, frame, err := parseCommand(frame)
	if err != nil {
		return nil, frame, err.extend(frame, "parsing if body")
	}
	ok, frame = parseExact("}", frame)
	if !ok {
		return nil, frame, newParserError(frame, "} expected")
	}
	keyword, frame := parseKeyWord([]string{"else", "elif"}, skipWhiteSpaces(frame))
	switch keyword {
	case "":
		return &ifNode{condition, body, nil}, frame, nil
	case "else":
		ok, frame := parseExact("{", skipWhiteSpaces(frame))
		if !ok {
			return nil, frame, newParserError(frame, "{ expected")
		}
		elseBody, frame, err := parseCommand(frame)
		if err != nil {
			return nil, frame, err.extend(frame, "parsing else body")
		}
		ok, frame = parseExact("}", skipWhiteSpaces(frame))
		if !ok {
			return nil, frame, newParserError(frame, "} expected")
		}
		return &ifNode{condition, body, elseBody}, frame, nil
	case "elif":
		frame := frame.new(frame.start-2, frame.end)
		elseBody, frame, err := parseIf(frame)
		if err != nil {
			return nil, frame, err.extend(frame, "parsing elif body")
		}
		return &ifNode{condition, body, elseBody}, frame, nil
	default:
		panic("unexpected return from parseKeyWord : " + keyword)
	}
}

func parseAlias(frame Frame) (node, Frame, *ParserError) {
	name, frame, err := parseWord(frame)
	if err != nil {
		return nil, frame, err
	}
	frame = skipWhiteSpaces(frame)
	ok, frame := parseExact("{", frame)
	if !ok {
		return nil, frame, newParserError(frame, "{ expected}")
	}
	frame = skipWhiteSpaces(frame)
	command, frame, err := parseCommand(frame)
	if err != nil {
		return nil, frame, err.extend(frame, "parsing alias body")
	}
	frame = skipWhiteSpaces(frame)
	ok, frame = parseExact("}", frame)
	if !ok {
		return nil, frame, newParserError(frame, "} expected")
	}
	return &funcDefNode{name, command}, frame, nil
}

func parseCallAlias(frame Frame) (node, Frame, *ParserError) {
	name, frame, err := parseWord(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing alias call")
	}
	return &funcCallNode{name}, frame, nil
}

func parseObjType(frame Frame) (string, Frame) {
	candidates := []string{}
	for command := range createObjDispatch {
		candidates = append(candidates, command)
	}
	return parseKeyWord(candidates, frame)
}

func parseCreate(frame Frame) (node, Frame, *ParserError) {
	objType, frame := parseObjType(frame)
	if objType == "" {
		return nil, frame, newParserError(frame, "parsing object type")
	}
	frame = skipWhiteSpaces(frame)
	if objType == "orphan" {
		return parseCreateOrphan(frame)
	}
	if frame.first() != ':' {
		return nil, frame, newParserError(frame, ": expected")
	}
	frame = skipWhiteSpaces(frame.forward(1))
	return createObjDispatch[objType](frame)
}

func parseColor(frame Frame) (node, Frame, *ParserError) {
	l := lexerFromFrame(frame)
	tok := l.nextToken(lexColor)
	if tok.t == tokColor {
		return &strLeaf{tok.str}, frame.from(tok.end), nil
	}
	color, newFrame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, newParserError(frame, "color expected")
	}
	return color, newFrame, nil
}

func parseKeyWordOrExpr(keywords []string, frame Frame) (node, Frame, *ParserError) {
	keyword, newFrame := parseKeyWord(keywords, frame)
	if keyword != "" {
		return &strLeaf{keyword}, newFrame, nil
	}
	expr, newFrame, err := parseExpr(frame)
	if err != nil {
		return nil, frame, newParserError(frame, "keyword or expr expected")
	}
	return expr, newFrame, nil
}

func parseRackOrientation(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"front", "rear", "left", "right"}, frame)
}

func parseAxisOrientation(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"+x+y", "+x-y", "-x-y", "-x+y"}, frame)
}

func parseFloorUnit(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"t", "m", "f"}, frame)
}

func parseSide(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"front", "rear", "frontflipped", "rearflipped"}, frame)
}

func parseTemperature(frame Frame) (node, Frame, *ParserError) {
	return parseKeyWordOrExpr([]string{"cold", "warm"}, frame)
}

func parseStringExpr(frame Frame) (node, Frame, *ParserError) {
	expr, nextFrame, err := parseExpr(frame)
	if err == nil && exprEnd(nextFrame) {
		return expr, nextFrame, nil
	}
	frame = skipWhiteSpaces(frame)
	str, frame, err := parseRawText(lexUnquotedString, frame)
	if err != nil {
		return nil, frame, err.extend(frame, "parsing string expression")
	}
	return str, skipWhiteSpaces(frame), nil
}

type objParam struct {
	name string
	t    string
}

func parseObjectParams(sig []objParam, frame Frame) (map[string]node, Frame, *ParserError) {
	values := map[string]node{}
	for i, param := range sig {
		if i != 0 {
			ok, nextFrame := parseExact("@", frame)
			if !ok {
				return nil, frame, newParserError(frame, "@ expected")
			}
			frame = nextFrame
		}
		var value node
		var err *ParserError
		switch param.t {
		case "path":
			value, frame, err = parsePath(frame)
		case "expr":
			value, frame, err = parseExpr(frame)
		case "stringexpr":
			value, frame, err = parseStringExpr(frame)
		case "axisOrientation":
			value, frame, err = parseAxisOrientation(frame)
		case "rackOrientation":
			value, frame, err = parseRackOrientation(frame)
		case "floorUnit":
			value, frame, err = parseFloorUnit(frame)
		case "side":
			value, frame, err = parseSide(frame)
		case "color":
			value, frame, err = parseColor(frame)
		}
		if err != nil {
			return nil, frame, err.extend(frame, "parsing "+param.name)
		}
		values[param.name] = value
	}
	return values, frame, nil
}

func parseCreateTenant(frame Frame) (node, Frame, *ParserError) {
	sig := []objParam{{"path", "path"}, {"color", "color"}}
	params, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing tenant parameters")
	}
	return &createTenantNode{params["path"], params["color"]}, frame, nil
}

func parseCreateSite(frame Frame) (node, Frame, *ParserError) {
	sig := []objParam{{"path", "path"}}
	params, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing site parameters")
	}
	return &createSiteNode{params["path"]}, frame, nil
}

func parseCreateBuilding(frame Frame) (node, Frame, *ParserError) {
	sig := []objParam{{"path", "path"}, {"posXY", "expr"}, {"rotation", "expr"}, {"sizeOrTemplate", "stringexpr"}}
	params, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing building parameters")
	}
	return &createBuildingNode{params["path"], params["posXY"], params["rotation"], params["sizeOrTemplate"]}, frame, nil
}

func parseCreateRoom(frame Frame) (node, Frame, *ParserError) {
	sig := []objParam{{"path", "path"}, {"posXY", "expr"}, {"rotation", "expr"}, {"sizeOrTemplate", "stringexpr"}}
	params1, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing room parameters")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return &createRoomNode{
			params1["path"],
			params1["posXY"],
			params1["rotation"],
			nil, nil, nil,
			params1["sizeOrTemplate"]}, frame, nil
	}
	sig = []objParam{{"axisOrientation", "axisOrientation"}}
	params2, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing room parameters")
	}
	ok, frame = parseExact("@", frame)
	if !ok {
		return &createRoomNode{
			params1["path"],
			params1["posXY"],
			params1["rotation"],
			params1["sizeOrTemplate"],
			params2["axisOrientation"], nil, nil}, frame, nil
	}
	sig = []objParam{{"floorUnit", "floorUnit"}}
	params3, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing room parameters")
	}
	return &createRoomNode{
		params1["path"],
		params1["posXY"],
		params1["rotation"],
		params1["sizeOrTemplate"],
		params2["axisOrientation"],
		params3["floorUnit"], nil}, frame, nil
}

func parseCreateRack(frame Frame) (node, Frame, *ParserError) {
	sig := []objParam{{"path", "path"}, {"pos", "expr"},
		{"sizeOrTemplate", "stringexpr"}, {"orientation", "rackOrientation"}}
	params, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing rack parameters")
	}
	return &createRackNode{params["path"], params["pos"], params["sizeOrTemplate"], params["orientation"]}, frame, nil
}

func parseCreateDevice(frame Frame) (node, Frame, *ParserError) {
	sig := []objParam{{"path", "path"}, {"posUOrSlot", "stringexpr"}, {"sizeUOrTemplate", "stringexpr"}}
	params1, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing device parameters")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return &createDeviceNode{params1["path"], params1["posUOrSlot"], params1["sizeUOrTemplate"], nil}, frame, nil
	}
	sig = []objParam{{"side", "side"}}
	params2, frame, err := parseObjectParams(sig, frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing device parameters")
	}
	return &createDeviceNode{params1["path"], params1["posUOrSlot"], params1["sizeUOrTemplate"], params2["side"]}, frame, nil
}

func parseCreateGroup(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing group physical path")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return nil, frame, newParserError(frame, "@ expected")
	}
	childs, frame, err := parsePathGroup(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing group childs")
	}
	return &createGroupNode{path, childs}, frame, nil
}

func parseCreateCorridor(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing group physical path")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return nil, frame, newParserError(frame, "@ expected")
	}
	racks, frame, err := parsePathGroup(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing group childs")
	}
	if len(racks) != 2 {
		return nil, frame, newParserError(frame, "only 2 racks expected")
	}
	ok, frame = parseExact("@", frame)
	if !ok {
		return nil, frame, newParserError(frame, "@ expected")
	}
	temperature, frame, err := parseTemperature(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing corridor temperature")
	}
	return &createCorridorNode{path, racks[0], racks[1], temperature}, frame, nil
}

func parseCreateOrphan(frame Frame) (node, Frame, *ParserError) {
	for _, word := range []string{"device", "dv"} {
		ok, newFrame := parseExact(word, frame)
		if ok {
			return parseCreateOrphanAux(skipWhiteSpaces(newFrame), false)
		}
	}
	for _, word := range []string{"sensor", "sr"} {
		ok, newFrame := parseExact(word, frame)
		if ok {
			return parseCreateOrphanAux(skipWhiteSpaces(newFrame), true)
		}
	}
	return nil, frame, newParserError(frame, "device or sensor keyword expected")
}

func parseCreateOrphanAux(frame Frame, sensor bool) (node, Frame, *ParserError) {
	if frame.first() != ':' {
		return nil, frame, newParserError(frame, ": expected")
	}
	frame = skipWhiteSpaces(frame.forward(1))
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing orphan physical path")
	}
	ok, frame := parseExact("@", frame)
	if !ok {
		return nil, frame, newParserError(frame, "@ expected")
	}
	template, frame, err := parseStringExpr(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing orphan template")
	}
	return &createOrphanNode{path, template, sensor}, frame, nil
}

func parseUpdate(frame Frame) (node, Frame, *ParserError) {
	path, frame, err := parsePath(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing update")
	}
	frame = skipWhiteSpaces(frame)
	ok, frame := parseExact(":", frame)
	if !ok {
		return nil, frame, newParserError(frame, ": expected")
	}
	attr, frame, err := parseAssign(frame)
	if err != nil {
		return nil, frame, err.extendMessage("parsing update")
	}
	frame = skipWhiteSpaces(frame)
	sharpe, frame := parseExact("#", frame)
	values := []node{}
	moreValues := true
	for moreValues {
		var val node
		val, frame, err = parseStringExpr(frame)
		if err != nil {
			return nil, frame, err.extend(frame, "parsing update new value")
		}
		values = append(values, val)
		moreValues, frame = parseExact("@", frame)
	}
	return &updateObjNode{path, attr, values, sharpe}, frame, nil
}

func parseCommandKeyWord(frame Frame) (string, Frame) {
	candidates := []string{}
	for command := range commandDispatch {
		candidates = append(candidates, command)
	}
	for command := range noArgsCommands {
		candidates = append(candidates, command)
	}
	candidates = append(candidates, lsCommands...)
	return parseKeyWord(candidates, frame)
}

func parseSingleCommand(frame Frame) (node, Frame, *ParserError) {
	frame = skipWhiteSpaces(frame)
	if commandEnd(frame) {
		return nil, frame, nil
	}
	startFrame := frame
	commandKeyWord, frame := parseCommandKeyWord(frame)
	if commandKeyWord != "" {
		if lsIdx := indexOf(lsCommands, commandKeyWord); lsIdx != -1 {
			return parseLsObj(lsIdx, skipWhiteSpaces(frame))
		}
		parseFunc, ok := commandDispatch[commandKeyWord]
		if ok {
			return parseFunc(skipWhiteSpaces(frame))
		}
		result, ok := noArgsCommands[commandKeyWord]
		if ok {
			return result, frame, nil
		}
	}
	_, frame, err := parsePath(frame)
	ok, _ := parseExact(":", frame)
	if err == nil && ok {
		return parseUpdate(startFrame)
	}
	return parseCallAlias(startFrame)
}

func parseCommand(frame Frame) (node, Frame, *ParserError) {
	if commandDispatch == nil {
		commandDispatch = map[string]parseCommandFunc{
			"ls":         parseLs,
			"get":        parseGet,
			"getu":       parseGetU,
			"getslot":    parseGetSlot,
			"undraw":     parseUndraw,
			"draw":       parseDraw,
			"drawable":   parseDrawable,
			"hc":         parseHc,
			"unset":      parseUnset,
			"env":        parseEnv,
			"+":          parseCreate,
			"-":          parseDelete,
			"=":          parseEqual,
			".var:":      parseVar,
			".cmds:":     parseLoad,
			".template:": parseTemplate,
			"len":        parseLen,
			"link:":      parseLink,
			"unlink":     parseUnlink,
			"print":      parsePrint,
			"man":        parseMan,
			"cd":         parseCd,
			"tree":       parseTree,
			"ui.":        parseUi,
			"camera.":    parseCamera,
			">":          parseFocus,
			"while":      parseWhile,
			"for":        parseFor,
			"if":         parseIf,
			"alias":      parseAlias,
		}
		createObjDispatch = map[string]parseCommandFunc{
			"tenant":   parseCreateTenant,
			"tn":       parseCreateTenant,
			"site":     parseCreateSite,
			"si":       parseCreateSite,
			"bldg":     parseCreateBuilding,
			"building": parseCreateBuilding,
			"bd":       parseCreateBuilding,
			"room":     parseCreateRoom,
			"ro":       parseCreateRoom,
			"rack":     parseCreateRack,
			"rk":       parseCreateRack,
			"device":   parseCreateDevice,
			"dv":       parseCreateDevice,
			"corridor": parseCreateCorridor,
			"co":       parseCreateCorridor,
			"group":    parseCreateGroup,
			"gr":       parseCreateGroup,
			"orphan":   parseCreateOrphan,
		}
		noArgsCommands = map[string]node{
			"selection":    &selectNode{},
			"clear":        &clrNode{},
			"grep":         &grepNode{},
			"lsog":         &lsogNode{},
			"lsenterprise": &lsenterpriseNode{},
			"pwd":          &pwdNode{},
			"exit":         &exitNode{},
		}
	}
	commands := []node{}
	var command node
	var err *ParserError
	var ok bool
	for {
		command, frame, err = parseSingleCommand(frame)
		if err != nil {
			return nil, frame, err.extend(frame, "parsing command")
		}
		commands = append(commands, command)
		frame = skipWhiteSpaces(frame)
		ok, frame = parseExact(";", frame)
		if !ok {
			if len(commands) > 1 {
				return &ast{commands}, frame, nil
			}
			return command, frame, nil
		}
		frame = skipWhiteSpaces(frame)
	}
}

func Parse(buffer string) (node, *ParserError) {
	commentIdx := strings.Index(buffer, "//")
	if commentIdx != -1 {
		buffer = buffer[:commentIdx]
	}
	frame := newFrame(buffer)
	node, frame, err := parseCommand(frame)
	if err != nil {
		return nil, err
	}
	if frame.start != frame.end {
		return nil, newParserError(frame, "unexpected characters")
	}
	return node, err
}
