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

type RelativePathArgs struct {
	FromPath             string
	fromPathWithoutLayer string
}

func (args RelativePathArgs) getFromPath() string {
	if args.fromPathWithoutLayer == "" {
		if models.PathIsLayer(args.FromPath) {
			args.fromPathWithoutLayer = models.PathRemoveLast(args.FromPath, 1) // remove layer
		} else {
			args.fromPathWithoutLayer = args.FromPath
		}
	}

	return args.fromPathWithoutLayer
}

func (args RelativePathArgs) Get(object map[string]any) string {
	id, idPresent := object["id"].(string)
	if !idPresent {
		return utils.NameOrSlug(object)
	}

	return models.ObjectIDToRelativePath(id, args.getFromPath())
}

func ListObjects(objects []map[string]any, sortAttr string, relativePath *RelativePathArgs) ([]string, error) {
	var stringList []string

	objects, err := SortObjects(objects, sortAttr)
	if err != nil {
		return nil, err
	}

	stringList = []string{}

	for _, object := range objects {
		stringList = append(stringList, getObjectNameOrPath(object, relativePath))
	}

	return stringList, nil
}

func getObjectNameOrPath(object map[string]any, relativePath *RelativePathArgs) string {
	if relativePath == nil {
		return utils.NameOrSlug(object)
	}

	return relativePath.Get(object)
}

func Ls(objects []map[string]any, sortAttr string, relativePath *RelativePathArgs) (string, error) {
	stringList, err := ListObjects(objects, sortAttr, relativePath)
	if err != nil {
		return "", err
	}

	toPrint := strings.Join(stringList, "\n")
	if len(toPrint) > 0 {
		toPrint = toPrint + "\n"
	}

	return toPrint, nil
}

func LsWithFormat(objects []map[string]any, sortAttr string, relativePath *RelativePathArgs, attributes []string) (string, error) {
	if sortAttr != "" {
		attributes = append([]string{sortAttr}, attributes...)
	}

	attributes = pie.Unique(attributes)

	objects, err := SortObjects(objects, sortAttr)
	if err != nil {
		return "", err
	}

	printAll := ""

	for _, obj := range objects {
		objectName, hasName := obj["name"].(string)
		if !hasName || !models.IsIDElementLayer(objectName) {
			printObject := "%s"
			attrVals := []any{getObjectNameOrPath(obj, relativePath)}

			for _, attr := range attributes {
				attrVal, hasAttr := utils.GetValFromObj(obj, attr)
				if !hasAttr {
					attrVal = "-"
				}
				attrVals = append(attrVals, attr, attrVal)
				printObject += "    %v: %v"
			}

			printObject += "\n"
			printAll = printAll + fmt.Sprintf(printObject, attrVals...)
		}
	}

	return printAll, nil
}

func idOrName(object map[string]any) string {
	id, okId := object["id"].(string)
	if okId {
		return id
	}

	return utils.NameOrSlug(object)
}

func SortObjects(objects []map[string]any, sortAttr string) ([]map[string]any, error) {
	if sortAttr == "" {
		sortAttr = "id"
	}

	if sortAttr == "id" {
		orderObjectsBy(objects, idOrName)
	} else {
		objects = pie.Filter(objects, func(object map[string]any) bool {
			_, hasAttr := utils.GetValFromObj(object, sortAttr)
			return hasAttr
		})

		if !objectsAreSortable(objects, sortAttr) {
			return nil, errors.New("objects cannot be sorted according to this attribute")
		}

		sort.Slice(objects, func(i, j int) bool {
			vali, _ := utils.GetValFromObj(objects[i], sortAttr)
			valj, _ := utils.GetValFromObj(objects[j], sortAttr)
			res, _ := utils.CompareVals(vali, valj)
			return res
		})
	}

	return objects, nil
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

func objectsAreSortable(objects []map[string]any, attr string) bool {
	for i := 1; i < len(objects); i++ {
		val0, _ := utils.GetValFromObj(objects[0], attr)
		vali, _ := utils.GetValFromObj(objects[i], attr)
		_, comparable := utils.CompareVals(val0, vali)
		if !comparable {
			return false
		}
	}
	return true
}
