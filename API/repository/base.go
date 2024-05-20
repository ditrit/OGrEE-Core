package repository

import (
	"context"
	"fmt"
	"net/url"
	u "p3/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Database
var globalDB *mongo.Database
var globalClient *mongo.Client

func ConnectToDB(host, port, user, pass, dbName, tenantName string) error {
	client, err := ConnectToMongo(host, port, user, pass, dbName)
	if err != nil {
		return err
	}

	globalClient = client

	db, err := GetDatabase(client, dbName)
	if err != nil {
		return err
	}

	globalDB = db

	err = SetupDB(db)
	if err != nil {
		return err
	}

	err = createInitialData(db, tenantName)
	if err != nil {
		return err
	}

	return nil
}

func SetupDB(db *mongo.Database) error {
	// Indexes creation
	// Enforce unique children
	for _, entity := range []int{u.DOMAIN, u.SITE, u.BLDG, u.ROOM, u.RACK, u.DEVICE, u.AC, u.PWRPNL, u.CABINET, u.CORRIDOR, u.GROUP, u.STRAYOBJ, u.GENERIC} {
		if err := createUniqueIndex(db, u.EntityToString(entity), bson.M{"id": 1}); err != nil {
			return err
		}
	}

	// Make slugs unique identifiers for templates
	for _, entity := range []int{u.ROOMTMPL, u.OBJTMPL, u.BLDGTMPL, u.TAG, u.LAYER} {
		if err := createUniqueIndex(db, u.EntityToString(entity), bson.M{"slug": 1}); err != nil {
			return err
		}
	}

	if err := createUniqueIndex(db, "account", bson.M{"email": 1}); err != nil {
		return err
	}
	if err := createUniqueIndex(db, "application", bson.M{"name": 1}); err != nil {
		return err
	}

	return nil
}

// Initial data creation
func createInitialData(db *mongo.Database, tenantName string) error {
	// Create a default domain
	ctx, cancel := u.Connect()
	defer cancel()

	_, err := CreateObject(ctx, u.EntityToString(u.DOMAIN), map[string]any{
		"id":          tenantName,
		"name":        tenantName,
		"category":    "domain",
		"description": "",
		"attributes": map[string]any{
			"color": "ffffff",
		},
	})
	if err != nil && err.Type != u.ErrDuplicate {
		return err
	}

	return nil
}

func createUniqueIndex(db *mongo.Database, collection string, on bson.M) error {
	indexCtx, indexCancel := u.Connect()
	defer indexCancel()

	_, err := db.Collection(collection).Indexes().CreateOne(
		indexCtx,
		mongo.IndexModel{
			Keys:    on,
			Options: options.Index().SetUnique(true),
		},
	)

	return err
}

func GetDatabase(client *mongo.Client, name string) (*mongo.Database, error) {
	if name == "admin" || name == "config" || name == "local" {
		return nil, fmt.Errorf("database %s not accessible", name)
	}

	// Check if API is authenticated
	if exists := databaseExists(client, name); !exists {
		return nil, fmt.Errorf("database %s not found, check that you are authorized", name)
	}

	db := client.Database(name)
	if db == nil {
		return nil, fmt.Errorf("error while getting database %s", name)
	}

	return db, nil
}

func databaseExists(client *mongo.Client, name string) bool {
	databaseList, err := client.ListDatabaseNames(context.Background(), bson.D{})
	if err != nil {
		return false
	}

	for _, databaseName := range databaseList {
		if databaseName == name {
			return true
		}
	}

	return false
}

func ConnectToMongo(host, port, user, pass, authDB string) (*mongo.Client, error) {
	params := "readPreference=primary"

	if host == "" {
		host = "localhost"
		params = params + "&directConnection=true"
	}

	if port == "" {
		port = "27017"
	}

	var dbUri string

	if user == "" || pass == "" {
		params = params + "&ssl=false"
		dbUri = fmt.Sprintf(
			"mongodb://%s:%s/?%s",
			host, port,
			params,
		)
	} else {
		dbUri = fmt.Sprintf(
			"mongodb://%s:%s@%s:%s/%s?%s",
			user, url.QueryEscape(pass),
			host, port, authDB,
			params,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(
		ctx,
		options.Client().ApplyURI(dbUri),
	)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetDB() *mongo.Database {
	return globalDB
}

func GetClient() *mongo.Client {
	return globalClient
}
