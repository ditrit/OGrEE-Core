/*
 The preprocessor processes OCLI scripts mainly
 to change the older syntax to new syntax

*/
package preprocessor

import (
	"bufio"
	l "cli/logger"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ProcessFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		println("Error:", err.Error())
		l.GetErrorLogger().Println("Error:", err)
		return ""
	}
	defer file.Close()
	var lines []string
	scanner := bufio.NewScanner(file)

	ocr := regexp.MustCompile(`^\+[a-z]+:([A-Za-z0-9_]+(\.|@|=))+`)
	editr := regexp.MustCompile(`^    ([A-Za-z0-9_]+(\.))+[A-Za-z0-9_]+=`)
	editr2 := regexp.MustCompile(`^([A-Za-z0-9_]+(\.))+[A-Za-z0-9_]+=`)
	editr3 := regexp.MustCompile(`^    ([A-Za-z0-9_]+(\.))+[A-Za-z0-9_]+:[A-Za-z0-9_]+=`)
	selr := regexp.MustCompile(`^=([A-Za-z0-9_]+(\.))+`)
	delr := regexp.MustCompile(`^-([A-Za-z0-9_]+(\.))+`)

	for scanner.Scan() {
		x := scanner.Text()

		ocMatch := ocr.FindString(x)
		editrMatch := editr.FindString(x)
		editr2Match := editr2.FindString(x)
		editr3Match := editr3.FindString(x)
		selrMatch := selr.FindString(x)
		delrMatch := delr.FindString(x)

		if ocMatch != "" {
			modifiedSubStr := strings.ReplaceAll(ocMatch, ".", "/")
			x = strings.Replace(x, ocMatch, modifiedSubStr, 1)
			lines = append(lines, x)
			continue
		}

		if editrMatch != "" {
			count := strings.Count(editrMatch, ".")
			modifiedSubStr := strings.Replace(editrMatch, ".", "/", count-1)
			modifiedSubStr = strings.Replace(modifiedSubStr, ".", ":", 1)
			x = strings.Replace(x, editrMatch, modifiedSubStr, 1)
			lines = append(lines, x)
			continue
		}

		if editr2Match != "" {
			if !strings.Contains(editr2Match, "ui") && !strings.Contains(editr2Match, "camera") {
				count := strings.Count(editr2Match, ".")
				modifiedSubStr := strings.Replace(editr2Match, ".", "/", count-1)
				modifiedSubStr = strings.Replace(modifiedSubStr, ".", ":", 1)
				x = strings.Replace(x, editr2Match, modifiedSubStr, 1)
				lines = append(lines, x)
				continue
			}
		}

		if editr3Match != "" {
			if !strings.Contains(editr3Match, "ui") && !strings.Contains(editr3Match, "camera") {
				count := strings.Count(editr3Match, ".")
				modifiedSubStr := strings.Replace(editr3Match, ".", "/", count)
				//modifiedSubStr = strings.Replace(modifiedSubStr, ".", ":", 1)
				x = strings.Replace(x, editr3Match, modifiedSubStr, 1)
				lines = append(lines, x)
				continue
			}
		}

		if selrMatch != "" {
			modifiedSubStr := strings.ReplaceAll(selrMatch, ".", "/")
			x = strings.Replace(x, selrMatch, modifiedSubStr, 1)
			lines = append(lines, x)
			continue
		}

		if delrMatch != "" {
			modifiedSubStr := strings.ReplaceAll(delrMatch, ".", "/")
			x = strings.Replace(x, delrMatch, modifiedSubStr, 1)
			lines = append(lines, x)
			continue
		}

		//Otherwise just append the non matching line
		lines = append(lines, x)

	}

	//Merge the lines array
	//Generate new script to execute
	prefix := filepath.Dir(path)
	if prefix != "/" {
		prefix = prefix + "/"
	}

	fileName := prefix + filepath.Base(path) + ".new"
	f, e := os.Create(fileName)
	if e != nil {
		println("Error:", err.Error())
		l.GetWarningLogger().Println("Error:", err)
		return ""
	}

	f.WriteString(strings.Join(lines, "\n"))
	return fileName
}

/*

	^\+[a-z]+:([A-Za-z0-9_]+(\.|@|=))+


	^    ([A-Za-z0-9_]+(\.))+[A-Za-z0-9_]+=
	^([A-Za-z0-9_]+(\.))+[A-Za-z0-9_]+= //Problematic, catches ui.delay= etc
	^=([A-Za-z0-9_]+(\.))+ //Catch selections

*/
