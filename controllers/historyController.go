package controllers

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

func AddHistory(rl *readline.Instance) {
	readFile, err := os.Open(".resources/.history")

	if err != nil {
		log.Fatalf("failed to open file: %s", err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	var fileTextLines []string

	for fileScanner.Scan() {
		fileTextLines = append(fileTextLines, fileScanner.Text())
	}

	readFile.Close()

	for _, eachline := range fileTextLines {
		rl.SaveHistory(strings.TrimSuffix(eachline, "\n"))
	}

	return
}
