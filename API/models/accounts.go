package models

import (
	"errors"
	"os"
	"p3/repository"
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
	Roles    map[string]Role    `bson:"roles" json:"roles"`
	Token    string             `bson:"token,omitempty" json:"token,omitempty"`
}

// Validate incoming user
func (account *Account) Validate() *u.Error {
	valid := regexp.MustCompile("(\\w)+@(\\w)+\\.(\\w)+").MatchString(account.Email)

	if !valid {
		return &u.Error{Type: u.ErrBadFormat, Message: "A valid email address is required"}
	}

	if e := validatePasswordFormat(account.Password); e != nil {
		return &u.Error{Type: u.ErrBadFormat, Message: e.Error()}
	}

	//Error checking and duplicate emails
	ctx, cancel := u.Connect()
	err := repository.GetDB().Collection("account").FindOne(ctx, bson.M{"email": account.Email}).Err()
	if err != nil && err != mongo.ErrNoDocuments {
		println("Error while creating account:", err.Error())
		return &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}

	//User already exists
	if err == nil {
		return &u.Error{Type: u.ErrDuplicate, Message: "Error: User already exists"}
	}
	defer cancel()

	// Validate domains and roles
	if e := validateDomainRoles(account.Roles); e != nil {
		return &u.Error{Type: u.ErrInvalidValue, Message: e.Error()}
	}

	return nil
}

func validateDomainRoles(roles map[string]Role) error {
	// Validate domains and roles
	if len(roles) <= 0 {
		return errors.New("Object 'roles' with domain names as keys and roles as values is mandatory")
	}
	for domain, role := range roles {
		if !CheckDomainExists(domain) {
			return errors.New("Domain does not exist: " + domain)
		}
		switch role {
		case Manager:
		case Viewer:
		case User:
			break
		default:
			return errors.New("Role assigned is not valid: ")
		}
	}
	return nil
}

func (account *Account) Create(callerRoles map[string]Role) (*Account, *u.Error) {
	// Check if user is allowed to create new users
	if !CheckCanManageUser(callerRoles, account.Roles) {
		return nil, &u.Error{Type: u.ErrUnauthorized,
			Message: "Invalid credentials for creating an account." +
				" Manager role in requested domains is needed."}
	}

	// Validate new user
	if e := account.Validate(); e != nil {
		return nil, e
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword(
		[]byte(account.Password), bcrypt.DefaultCost)

	account.Password = string(hashedPassword)

	ctx, cancel := u.Connect()
	res, err := repository.GetDB().Collection("account").InsertOne(ctx, account)
	if err != nil {
		return nil, &u.Error{Type: u.ErrDBError,
			Message: "DB error when creating user: " + err.Error()}
	} else {
		account.ID = res.InsertedID.(primitive.ObjectID)
	}
	defer cancel()

	//Create new JWT token for the newly created account
	account.Token = GenerateToken(account.Email, account.ID, token_expiration)
	account.Password = ""
	return account, nil
}

func validatePasswordFormat(password string) error {
	if len(password) < 7 {
		return errors.New("Please provide a password with a length greater than 6")
	}
	return nil
}

func comparePasswordToAccount(account Account, inputPassword string) *u.Error {
	err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(inputPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return &u.Error{Type: u.ErrUnauthorized, Message: "Invalid login credentials"}
	} else if err == bcrypt.ErrHashTooShort {
		if account.Email == "admin" &&
			account.Password == inputPassword && account.Password == "admin" {
			return &u.Error{Type: u.WarnShouldChangePass}
		} else {
			return &u.Error{Type: u.ErrUnauthorized, Message: "Invalid login credentials"}
		}
	} else if err != nil {
		return &u.Error{Type: u.ErrInternal, Message: "Internal error comparing passwords"}
	}
	return nil
}

func (account *Account) ChangePassword(password string, newPassword string, isReset bool) (string, *u.Error) {
	if !isReset {
		// Check if current password is correct
		err := comparePasswordToAccount(*account, password)
		if err != nil && err.Type != u.WarnShouldChangePass {
			return "", err
		}
	}

	// Validate new password
	if e := validatePasswordFormat(newPassword); e != nil {
		return "", &u.Error{Type: u.ErrBadFormat, Message: "New password not valid: " + e.Error()}
	}

	// Update user
	ctx, cancel := u.Connect()
	defer cancel()
	user := map[string]interface{}{}
	hashedPassword, _ := bcrypt.GenerateFromPassword(
		[]byte(newPassword), bcrypt.DefaultCost)
	user["password"] = string(hashedPassword)
	err := repository.GetDB().Collection("account").FindOneAndUpdate(ctx, bson.M{"_id": account.ID}, bson.M{"$set": user}).Err()
	if err != nil {
		return "", &u.Error{Type: u.ErrDBError, Message: "Error updating user password: " + err.Error()}
	}

	return GenerateToken(account.Email, account.ID, token_expiration), nil
}

func Login(email, password string) (*Account, *u.Error) {
	account := &Account{}
	resp := u.Message("Logged In")

	ctx, cancel := u.Connect()
	err := repository.GetDB().Collection("account").FindOne(ctx, bson.M{"email": email}).Decode(account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &u.Error{Type: u.ErrNotFound, Message: "User does not exist"}
		}
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	defer cancel()

	//Check password
	e := comparePasswordToAccount(*account, password)
	if e != nil {
		if e.Type == u.WarnShouldChangePass {
			resp["shouldChange"] = true
		} else {
			return nil, e
		}
	}

	//Success
	account.Password = ""

	//Create JWT token
	account.Token = GenerateToken(account.Email, account.ID, token_expiration)

	return account, nil
}

func GenerateToken(email string, id primitive.ObjectID, expire time.Duration) string {
	// Create JWT token
	tk := &Token{Email: email, UserId: id}
	tk.ExpiresAt = time.Now().Add(expire).Unix()
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(os.Getenv("token_password")))
	return tokenString
}

