package models

import (
	"fmt"
	"os"
	u "p3/utils"
	"regexp"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// JWT Claims struct
type Token struct {
	Email  string             `json:"email"`
	UserId primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	jwt.StandardClaims
}

// a struct for rep user account
type Account struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
	Roles    map[string]string  `bson:"roles" json:"roles"`
	Token    string             `bson:"token,omitempty" json:"token,omitempty"`
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
	err := GetDB().Collection("account").FindOne(ctx, bson.M{"email": account.Email}).Err()
	if err != nil && err != mongo.ErrNoDocuments {
		println("Error while creating account:", err.Error())
		return u.Message(false, "Connection error. Please retry"), false
	}

	//User already exists
	if err == nil {
		return u.Message(false, "Error: User already exists"), false
	}
	defer cancel()

	// Validate domains and roles
	if len(account.Roles) <= 0 {
		return u.Message(false, "Object 'roles' with domain names as keys and roles as values is mandatory"), false
	}
	for domain, role := range account.Roles {
		if !CheckDomainExists(domain) {
			return u.Message(false, "Domain does not exist: "+domain), false
		}
		switch role {
		case Manager:
		case Viewer:
		case User:
			break
		default:
			return u.Message(false, "Role assigned is not valid: "+role), false
		}
	}

	return u.Message(false, "Requirement passed"), true
}

func (account *Account) Create(callerRoles map[string]string) (map[string]interface{}, string) {
	// Check if user is allowed to create new users
	if !CheckCanCreateUser(callerRoles, account.Roles) {
		return u.Message(false,
			"Invalid credentials for creating an account."+
				"Manager role in requested domains is needed"), "unauthorised"
	}

	// Validate new user
	if resp, ok := account.Validate(); !ok {
		return resp, "validate"
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword(
		[]byte(account.Password), bcrypt.DefaultCost)

	account.Password = string(hashedPassword)

	ctx, cancel := u.Connect()
	res, err := GetDB().Collection("account").InsertOne(ctx, account)
	if err != nil {
		return u.Message(false,
			"DB error when creating user: "+err.Error()), "internal"
	} else {
		account.ID = res.InsertedID.(primitive.ObjectID)
	}
	defer cancel()

	//Create new JWT token for the newly created account
	tk := &Token{Email: account.Email, UserId: account.ID}
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
	resp := u.Message(true, "Logged In")

	ctx, cancel := u.Connect()
	err := GetDB().Collection("account").FindOne(ctx, bson.M{"email": email}).Decode(account)
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
	} else if err == bcrypt.ErrHashTooShort {
		if account.Email == "admin" &&
			account.Password == password && account.Password == "admin" {
			resp["shouldChange"] = true
		} else {
			return u.Message(false,
				"Invalid login credentials. Please try again"), "invalid"
		}
	}

	//Success
	account.Password = ""

	//Create JWT token
	tk := &Token{Email: account.Email, UserId: account.ID}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	account.Token = tokenString

	resp["account"] = account
	return resp, ""
}

func GetUser(userId primitive.ObjectID) *Account {
	acc := &Account{}
	ctx, cancel := u.Connect()
	fmt.Println(userId)
	err := GetDB().Collection("account").FindOne(ctx, bson.M{"_id": userId}).Decode(acc)
	if err != nil || acc.Email == "" {
		return nil
	}
	defer cancel()

	acc.Password = ""
	return acc
}

func GetAllUsers() ([]Account, string) {
	ctx, cancel := u.Connect()
	c, err := GetDB().Collection("account").Find(ctx, bson.M{})
	if err != nil {
		println(err.Error())
		return nil, err.Error()
	}
	users := []Account{}
	err = c.All(ctx, &users)
	if err != nil {
		println(err.Error())
		return nil, err.Error()
	}

	defer cancel()
	return users, ""
}

func DeleteUser(id string) string {
	ctx, cancel := u.Connect()
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "User ID not valid"
	}
	req := bson.M{"_id": objID}
	c, _ := GetDB().Collection("account").DeleteOne(ctx, req)
	if c.DeletedCount == 0 {
		return "Internal error try to delete user"
	}
	defer cancel()
	return ""
}
