package app

import (
	"net/http"
	u "p3/utils"
	"strings"

	"context"
	"os"
	"p3/models"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var Log = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("Selected Path: ", r.URL.Path)
		println("Raw Query: ", r.URL.RawQuery)
		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

var JwtAuthentication = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//Endpoints that don't require auth
		notAuth := []string{"/api", "/api/login", "/api/users/password/forgot"}
		requestPath := r.URL.Path //current request path
		println(requestPath)

		//check if request needs auth
		//serve the request if not needed
		for _, value := range notAuth {
			if value == requestPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		//Grab the token from the header
		response := make(map[string]interface{})
		tokenHeader := r.Header.Get("Authorization")

		//Token is missing return 403
		if tokenHeader == "" {
			response = u.Message(false, "Missing auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		//Token format `Bearer {token-body}`
		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			response = u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		//Grab the token body
		tokenPart := splitted[1]
		tk := &models.Token{}

		token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("token_password")), nil
		})

		//Malformed token
		if err != nil {
			response = u.Message(false, "Malformed authentication token")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		//Token is invalid
		if !token.Valid || ((tk.Email == "RESET") != (requestPath == "/api/users/password/reset")) {
			response = u.Message(false, "Token is not valid.")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		//Success
		//set the caller to the user retrieved from the parsed token
		//Useful for monitoring
		userData := map[string]interface{}{"email": tk.Email, "userID": tk.UserId}
		ctx := context.WithValue(r.Context(), "user", userData)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

func ParseToken(w http.ResponseWriter, r *http.Request) map[string]primitive.ObjectID {
	//Grab the token from the header
	tokenHeader := r.Header.Get("Authorization")

	//Token is missing return 403
	if tokenHeader == "" {
		return nil
	}

	//Token format `Bearer {token-body}`
	splitted := strings.Split(tokenHeader, " ")
	if len(splitted) != 2 {
		return nil
	}

	//Grab the token body
	tokenPart := splitted[1]
	tk := &models.Token{}

	token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("token_password")), nil
	})

	//Malformed token
	if err != nil {
		return nil
	}

	//Token is invalid
	if !token.Valid {
		return nil
	}

	//Success
	return map[string]primitive.ObjectID{
		"userID": tk.UserId}
}
