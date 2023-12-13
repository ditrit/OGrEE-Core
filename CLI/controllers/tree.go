package controllers

import (
	"cli/models"
	"cli/utils"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/mitchellh/mapstructure"
)

const updatedThreshold time.Duration = 10 * time.Minute

type FillFunc func(n *HierarchyNode, path string, depth int) error

type HierarchyNode struct {
	Name             string
	Children         map[string]*HierarchyNode
	FillFn           FillFunc
	Obj              any
	IsLayer          bool      // Indicates whether the node represents a layer, so Obj will be of type Layer
	IsCached         bool      // Indicates if the node is cached (not getting its information from the api when consulted)
	LastHierarchyGet time.Time // Indicates when the children of this node have been loaded from the API
}

// Creates a HierarchyNode with a name
func NewNode(name string) *HierarchyNode {
	return &HierarchyNode{Name: name, Children: map[string]*HierarchyNode{}}
}

// Creates a HierarchyNode from an object.
// object is initial a map and it's decoded into the object's type T.
func NewNodeFromObject[T any](name string, object map[string]any) (*HierarchyNode, error) {
	newNode := NewNode(name)

	var objStruct T
	err := mapstructure.Decode(object, &objStruct)
	if err != nil {
		return nil, err
	}

	newNode.Obj = objStruct

	return newNode, nil
}

// Creates a HierarchyNode that represents a layer in the Hierarchy
func NewLayerNode(name string, layer any) *HierarchyNode {
	newNode := NewNode(name)
	newNode.Obj = layer
	newNode.IsLayer = true

	return newNode
}

// Creates a HierarchyNode that can be cached
func NewCachedNode(name string) *HierarchyNode {
	newNode := NewNode(name)
	newNode.IsCached = true

	return newNode
}

