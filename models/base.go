package models

//https://www.cockroachlabs.com/blog/upperdb-cockroachdb/
//https://www.cockroachlabs.com/docs/stable/build-a-go-app-with-cockroachdb-gorm.html
import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

//Database
var db *gorm.DB

func init() {

	e := godotenv.Load()

	if e != nil {
		fmt.Print(e)
	}

	username := os.Getenv("db_user")
	//password := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")

	dbUri := fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=disable",
		username, dbHost, dbPort, dbName)

	fmt.Println(dbUri)

	conn, err := gorm.Open("postgres", dbUri)
	println("GOT HERE 2")
	if err != nil {
		fmt.Println("FAILURE!!!")
		fmt.Print(err)
	}

	fmt.Print(err)
	db = conn
	db.Debug().SingularTable(true)
}

func GetDB() *gorm.DB {
	return db
}
