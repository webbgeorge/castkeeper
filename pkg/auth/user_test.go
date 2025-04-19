package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/webbgeorge/castkeeper/pkg/auth"
	"github.com/webbgeorge/castkeeper/pkg/fixtures"
)

func TestUserBeforeSave(t *testing.T) {
	testCases := map[string]struct {
		user        auth.User
		expectedErr string
	}{
		"valid": {
			user: auth.User{
				Username: "someuser",
				Password: "aaaaaa",
			},
			expectedErr: "",
		},
		"usernameMissing": {
			user: auth.User{
				Password: "aaaaaa",
			},
			expectedErr: "user not valid: Key: 'User.Username' Error:Field validation for 'Username' failed on the 'required' tag",
		},
		"usernameTooLong": {
			user: auth.User{
				Username: fixtures.StrOfLen(100),
				Password: "aaaaaa",
			},
			expectedErr: "user not valid: Key: 'User.Username' Error:Field validation for 'Username' failed on the 'lte' tag",
		},
		"passwordMissing": {
			user: auth.User{
				Username: "someuser",
			},
			expectedErr: "user not valid: Key: 'User.Password' Error:Field validation for 'Password' failed on the 'required' tag",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.user.BeforeSave(nil)
			if tc.expectedErr == "" {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tc.expectedErr, err.Error())
			}
		})
	}
}

func TestGetByUsername_Exists(t *testing.T) {
	db, resetDB := fixtures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	// from fixture
	user, err := auth.GetUserByUsername(context.Background(), db, "unittest")

	assert.Nil(t, err)
	assert.Equal(t, 123, int(user.ID))
	assert.Equal(t, "unittest", user.Username)
	assert.NotEqual(t, "unittestpw", user.Password)
	assert.NotEmpty(t, user.Password)
}

func TestGetByUsername_NotFound(t *testing.T) {
	db, resetDB := fixtures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	user, err := auth.GetUserByUsername(context.Background(), db, "notauser")

	assert.Equal(t, "record not found", err.Error())
	assert.Zero(t, user)
}

func TestUserCheckPassword(t *testing.T) {
	db, resetDB := fixtures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	user, err := auth.GetUserByUsername(context.Background(), db, "unittest")
	assert.Nil(t, err)

	t.Run("correct", func(t *testing.T) {
		err = user.CheckPassword("unittestpw")
		assert.Nil(t, err)
	})

	t.Run("noPassword", func(t *testing.T) {
		err = user.CheckPassword("")
		assert.NotNil(t, err)
	})

	t.Run("wrongPassword", func(t *testing.T) {
		err = user.CheckPassword("wrong")
		assert.NotNil(t, err)
	})
}

func TestCreateUser_Valid(t *testing.T) {
	db, resetDB := fixtures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	err := auth.CreateUser(context.Background(), db, "user1", "pass1")
	assert.Nil(t, err)

	user, err := auth.GetUserByUsername(context.Background(), db, "user1")
	assert.Nil(t, err)
	err = user.CheckPassword("pass1")
	assert.Nil(t, err)
}

func TestCreateUser_InvalidUsername(t *testing.T) {
	db, resetDB := fixtures.ConfigureDBForTestWithFixtures()
	defer resetDB()

	err := auth.CreateUser(context.Background(), db, "", "pass1")
	assert.Equal(t, "user not valid: Key: 'User.Username' Error:Field validation for 'Username' failed on the 'required' tag", err.Error())
}