// Returns user omitting password
func GetUser(userId primitive.ObjectID) *Account {
	acc := &Account{}
	ctx, cancel := u.Connect()
	err := repository.GetDB().Collection("account").FindOne(ctx, bson.M{"_id": userId}).Decode(acc)
	if err != nil || acc.Email == "" {
		return nil
	}
	defer cancel()

	acc.Password = ""
	return acc
}

// Returns user with password in clear text
func GetUserByEmail(email string) *Account {
	acc := &Account{}
	ctx, cancel := u.Connect()
	err := repository.GetDB().Collection("account").FindOne(ctx, bson.M{"email": email}).Decode(acc)
	if err != nil || acc.Email == "" {
		return nil
	}
	defer cancel()

	return acc
}

func GetAllUsers(callerRoles map[string]Role) ([]Account, *u.Error) {
	// Get all users
	ctx, cancel := u.Connect()
	c, err := repository.GetDB().Collection("account").Find(ctx, bson.M{})
	if err != nil {
		println(err.Error())
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}
	users := []Account{}
	err = c.All(ctx, &users)
	if err != nil {
		println(err.Error())
		return nil, &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}

	// Return allowed users according to caller permissions
	allowedUsers := []Account{}
	for _, user := range users {
		if CheckCanManageUser(callerRoles, user.Roles) {
			allowedUsers = append(allowedUsers, user)
		}
	}

	defer cancel()
	return allowedUsers, nil
}

func DeleteUser(userId primitive.ObjectID) *u.Error {
	ctx, cancel := u.Connect()
	req := bson.M{"_id": userId}
	c, _ := repository.GetDB().Collection("account").DeleteOne(ctx, req)
	if c.DeletedCount == 0 {
		return &u.Error{Type: u.ErrDBError, Message: "Unable to delete user"}
	}
	defer cancel()
	return nil
}

func ModifyUser(id string, roles map[string]Role) *u.Error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return &u.Error{Type: u.ErrInvalidValue, Message: "User ID not valid"}
	}

	if err := validateDomainRoles(roles); err != nil {
		return &u.Error{Type: u.ErrInvalidValue, Message: err.Error()}
	}

	ctx, cancel := u.Connect()
	defer cancel()
	user := map[string]interface{}{}
	user["roles"] = roles
	err = repository.GetDB().Collection("account").FindOneAndUpdate(ctx, bson.M{"_id": objID}, bson.M{"$set": user}).Err()
	if err != nil {
		return &u.Error{Type: u.ErrDBError, Message: err.Error()}
	}

	return nil
}
