package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	m "ogree_app_backend/models"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var OGREE_URL string
var OGREE_TOKEN string

const WEB_PROJECTS = "web_project"

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	OGREE_URL = os.Getenv("OGREE_URL")
	OGREE_TOKEN = os.Getenv("OGREE_TOKEN")
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/projects", getProjectsByUserEmail)
	router.POST("/projects", addProject)
	router.PUT("/projects/:id", updateProject)
	router.DELETE("/projects/:id", deleteProject)

	router.GET("/hierarchy", getPhysicalHierarchy)
	router.GET("/attributes/all", getAllPhysicalAttributes)
	router.GET("/attributes", getPhysicalAttributesById)

	router.Run(":8080")
}

// PROJECTS

// project represents data about a recorded web project.
type project struct {
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

func getProjectsByUserEmail(c *gin.Context) {
	data := map[string]interface{}{}
	response := make(map[string]interface{})
	response["projects"] = make([]interface{}, 0)
	// Get query params
	userId := c.Query("userid")
	println("Get projects for " + userId)

	// Get user's email
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid user ID format: "+err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = m.GetDB().Collection("account").FindOne(ctx, bson.M{"_id": objId}).Decode(&data)

	if err != nil {
		c.JSON(http.StatusBadRequest, "Unable to find user: "+err.Error())
	} else {
		// Get projects with user permitted
		var results []project
		filter := bson.D{
			{Key: "$or",
				Value: bson.A{
					bson.D{{Key: "permissions", Value: data["email"]}},
					bson.D{{Key: "isPublic", Value: true}},
				},
			},
		}
		cursor, err := m.GetDB().Collection(WEB_PROJECTS).Find(ctx, filter)
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
	}

	defer cancel()

	c.IndentedJSON(http.StatusOK, response)
}

// IF USERS HAVE LIST OF PROJECTS
func getProjectsByUserId(c *gin.Context) {
	data := map[string]interface{}{}
	response := make(map[string]interface{})
	response["projects"] = make([]interface{}, 0)
	// Get query params
	userId := c.Query("userid")
	println("Get projects for " + userId)

	// Get user project ids
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid user ID format: "+err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	err = m.GetDB().Collection("account").FindOne(ctx, bson.M{"_id": objId}).Decode(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Unable to find user: "+err.Error())
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
			var results []project
			cursor, err := m.GetDB().Collection("web_project").Find(ctx, bson.M{"_id": bson.M{"$in": objIds}})
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

	c.IndentedJSON(http.StatusOK, response)
}

func addProject(c *gin.Context) {
	var newProject project

	// Call BindJSON to bind the received JSON to newProject
	if err := c.BindJSON(&newProject); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Add the new project
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	_, err := m.GetDB().Collection(WEB_PROJECTS).InsertOne(ctx, newProject)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		defer cancel()
		return
	}

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

	defer cancel()
	c.IndentedJSON(http.StatusCreated, newProject)
}

func updateProject(c *gin.Context) {
	var newProject project
	id := c.Param("id")
	println(id)

	// Call BindJSON to bind the received JSON to newProject
	if err := c.BindJSON(&newProject); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Update existing project, if exists
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	objId, _ := primitive.ObjectIDFromHex(id)
	res, err := m.GetDB().Collection(WEB_PROJECTS).UpdateOne(ctx,
		bson.M{"_id": objId}, bson.M{"$set": newProject})
	defer cancel()
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	if res.MatchedCount <= 0 {
		c.IndentedJSON(http.StatusNotFound, "No project found with this ID")
		return
	}

	c.IndentedJSON(http.StatusOK, newProject)
}

func deleteProject(c *gin.Context) {
	id := c.Param("id")
	println(id)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	objId, _ := primitive.ObjectIDFromHex(id)
	res, err := m.GetDB().Collection(WEB_PROJECTS).DeleteOne(ctx, bson.M{"_id": objId})
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	} else if res.DeletedCount <= 0 {
		c.IndentedJSON(http.StatusNotFound, "Project not found")
	} else {
		c.IndentedJSON(http.StatusOK, "")
	}
	defer cancel()
}

// ATTRIBUTES

