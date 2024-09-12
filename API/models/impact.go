package models

import (
	u "p3/utils"
	"reflect"
	"strings"

	"github.com/elliotchance/pie/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type ImpactFilters struct {
	Categories []string `schema:"categories"`
	Ptypes     []string `schema:"ptypes"`
	Vtypes     []string `schema:"vtypes"`
}

func GetImpact(id string, userRoles map[string]Role, filters ImpactFilters) (map[string]any, *u.Error) {
	directChildren := map[string]any{}
	indirectChildren := map[string]any{}
	clusterRelations := map[string][]string{} // map of clusterId and list of objIds linked to that cluster

	// Get target object for impact analysis
	target, err := GetObjectById(id, u.HIERARCHYOBJS_ENT, u.RequestFilters{}, userRoles)
	if err != nil {
		return nil, err
	}

	// Get all children
	allChildren, _, err := getChildren(target["category"].(string), target["id"].(string), 999, u.RequestFilters{})
	if err != nil {
		return nil, err
	}

	// Find relations
	// Cluster associated to this target
	if targetAttrs, ok := target["attributes"].(map[string]any); ok {
		setClusterRelation(id, targetAttrs, clusterRelations)
	}
	// Direct/indirect children and associated clusters
	targetLevel := strings.Count(id, ".")
	for childId, childData := range allChildren {
		childAttrs := childData.(map[string]any)["attributes"].(map[string]any)
		if strings.Count(childId, ".") == targetLevel+1 {
			// direct child
			directChildren[childId] = childData
			// check if linked to a cluster
			setClusterRelation(childId, childAttrs, clusterRelations)
			continue
		}
		// indirect child
		setIndirectChildren(filters, childData.(map[string]any), childAttrs, indirectChildren, clusterRelations)
	}

	// handle cluster relations
	if err := clusterRelationsToIndirect(filters, clusterRelations, indirectChildren, userRoles); err != nil {
		return nil, err
	}

	// send response
	data := map[string]any{"direct": directChildren, "indirect": indirectChildren, "relations": clusterRelations}
	return data, nil
}

func setClusterRelation(childId string, childAttrs map[string]any, clusterRelations map[string][]string) {
	vconfig, hasVconfig := childAttrs["virtual_config"].(map[string]any)
	if hasVconfig && vconfig["clusterId"] != nil {
		clusterId := vconfig["clusterId"].(string)
		if clusterId == "" {
			return
		}
		clusterRelations[clusterId] = append(clusterRelations[clusterId], childId)
	}
}

func setIndirectChildren(filters ImpactFilters, childData, childAttrs, indirectChildren map[string]any,
	clusterRelations map[string][]string) {
	childId := childData["id"].(string)
	vconfig, hasVconfig := childAttrs["virtual_config"].(map[string]any)
	ptype, hasPtype := childAttrs["type"].(string)
	if pie.Contains(filters.Categories, childData["category"].(string)) ||
		(hasPtype && pie.Contains(filters.Ptypes, ptype)) ||
		(hasVconfig && reflect.TypeOf(vconfig["type"]).Kind() == reflect.String && pie.Contains(filters.Vtypes, vconfig["type"].(string))) {
		// indirect relation
		indirectChildren[childId] = childData
		// check if linked to a cluster
		setClusterRelation(childId, childAttrs, clusterRelations)
	}
}

func clusterRelationsToIndirect(filters ImpactFilters, clusterRelations map[string][]string, indirectChildren map[string]any, userRoles map[string]Role) *u.Error {
	if pie.Contains(filters.Vtypes, "application") {
		for clusterId := range clusterRelations {
			// check if linked cluster has apps
			entData, err := GetManyObjects("virtual_obj", bson.M{}, u.RequestFilters{}, "id="+clusterId+".**.*&"+"(virtual_config.type=kube-app|virtual_config.type=application)", userRoles)
			if err != nil {
				return err
			} else if len(entData) == 0 {
				// no apps, show only cluster
				indirectChildren[clusterId] = map[string]any{
					"category": "virtual_obj",
				}
			} else {
				// show apps
				for _, appData := range entData {
					indirectChildren[appData["id"].(string)] = appData
				}
			}
		}
	} else if pie.Contains(filters.Vtypes, "cluster") {
		for clusterId := range clusterRelations {
			// no apps, show only cluster
			indirectChildren[clusterId] = map[string]any{
				"category": "virtual_obj"}
		}
	}
	return nil
}
