package controllers_test

import (
	"cli/controllers"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests CreateUser
func TestCreateUserInvalidEmail(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockAPI.On(
		"Request", "POST",
		"/api/users",
		"mock.Anything", 201,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"message": "A valid email address is required",
			},
			Status: 400,
		}, errors.New("[Response From API] A valid email address is required"),
	).Once()

	err := controller.CreateUser("email", "manager", "*")
	assert.NotNil(t, err)
	assert.Equal(t, "[Response From API] A valid email address is required", err.Error())
}

func TestCreateUserWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockAPI.On("Request", "POST",
		"/api/users",
		"mock.Anything", 201,
	).Return(
		&controllers.Response{
			Body: map[string]any{
				"message": "Account has been created",
			},
		}, nil,
	).Once()

	err := controller.CreateUser("email@email.com", "manager", "*")
	assert.Nil(t, err)
}

// Tests AddRole
func TestAddRoleUserNotFound(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockAPI.On("Request", "GET", "/api/users", "mock.Anything", 200).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": []any{},
			},
		}, nil,
	).Once()

	err := controller.AddRole("email@email.com", "manager", "*")
	assert.NotNil(t, err)
	assert.Equal(t, "user not found", err.Error())
}

func TestAddRoleWorks(t *testing.T) {
	controller, mockAPI, _ := layersSetup(t)

	mockAPI.On("Request", "GET", "/api/users", "mock.Anything", 200).Return(
		&controllers.Response{
			Body: map[string]any{
				"data": []any{
					map[string]any{
						"_id":   "507f1f77bcf86cd799439011",
						"email": "email@email.com",
					},
				},
			},
		}, nil,
	).Once()

	mockAPI.On("Request", "PATCH", "/api/users/507f1f77bcf86cd799439011", "mock.Anything", 200).Return(
		&controllers.Response{
			Body: map[string]any{
				"message": "successfully updated user roles",
			},
		}, nil,
	).Once()

	err := controller.AddRole("email@email.com", "manager", "*")
	assert.Nil(t, err)
}
