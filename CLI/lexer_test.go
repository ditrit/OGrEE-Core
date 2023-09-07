package main

import (
	"testing"
)

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
