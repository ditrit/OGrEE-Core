package models

//https://www.cockroachlabs.com/blog/upperdb-cockroachdb/
//https://www.cockroachlabs.com/docs/stable/build-a-go-app-with-cockroachdb-gorm.html
import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	/*clientOptions := options.Client().ApplyURI(dbUri)
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
		println("Error while connecting")
	}

	fmt.Println("Connected to MongoDB!")

	//conn, err := gorm.Open("postgres", dbUri)
	println("GOT HERE 2")
	if err != nil {
		fmt.Println("FAILURE!!!")
		fmt.Print(err)
	}

	fmt.Print(err)
	db = client.Database("ogree")*/

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
		println("Printing collection names")
		x, _ := db.ListCollectionNames(GetCtx(), bson.M{})
		println(len(x))
		if len(x) == 0 {
			println("ERROR!")
			os.Exit(-1)
		}

	}
	//db.Debug().SingularTable(true)
}

// ConnectDB : This is helper function to connect mongoDB
/*func ConnectDB() *mongo.Collection {

	// Set client options
	clientOptions := options.Client().ApplyURI("your_cluster_endpoint")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("ogree").Collection("books")

	return collection
}*/

func GetDB() *mongo.Database {
	return db
}

func GetCtx() context.Context {
	return *(c)
}
