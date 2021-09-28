package models

//https://www.cockroachlabs.com/blog/upperdb-cockroachdb/
//https://www.cockroachlabs.com/docs/stable/build-a-go-app-with-cockroachdb-gorm.html
import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//Database
//var db *gorm.DB
var db *mongo.Database
var c *context.Context

func init() {

	e := godotenv.Load()

	if e != nil {
		fmt.Print(e)
	}

	//username := os.Getenv("db_user")
	//password := os.Getenv("db_pass")
	//dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")

	dbUri := fmt.Sprintf("mongodb://%s:%s/?readPreference=primary&ssl=false",
		dbHost, dbPort)

	fmt.Println(dbUri)

	client, err := mongo.NewClient(options.Client().ApplyURI(dbUri))
	if err != nil {
		log.Fatal(err)
		println("Error while generating client")
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
		println("Error while connecting")
	}
	//defer client.Disconnect(ctx)
	c = &ctx
	db = client.Database("ogree")
	if db == nil {
		println("Error while connecting")
	} else {
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			log.Fatal(err)
		} else {
			println("Successfully connected to DB")
		}

	}
}

func GetDB() *mongo.Database {
	return db
}

func GetCtx() context.Context {
	return *(c)
}
