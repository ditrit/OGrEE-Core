package models

import (
	"fmt"
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
	OrganisationPath      = "/Organisation/"
	DomainsPath           = OrganisationPath + "Domain/"
)

var pathPrefixes = []string{
	StayPath,
	PhysicalPath,
	ObjectTemplatesPath,
	RoomTemplatesPath,
	BuildingTemplatesPath,
	GroupsPath,
	TagsPath,
	DomainsPath,
}

func SplitPath(path string) (string, string, error) {
	for _, prefix := range pathPrefixes {
		if strings.HasPrefix(path, prefix) {
			id := path[len(prefix):]
			id = strings.ReplaceAll(id, "/", ".")
			return prefix, id, nil
		}
	}
	return "", "", fmt.Errorf("invalid object path")
}

func IsHierarchical(path string) bool {
	return !IsNonHierarchical(path)
}

func IsNonHierarchical(path string) bool {
	return IsObjectTemplate(path) || IsRoomTemplate(path) ||
		IsBuildingTemplate(path) || IsTag(path)
}

func IsObjectTemplate(path string) bool {
	return strings.HasPrefix(path, ObjectTemplatesPath)
}

func IsRoomTemplate(path string) bool {
	return strings.HasPrefix(path, RoomTemplatesPath)
}

func IsBuildingTemplate(path string) bool {
	return strings.HasPrefix(path, BuildingTemplatesPath)
}

func IsTag(path string) bool {
	return strings.HasPrefix(path, TagsPath)
}
