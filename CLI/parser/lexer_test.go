package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenTypeString(t *testing.T) {
	tokenStrings := map[tokenType]string{
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
	}

	for key, value := range tokenStrings {
		assert.Equal(t, value, key.String())
	}
}

func TestTokenTypePrecedence(t *testing.T) {
	precedenceMap := map[int][]tokenType{
		1: []tokenType{tokOr},
		2: []tokenType{tokAnd},
		3: []tokenType{tokEq, tokNeq, tokLss, tokLeq, tokGtr, tokGeq},
		4: []tokenType{tokAdd, tokSub},
		5: []tokenType{tokMul, tokDiv, tokIntDiv, tokMod},
		6: []tokenType{tokNot},
		0: []tokenType{tokEOF},
	}

	for key, value := range precedenceMap {
		for _, tokType := range value {
			tok := token{
				t: tokType,
			}
			assert.Equal(t, key, tok.precedence())
		}
	}
}

func checkTokSequence(lexFunc func() token, expectedTypes []tokenType, expectedVals []any, t *testing.T) {
	for i := 0; i < len(expectedTypes); i++ {
		tok := lexFunc()
		if tok.t != expectedTypes[i] {
			t.Errorf("Unexpected token : %s when %s was expected", tok.t.String(), expectedTypes[i].String())
		}
		if tok.val != expectedVals[i] {
			t.Errorf("Unexpected token value : %v when %v was expected", tok.val, expectedVals[i])
		}
	}
}

func TestLex(t *testing.T) {
	p := newParser("42 + (3 - 4) * + \" || false")
	expectedTypes := []tokenType{tokInt, tokAdd, tokLeftParen, tokInt, tokSub, tokInt, tokRightParen,
		tokMul, tokAdd, tokDoubleQuote, tokOr, tokBool, tokEOF}
	expectedVals := []any{42, nil, nil, 3, nil, 4, nil, nil, nil, nil, nil, false, nil}
	checkTokSequence(p.parseExprToken, expectedTypes, expectedVals, t)
}

func TestLexDoubleDot(t *testing.T) {
	p := newParser("42..")
	expectedTypes := []tokenType{tokInt, tokEOF}
	expectedVals := []any{42, nil}
	checkTokSequence(p.parseExprToken, expectedTypes, expectedVals, t)
}

func TestLexFormattedString(t *testing.T) {
	p := newParser("${a}a$ab")
	expectedTypes := []tokenType{tokDeref, tokText, tokDeref, tokEOF}
	expectedVals := []any{"a", nil, "ab", nil}
	checkTokSequence(p.parseUnquotedStringToken, expectedTypes, expectedVals, t)
}

func TestParserParseExprToken(t *testing.T) {
	p := newParser("\"[],()+-*/\\%")
	simpleTokens := []tokenType{tokDoubleQuote, tokLeftBrac, tokRightBrac, tokComma, tokLeftParen, tokRightParen, tokAdd, tokSub, tokMul, tokDiv, tokIntDiv, tokMod}
	for _, value := range simpleTokens {
		tok := p.parseExprToken()
		assert.Equal(t, value, tok.t)
	}

	// other cases
	tokensMap := map[string]tokenType{
		"${var}": tokDeref,
		"$((":    tokLeftEval,
		"||":     tokOr,
		"|":      tokEOF,
		"&":      tokEOF,
		"&&":     tokAnd,
		"=":      tokEOF,
		"==":     tokEq,
		"!=":     tokNeq,
		"!":      tokNot,
		"<":      tokLss,
		"<=":     tokLeq,
		">":      tokGtr,
		">=":     tokGeq,
		"10":     tokInt,
		"1.0":    tokFloat,
		"true":   tokBool,
		"format": tokFormat,
		"abc":    tokEOF,
	}
	for key, value := range tokensMap {
		p = newParser(key)
		tok := p.parseExprToken()
		assert.Equal(t, value, tok.t)
	}
}
