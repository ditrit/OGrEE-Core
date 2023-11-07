package controllers

import (
	"cli/models"
	"cli/utils"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/elliotchance/pie/v2"
)

type FillFunc func(n *HierarchyNode, path string, depth int) error

type HierarchyNode struct {
	Name     string
	Children map[string]*HierarchyNode
	FillFn   FillFunc
	Obj      any
	IsLayer  bool // Indicates whether the node represents a layer, so Obj will be of type Layer
}

func NewNode(name string) *HierarchyNode {
	return &HierarchyNode{name, map[string]*HierarchyNode{}, nil, nil, false}
}

func NewLayerNode(name string, layer models.Layer) *HierarchyNode {
	newNode := &HierarchyNode{name, map[string]*HierarchyNode{}, nil, nil, true}
	newNode.Obj = layer

	return newNode
}

func (n *HierarchyNode) StringAux(prefix string, sb *strings.Builder, depth int) {
	if depth == 0 {
		return
	}
	children := []*HierarchyNode{}
	for _, child := range n.Children {
		children = append(children, child)
	}
	sort.SliceStable(children, func(i, j int) bool {
		return children[i].Name < children[j].Name
	})

	// do not show layers in tree
	children = pie.Filter(children, func(child *HierarchyNode) bool {
		return !child.IsLayer
	})

	for i, child := range children {
		if i == len(children)-1 {
			sb.WriteString(prefix + "└── " + child.Name + "\n")
			child.StringAux(prefix+"    ", sb, depth-1)
		} else {
			sb.WriteString(prefix + "├── " + child.Name + "\n")
			child.StringAux(prefix+"│   ", sb, depth-1)
		}
	}
}

func (n *HierarchyNode) String(depth int) string {
	var sb strings.Builder
	n.StringAux("", &sb, depth)
	s := sb.String()
	for len(s) >= 1 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	return s
}

func (n *HierarchyNode) CanBeFilled() bool {
	return n.FillFn != nil
}

func (n *HierarchyNode) AddChild(child *HierarchyNode) {
	n.Children[child.Name] = child
}

func (n *HierarchyNode) AddChildInPath(child *HierarchyNode, path string) {
	nearest, remainingPath := n.FindNearestNode(path)
	if remainingPath != "" {
		intermediateObjects := models.SplitPath(remainingPath)

		for i, intermediateObject := range intermediateObjects {
			if i != len(intermediateObjects)-1 {
				newNode := NewNode(intermediateObject)
				nearest.AddChild(newNode)
				nearest = newNode
			}
		}

		nearest.AddChild(child)
	} else {
		// child is already in hierarchy, add children
		nearest.Children = child.Children
		nearest.Obj = child.Obj
	}
}

