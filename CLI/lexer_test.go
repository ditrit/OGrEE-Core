package main

import (
	"testing"
)

func checkTokSequence(lexFunc stateFn, expectedTypes []tokenType, expectedVals []any, str string, t *testing.T) {
	l := newLexer(str, 0, len(str))
	for i := 0; i < len(expectedTypes); i++ {
		tok := l.nextToken(lexFunc)
		if tok.t != expectedTypes[i] {
			t.Errorf("Unexpected token : %s when %s was expected", tok.t.String(), expectedTypes[i].String())
		}
		if tok.val != expectedVals[i] {
			t.Errorf("Unexpected token value : %v when %v was expected", tok.val, expectedVals[i])
		}
	}
}

func TestLex(t *testing.T) {
	str := "false42 + (3 - 4) * plouf42 + \"plouf\" || false"
	expectedTypes := []tokenType{tokWord, tokAdd, tokLeftParen, tokInt, tokSub, tokInt, tokRightParen,
		tokMul, tokWord, tokAdd, tokString, tokOr, tokBool, tokEOF}
	expectedVals := []any{nil, nil, nil, 3, nil, 4, nil, nil, nil, nil, "plouf", nil, false, nil}
	checkTokSequence(lexExpr, expectedTypes, expectedVals, str, t)
}

func TestLexDoubleDot(t *testing.T) {
	str := "42.."
	expectedTypes := []tokenType{tokInt, tokEOF}
	expectedVals := []any{42, nil}
	checkTokSequence(lexExpr, expectedTypes, expectedVals, str, t)
}

func TestLexFormattedString(t *testing.T) {
	str := "${a}a$ab"
	expectedTypes := []tokenType{tokDeref, tokText, tokDeref, tokEOF}
	expectedVals := []any{"a", nil, "ab", nil}
	checkTokSequence(lexUnquotedString, expectedTypes, expectedVals, str, t)
}
