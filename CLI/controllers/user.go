package controllers

import (
	"cli/readline"
	"fmt"
	"math/rand"
	"net/http"
)

type User struct {
	Email string
	ID    string
}

func (controller Controller) CreateUser(email string, role string, domain string) error {
	password := randPassword(14)
	response, err := controller.API.Request(
		"POST",
		"/api/users",
		map[string]any{
			"email":    email,
			"password": password,
			"roles": map[string]any{
				domain: role,
			},
		},
		http.StatusCreated,
	)
	if err != nil {
		return err
	}
	println(response.message)
	println("password:" + password)
	return nil
}

func (controller Controller) AddRole(email string, role string, domain string) error {
	response, err := controller.API.Request("GET", "/api/users", nil, http.StatusOK)
	if err != nil {
		return err
	}
	userList, userListOk := response.Body["data"].([]any)
	if !userListOk {
		return fmt.Errorf("response contains no user list")
	}
	userID := ""
	for _, user := range userList {
		userMap, ok := user.(map[string]any)
		if !ok {
			continue
		}
		userEmail, emailOk := userMap["email"].(string)
		id, idOk := userMap["_id"].(string)
		if emailOk && idOk && userEmail == email {
			userID = id
			break
		}
	}
	if userID == "" {
		return fmt.Errorf("user not found")
	}
	response, err = controller.API.Request("PATCH", fmt.Sprintf("/api/users/%s", userID),
		map[string]any{
			"roles": map[string]any{
				domain: role,
			},
		},
		http.StatusOK,
	)
	if err != nil {
		return err
	}
	println(response.message)
	return nil
}

func ChangePassword() error {
	currentPassword, err := readline.Password("Current password: ")
	if err != nil {
		return err
	}
	newPassword, err := readline.Password("New password: ")
	if err != nil {
		return err
	}
	response, err := API.Request("POST", "/api/users/password/change",
		map[string]any{
			"currentPassword": string(currentPassword),
			"newPassword":     string(newPassword),
		},
		http.StatusOK,
	)
	if err != nil {
		return err
	}
	println(response.message)
	return nil
}

func randPassword(n int) string {
	const passChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = passChars[rand.Intn(len(passChars))]
	}
	return string(b)
}
