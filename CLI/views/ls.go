package views

import (
	"cli/models"
	"cli/utils"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/elliotchance/pie/v2"
)

func OrderObjects(objects []map[string]any, byID bool) {
	if byID {
		orderObjectsBy(objects, func(object map[string]any) string {
			return object["id"].(string)
		})
	} else {
		orderObjectsBy(objects, utils.NameOrSlug)
	}
}

func orderObjectsBy(objects []map[string]any, attributeGetter func(map[string]any) string) {
	sort.Slice(objects, func(i, j int) bool {
		// layers in last place
		if isObjectLayer(objects[i]) {
			if !isObjectLayer(objects[j]) {
				return false
			}
		} else if isObjectLayer(objects[j]) {
			return true
		}

		return attributeGetter(objects[i]) < attributeGetter(objects[j])
	})
}

func isObjectLayer(object map[string]any) bool {
	name, hasName := object["name"].(string)
	if !hasName {
		return false
	}

	return models.IsObjectIDLayer(name)
}

func ListObjects(objects []map[string]any, showRelativePath bool, fromPath string) []string {
	var stringsToShow []string

	fromPath = models.PathRemoveLast(fromPath, 1) // remove layer

	OrderObjects(objects, showRelativePath)

	stringsToShow = pie.Map(objects, func(object map[string]any) string {
		if !showRelativePath {
			return utils.NameOrSlug(object)
		}

		return models.ObjectIDToRelativePath(object["id"].(string), fromPath)
	})

	return stringsToShow
}

func getObjectNameOrPath(object map[string]any, showRelativePath bool, fromPath string) string {
	if !showRelativePath {
		return utils.NameOrSlug(object)
	}

	return models.ObjectIDToRelativePath(object["id"].(string), fromPath)
}

func Objects(objects []map[string]any, showRelativePath bool, fromPath string) string {
	toPrint := strings.Join(ListObjects(objects, showRelativePath, fromPath), "\n")
	if len(toPrint) > 0 {
		toPrint = toPrint + "\n"
	}

	return toPrint
}

func SortedObjects(objects []map[string]any, sortAttr string, attributes []string, showRelativePath bool, fromPath string) (string, error) {
	attributes = append([]string{sortAttr}, attributes...)

	fromPath = models.PathRemoveLast(fromPath, 1) // remove layer

	objects = pie.Filter(objects, func(object map[string]any) bool {
		_, hasAttr := utils.ObjectAttr(object, sortAttr)
		return hasAttr
	})

	if !objectsAreSortable(objects, sortAttr) {
		return "", errors.New("objects cannot be sorted according to this attribute")
	}

	sort.Slice(objects, func(i, j int) bool {
		vali, _ := utils.ObjectAttr(objects[i], sortAttr)
		valj, _ := utils.ObjectAttr(objects[j], sortAttr)
		res, _ := utils.CompareVals(vali, valj)
		return res
	})

	printAll := ""

	for _, obj := range objects {
		printObject := "%s"
		attrVals := []any{getObjectNameOrPath(obj, showRelativePath, fromPath)}

		for _, attr := range attributes {
			attrVal, hasAttr := utils.ObjectAttr(obj, attr)
			if !hasAttr {
				attrVal = "-"
			}
			attrVals = append(attrVals, attr, attrVal)
			printObject += "    %v: %v"
		}
		printObject += "\n"
		printAll = printAll + fmt.Sprintf(printObject, attrVals...)
	}

	return printAll, nil
}

func objectsAreSortable(objects []map[string]any, attr string) bool {
	for i := 1; i < len(objects); i++ {
		val0, _ := utils.ObjectAttr(objects[0], attr)
		vali, _ := utils.ObjectAttr(objects[i], attr)
		_, comparable := utils.CompareVals(val0, vali)
		if !comparable {
			return false
		}
	}
	return true
}
