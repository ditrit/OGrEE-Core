package integration

import (
	"log"
	"p3/repository"
	"p3/utils"
)

const testDBPort = "27018"
const TestDBName = "ogreeAutoTest"
const testDBUser = TestDBName + "Admin"

func init() {
	log.Println("------------------ejecutando integration init-----------------------")
	recreateTestDB()

	log.Println("database recreated: ", TestDBName)

	err := repository.ConnectToDB("", testDBPort, testDBUser, "123", TestDBName, TestDBName)
	if err != nil {
		log.Fatalln("Error connecting to", TestDBName, ":", err.Error())
	}
}

func recreateTestDB() {
	client, err := repository.ConnectToMongo("", testDBPort, "admin", "adminpassword", "admin")
	if err != nil {
		log.Fatalln(err.Error())
	}

	db := client.Database(TestDBName)

	ctx, _ := utils.Connect()

	err = db.Drop(ctx)
	for err != nil {
		log.Println("Error while doing drop:", err.Error())
		err = db.Drop(ctx)
	}

	db = client.Database(TestDBName)

	err = repository.SetupDB(db)
	for err != nil {
		// An error can occur if the database was not dropped yet (not synchronic)
		log.Println("Error while doing setup:", err.Error())
		err = repository.SetupDB(db)
	}
}
