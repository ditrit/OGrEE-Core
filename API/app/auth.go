package app

import (
	"net/http"
	u "p3/utils"
	"strings"

	"context"
	"os"
	"p3/models"

	jwt "github.com/dgrijalva/jwt-go"
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
		notAuth := []string{"/api", "/api/login"}
		requestPath := r.URL.Path //current request path

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
		if !token.Valid {
			response = u.Message(false, "Token is not valid.")
			w.WriteHeader(http.StatusForbidden)
			w.Header().Add("Content-Type", "application/json")
			u.Respond(w, response)
			return
		}

		//Success
		//set the caller to the user retrieved from the parsed token
		//Useful for monitoring

		//fmt.Sprintf("User %", tk.UserId)
		userData := map[string]interface{}{"email": tk.Email}
		ctx := context.WithValue(r.Context(), "user", userData)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
