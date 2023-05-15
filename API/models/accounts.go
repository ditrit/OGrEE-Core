package models

import (
	"os"
	u "p3/utils"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

const token_expiration = time.Hour * 72

// JWT Claims struct
type Token struct {
	Email  string             `json:"email"`
	UserId primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	jwt.StandardClaims
}

// a struct for rep user account
type Account struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name     string             `bson:"name" json:"name"`
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

	if e := validatePasswordFormat(account.Password); e != "" {
		return u.Message(false, e), false
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
	if e := validateDomainRoles(account.Roles); e != "" {
		return u.Message(false, e), false
	}

	return u.Message(false, "Requirement passed"), true
}

func validateDomainRoles(roles map[string]string) string {
	// Validate domains and roles
	if len(roles) <= 0 {
		return "Object 'roles' with domain names as keys and roles as values is mandatory"
	}
	for domain, role := range roles {
		if !CheckDomainExists(domain) {
			return "Domain does not exist: " + domain
		}
		switch role {
		case Manager:
		case Viewer:
		case User:
			break
		default:
			return "Role assigned is not valid: "
		}
	}
	return ""
}

func (account *Account) Create(callerRoles map[string]string) (map[string]interface{}, string) {
	// Check if user is allowed to create new users
	if !CheckCanManageUser(callerRoles, account.Roles) {
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
	account.Token = GenerateToken(account.Email, account.ID, token_expiration)
	account.Password = ""

	response := u.Message(true, "Account has been created")
	response["account"] = account
	return response, ""
}

func validatePasswordFormat(password string) string {
	if len(password) < 7 {
		return "Please provide a password with a length greater than 6"
	}
	return ""
}

func comparePasswordToAccount(account Account, inputPassword string) (string, string) {
	println(inputPassword)
	println(account.Password)
	err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(inputPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return "Password is not correct", "validate"
	} else if err == bcrypt.ErrHashTooShort {
		if account.Email == "admin" &&
			account.Password == inputPassword && account.Password == "admin" {
			return "", "change"
		} else {
			return "Password is not correct", "validate"
		}
	} else if err != nil {
		return "Internal error comparing passwords", "internal"
	}
	return "", ""
}

func (account *Account) ChangePassword(password string, newPassword string, isReset bool) (string, string) {
	if !isReset {
		// Check if current password is correct
		errStr, errType := comparePasswordToAccount(*account, password)
		if errStr != "" {
			return errStr, errType
		}
	}

	// Validate new password
	if e := validatePasswordFormat(newPassword); e != "" {
		return "New password not valid: " + e, "validate"
	}

	// Update user
	ctx, cancel := u.Connect()
	defer cancel()
	user := map[string]interface{}{}
	hashedPassword, _ := bcrypt.GenerateFromPassword(
		[]byte(newPassword), bcrypt.DefaultCost)
	user["password"] = string(hashedPassword)
	err := GetDB().Collection("account").FindOneAndUpdate(ctx, bson.M{"_id": account.ID}, bson.M{"$set": user}).Err()
	if err != nil {
		return "Internal error while updating user password", "internal"
	}

	return GenerateToken(account.Email, account.ID, token_expiration), ""
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

	//Check password
	errStr, errType := comparePasswordToAccount(*account, password)
	if errStr != "" {
		return u.Message(false,
			"Invalid login credentials. Please try again"), errType
	} else if errType != "" {
		resp["shouldChange"] = true
	}

	//Success
	account.Password = ""

	//Create JWT token
	account.Token = GenerateToken(account.Email, account.ID, token_expiration)

	resp["account"] = account
	return resp, ""
}

func GenerateToken(email string, id primitive.ObjectID, expire time.Duration) string {
	// Create JWT token
	tk := &Token{Email: email, UserId: id}
	tk.ExpiresAt = time.Now().Add(expire).Unix()
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	return tokenString
}

func GetUser(userId primitive.ObjectID) *Account {
	acc := &Account{}
	ctx, cancel := u.Connect()
	err := GetDB().Collection("account").FindOne(ctx, bson.M{"_id": userId}).Decode(acc)
	if err != nil || acc.Email == "" {
		return nil
	}
	defer cancel()

	acc.Password = ""
	return acc
}

func GetUserByEmail(email string) *Account {
	acc := &Account{}
	ctx, cancel := u.Connect()
	err := GetDB().Collection("account").FindOne(ctx, bson.M{"email": email}).Decode(acc)
	if err != nil || acc.Email == "" {
		return nil
	}
	defer cancel()

	return acc
}

func GetAllUsers(callerRoles map[string]string) ([]Account, string) {
	// Get all users
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

	// Return allowed users according to caller permissions
	allowedUser := []Account{}
	for _, user := range users {
		if CheckCanManageUser(callerRoles, user.Roles) {
			allowedUser = append(allowedUser, user)
		}
	}

	defer cancel()
	return allowedUser, ""
}

func DeleteUser(userId primitive.ObjectID) string {
	ctx, cancel := u.Connect()
	req := bson.M{"_id": userId}
	c, _ := GetDB().Collection("account").DeleteOne(ctx, req)
	if c.DeletedCount == 0 {
		return "Internal error try to delete user"
	}
	defer cancel()
	return ""
}

func ModifyUser(id string, roles map[string]string) (string, string) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "User ID not valid", "validate"
	}

	if e := validateDomainRoles(roles); e != "" {
		return e, "validate"
	}

	println("UPDATE!")
	ctx, cancel := u.Connect()
	defer cancel()
	user := map[string]interface{}{}
	user["roles"] = roles
	err = GetDB().Collection("account").FindOneAndUpdate(ctx, bson.M{"_id": objID}, bson.M{"$set": user}).Err()
	if err != nil {
		return "Internal error while updating user roles", "internal"
	}

	return "", ""
}
