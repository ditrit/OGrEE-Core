package models

import (
	"os"
	u "p3/utils"
	"regexp"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// JWT Claims struct
type Token struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// a struct for rep user account
type Account struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token" sql:"-"`
}

// Validate incoming user
func (account *Account) Validate() (map[string]interface{}, bool) {
	valid := regexp.MustCompile("(\\w)+@(\\w)+\\.(\\w)+").MatchString(account.Email)

	if !valid {
		return u.Message(false, "A valid email address is required"), false
	}

	if len(account.Password) < 7 {
		return u.Message(false,
			"Please provide a Password with a length greater than 6"), false
	}

	//Error checking and duplicate emails
	ctx, cancel := u.Connect()
	err := GetDB().Collection("account").FindOne(ctx, bson.M{"email": account.Email}).Err() //.Where("email = ?", account.Email).First(temp).Error
	if err != nil && err != mongo.ErrNoDocuments {
		return u.Message(false, "Connection error : "+err.Error()), false
	}
	//User already exists
	if err == nil {
		return u.Message(false, "Error: User already exists"), false
	}
	defer cancel()
	return u.Message(false, "Requirement passed"), true
}

func (account *Account) Create() (map[string]interface{}, string) {

	if resp, ok := account.Validate(); !ok {
		return resp, "exists"
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword(
		[]byte(account.Password), bcrypt.DefaultCost)

	account.Password = string(hashedPassword)

	ctx, cancel := u.Connect()
	_, e := GetDB().Collection("account").InsertOne(ctx, account)
	if e != nil {
		return u.Message(false, "Connection error please retry again later"), "internal"
	}

	defer cancel()

	//Create new JWT token for the newly created account
	tk := &Token{Email: account.Email}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))

	account.Token = tokenString

	account.Password = ""

	response := u.Message(true, "Account has been created")
	response["account"] = account
	return response, ""
}

func Login(email, password string) (map[string]interface{}, string) {
	account := &Account{}

	ctx, cancel := u.Connect()
	err := GetDB().Collection("account").FindOne(ctx, bson.M{"email": email}).Decode(account)
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
	tk := &Token{Email: account.Email}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString

	resp := u.Message(true, "Logged In")
	resp["account"] = account
	return resp, ""
}

func GetUser(user int) *Account {

	acc := &Account{}
	ctx, cancel := u.Connect()
	GetDB().Collection("account").FindOne(ctx, bson.M{"_id": user}).Decode(acc)
	if acc.Email == "" {
		return nil
	}
	defer cancel()

	acc.Password = ""
	return acc
}
