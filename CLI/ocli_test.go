package main

import (
	"cli/controllers"
	"fmt"
	"os"
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

func TestParseFileError(t *testing.T) {
	invalidPath := "/invalid/path/file.ocli"
	_, err := parseFile(invalidPath)
	assert.ErrorContains(t, err, "open "+invalidPath+": no such file or directory")
}

func TestParseFile(t *testing.T) {
	basePath := t.TempDir() // temporary directory that will be deleted after the tests have finished
	fileContent := ".var:siteName=siteB\n"
	fileContent += "+site:$siteName\n"
	fileContent += "+bd:/P/$siteName/blgdB@[0,0]@-90@[25,29.4,1]\n\n"
	fileContent += "//This is a comment line\n"
	fileContent += "+ro:/P/$siteName/blgdB/R2@[0,0]@0@[22.8,19.8,0.5]@+x+y\n\n"
	fileContent += "for i in 1..2 {                                                            \\\n"
	fileContent += "	.var:multbyten=eval 10*$i;                                             \\\n"
	fileContent += "	+rk:/P/$siteName/blgdB/R2/A${multbyten}@[$i,2]@t@[0,0,180]@[60,120,42] \\\n"
	fileContent += "}\n"

	filename := "parse_test_file.ocli"
	filePath := basePath + "/" + filename

	err := os.WriteFile(filePath, []byte(fileContent), 0644)

	if err != nil {
		t.Errorf("an error ocurred while creating the test file: %s", err)
	}
	parsedLines, err := parseFile(filePath)
	if err != nil {
		t.Errorf("an error ocurred parsing the file: %s", err)
	}
	assert.Len(t, parsedLines, 5)
	assert.Equal(t, "siteB", parsedLines[0].root.(*assignNode).val.(*valueNode).val)
	assert.Equal(t, "siteName", parsedLines[0].root.(*assignNode).variable)
	assert.IsType(t, &createSiteNode{}, parsedLines[1].root)
	assert.IsType(t, &createBuildingNode{}, parsedLines[2].root)
	assert.IsType(t, &createRoomNode{}, parsedLines[3].root)
	assert.IsType(t, &forRangeNode{}, parsedLines[4].root)
}

func TestNewStackTraceError(t *testing.T) {
	err := fmt.Errorf("my-error")
	stackTrace := newStackTraceError(err, "my_file", "line", 1)
	msg := "Stack trace (most recent call last):\n"
	msg += stackTrace.history + "Error : " + err.Error()
	assert.Equal(t, msg, stackTrace.Error())
}

func TestLoadFile(t *testing.T) {
	_, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	basePath := t.TempDir() // temporary directory that will be deleted after the tests have finished
	fileContent := ".var:siteName=siteB\n"

	filename := "load_test_file.ocli"
	filePath := basePath + "/" + filename

	err := os.WriteFile(filePath, []byte(fileContent), 0644)

	if err != nil {
		t.Errorf("an error ocurred while creating the test file: %s", err)
	}
	err = LoadFile(filePath)
	if err != nil {
		t.Errorf("an error ocurred parsing the file: %s", err)
	}

	assert.Contains(t, controllers.State.DynamicSymbolTable, "siteName")
	assert.Equal(t, "siteB", controllers.State.DynamicSymbolTable["siteName"])
}

func TestLoadFileParseError(t *testing.T) {
	_, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	basePath := t.TempDir() // temporary directory that will be deleted after the tests have finished
	fileContent := "siteName=siteB\n"

	filename := "load_test_file.ocli"
	filePath := basePath + "/" + filename

	err := os.WriteFile(filePath, []byte(fileContent), 0644)

	if err != nil {
		t.Errorf("an error ocurred while creating the test file: %s", err)
	}
	err = LoadFile(filePath)
	assert.NotNil(t, err)
	assert.IsType(t, &fileParseError{}, err)

	errMsg := "Syntax errors were found in the file: " + filename + "\nThe following commands were invalid\n  LINE#: 1\tCOMMAND:siteName=siteB"
	assert.ErrorContains(t, err, errMsg)
}

func TestLoadFileStackError(t *testing.T) {
	_, _, deferFunction := setMainEnvironmentMock(t)
	defer deferFunction()

	basePath := t.TempDir() // temporary directory that will be deleted after the tests have finished
	fileContent := ".var: i = eval 10/0\n"

	filename := "load_test_file.ocli"
	filePath := basePath + "/" + filename

	err := os.WriteFile(filePath, []byte(fileContent), 0644)

	if err != nil {
		t.Errorf("an error ocurred while creating the test file: %s", err)
	}
	err = LoadFile(filePath)
	assert.NotNil(t, err)
	assert.IsType(t, &stackTraceError{}, err)

	errMsg := "Stack trace (most recent call last):\n  File \"" + filename + "\", line 1\n    .var: i = eval 10/0\nError : cannot divide by 0"
	assert.ErrorContains(t, err, errMsg)
}
