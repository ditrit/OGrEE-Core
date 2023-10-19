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
	recreateTestDB()

	log.Println("database recreated")

	err := repository.ConnectToDB("", testDBPort, testDBUser, "123", TestDBName)
	if err != nil {
		log.Fatalln(err)
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
	if err != nil {
		log.Fatalln(err.Error())
	}

	db = client.Database(TestDBName)
	repository.SetupDB(db)
}
