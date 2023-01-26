package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var OGREE_URL string
var OGREE_TOKEN string

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
	router.GET("/hierarchy", getPhysicalHierarchy)
	router.GET("/attributes/all", getAllPhysicalAttributes)
	router.GET("/attributes", getPhysicalAttributesById)

	router.Run("localhost:8080")
}

// ATTRIBUTES

func getAllPhysicalAttributes(c *gin.Context) {
	attributes := make(map[string]interface{})

	// Get all sites
	data := getFromAPI("/api/sites")

	// Get data for each site
	println(data["objects"])
	for _, site := range data["objects"].([]interface{}) {
		name := (site.(map[string]interface{}))["name"].(string)
		id := (site.(map[string]interface{}))["id"].(string)
		// execution tip: remove the query params (after ?) if API not in special rbac branch
		data = getFromAPI("/api/sites/" + id + "/all?field=name&field=attributes")
		fmt.Println(data)

		getAllChildrenAttrs(data, &attributes, name)
		fmt.Println("### getAllChildrenAttrs ###")
		bs, _ := json.Marshal(attributes)
		fmt.Println(string(bs))
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
			bs, _ := json.Marshal(attributes)
			fmt.Println(string(bs))
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
	println(data["objects"])
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
	fmt.Printf("%s", data["data"])
	return data["data"].(map[string]interface{})
}
