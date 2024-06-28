package parser

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
	line := ""
	startLineNumber := 1
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		newLine := strings.TrimRight(scanner.Text(), " ")
		if len(newLine) >= 1 && newLine[len(newLine)-1] == '\\' {
			newLine = newLine[:len(newLine)-1]
			line += newLine + "\n"
		} else {
			line += newLine
			root, err := Parse(line)
			if err != nil {
				fileErr = addLineError(fileErr, err, filename, startLineNumber, line)
			}
			if root != nil {
				result = append(result, parsedLine{line, startLineNumber, root})
			}
			line = ""
			startLineNumber = lineNumber + 1
		}
	}
	if fileErr != nil {
		return result, fileErr
	}
	return result, nil
}

type StackTraceError struct {
	err     error
	history string
}

func newStackTraceError(err error, filename string, line string, lineNumber int) *StackTraceError {
	stackErr := &StackTraceError{err: err}
	stackErr.extend(filename, line, lineNumber)
	return stackErr
}

func (s *StackTraceError) extend(filename string, line string, lineNumber int) {
	trace := fmt.Sprintf("  File \"%s\", line %d\n", filename, lineNumber)
	trace += "    " + line + "\n"
	s.history = trace + s.history
}

func (s *StackTraceError) Error() string {
	msg := "Stack trace (most recent call last):\n"
	return msg + s.history + "Error : " + s.err.Error()
}

func LoadFile(path string) error {
	filename := filepath.Base(path)
	file, err := parseFile(path)
	if err != nil && !c.State.DryRun {
		// if c.State.DryRun {
		fmt.Println(err)
		// 	// c.State.DryRunErrors = append(c.State.DryRunErrors, err)
		// } else {
		return err
		// }
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
			stackTraceErr, ok := err.(*StackTraceError)
			if ok {
				stackTraceErr.extend(filename, file[i].line, file[i].lineNumber)
			} else {
				stackTraceErr = newStackTraceError(err, filename, file[i].line, file[i].lineNumber)
			}
			if c.State.DryRun {
				fmt.Println(stackTraceErr)
				c.State.DryRunErrors = append(c.State.DryRunErrors, stackTraceErr)
			} else {
				return stackTraceErr
			}
		}
	}
	// fmt.Println("END LOAD ", errCount)
	return err
}
