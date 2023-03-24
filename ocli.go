package main

//This file loads and executes OCLI script files

import (
	"bufio"
	c "cli/controllers"
	l "cli/logger"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fileParseError struct {
	filename   string
	lineErrors []string
}

func (e *fileParseError) Error() string {
	msg := "Syntax errors were found in the file: " + e.filename
	msg += "\nThe following commands were invalid"
	for _, err := range e.lineErrors {
		msg += "\n" + err
	}
	return msg
}

func addLineError(
	fileErr *fileParseError,
	lineErr error,
	filename string,
	lineNumber int,
	line string,
) *fileParseError {
	msg := fmt.Sprintf("  LINE#: %d\tCOMMAND:%s", lineNumber, line)
	if fileErr == nil {
		return &fileParseError{filename, []string{msg}}
	}
	fileErr.lineErrors = append(fileErr.lineErrors, msg)
	return fileErr
}

type parsedLine struct {
	line       string
	lineNumber int
	root       node
}

func parseFile(path string) ([]parsedLine, error) {
	filename := filepath.Base(path)
	file, openErr := os.Open(path)
	if openErr != nil {
		return nil, openErr
	}
	result := []parsedLine{}
	var fileErr *fileParseError
	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if len(line) > 0 {
			root, err := Parse(line)
			if err != nil {
				fileErr = addLineError(fileErr, err, filename, lineNumber, line)
			}
			if root != nil {
				result = append(result, parsedLine{line, lineNumber, root})
			}
		}
	}
	if fileErr != nil {
		return result, fileErr
	}
	return result, nil
}

type stackTraceError struct {
	err     error
	history string
}

func newStackTraceError(err error, filename string, line string, lineNumber int) *stackTraceError {
	stackErr := &stackTraceError{err: err}
	stackErr.extend(filename, line, lineNumber)
	return stackErr
}

func (s *stackTraceError) extend(filename string, line string, lineNumber int) {
	trace := fmt.Sprintf("  File \"%s\", line %d\n", filename, lineNumber)
	trace += "    " + line + "\n"
	s.history = trace + s.history
}

func (s *stackTraceError) Error() string {
	msg := "Stack trace (most recent call last):\n"
	return msg + s.history + "Error : " + s.err.Error()
}

func LoadFile(path string) error {
	filename := filepath.Base(path)
	file, err := parseFile(path)
	if err != nil {
		return err
	}
	for i := range file {
		fmt.Println(file[i].line)
		_, err := file[i].root.execute()
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "Duplicate") || strings.Contains(errMsg, "duplicate") {
				l.GetWarningLogger().Println(errMsg)
				if c.State.DebugLvl > c.NONE {
					fmt.Println(errMsg)
				}
				continue
			}
			stackTraceErr, ok := err.(*stackTraceError)
			if ok {
				stackTraceErr.extend(filename, file[i].line, file[i].lineNumber)
			} else {
				stackTraceErr = newStackTraceError(err, filename, file[i].line, file[i].lineNumber)
			}
			return stackTraceErr
		}
	}
	return nil
}
