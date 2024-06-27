package models

import (
	"errors"
	"fmt"
	"strings"

	"github.com/elliotchance/pie/v2"
)

const (
	PhysicalPath          = "/Physical/"
	StrayPath             = PhysicalPath + "Stray/"
	LogicalPath           = "/Logical/"
	ObjectTemplatesPath   = LogicalPath + "ObjectTemplates/"
	RoomTemplatesPath     = LogicalPath + "RoomTemplates/"
	BuildingTemplatesPath = LogicalPath + "BldgTemplates/"
	GroupsPath            = LogicalPath + "Groups/"
	TagsPath              = LogicalPath + "Tags/"
	LayersPath            = LogicalPath + "Layers/"
	VirtualObjsNode       = "VirtualObjects"
	VirtualObjsPath       = LogicalPath + VirtualObjsNode + "/"
	OrganisationPath      = "/Organisation/"
	DomainsPath           = OrganisationPath + "Domain/"
)

var PathPrefixes = []string{
	StrayPath,
	PhysicalPath,
	ObjectTemplatesPath,
	RoomTemplatesPath,
	BuildingTemplatesPath,
	GroupsPath,
	TagsPath,
	LayersPath,
	DomainsPath,
	VirtualObjsPath,
}

type Path struct {
	Prefix   string // The prefix indicating to which entity class it belongs (physical, template, group, etc.)
	ObjectID string
	Layer    Layer // If the path is inside a layer
}

const UnlimitedDepth = -1

var ErrMaxLessMin = errors.New("max depth cannot be less than the min depth")

// Transforms the path into a recursive path, transforming the * wildcard into **.
// minDepth and mexDepth are use to set the minimum and maximum amount of children between the path and the results
func (path *Path) MakeRecursive(minDepth, maxDepth int, fromPath string) error {
	depth := ""
	if maxDepth > UnlimitedDepth {
		if minDepth > maxDepth {
			return ErrMaxLessMin
		}

		depth = fmt.Sprintf("{%v,%v}", minDepth, maxDepth)
	} else if minDepth > 0 {
		depth = fmt.Sprintf("{%v,}", minDepth)
	}

	recursiveWildcard := "**" + depth

	if strings.HasSuffix(path.ObjectID, ".*") {
		// finishes in .*, meaning all the children
		path.ObjectID = path.ObjectID[:len(path.ObjectID)-2] + "." + recursiveWildcard + ".*"
		return nil
	}

	fromID := PhysicalPathToObjectID(fromPath)

	if !pie.Contains([]string{"", ".", "_", "-"}, fromID) {
		index := strings.Index(path.ObjectID, fromID)
		if index != -1 {
			path.ObjectID = path.ObjectID[:index] + recursiveWildcard + "." + path.ObjectID[index:]
		}
	}

	return nil
}

func IsPhysical(path string) bool {
	return pathIs(path, PhysicalPath)
}

func IsStray(path string) bool {
	return pathIs(path, StrayPath)
}

func IsObjectTemplate(path string) bool {
	return pathIs(path, ObjectTemplatesPath)
}

func IsRoomTemplate(path string) bool {
	return pathIs(path, RoomTemplatesPath)
}

func IsBuildingTemplate(path string) bool {
	return pathIs(path, BuildingTemplatesPath)
}

func IsTag(path string) bool {
	return pathIs(path, TagsPath)
}

func IsLayer(path string) bool {
	return pathIs(path, LayersPath)
}

func IsGroup(path string) bool {
	return pathIs(path, GroupsPath)
}

func IsVirtual(path string) bool {
	return pathIs(path, VirtualObjsPath)
}

func pathIs(path, prefix string) bool {
	return strings.HasPrefix(addLastSlash(path), prefix)
}

func addLastSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}

	return path
}

func SplitPath(path string) []string {
	return strings.Split(path, "/")
}

func JoinPath(path []string) string {
	return strings.Join(path, "/")
}

func PhysicalPathToObjectID(path string) string {
	return strings.TrimSuffix(
		strings.ReplaceAll(
			strings.TrimPrefix(
				addLastSlash(path),
				PhysicalPath,
			),
			"/", ".",
		),
		".",
	)
}

// Transforms the id of a physical object to its path
func PhysicalIDToPath(id string) string {
	return PhysicalPath + strings.ReplaceAll(id, ".", "/")
}

func GetObjectIDFromPath(pathStr string) string {
	for _, prefix := range PathPrefixes {
		if strings.HasPrefix(pathStr, string(prefix)) {
			id := pathStr[len(prefix):]
			id = strings.ReplaceAll(id, "/", ".")
			return id
		}
	}
	return ""
}

// Removes last "amount" elements from the "path"
func PathRemoveLast(path string, amount int) string {
	pathSplit := SplitPath(path)

	return JoinPath(pathSplit[:len(pathSplit)-amount])
}

// Transform an object id into a relative path from the path "fromPath"
// Example: BASIC.A.R1 is A/R1 from /Physical/BASIC
func ObjectIDToRelativePath(objectID, fromPath string) string {
	objectIDElements := strings.Split(objectID, ".")
	fromPathLast := pie.Last(SplitPath(fromPath))

	index := pie.FindFirstUsing(objectIDElements, func(element string) bool {
		return element == fromPathLast
	})

	remainingElements := objectIDElements[index+1:]

	return JoinPath(remainingElements)
}
