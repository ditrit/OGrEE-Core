package utils

import (
	"p3/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CreateTestUser(t *testing.T, role models.Role) (string, string) {
	// It creates a temporary user that will be deleted at the end of the test t
	email := "temporary_user@test.com"
	password := "fake_password"
	account := &models.Account{
		Name:     "Temporary User",
		Email:    email,
		Password: password,
		Roles: map[string]models.Role{
			"*": role,
		},
	}
	acc, err := account.Create(map[string]models.Role{"*": "manager"})
	assert.Nil(t, err)

	t.Cleanup(func() {
		// we get the user again as the user may have been deleted in a test
		user := models.GetUserByEmail(acc.Email)
		if user != nil {
			err := models.DeleteUser(user.ID)
			assert.Nil(t, err)
		}
	})
	return email, password
}
