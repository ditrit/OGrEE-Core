package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type FillFunc func(n *HierarchyNode, path string, depth int) error

type HierarchyNode struct {
	Name   string
	Childs map[string]*HierarchyNode
	FillFn FillFunc
}

func NewNode(name string) *HierarchyNode {
	return &HierarchyNode{name, map[string]*HierarchyNode{}, nil}
}

func (n *HierarchyNode) StringAux(prefix string, sb *strings.Builder) {
	childs := []*HierarchyNode{}
	for _, child := range n.Childs {
		childs = append(childs, child)
	}
	sort.SliceStable(childs, func(i, j int) bool {
		return childs[i].Name < childs[j].Name
	})
	for i, child := range childs {
		if i == len(childs)-1 {
			sb.WriteString(prefix + "└── " + child.Name + "\n")
			child.StringAux(prefix+"    ", sb)
		} else {
			sb.WriteString(prefix + "├── " + child.Name + "\n")
			child.StringAux(prefix+"│   ", sb)
		}
	}
}

func (n *HierarchyNode) String() string {
	var sb strings.Builder
	n.StringAux("", &sb)
	s := sb.String()
	for len(s) >= 1 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s
}

func (n *HierarchyNode) AddChild(child *HierarchyNode) {
	n.Childs[child.Name] = child
}

func (n *HierarchyNode) findNodeAux(path []string) (r *HierarchyNode, remainingPath []string) {
	if len(path) == 0 {
		return n, []string{}
	}
	child := n.Childs[path[0]]
	if child != nil {
		return child.findNodeAux(path[1:])
	}
	return n, path
}

func (n *HierarchyNode) FindNode(path string) *HierarchyNode {
	pathList := strings.Split(path, "/")[1:]
	if pathList[0] == "" {
		pathList = pathList[1:]
	}
	r, remainingPath := n.findNodeAux(pathList)
	if len(remainingPath) > 0 {
		return nil
	}
	return r
}

func (n *HierarchyNode) FindNearestNode(path string) (r *HierarchyNode, remainingPath string) {
	r, remainingPathList := n.findNodeAux(strings.Split(path, "/")[1:])
	return r, strings.Join(remainingPathList, "/")
}

func BuildBaseTree() *HierarchyNode {
	root := NewNode("")
	physical := NewNode("Physical")
	physical.FillFn = FillUrlTreeFn("/api/sites", FillObjectTree)
	root.AddChild(physical)
	stray := NewNode("Stray")
	stray.FillFn = FillUrlTreeFn("/api/stray-objects", FillObjectTree)
	physical.AddChild(stray)
	logical := NewNode("Logical")
	root.AddChild(logical)
	objectTemplates := NewNode("ObjectTemplates")
	objectTemplates.FillFn = FillUrlTreeFn("/api/obj-templates", nil)
	logical.AddChild(objectTemplates)
	roomTemplates := NewNode("RoomTemplates")
	roomTemplates.FillFn = FillUrlTreeFn("/api/room-templates", nil)
	logical.AddChild(roomTemplates)
	bldgTemplates := NewNode("BldgTemplates")
	bldgTemplates.FillFn = FillUrlTreeFn("/api/bldg-templates", nil)
	logical.AddChild(bldgTemplates)
	groups := NewNode("Groups")
	groups.FillFn = FillUrlTreeFn("/api/groups", nil)
	logical.AddChild(groups)
	organisation := NewNode("Organisation")
	root.AddChild(organisation)
	domain := NewNode("Domain")
	domain.FillFn = FillUrlTreeFn("/api/domains", FillObjectTree)
	organisation.AddChild(domain)
	organisation.AddChild(NewNode("Enterprise"))
	return root
}

func nameOrSlug(obj map[string]any) (string, error) {
	name, okName := obj["name"].(string)
	if !okName {
		name, okName = obj["slug"].(string)
	}
	if !okName {
		return "", fmt.Errorf("child has no name/slug")
	}
	return name, nil
}

func FillMapTree(n *HierarchyNode, obj map[string]any) error {
	children, ok := obj["children"].([]any)
	if !ok {
		return nil
	}
	for _, childAny := range children {
		childMap, ok := childAny.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid child format")
		}
		name, err := nameOrSlug(childMap)
		if err != nil {
			return err
		}
		child := NewNode(name)
		err = FillMapTree(child, childMap)
		if err != nil {
			return err
		}
		n.AddChild(child)
	}
	return nil
}

func FillObjectTree(n *HierarchyNode, path string, depth int) error {
	obj, err := PollObjectWithChildren(path, depth)
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("location not found")
	}
	return FillMapTree(n, obj)
}

func FillUrlTree(n *HierarchyNode, path string, depth int, url string, followFillFn FillFunc) error {
	resp, err := RequestAPI("GET", url, nil, http.StatusOK)
	if err != nil {
		return err
	}
	invalidRespErr := fmt.Errorf("invalid response from API on GET %s", url)
	obj, ok := resp.body["data"].(map[string]any)
	if !ok {
		return invalidRespErr
	}
	objects, hasObjects := obj["objects"].([]any)
	if !hasObjects {
		return invalidRespErr
	}
	for _, objAny := range objects {
		obj, ok := objAny.(map[string]any)
		if !ok {
			return invalidRespErr
		}
		objName, err := nameOrSlug(obj)
		if err != nil {
			return invalidRespErr
		}
		subTree := NewNode(objName)
		subTree.FillFn = followFillFn
		err = FillTree(subTree, path+"/"+objName, depth-1)
		if err != nil {
			return err
		}
		n.AddChild(subTree)
	}
	return nil
}

func FillUrlTreeFn(url string, followFn FillFunc) FillFunc {
	return func(n *HierarchyNode, path string, depth int) error {
		return FillUrlTree(n, path, depth, url, followFn)
	}
}

func FillTree(n *HierarchyNode, path string, depth int) error {
	if depth == 0 {
		return nil
	}
	if len(n.Childs) != 0 {
		for _, child := range n.Childs {
			newPath := path
			if path != "/" {
				newPath += "/"
			}
			newPath += child.Name
			err := FillTree(child, newPath, depth-1)
			if err != nil {
				return err
			}
		}
	}
	if n.FillFn != nil {
		return n.FillFn(n, path, depth)
	}
	return nil
}

func Tree(path string, depth int) (*HierarchyNode, error) {
	n := State.Hierarchy.FindNode(path)
	if n != nil {
		tempHierarchy := BuildBaseTree()
		root := tempHierarchy.FindNode(path)
		err := FillTree(root, path, depth)
		if err != nil {
			return nil, err
		}
		return root, nil
	}
	obj, err := PollObjectWithChildren(path, depth)
	if err != nil {
		return nil, err
	}
	name, err := nameOrSlug(obj)
	if err != nil {
		return nil, err
	}
	n = NewNode(name)
	err = FillMapTree(n, obj)
	if err != nil {
		return nil, err
	}
	return n, nil
}
