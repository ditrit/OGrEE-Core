package models

import (
	"os"
	u "p3/utils"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

//JWT Claims struct
type Token struct {
	UserId   uint
	Customer string
	jwt.StandardClaims
}

//a struct for rep user account
type Account struct {
	ID       uint   ``
	Database string `json:"customer"`
	Email    string `json: "email"`
	Password string `json: "password"`
	Token    string `json:"token";sql:"-"`
}

//Validate incoming user
func (account *Account) Validate() (map[string]interface{}, bool) {
	if !strings.Contains(account.Email, "@") {
		return u.Message(false, "Email address is required"), false
	}

	if len(account.Password) < 6 {
		return u.Message(false, "Password is required"), false
	}

	//Email is unique
	temp := &Account{}

	//Error checking and duplicate emails
	ctx, cancel := u.Connect()
	//err := GetDB().Collection("accounts").FindOne(ctx, bson.M{"email": account.Email}).Decode(&temp) //.Where("email = ?", account.Email).First(temp).Error
	err := GetDBByName("ogree"+account.Database).Collection("account").FindOne(ctx, bson.M{"email": account.Email}).Decode(&temp)
	if err != nil && err != mongo.ErrNoDocuments {
		println("Error while creating account:", err.Error())
		return u.Message(false, "Connection error. Please retry"), false
	}
	defer cancel()
	return u.Message(false, "Requirement passed"), true
}

func (account *Account) Create() (map[string]interface{}, string) {

	if account.Database == "" || account.Database == "admin" ||
		account.Database == "config" || account.Database == "local" {
		account.Database = "ogreeDevelop"
	}

	if resp, ok := account.Validate(); !ok {
		return resp, ""
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword(
		[]byte(account.Password), bcrypt.DefaultCost)

	account.Password = string(hashedPassword)

	//If the customer/db doesn't exist let's create one
	if exists, _ := CheckIfDBExists("ogree" + account.Database); !exists {
		CreateTenantDB("ogree" + account.Database)

		customer := map[string]interface{}{"name": account.Database}

		//Update customer record
		ctx, cancel := u.Connect()
		_, e := GetDBByName("ogree").Collection("customer").InsertOne(ctx, customer)
		if e != nil {
			return u.Message(false,
				"Internal error while updating customer record: "+e.Error()), "internal"
		}
		defer cancel()
	}

	ctx, cancel := u.Connect()
	search := GetDBByName("ogree"+account.Database).Collection("account").FindOne(ctx, bson.M{"email": account.Email})
	if search.Err() != nil {
		GetDBByName("ogree"+account.Database).Collection("account").InsertOne(ctx, account)
	} else {
		return u.Message(false,
			"Error: User already exists:"), "clientError"
	}
	defer cancel()

	//Create new JWT token for the newly created account
	tk := &Token{UserId: account.ID, Customer: account.Database}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))

	account.Token = tokenString

	account.Password = ""

	response := u.Message(true, "Account has been created")
	response["account"] = account
	return response, ""
}

func Login(email, password, db string) (map[string]interface{}, string) {
	account := &Account{}

	if db == "" {
		return u.Message(false, "Error, please provide customer name in your request"), "clientError"
	}

	ctx, cancel := u.Connect()
	err := GetDBByName("ogree"+db).Collection("account").FindOne(ctx, bson.M{"email": email}).Decode(account)
	//err := GetDB().Collection("accounts").FindOne(ctx, bson.M{"email": email}).Decode(account)
	//err := GetDB().Table("account").Where("email = ?", email).First(account).Error
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return u.Message(false, "Error, email not found"), "internal"
		}
		return u.Message(false, "Connection error. Please try again later"),
			"internal"
	}
	defer cancel()

	//Should investigate if the password is sent in
	//cleartext over the wire
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))

	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return u.Message(false,
			"Invalid login credentials. Please try again"), "invalid"
	}

	//Success
	account.Password = ""

	//Create JWT token
	tk := &Token{UserId: account.ID, Customer: account.Database}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString

	resp := u.Message(true, "Logged In")
	resp["account"] = account
	return resp, ""
}

func GetUser(user int, db string) *Account {

	acc := &Account{}
	ctx, cancel := u.Connect()
	GetDBByName(db).Collection("account").FindOne(ctx, bson.M{"_id": user}).Decode(acc)
	if acc.Email == "" {
		return nil
	}
	defer cancel()

	acc.Password = ""
	return acc
}
