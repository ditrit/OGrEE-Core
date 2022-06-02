package models

//https://www.cockroachlabs.com/blog/upperdb-cockroachdb/
//https://www.cockroachlabs.com/docs/stable/build-a-go-app-with-cockroachdb-gorm.html
import (
	"context"
	"fmt"
	"log"
	"os"
	u "p3/utils"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

//Database
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
	dbName := os.Getenv("db")

	if user == "" || pass == "" {
		println("USER:", user)
		println("PASS:", pass)
		dbUri = fmt.Sprintf("mongodb://%s:%s/?readPreference=primary&ssl=false",
			dbHost, dbPort)
	} else {
		dbUri = fmt.Sprintf("mongodb://%s:%s@%s:%s/?readPreference=primary",
			user, pass, dbHost, dbPort)
	}

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

	return false, e

}

//This function shall execute the same
//commands as createdb.js found in the
//root dir of this API
func CreateTenantDB(name string) {
	ctx, cancel := u.Connect()
	newDB := GetDB().Client().Database(name, nil)
	defer cancel()
	//TODO
	//we can move the schema validation to the DB
	//options.CreateCollectionOptions{}
	newDB.CreateCollection(ctx, "account", nil)
	newDB.CreateCollection(ctx, "domain", nil)
	newDB.CreateCollection(ctx, "site", nil)
	newDB.CreateCollection(ctx, "building", nil)
	newDB.CreateCollection(ctx, "room", nil)
	newDB.CreateCollection(ctx, "rack", nil)
	newDB.CreateCollection(ctx, "device", nil)

	//Template Collections
	newDB.CreateCollection(ctx, "room_template")
	newDB.CreateCollection(ctx, "obj_template")

	//Group Collections
	newDB.CreateCollection(ctx, "group")

	//Nonhierarchal objects
	newDB.CreateCollection(ctx, "ac")
	newDB.CreateCollection(ctx, "panel")
	newDB.CreateCollection(ctx, "separator")
	newDB.CreateCollection(ctx, "row")
	newDB.CreateCollection(ctx, "tile")
	newDB.CreateCollection(ctx, "cabinet")
	newDB.CreateCollection(ctx, "corridor")

	//Sensors
	newDB.CreateCollection(ctx, "sensor")

	//Stray Objects
	newDB.CreateCollection(ctx, "stray_device")
	newDB.CreateCollection(ctx, "stray_sensor")

	//Create Index variables
	d := bsonx.Doc{{Key: "parentId", Value: bsonx.Int32(1)},
		{Key: "name", Value: bsonx.Int32(1)}}

	sd := bsonx.Doc{{Key: "parentId", Value: bsonx.Int32(1)},
		{Key: "name", Value: bsonx.Int32(1)},
		{Key: "type", Value: bsonx.Int32(1)}}

	genericIdx := mongo.IndexModel{Keys: d, Options: options.Index().SetUnique(true)}
	templateIdx := mongo.IndexModel{Keys: bson.M{"slug": 1}, Options: options.Index().SetUnique(true)}
	sensorIdx := mongo.IndexModel{Keys: sd, Options: options.Index().SetUnique(true)}

	//Setup Indexes
	newDB.Collection("domain").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("site").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("building").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("room").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("rack").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("device").Indexes().CreateOne(ctx, genericIdx)

	newDB.Collection("room_template").Indexes().CreateOne(ctx, templateIdx)
	newDB.Collection("obj_template").Indexes().CreateOne(ctx, templateIdx)

	newDB.Collection("sensor").Indexes().CreateOne(ctx, sensorIdx)

	newDB.Collection("ac").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("panel").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("separator").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("row").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("tile").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("cabinet").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("corridor").Indexes().CreateOne(ctx, genericIdx)

	newDB.Collection("group").Indexes().CreateOne(ctx, genericIdx)

	newDB.Collection("stray_device").Indexes().CreateOne(ctx, genericIdx)
	newDB.Collection("stray_device").Indexes().CreateOne(ctx, genericIdx)
}
