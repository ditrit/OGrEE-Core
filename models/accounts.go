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
	UserId uint
	jwt.StandardClaims
}

//a struct for rep user account
type Account struct {
	ID       uint
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
	err := GetDB().Collection("accounts").FindOne(GetCtx(), bson.M{"email": account.Email}).Decode(&temp) //.Where("email = ?", account.Email).First(temp).Error
	if err != nil && err != mongo.ErrNoDocuments {
		return u.Message(false, "Connection error. Please retry"), false
	}
	return u.Message(false, "Requirement passed"), true
}

func (account *Account) Create() map[string]interface{} {

	if resp, ok := account.Validate(); !ok {
		return resp
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword(
		[]byte(account.Password), bcrypt.DefaultCost)

	account.Password = string(hashedPassword)

	GetDB().Collection("accounts").InsertOne(GetCtx(), account)

	if account.ID <= 0 {
		return u.Message(false, "Failed to create account, connection error.")
	}

	//Create new JWT token for the newly created account
	tk := &Token{UserId: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))

	account.Token = tokenString

	account.Password = ""

	response := u.Message(true, "Account has been created")
	response["account"] = account
	return response
}

func Login(email, password string) (map[string]interface{}, string) {
	account := &Account{}

	err := GetDB().Collection("accounts").FindOne(GetCtx(), bson.M{"email": email}).Decode(account)
	//err := GetDB().Table("account").Where("email = ?", email).First(account).Error
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return u.Message(false, "Error, email not found"), "internal"
		}
		return u.Message(false, "Connection error. Please try again later"),
			"internal"
	}

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
	tk := &Token{UserId: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString

	resp := u.Message(true, "Logged In")
	resp["account"] = account
	return resp, ""
}

func GetUser(u int) *Account {

	acc := &Account{}
	GetDB().Collection("accounts").FindOne(GetCtx(), bson.M{"_id": u}).Decode(acc)
	if acc.Email == "" {
		return nil
	}

	acc.Password = ""
	return acc
}
