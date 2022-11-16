package controllers

import (
	"encoding/json"
	"sort"
)

// Display contents of []map[string]inf array
func DispMapArr(x []map[string]interface{}) {
	for idx := range x {
		println()
		println()
		println("OBJECT: ", idx)
		displayObject(x[idx])
		println()
	}
}

// displays contents of maps
func Disp(x map[string]interface{}) {

	jx, _ := json.Marshal(x)

	println("JSON: ", string(jx))
}

/*
// Ensure it satisfies sort.Interface
func (d Deals) Len() int           { return len(d) }
func (d Deals) Less(i, j int) bool { return d[i].Id < d[j].Id }
func (d Deals) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
*/

type sortable interface {
	getData() []interface{}
	Print()
}

// Helper Struct for sorting
type SortableMArr struct {
	data     []interface{}
	attr     string //Desired attr the user wants to use for sorting
	isNested bool   //If the attr is in "attributes" map
}

func (s SortableMArr) getData() []interface{} { return s.data }
func (s SortableMArr) Len() int               { return len(s.data) }
func (s SortableMArr) Swap(i, j int)          { s.data[i], s.data[j] = s.data[j], s.data[i] }
func (s SortableMArr) Less(i, j int) bool {
	var lKey string
	var rKey string
	var lmap map[string]interface{}
	var rmap map[string]interface{}

	//Check if the attribute is in the 'attributes' map
	if s.isNested {
		lKey = determineStrKey(s.data[i].(map[string]interface{})["attributes"].(map[string]interface{}), []string{s.attr})
		rKey = determineStrKey(s.data[j].(map[string]interface{})["attributes"].(map[string]interface{}), []string{s.attr})
		lmap = s.data[i].(map[string]interface{})["attributes"].(map[string]interface{})
		rmap = s.data[j].(map[string]interface{})["attributes"].(map[string]interface{})
	} else {
		lKey = determineStrKey(s.data[i].(map[string]interface{}), []string{s.attr})
		rKey = determineStrKey(s.data[j].(map[string]interface{}), []string{s.attr})
		lmap = s.data[i].(map[string]interface{})
		rmap = s.data[j].(map[string]interface{})
	}

	//We want the objs with non existing attribute at the end of the array
	if lKey == "" && rKey != "" {
		return false
	}

	if rKey == "" && lKey != "" {
		return true
	}

	lH := lmap[s.attr]
	rH := rmap[s.attr]

	//We must ensure that they are strings, non strings will be
	//placed at the end of the array
	var lOK, rOK bool
	_, lOK = lH.(string)
	_, rOK = rH.(string)

	if !lOK && rOK || lH == nil && rH != nil {
		return false
	}

	if lOK && !rOK || lH != nil && rH == nil {
		return true
	}

	if lH == nil && rH == nil {
		return false
	}

	return lH.(string) < rH.(string)

}

func (s SortableMArr) Print() {
	objs := s.getData()
	if s.isNested {
		for i := range objs {
			attr := objs[i].(map[string]interface{})["attributes"].(map[string]interface{})[s.attr]
			if attr == nil {
				attr = "NULL"
			}
			println(s.attr, ":",
				attr.(string),
				"  Name: ", objs[i].(map[string]interface{})["name"].(string))
		}
	} else {
		for i := range objs {
			println(s.attr, ":", objs[i].(map[string]interface{})[s.attr],
				"  Name: ", objs[i].(map[string]interface{})["name"].(string))
		}
	}

}

func SortObjects(objs *[]interface{}, attr string) *SortableMArr {
	var x SortableMArr
	var nested bool
	switch attr {
	case "id", "name", "category", "parentID",
		"description", "domain", "parentid", "parentId":
		nested = false
	default:
		nested = true
	}

	x = SortableMArr{*objs, attr, nested}
	sort.Sort(x)
	return &x
}
