package models

import (
	"strings"

	"github.com/elliotchance/pie/v2"
)

const (
	PhysicalPath          = "/Physical/"
	StayPath              = PhysicalPath + "Stray/"
	LogicalPath           = "/Logical/"
	ObjectTemplatesPath   = LogicalPath + "ObjectTemplates/"
	RoomTemplatesPath     = LogicalPath + "RoomTemplates/"
	BuildingTemplatesPath = LogicalPath + "BldgTemplates/"
	GroupsPath            = LogicalPath + "Groups/"
	TagsPath              = LogicalPath + "Tags/"
	LayersPath            = LogicalPath + "Layers/"
	OrganisationPath      = "/Organisation/"
	DomainsPath           = OrganisationPath + "Domain/"
)

var PathPrefixes = []string{
	StayPath,
	PhysicalPath,
	ObjectTemplatesPath,
	RoomTemplatesPath,
	BuildingTemplatesPath,
	GroupsPath,
	TagsPath,
	LayersPath,
	DomainsPath,
}

type Path struct {
	Prefix   string // The prefix indicating to which entity class it belongs (physical, template, group, etc.)
	ObjectID string
	Layer    Layer // If the path is inside a layer
}

// Transforms the path into a recursive path, transforming the * wildcard into **
func (path *Path) MakeRecursive() {
	index := strings.LastIndex(path.ObjectID, ".*")
	if index != -1 {
		// finishes in .*, meaning all the children
		path.ObjectID = path.ObjectID[:index] + strings.Replace(path.ObjectID[index:], ".*", ".**", 1)
		return
	}

	// all the children that are called as the last element of the id
	idElements := strings.Split(path.ObjectID, ".")

	idElements[len(idElements)-1] = "**" + idElements[len(idElements)-1]
	path.ObjectID = strings.Join(idElements, ".")
}

func IsPhysical(path string) bool {
	return pathIs(path, PhysicalPath)
}

func IsStray(path string) bool {
	return pathIs(path, StayPath)
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
