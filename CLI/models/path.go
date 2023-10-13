package models

import (
	"cli/utils"
	"fmt"
)

const (
	PhysicalPath          = "/Physical/"
	StayPath              = PhysicalPath + "Stray/"
	LogicalPath           = "/Logical/"
	ObjectTemplatesPath   = LogicalPath + "ObjectTemplates/"
	RoomTemplatesPath     = LogicalPath + "RoomTemplates/"
	BuildingTemplatesPath = LogicalPath + "BldgTemplates/"
	TagsPath              = LogicalPath + "Tags/"
	GroupsPath            = LogicalPath + "Groups/"
	OrganisationPath      = "/Organisation/"
	DomainsPath           = OrganisationPath + "Domain/"
)

func IsHierarchical(path string) bool {
	return !IsNonHierarchical(path)
}

func IsNonHierarchical(path string) bool {
	return IsObjectTemplate(path, nil) || IsRoomTemplate(path, nil) ||
		IsBuildingTemplate(path, nil) || IsTag(path, nil)
}

func IsPhysical(path string, suffix *string) bool {
	return utils.StartsWith(path, PhysicalPath, suffix)
}

func IsStray(path string, suffix *string) bool {
	return utils.StartsWith(path, StayPath, suffix)
}

func IsObjectTemplate(path string, suffix *string) bool {
	return utils.StartsWith(path, ObjectTemplatesPath, suffix)
}

func IsRoomTemplate(path string, suffix *string) bool {
	return utils.StartsWith(path, RoomTemplatesPath, suffix)
}

func IsBuildingTemplate(path string, suffix *string) bool {
	return utils.StartsWith(path, BuildingTemplatesPath, suffix)
}

func IsTag(path string, suffix *string) bool {
	return utils.StartsWith(path, TagsPath, suffix)
}

func IsGroup(path string, suffix *string) bool {
	return utils.StartsWith(path, GroupsPath, suffix)
}

func IsDomain(path string, suffix *string) bool {
	return utils.StartsWith(path, DomainsPath, suffix)
}

func ObjectId(path string) (string, error) {
	var suffix string
	if IsPhysical(path, &suffix) {
		return suffix, nil
	}

	return "", fmt.Errorf("path does not point to a physical object")
}

func ObjectSlug(path string) (string, error) {
	var suffix string
	if IsObjectTemplate(path, &suffix) ||
		IsRoomTemplate(path, &suffix) ||
		IsBuildingTemplate(path, &suffix) ||
		IsTag(path, &suffix) {
		return suffix, nil
	}

	return "", fmt.Errorf("path does not point to a logical object")
}