// Creates a HierarchyNode with a object in map format.
// This is the legacy way to store objects in the hierarchy, storing them as maps.
// It should progressively change to NewNodeFromObject, where the objects are stored
// in the corresponding type, avoiding working with maps all the time.
func NewNodeFromMap(name string, object map[string]any) (*HierarchyNode, error) {
	newNode := NewNode(name)

	return newNode, newNode.FillWithMap(object)
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

// Adds to hierarchy a node in the path.
// Creates all the intermediate nodes if the new node cannot be added directly to the current hierarchy.
// The new node will have a object of the type that corresponds to the path.
func (n *HierarchyNode) AddObjectInPath(object map[string]any, path string) (*HierarchyNode, error) {
	var newNode *HierarchyNode
	var err error

	if models.IsLayer(path) {
		newNode, err = NewNodeFromObject[models.UserDefinedLayer](utils.NameOrSlug(object), object)
	}

	if err != nil {
		return nil, err
	}

	if newNode != nil {
		n.AddChildInPath(newNode, path)
	}

	return newNode, nil
}

// Adds to hierarchy a node in the path.
// Creates all the intermediate nodes if the new node cannot be added directly to the current hierarchy.
// The new node will have a object in map format.
// All the children in the map will also be added to the hierarchy.
// This is the legacy way to store objects in the hierarchy, storing them as maps.
// It should progressively change to AddObjectInPath, where the objects are stored
// in the corresponding type, avoiding working with maps all the time.
func (n *HierarchyNode) AddMapInPath(object map[string]any, path string) (*HierarchyNode, error) {
	newNode, err := NewNodeFromMap(utils.NameOrSlug(object), object)
	if err != nil {
		return nil, err
	}

	n.AddChildInPath(newNode, path)

	return newNode, nil
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
	if len(pathList) > 0 && pathList[len(pathList)-1] == "" {
		pathList = pathList[:len(pathList)-1]
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

func (n *HierarchyNode) Fill(path string, depth int, now time.Time) error {
	if depth == 0 {
		return nil
	}

	if depth == 1 && n.IsUpdated(now) {
		// return cached node
		return nil
	}

	n.LastHierarchyGet = now

	if n.Name == "Layers" {
		n.IsCached = true
	}

	if n.FillFn != nil {
		return n.FillFn(n, path, depth)
	}

	return nil
}

func (n *HierarchyNode) FillWithMap(obj map[string]any) error {
	children, hasChildren := obj["children"].([]any)

	if len(n.Children) != 0 {
		n.Children = map[string]*HierarchyNode{}
	}

	if hasChildren {
		delete(obj, "children")

		for _, childAny := range children {
			var childNode *HierarchyNode
			var err error

			switch child := childAny.(type) {
			case map[string]any:
				childNode, err = NewNodeFromMap(utils.NameOrSlug(child), child)
			default:
				return fmt.Errorf("invalid child format")
			}

			if err != nil {
				return err
			}

			n.AddChild(childNode)
		}
	}

	n.Obj = obj

	return nil
}

func (n *HierarchyNode) IsUpdated(now time.Time) bool {
	return n.IsCached && now.Sub(n.LastHierarchyGet) < updatedThreshold
}

func BuildBaseTree(controller Controller) *HierarchyNode {
	root := NewNode("")
	root.FillFn = FillChildren

	physical := NewNode("Physical")
	physical.FillFn = FillUrlTreeFnAndFillChildren[map[string]any]("/api/sites", controller.FillObjectTree, false, controller.API)
	root.AddChild(physical)

	stray := NewNode("Stray")
	stray.FillFn = FillUrlTreeFn[map[string]any]("/api/stray-objects", controller.FillObjectTree, false, controller.API)
	physical.AddChild(stray)

	logical := NewNode("Logical")
	logical.FillFn = FillChildren
	root.AddChild(logical)

	objectTemplates := NewNode("ObjectTemplates")
	objectTemplates.FillFn = FillUrlTreeFn[map[string]any]("/api/obj-templates", nil, false, controller.API)
	logical.AddChild(objectTemplates)

	roomTemplates := NewNode("RoomTemplates")
	roomTemplates.FillFn = FillUrlTreeFn[map[string]any]("/api/room-templates", nil, false, controller.API)
	logical.AddChild(roomTemplates)

	bldgTemplates := NewNode("BldgTemplates")
	bldgTemplates.FillFn = FillUrlTreeFn[map[string]any]("/api/bldg-templates", nil, false, controller.API)
	logical.AddChild(bldgTemplates)

	layers := NewCachedNode("Layers")
	layers.FillFn = FillUrlTreeFn[models.UserDefinedLayer](LayersURL, nil, false, controller.API)
	logical.AddChild(layers)

	tags := NewNode("Tags")
	tags.FillFn = FillUrlTreeFn[map[string]any]("/api/tags", nil, false, controller.API)
	logical.AddChild(tags)

	groups := NewNode("Groups")
	groups.FillFn = FillUrlTreeFn[map[string]any]("/api/groups", nil, true, controller.API)
	logical.AddChild(groups)

	organisation := NewNode("Organisation")
	organisation.FillFn = FillChildren
	root.AddChild(organisation)

	domain := NewNode("Domain")
	domain.FillFn = FillUrlTreeFn[map[string]any]("/api/domains", controller.FillObjectTree, false, controller.API)
	organisation.AddChild(domain)

	organisation.AddChild(NewNode("Enterprise"))

	return root
}

func FillChildren(n *HierarchyNode, path string, depth int) error {
	for _, child := range n.Children {
		newPath := path
		if path != "/" {
			newPath += "/"
		}
		newPath += child.Name
		err := child.Fill(newPath, depth-1, time.Now())
		if err != nil {
			return err
		}
	}

	return nil
}

func (controller Controller) FillObjectTree(n *HierarchyNode, path string, depth int) error {
	obj, err := controller.PollObjectWithChildren(path, depth)
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("location not found")
	}

	return n.FillWithMap(obj)
}

func FillUrlTree[T any](n *HierarchyNode, api APIPort, path string, depth int, url string, followFillFn FillFunc, fullId bool) error {
	resp, err := api.Request(http.MethodGet, url, nil, http.StatusOK)
	if err != nil {
		return err
	}

	invalidRespErr := fmt.Errorf("invalid response from API on GET %s", url)
	data, hasData := resp.Body["data"].(map[string]any)
	if !hasData {
		return invalidRespErr
	}

	objects, hasObjects := data["objects"].([]any)
	if !hasObjects {
		return invalidRespErr
	}

	if _, ok := n.Children["Stray"]; ok && n.Name == "Physical" {
		n.Children = map[string]*HierarchyNode{"Stray": n.Children["Stray"]}
	} else {
		n.Children = map[string]*HierarchyNode{}
	}

	for _, objAny := range objects {
		obj, isMap := objAny.(map[string]any)
		if !isMap {
			return invalidRespErr
		}

		var objName string
		if fullId {
			objName = strings.Replace(obj["id"].(string), ".", "/", -1)
		} else {
			objName = utils.NameOrSlug(obj)
			objId, hasID := obj["id"].(string)
			if hasID && objId != objName {
				continue
			}
		}

		delete(obj, "children")
		subTree, err := NewNodeFromObject[T](objName, obj)
		if err != nil {
			return err
		}

		subTree.FillFn = followFillFn

		err = subTree.Fill(path+"/"+objName, depth-1, time.Now())
		if err != nil {
			return err
		}

		n.AddChild(subTree)
	}

	return nil
}

func FillUrlTreeFnAndFillChildren[T any](url string, followFn FillFunc, fullId bool, api APIPort) FillFunc {
	return func(n *HierarchyNode, path string, depth int) error {
		err := FillUrlTree[T](n, api, path, depth, url, followFn, fullId)
		if err != nil {
			return err
		}

		return FillChildren(n, path, depth)
	}
}

func FillUrlTreeFn[T any](url string, followFn FillFunc, fullId bool, api APIPort) FillFunc {
	return func(n *HierarchyNode, path string, depth int) error {
		return FillUrlTree[T](n, api, path, depth, url, followFn, fullId)
	}
}

func (controller Controller) Tree(path string, depth int) (*HierarchyNode, error) {
	if models.PathIsLayer(path) {
		return nil, errors.New("it is not possible to tree a layer")
	}

	n := State.Hierarchy.FindNode(path)
	if n != nil && n.CanBeFilled() {
		err := n.Fill(path, depth, controller.Clock.Now())
		if err != nil {
			return nil, err
		}

		return n, nil
	}

	obj, err := controller.GetObjectWithChildren(path, depth)
	if err != nil {
		return nil, err
	}

	// add object to the stored hierarchy
	return State.Hierarchy.AddMapInPath(obj, path)
}
