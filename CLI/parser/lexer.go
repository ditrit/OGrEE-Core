package parser

import (
	"strconv"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokEOF         tokenType = iota
	tokDeref                 // variable dereferenciation
	tokInt                   // integer constant
	tokFloat                 // float constant
	tokBool                  // boolean constant
	tokDoubleQuote           // '"'
	tokLeftBrac              // '['
	tokRightBrac             // ']'
	tokComma                 // ','
	tokSemiCol               // ';'
	tokAt                    // '@'
	tokLeftParen             // '('
	tokRightParen            // ')'
	tokNot                   // '!'
	tokAdd                   // '+'
	tokSub                   // '-'
	tokMul                   // '*'
	tokDiv                   // '/'
	tokIntDiv                // '\'
	tokMod                   // '%'
	tokOr                    // '||'
	tokAnd                   // '&&'
	tokEq                    // '=='
	tokNeq                   // '!='
	tokLeq                   // '<=
	tokGeq                   // '>='
	tokGtr                   // '>'
	tokLss                   // '<'
	tokColor
	tokText
	tokLeftEval // '$(('
	tokFormat   // 'format'
)

func (s tokenType) String() string {
	return map[tokenType]string{
		tokEOF:         "eof",
		tokDeref:       "deref",
		tokInt:         "int",
		tokFloat:       "float",
		tokBool:        "bool",
		tokDoubleQuote: "doublQuote",
		tokLeftBrac:    "leftBrac",
		tokRightBrac:   "rightBrac",
		tokComma:       "comma",
		tokSemiCol:     "semicol",
		tokAt:          "at",
		tokLeftParen:   "leftParen",
		tokRightParen:  "rightParen",
		tokNot:         "not",
		tokAdd:         "add",
		tokSub:         "sub",
		tokMul:         "mul",
		tokDiv:         "div",
		tokIntDiv:      "intdiv",
		tokMod:         "mod",
		tokOr:          "or",
		tokAnd:         "and",
		tokEq:          "eq",
		tokNeq:         "neq",
		tokLeq:         "leq",
		tokGeq:         "geq",
		tokGtr:         "gtr",
		tokLss:         "lss",
		tokColor:       "color",
		tokText:        "text",
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
	case tokMul, tokDiv, tokIntDiv, tokMod:
		return 5
	case tokNot:
		return 6
	}
	return 0
}

const eof = 0

func (p *parser) emit(t tokenType, val interface{}) token {
	str := p.item(false)
	if t == tokEOF {
		str = "eof"
	}
	p.tok = token{
		t:     t,
		start: p.startCursor,
		end:   p.cursor,
		str:   str,
		val:   val,
	}
	return p.tok
}

func (p *parser) acceptRun(valid string) {
	for strings.Contains(valid, string(p.next())) {
	}
	p.backward(1)
}

