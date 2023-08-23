package database

import (
	"arango-api/models"
	"encoding/json"
	h "net/http"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/gin-gonic/gin"
)

func GetDBConn(c *gin.Context) (*driver.Database, *models.ErrorMessage) {
	dbConn, ok := c.Value("database").(*driver.Database)
	//dbConn, ok := c.MustGet("database").(driver.Database)
	if !ok {
		return nil, &models.ErrorMessage{StatusCode: h.StatusInternalServerError, Message: "Failed to get database"}
	}
	if *dbConn == nil {
		return nil, &models.ErrorMessage{StatusCode: h.StatusNotFound, Message: "Failed to get database"}
	}
	return dbConn, nil
}

func ParseToString(obj interface{}) (string, *models.ErrorMessage) {

	asJson, err := json.Marshal(obj)
	if err != nil {
		return "", &models.ErrorMessage{StatusCode: h.StatusInternalServerError, Message: "Failed to parse query string"}
	}
	return string(asJson), nil

}

func ExecQuerry(db driver.Database, query string) ([]interface{}, *models.ErrorMessage) {

	var result []interface{}
	cursor, err := db.Query(nil, query, nil)

	if err != nil {
		return result, &models.ErrorMessage{StatusCode: h.StatusInternalServerError, Message: err.Error()}
	}

	defer cursor.Close()

	for {
		var doc interface{}
		_, err = cursor.ReadDocument(nil, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			return result, &models.ErrorMessage{StatusCode: h.StatusInternalServerError, Message: err.Error()}
		} else {
			result = append(result, doc)
		}
	}
	return result, nil

}

func GetAll(c *gin.Context, col string) ([]interface{}, *models.ErrorMessage) {
	db, err := GetDBConn(c)
	if err != nil {
		return nil, err
	}
	values := c.Request.URL.Query()

	querystring := "FOR doc IN " + col

	for key, value := range values {
		querystring += " FILTER doc." + key + " LIKE \"" + value[0] + "\" "
	}
	querystring += " RETURN doc"
	result, err := ExecQuerry(*db, querystring)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Delete(c *gin.Context, key, col string) ([]interface{}, *models.ErrorMessage) {

	db, err := GetDBConn(c)
	if err != nil {
		return nil, err
	}

	querystring := "FOR doc IN " + col + " FILTER doc.`_key`== \"" + key + "\" REMOVE doc IN " + col

	result, err := ExecQuerry(*db, querystring)
	if err != nil {
		return nil, err
	}
	return result, nil

}

func Update(c *gin.Context, doc interface{}, key, col string) ([]interface{}, *models.ErrorMessage) {
	db, err := GetDBConn(c)
	if err != nil {
		return nil, err
	}

	docStr, err := ParseToString(doc)
	if err != nil {
		return nil, err
	}
	querystring := "UPDATE \"" + key + "\" WITH " + docStr + " IN " + col + " RETURN " + key

	result, err := ExecQuerry(*db, querystring)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ConnectToArango(addr, database, user, password string) (driver.Database, *models.ErrorMessage) {

	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{addr},
	})
	if err != nil {
		return nil, &models.ErrorMessage{StatusCode: h.StatusBadRequest, Message: err.Error()}
	}
	client, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(user, password),
	})
	if err != nil {
		return nil, &models.ErrorMessage{StatusCode: h.StatusBadRequest, Message: err.Error()}
	}
	db, err := client.Database(nil, database)
	if err != nil {
		return nil, &models.ErrorMessage{StatusCode: h.StatusBadRequest, Message: err.Error()}
	}
	return db, nil

}

func CreateCollection(db driver.Database, collectionName string) (driver.Collection, error) {
	var col driver.Collection
	coll_exists, err := db.CollectionExists(nil, collectionName)

	if !coll_exists {
		col, err = db.CreateCollection(nil, collectionName, nil)

		if err != nil {
			return nil, err
		}
	}

	return col, nil

}
