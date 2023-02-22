package models

import (
	"context"
	"fmt"
	u "p3/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const WEB_PROJECTS = "web_project"

// Project represents data about a recorded web project
type Project struct {
	Id          string   `bson:"_id,omitempty"`
	Name        string   `json:"name" binding:"required"`
	DateRange   string   `json:"dateRange" binding:"required"`
	Namespace   string   `json:"namespace" binding:"required"`
	Attributes  []string `json:"attributes" binding:"required"`
	Objects     []string `json:"objects" binding:"required"`
	Permissions []string `json:"permissions" binding:"required,dive,email"`
	Author      string   `json:"authorLastUpdate" binding:"required"`
	LastUpdate  string   `json:"lastUpdate" binding:"required"`
	ShowAvg     bool     `json:"showAvg"`
	ShowSum     bool     `json:"showSum"`
	IsPublic    bool     `json:"isPublic"`
}

// PROJECTS
// GET
func GetProjectsByUserEmail(userEmail string) (map[string]interface{}, string) {
	response := make(map[string]interface{})
	response["projects"] = make([]interface{}, 0)
	println("Get projects for " + userEmail)

	// Get projects with user permitted
	var results []Project
	filter := bson.D{
		{Key: "$or",
			Value: bson.A{
				bson.D{{Key: "permissions", Value: userEmail}},
				bson.D{{Key: "isPublic", Value: true}},
			},
		},
	}
	ctx, cancel := u.Connect()
	cursor, err := GetDB().Collection(WEB_PROJECTS).Find(ctx, filter)
	if err != nil {
		fmt.Println(err)
	} else {
		// response["projects"], _ = m.ExtractCursor(cursor, ctx)
		if err = cursor.All(ctx, &results); err != nil {
			fmt.Println(err)
		} else if len(results) > 0 {
			response["projects"] = results
		}
	}

	defer cancel()

	return response, ""
}

// POST
func AddProject(newProject Project) string {
	// Add the new project
	ctx, cancel := u.Connect()
	_, err := GetDB().Collection(WEB_PROJECTS).InsertOne(ctx, newProject)
	if err != nil {
		println(err.Error())
		return err.Error()
	}

	defer cancel()
	return ""
}

// PUT
func UpdateProject(newProject Project, projectId string) string {
	// Update existing project, if exists
	ctx, cancel := u.Connect()
	objId, _ := primitive.ObjectIDFromHex(projectId)
	res, err := GetDB().Collection(WEB_PROJECTS).UpdateOne(ctx,
		bson.M{"_id": objId}, bson.M{"$set": newProject})
	defer cancel()

	if err != nil {
		return err.Error()
	}
	if res.MatchedCount <= 0 {
		return "No project found with this ID"
	}
	return ""
}

// DELETE
func DeleteProject(projectId string) string {
	println(projectId)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	objId, _ := primitive.ObjectIDFromHex(projectId)
	res, err := GetDB().Collection(WEB_PROJECTS).DeleteOne(ctx, bson.M{"_id": objId})
	defer cancel()

	if err != nil {
		return err.Error()
	} else if res.DeletedCount <= 0 {
		return "Project not found"
	}
	return ""
}

// IF USERS HAVE LIST OF PROJECTS
// CURRENTLY NOT USED
func getProjectsFromUser() {
	data := map[string]interface{}{}
	response := make(map[string]interface{})
	response["projects"] = make([]interface{}, 0)
	// Get query params
	userId := "test" //c.Query("userid")
	println("Get projects for " + userId)

	// Get user project ids
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		// c.JSON(http.StatusBadRequest, "Invalid user ID format: "+err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = GetDB().Collection("account").FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		// c.JSON(http.StatusBadRequest, "Unable to find user: "+err.Error())
	} else {
		println(data["web_projects"])
		projects, _ := data["web_projects"].(primitive.A)
		projectIds := []interface{}(projects)
		if len(projectIds) > 0 {
			// Convert IDs to good format
			println(projectIds)
			var objIds []primitive.ObjectID
			for _, id := range projectIds {
				println("ID " + id.(string))
				objId, err := primitive.ObjectIDFromHex(id.(string))
				println(err != nil)
				if err == nil {
					objIds = append(objIds, objId)
				}
			}
			// Get projects
			println(objIds)
			var results []Project
			cursor, err := GetDB().Collection("web_project").Find(ctx, bson.M{"_id": bson.M{"$in": objIds}})
			if err != nil {
				fmt.Println(err)
			} else {
				if err = cursor.All(ctx, &results); err != nil {
					fmt.Println(err)
				} else {
					response["projects"] = results
				}
			}
		}
	}

	defer cancel()

	// c.IndentedJSON(http.StatusOK, response)

	// FOR ADD
	// Add project id to users with permissions
	// addedPermissions := []string{}
	// for _, userEmail := range newProject.Permissions {
	// 	println(userEmail)
	// 	res, err := m.GetDB().Collection("account").UpdateOne(ctx,
	// 		bson.M{"email": userEmail}, bson.M{"$push": bson.M{"web_projects": result.InsertedID.(primitive.ObjectID).Hex()}})
	// 	if err == nil && res.MatchedCount > 0 {
	// 		addedPermissions = append(addedPermissions, userEmail)
	// 	}
	// }
	// newProject.Permissions = addedPermissions
}
