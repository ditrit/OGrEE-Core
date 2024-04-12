package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileParseErrorError(t *testing.T) {
	fileParser := fileParseError{
		filename:   "file",
		lineErrors: []string{"line1", "line2"},
	}
	message := fileParser.Error()
	expectedMessage := "Syntax errors were found in the file: " + fileParser.filename
	expectedMessage += "\nThe following commands were invalid"
	expectedMessage += "\n" + strings.Join(fileParser.lineErrors, "\n")
	assert.Equal(t, expectedMessage, message)
}

func TestAddLineError(t *testing.T) {
	err := fmt.Errorf("my error message")
	filename := "my-file"
	lineNumber := 3
	line := "line1"
	fileParser := addLineError(nil, err, filename, lineNumber, line)
	assert.Equal(t, filename, fileParser.filename)
	assert.Len(t, fileParser.lineErrors, 1)
	assert.Equal(t, fmt.Sprintf("  LINE#: %d\tCOMMAND:%s", lineNumber, line), fileParser.lineErrors[0])

	lineNumber = 10
	line = "line2"
	addLineError(fileParser, err, filename, lineNumber, line)
	assert.Equal(t, filename, fileParser.filename)
	assert.Len(t, fileParser.lineErrors, 2)
	assert.Equal(t, fmt.Sprintf("  LINE#: %d\tCOMMAND:%s", lineNumber, line), fileParser.lineErrors[1])
}
