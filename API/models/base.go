package models

//https://www.cockroachlabs.com/blog/upperdb-cockroachdb/
//https://www.cockroachlabs.com/docs/stable/build-a-go-app-with-cockroachdb-gorm.html
import (
	"context"
	"fmt"
	"log"
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
		fmt.Print(e)
	}

	var dbUri string

	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")
	user := os.Getenv("db_user")
	pass := os.Getenv("db_pass")
	dbName := "ogree" + os.Getenv("db")
	if strings.HasSuffix(os.Args[0], ".test") {
		dbName = "autoTest"
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
		dbUri = fmt.Sprintf("mongodb://%s:%s/?readPreference=primary&ssl=false",
			dbHost, dbPort)
	} else {
		dbUri = fmt.Sprintf("mongodb://ogree%sAdmin:%s@%s:%s/%s?readPreference=primary&authSource=%s",
			user, pass, dbHost, dbPort, dbName, dbName)
	}

	fmt.Println(dbUri)

	client, err := mongo.NewClient(options.Client().ApplyURI(dbUri))
	if err != nil {
		log.Fatal(err)
		println("Error while generating client")
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		println("Error while connecting")
	}
	//defer client.Disconnect(ctx)
	if dbName != "" {
		db = client.Database(dbName)
	} else {
		db = client.Database("ogree")
	}

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

func CheckIfDBExists(name string) (bool, error) {
	//options.ListDatabasesOptions{}
	if name == "admin" || name == "config" || name == "local" {
		return false, nil
	}

	ldr, e := GetDB().Client().ListDatabaseNames(context.TODO(), bson.D{{}})
	if e == nil {
		for i := range ldr {
			if ldr[i] == name {
				return true, nil
			}
		}
	}
	//`GetDB().Client().

	return false, e

}
