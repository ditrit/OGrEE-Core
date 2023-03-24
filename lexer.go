package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokEOF tokenType = iota
	tokError
	tokWord       // identifier
	tokDeref      // variable dereferenciation
	tokInt        // integer constant
	tokFloat      // float constant
	tokBool       // boolean constant
	tokString     // quoted string
	tokLeftBrac   // '['
	tokRightBrac  // ']'
	tokComma      // ','
	tokSemiCol    // ';'
	tokAt         // '@'
	tokLeftParen  // '('
	tokRightParen // ')'
	tokNot        // '!'
	tokAdd        // '+'
	tokSub        // '-'
	tokMul        // '*'
	tokDiv        // '/'
	tokMod        // '%'
	tokOr         // '||'
	tokAnd        // '&&'
	tokEq         // '=='
	tokNeq        // '!='
	tokLeq        // '<=
	tokGeq        // '>='
	tokGtr        // '>'
	tokLss        // '<'
	tokColor
	tokText
)

func (s tokenType) String() string {
	return map[tokenType]string{
		tokEOF:        "eof",
		tokError:      "error",
		tokWord:       "word",
		tokDeref:      "deref",
		tokInt:        "int",
		tokFloat:      "float",
		tokBool:       "bool",
		tokString:     "string",
		tokLeftBrac:   "leftBrac",
		tokRightBrac:  "rightBrac",
		tokComma:      "comma",
		tokSemiCol:    "semicol",
		tokAt:         "at",
		tokLeftParen:  "leftParen",
		tokRightParen: "rightParen",
		tokNot:        "not",
		tokAdd:        "add",
		tokSub:        "sub",
		tokMul:        "mul",
		tokDiv:        "div",
		tokMod:        "mod",
		tokOr:         "or",
		tokAnd:        "and",
		tokEq:         "eq",
		tokNeq:        "neq",
		tokLeq:        "leq",
		tokGeq:        "geq",
		tokGtr:        "gtr",
		tokLss:        "lss",
		tokColor:      "color",
		tokText:       "text",
	}[s]
}

type token struct {
	t     tokenType
	start int
	end   int
	str   string
	val   any
}

func (t token) precedence() int {
	switch t.t {
	case tokOr:
		return 1
	case tokAnd:
		return 2
	case tokEq, tokNeq, tokLss, tokLeq, tokGtr, tokGeq:
		return 3
	case tokAdd, tokSub:
		return 4
	case tokMul, tokDiv, tokMod:
		return 5
	case tokNot:
		return 6
	}
	return 0
}

const eof = 0

type lexer struct {
	input string
	pos   int
	start int
	end   int
	tok   token
	atEOF bool
}

type stateFn func(*lexer) stateFn

func (l *lexer) emit(t tokenType, val interface{}) stateFn {
	str := l.input[l.start:l.pos]
	if t == tokEOF {
		str = "eof"
	}
	l.tok = token{
		t:     t,
		start: l.start,
		end:   l.pos,
		str:   str,
		val:   val,
	}
	l.start = l.pos
	return nil
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	val := fmt.Sprintf(format, args...)
	l.tok = token{
		t:     tokError,
		start: l.start,
		str:   val,
		val:   val,
	}
	//l.start = 0
	//l.pos = 0
	//l.input = l.input[:0]
	return nil
}

func (l *lexer) next() byte {
	if l.pos >= l.end {
		l.atEOF = true
		return eof
	}
	char := l.input[l.pos]
	l.pos++
	return char
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	if l.pos > 0 && !l.atEOF {
		l.pos--
	}
}

func (l *lexer) accept(valid string) bool {
	if strings.Contains(valid, string(l.next())) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.Contains(valid, string(l.next())) {
	}
	l.backup()
}