func (n *HierarchyNode) findNodeAux(path []string) (r *HierarchyNode, remainingPath []string) {
	if len(path) == 0 {
		return n, []string{}
	}
	child := n.Children[path[0]]
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

func (n *HierarchyNode) Fill(path string, depth int) error {
	if depth == 0 {
		return nil
	}

	if n.FillFn != nil {
		return n.FillFn(n, path, depth)
	}

	return nil
}

func BuildBaseTree() *HierarchyNode {
	root := NewNode("")
	root.FillFn = FillChildren

	physical := NewNode("Physical")
	physical.FillFn = FillUrlTreeFnAndFillChildren("/api/sites", FillObjectTree, false)
	root.AddChild(physical)

	stray := NewNode("Stray")
	stray.FillFn = FillUrlTreeFn("/api/stray-objects", FillObjectTree, false)
	physical.AddChild(stray)

	logical := NewNode("Logical")
	logical.FillFn = FillChildren
	root.AddChild(logical)

	objectTemplates := NewNode("ObjectTemplates")
	objectTemplates.FillFn = FillUrlTreeFn("/api/obj-templates", nil, false)
	logical.AddChild(objectTemplates)

	roomTemplates := NewNode("RoomTemplates")
	roomTemplates.FillFn = FillUrlTreeFn("/api/room-templates", nil, false)
	logical.AddChild(roomTemplates)

	bldgTemplates := NewNode("BldgTemplates")
	bldgTemplates.FillFn = FillUrlTreeFn("/api/bldg-templates", nil, false)
	logical.AddChild(bldgTemplates)

	tags := NewNode("Tags")
	tags.FillFn = FillUrlTreeFn("/api/tags", nil, false)
	logical.AddChild(tags)

	groups := NewNode("Groups")
	groups.FillFn = FillUrlTreeFn("/api/groups", nil, true)
	logical.AddChild(groups)

	organisation := NewNode("Organisation")
	organisation.FillFn = FillChildren
	root.AddChild(organisation)

	domain := NewNode("Domain")
	domain.FillFn = FillUrlTreeFn("/api/domains", FillObjectTree, false)
	organisation.AddChild(domain)

	organisation.AddChild(NewNode("Enterprise"))

	return root
}

func FillMapTree(n *HierarchyNode, obj map[string]any) error {
	children, ok := obj["children"].([]any)
	if !ok {
		return nil
	}

	for _, childAny := range children {
		var childNode *HierarchyNode

		switch child := childAny.(type) {
		case map[string]any:
			childNode = NewNode(utils.NameOrSlug(child))

			err := FillMapTree(childNode, child)
			if err != nil {
				return err
			}

			delete(child, "children")
		case models.Layer:
			childNode = NewLayerNode(child.Name, child)
		default:
			return fmt.Errorf("invalid child format")
		}

		childNode.Obj = childAny
		n.AddChild(childNode)
	}

	delete(obj, "children")
	n.Obj = obj

	return nil
}

func FillChildren(n *HierarchyNode, path string, depth int) error {
	for _, child := range n.Children {
		newPath := path
		if path != "/" {
			newPath += "/"
		}
		newPath += child.Name
		err := child.Fill(newPath, depth-1)
		if err != nil {
			return err
		}
	}

	return nil
}

func FillObjectTree(n *HierarchyNode, path string, depth int) error {
	obj, err := C.PollObjectWithChildren(path, depth)
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("location not found")
	}
	return FillMapTree(n, obj)
}

func FillUrlTree(n *HierarchyNode, path string, depth int, url string, followFillFn FillFunc, fullId bool) error {
	resp, err := API.Request("GET", url, nil, http.StatusOK)
	if err != nil {
		return err
	}
	invalidRespErr := fmt.Errorf("invalid response from API on GET %s", url)
	obj, ok := resp.Body["data"].(map[string]any)
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
		var objName string
		if fullId {
			objName = strings.Replace(obj["id"].(string), ".", "/", -1)
		} else {
			objName = utils.NameOrSlug(obj)
			objId, okId := obj["id"].(string)
			if okId && objId != objName {
				continue
			}
		}
		subTree := NewNode(objName)
		delete(obj, "children")
		subTree.Obj = obj
		subTree.FillFn = followFillFn

		err = subTree.Fill(path+"/"+objName, depth-1)
		if err != nil {
			return err
		}
		n.AddChild(subTree)
	}
	return nil
}

func FillUrlTreeFnAndFillChildren(url string, followFn FillFunc, fullId bool) FillFunc {
	return func(n *HierarchyNode, path string, depth int) error {
		err := FillUrlTree(n, path, depth, url, followFn, fullId)
		if err != nil {
			return err
		}

		return FillChildren(n, path, depth)
	}
}

func FillUrlTreeFn(url string, followFn FillFunc, fullId bool) FillFunc {
	return func(n *HierarchyNode, path string, depth int) error {
		return FillUrlTree(n, path, depth, url, followFn, fullId)
	}
}

func (controller Controller) Tree(path string, depth int) (*HierarchyNode, error) {
	if models.PathIsLayer(path) {
		return nil, errors.New("it is not possible to tree a layer")
	}

	n := State.Hierarchy.FindNode(path)
	if n != nil && n.CanBeFilled() {
		err := n.Fill(path, depth)
		if err != nil {
			return nil, err
		}

		return n, nil
	}

	obj, err := controller.GetObjectWithChildren(path, depth)
	if err != nil {
		return nil, err
	}

	n = NewNode(utils.NameOrSlug(obj))

	err = FillMapTree(n, obj)
	if err != nil {
		return nil, err
	}

	// add node to the stored hierarchy
	State.Hierarchy.AddChildInPath(n, path)

	return n, nil
}
