package models

import (
	"strings"
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
	Layer    *Layer // If the path is inside a layer
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

// Transforms the id of a physical object to its path
func PhysicalIDToPath(id string) string {
	return PhysicalPath + strings.ReplaceAll(strings.ReplaceAll(id, ".", "/"), "/*", "")
}
