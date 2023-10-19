package e2e

import (
	"log"
	"p3/models"
	_ "p3/test/integration"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	createAdminAccount()
}

var AdminId primitive.ObjectID
var AdminToken string

func createAdminAccount() {
	// Create admin account
	admin := models.Account{}
	admin.Email = "admin@admin.com"
	admin.Password = "admin123"
	admin.Roles = map[string]models.Role{"*": "manager"}

	newAcc, err := admin.Create(map[string]models.Role{"*": "manager"})
	if err != nil {
		log.Fatal(err.Error())
	}

	if newAcc != nil {
		AdminId = newAcc.ID
		AdminToken = newAcc.Token
	}
}
