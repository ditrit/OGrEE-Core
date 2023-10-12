package models

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Database
var db *mongo.Database
var globalClient *mongo.Client

func init() {
	e := godotenv.Load()
	if e != nil {
		fmt.Println(e)
	}

	var dbUri string

	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")
	user := os.Getenv("db_user")
	pass := os.Getenv("db_pass")
	dbName := "ogree" + os.Getenv("db")
	params := "readPreference=primary"
	if strings.HasSuffix(os.Args[0], ".test") {
		dbName = "ogreeAutoTest"
		user = "AutoTest"
		pass = "123"
		dbPort = "27018"
		params = params + "&directConnection=true"
	}

	println("USER:", user)
	println("DB:", dbName)

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "27017"
	}

	if user == "" || pass == "" {
		params = params + "&ssl=false"
		dbUri = fmt.Sprintf(
			"mongodb://%s:%s/?%s",
			dbHost, dbPort,
			params,
		)
	} else {
		params = params + fmt.Sprintf("&authSource=%s", dbName)
		dbUri = fmt.Sprintf("mongodb://ogree%sAdmin:%s@%s:%s/%s?%s",
			user, url.QueryEscape(pass), dbHost, dbPort, dbName, params)
	}

	fmt.Println(dbUri)

	client, err := mongo.NewClient(options.Client().ApplyURI(dbUri))
	if err != nil {
		println("Error while generating client")
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		println("Error while connecting")
		log.Fatal(err)
	}

	//Check if API is authenticated
	if found, err1 := CheckIfDBExists(dbName, client); !found || err1 != nil {
		if err1 != nil {
			if strings.Contains(err1.Error(), "listDatabases requires authentication") {
				log.Fatal("Error! Authentication failed")
			}
			log.Fatal(err1.Error())
		}
		log.Fatal("Target DB not found. Please check that you are authorized")
	}

	//defer client.Disconnect(ctx)
	db = client.Database(dbName)

	if db == nil {
		println("Error while connecting")
	} else {
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.Fatal(err)
		} else {
			println("Successfully connected to DB")
		}
		globalClient = client
	}
}

func GetDB() *mongo.Database {
	return db
}

func GetClient() *mongo.Client {
	return globalClient
}

func GetDBByName(name string) *mongo.Database {
	return GetClient().Database(name)
}

func CheckIfDBExists(name string, client *mongo.Client) (bool, error) {
	//options.ListDatabasesOptions{}
	if name == "admin" || name == "config" || name == "local" {
		return false, nil
	}

	ldr, e := client.ListDatabaseNames(context.TODO(), bson.D{{}})
	if e == nil {
		for i := range ldr {
			if ldr[i] == name {
				return true, nil
			}
		}
	}

	return false, e
}