func (l *lexer) acceptRunAlphaNumeric() {
	for isAlphaNumeric(l.next()) {
	}
	l.backup()
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

func isLetter(c byte) bool {
	return c == '_' || unicode.IsLetter(rune(c))
}

func isDigit(c byte) bool {
	return unicode.IsDigit(rune(c))
}

func isAlphaNumeric(c byte) bool {
	return isLetter(c) || isDigit(c)
}

func lexExpr(l *lexer) stateFn {
	c := l.next()
	switch c {
	case ' ', '\t':
		l.ignore()
		return lexExpr
	case '$':
		return lexDeref
	case '"':
		return lexString
	case '[':
		return l.emit(tokLeftBrac, nil)
	case ']':
		return l.emit(tokRightBrac, nil)
	case ',':
		return l.emit(tokComma, nil)
	case '(':
		return l.emit(tokLeftParen, nil)
	case ')':
		return l.emit(tokRightParen, nil)
	case '+':
		return l.emit(tokAdd, nil)
	case '-':
		return l.emit(tokSub, nil)
	case '*':
		return l.emit(tokMul, nil)
	case '/':
		return l.emit(tokDiv, nil)
	case '%':
		return l.emit(tokMod, nil)
	case '|':
		if l.next() != '|' {
			return l.errorf("| expected")
		}
		return l.emit(tokOr, nil)
	case '&':
		if l.next() != '&' {
			return l.errorf("& expected")
		}
		return l.emit(tokAnd, nil)
	case '=':
		if l.next() != '=' {
			return l.errorf("= expected")
		}
		return l.emit(tokEq, nil)
	case '!':
		if l.next() == '=' {
			return l.emit(tokNeq, nil)
		}
		l.backup()
		return l.emit(tokNot, nil)
	case '<':
		if l.next() == '=' {
			return l.emit(tokLeq, nil)
		}
		l.backup()
		return l.emit(tokLss, nil)
	case '>':
		c = l.next()
		if c == '=' {
			return l.emit(tokGeq, nil)
		}
		l.backup()
		return l.emit(tokGtr, nil)
	}
	if isDigit(c) {
		return lexNumber
	}
	if c == '.' {
		if l.accept(".") {
			return l.emit(tokEOF, nil)
		}
		l.backup()
		return lexNumber
	}
	if isLetter(c) {
		return lexAlphaNumeric
	}
	return l.emit(tokEOF, nil)
}

func lexDeref(l *lexer) stateFn {
	if l.accept("{") {
		return lexDerefBracket
	}
	if !isAlphaNumeric(l.next()) {
		l.backup()
		return l.errorf("identifier expected")
	}
	l.acceptRunAlphaNumeric()
	return l.emit(tokDeref, l.input[l.start+1:l.pos])
}

func lexDerefBracket(l *lexer) stateFn {
	for isSpace(l.next()) {
	}
	l.backup()
	for isAlphaNumeric(l.next()) {
	}
	l.backup()
	for isSpace(l.next()) {
	}
	l.backup()
	if !l.accept("}") {
		return l.errorf("} expected")
	}
	return l.emit(tokDeref, l.input[l.start+2:l.pos-1])
}

func lexString(l *lexer) stateFn {
	for {
		switch l.next() {
		case eof:
			return l.errorf("unterminated string")
		case '"':
			return l.emit(tokString, l.input[l.start+1:l.pos-1])
		}
	}
}

func (l *lexer) endNumber(t tokenType, val any) stateFn {
	c := l.next()
	l.backup()
	if isLetter(c) {
		return lexAlphaNumeric
	}
	return l.emit(t, val)
}

func lexNumber(l *lexer) stateFn {
	digits := "0123456789_"
	l.acceptRun(digits)
	isFloat := false
	if l.accept(".") {
		if l.accept(".") {
			l.backup()
			l.backup()
			val, _ := strconv.Atoi(l.input[l.start:l.pos])
			return l.endNumber(tokInt, val)
		}
		isFloat = true
		l.acceptRun(digits)
	}
	if isFloat {
		val, _ := strconv.ParseFloat(l.input[l.start:l.pos], 64)
		return l.endNumber(tokFloat, val)
	}
	val, _ := strconv.Atoi(l.input[l.start:l.pos])
	return l.endNumber(tokInt, val)
}

func lexAlphaNumeric(l *lexer) stateFn {
	for {
		c := l.next()
		if isAlphaNumeric(c) {
			continue
		}
		l.backup()
		word := l.input[l.start:l.pos]
		if word == "true" {
			return l.emit(tokBool, true)
		}
		if word == "false" {
			return l.emit(tokBool, false)
		}
		return l.emit(tokWord, nil)
	}
}

func lexColor(l *lexer) stateFn {
	for i := 0; i < 6; i++ {
		if !l.accept("0123456789ABCDEFabcdef") {
			return l.errorf("invalid character in color")
		}
	}
	return l.emit(tokColor, nil)
}

func lexText(l *lexer, endCharacters string, caller stateFn) stateFn {
	c := l.next()
	if c == '$' {
		return lexDeref
	}
	for {
		if c == eof || strings.Contains(endCharacters+"$", string(c)) {
			l.backup()
			if l.pos == l.start {
				return l.emit(tokEOF, nil)
			}
			return l.emit(tokText, nil)
		}
		c = l.next()
	}
}

func lexUnquotedString(l *lexer) stateFn {
	return lexText(l, " @;,})", lexUnquotedString)
}

func lexQuotedString(l *lexer) stateFn {
	return lexText(l, "", lexQuotedString)
}

func lexPath(l *lexer) stateFn {
	return lexText(l, " @;,}):", lexPath)
}

func (l *lexer) nextToken(state stateFn) token {
	if l.next() == eof {
		state = l.emit(tokEOF, nil)
	} else {
		l.backup()
	}
	for {
		if state == nil {
			return l.tok
		}
		state = state(l)
	}
}

func newLexer(input string, start int, end int) *lexer {
	return &lexer{input: input, start: start, end: end, pos: start}
}
