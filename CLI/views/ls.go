package views

import (
	"cli/utils"
	"fmt"
)

func Objects(objects []map[string]any) {
	for _, obj := range objects {
		fmt.Println(utils.NameOrSlug(obj))
	}
}

func SortedObjects(objects []map[string]any, attributes []string) {
	for _, obj := range objects {
		printStr := "Name : %s"
		attrVals := []any{utils.NameOrSlug(obj)}
		for _, attr := range attributes {
			attrVal, hasAttr := utils.ObjectAttr(obj, attr)
			if !hasAttr {
				attrVal = "-"
			}
			attrVals = append(attrVals, attr)
			attrVals = append(attrVals, attrVal)
			printStr += "    %v : %v"
		}
		printStr += "\n"
		fmt.Printf(printStr, attrVals...)
	}
}