func (p *parser) acceptRunAlphaNumeric() {
	for isAlphaNumeric(p.next()) {
	}
	p.backward(1)
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

func (p *parser) parseExprToken() token {
	defer un(trace(p, ""))
	c := p.next()
	switch c {
	case ' ', '\t':
		return p.parseExprToken()
	case '$':
		return p.lexDeref()
	case '"':
		return p.emit(tokDoubleQuote, nil)
	case '[':
		return p.emit(tokLeftBrac, nil)
	case ']':
		return p.emit(tokRightBrac, nil)
	case ',':
		return p.emit(tokComma, nil)
	case '(':
		return p.emit(tokLeftParen, nil)
	case ')':
		return p.emit(tokRightParen, nil)
	case '+':
		return p.emit(tokAdd, nil)
	case '-':
		return p.emit(tokSub, nil)
	case '*':
		return p.emit(tokMul, nil)
	case '/':
		return p.emit(tokDiv, nil)
	case '\\':
		return p.emit(tokIntDiv, nil)
	case '%':
		return p.emit(tokMod, nil)
	case '|':
		if p.next() != '|' {
			p.backward(2)
			return p.emit(tokEOF, nil)
		}
		return p.emit(tokOr, nil)
	case '&':
		if p.next() != '&' {
			p.backward(2)
			return p.emit(tokEOF, nil)
		}
		return p.emit(tokAnd, nil)
	case '=':
		if p.next() != '=' {
			p.backward(2)
			return p.emit(tokEOF, nil)
		}
		return p.emit(tokEq, nil)
	case '!':
		if p.next() == '=' {
			return p.emit(tokNeq, nil)
		}
		p.backward(1)
		return p.emit(tokNot, nil)
	case '<':
		if p.next() == '=' {
			return p.emit(tokLeq, nil)
		}
		p.backward(1)
		return p.emit(tokLss, nil)
	case '>':
		c = p.next()
		if c == '=' {
			return p.emit(tokGeq, nil)
		}
		p.backward(1)
		return p.emit(tokGtr, nil)
	}
	if isDigit(c) {
		return p.lexNumber()
	}
	if c == '.' {
		p.backward(1)
		return p.lexNumber()
	}
	p.backward(1)
	if p.parseExact("true") {
		return p.emit(tokBool, true)
	}
	if p.parseExact("false") {
		return p.emit(tokBool, false)
	}
	if p.parseExact("format") {
		return p.emit(tokFormat, nil)
	}
	return p.emit(tokEOF, nil)
}

func (p *parser) lexDeref() token {
	if p.parseExact("{") {
		return p.lexDerefBracket()
	}
	if p.parseExact("((") {
		return p.emit(tokLeftEval, nil)
	}
	if !isAlphaNumeric(p.next()) {
		p.backward(1)
		p.error("identifier expected")
	}
	p.acceptRunAlphaNumeric()
	return p.emit(tokDeref, p.item(false)[1:])
}

func (p *parser) lexDerefBracket() token {
	for isSpace(p.next()) {
	}
	p.backward(1)
	for isAlphaNumeric(p.next()) {
	}
	p.backward(1)
	for isSpace(p.next()) {
	}
	p.backward(1)
	if !p.parseExact("}") {
		p.error("} expected")
	}
	s := p.item(false)
	return p.emit(tokDeref, s[2:len(s)-1])
}

func (p *parser) endNumber(t tokenType, val any) token {
	c := p.peek()
	if isLetter(c) {
		return p.lexUnquotedString()
	}
	return p.emit(t, val)
}

func (p *parser) lexNumber() token {
	digits := "0123456789_"
	p.acceptRun(digits)
	if p.parseExact(".") {
		if p.parseExact(".") {
			p.backward(2)
			if p.item(false) == "" {
				return p.emit(tokEOF, nil)
			}
			val, err := strconv.Atoi(p.item(false))
			if err != nil {
				panic("cannot convert string " + p.item(false) + " to int")
			}
			return p.endNumber(tokInt, val)
		}
		p.acceptRun(digits)
		val, err := strconv.ParseFloat(p.item(false), 64)
		if err != nil {
			p.error("invalid float")
		}
		return p.endNumber(tokFloat, val)
	}
	val, err := strconv.Atoi(p.item(false))
	if err != nil {
		p.error("invalid integer")
	}
	return p.endNumber(tokInt, val)
}

func (p *parser) lexText(endCharacters string) token {
	c := p.next()
	if c == '$' {
		return p.lexDeref()
	}
	for {
		if c == eof || strings.Contains(endCharacters+"$", string(c)) {
			p.backward(1)
			if p.cursor == p.startCursor {
				return p.emit(tokEOF, nil)
			}
			return p.emit(tokText, nil)
		}
		c = p.next()
	}
}

func (p *parser) lexUnquotedString() token {
	return p.lexText("@;,})")
}

func (p *parser) parseUnquotedStringToken() token {
	defer un(trace(p, ""))
	return p.lexUnquotedString()
}

func (p *parser) lexQuotedString() token {
	return p.lexText("\"")
}

func (p *parser) parseQuotedStringToken() token {
	defer un(trace(p, ""))
	return p.lexQuotedString()
}

func (p *parser) lexPath() token {
	return p.lexText(" @;,}):")
}

func (p *parser) parsePathToken() token {
	defer un(trace(p, ""))
	return p.lexPath()
}