func getAllPhysicalAttributes(c *gin.Context) {
	attributes := make(map[string]interface{})

	// Get all sites
	data := getFromAPI("/api/sites")

	// Get data for each site
	// println(data["objects"])
	for _, site := range data["objects"].([]interface{}) {
		name := (site.(map[string]interface{}))["name"].(string)
		id := (site.(map[string]interface{}))["id"].(string)
		// execution tip: remove the query params (after ?) if API not in special rbac branch
		data = getFromAPI("/api/sites/" + id + "/all") //?field=name&field=attributes")
		// fmt.Println(data)

		getAllChildrenAttrs(data, &attributes, name)
		fmt.Println("### getAllChildrenAttrs ###")
		// bs, _ := json.Marshal(attributes)
		// fmt.Println(string(bs))
	}

	c.IndentedJSON(http.StatusOK, attributes)
}

func getAllChildrenAttrs(data map[string]interface{}, attributes *map[string]interface{}, currentName string) {
	if data["attributes"] != nil {
		(*attributes)[currentName] = data["attributes"]
	}

	if data["children"] != nil {
		for _, u := range data["children"].([]interface{}) {
			child := u.(map[string]interface{})
			name := currentName + "." + child["name"].(string)
			getAllChildrenAttrs(child, attributes, name)
		}
	}
}

func getPhysicalAttributesById(c *gin.Context) {
	attributes := make(map[string]interface{})
	// Get query params
	category := c.Query("category")
	id := c.Query("id")

	if category != "" && id != "" {
		data := getFromAPI("/api/" + category + "/" + id)
		if data["name"] != nil && data["attributes"] != nil {
			attributes[data["name"].(string)] = data["attributes"]
			fmt.Println("### ATTRIBUTES ###")
			// bs, _ := json.Marshal(attributes)
			// fmt.Println(string(bs))
			c.IndentedJSON(http.StatusOK, attributes)
		} else {
			c.JSON(http.StatusInternalServerError, "Error getting object from OGrEE-API")
		}
	} else {
		c.JSON(http.StatusBadRequest, "Query params id and category not found")
	}
}

// HIERARCHY

func getPhysicalHierarchy(c *gin.Context) {
	hierarchy := make(map[string][]string)
	categories := make(map[string][]string)
	categories["KeysOrder"] = []string{"site", "building", "room", "rack"}
	hierarchy["Root"] = []string{}

	// Get all sites
	data := getFromAPI("/api/sites")

	// Get data for each site
	// println(data["objects"])
	for _, site := range data["objects"].([]interface{}) {
		name := (site.(map[string]interface{}))["name"].(string)
		id := (site.(map[string]interface{}))["id"].(string)
		data = getFromAPI("/api/sites/" + id + "/all")

		hierarchy["Root"] = append(hierarchy["Root"], name)
		hierarchy[name] = getChildren(data, &hierarchy, &categories, name)
		// bs, _ := json.Marshal(hierarchy)
		//fmt.Println(string(bs))
		// bs, _ = json.Marshal(categories)
		// fmt.Println(string(bs))
	}

	response := make(map[string]interface{})
	response["tree"] = hierarchy
	response["categories"] = categories

	c.IndentedJSON(http.StatusOK, response)
}

func getChildren(data map[string]interface{}, hierarchy *map[string][]string,
	categories *map[string][]string, prefix string) []string {
	children := []string{}

	// Create a list of object names for each category
	if data["category"] != nil {
		if (*categories)[data["category"].(string)] != nil {
			(*categories)[data["category"].(string)] = append((*categories)[data["category"].(string)], prefix)
		} else {
			(*categories)[data["category"].(string)] = []string{prefix}
		}
	}

	// Create a list of children names for each object
	if data["children"] != nil {
		for _, u := range data["children"].([]interface{}) {
			// add each child
			child := u.(map[string]interface{})
			children = append(children, child["name"].(string))
			childName := prefix + "." + child["name"].(string) // unique name

			// call recursively for each child's children
			grandchildren := []string{}
			grandchildren = getChildren(child, hierarchy, categories, childName)
			if len(grandchildren) > 0 {
				(*hierarchy)[childName] = grandchildren
			}
		}
	}

	return children
}

// OGrEE API

func getFromAPI(path string) map[string]interface{} {
	token := "Bearer " + OGREE_TOKEN
	client := &http.Client{}
	req, _ := http.NewRequest("GET", OGREE_URL+path, nil)
	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%s", data["data"])
	return data["data"].(map[string]interface{})
}
