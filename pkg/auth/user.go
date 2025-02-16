package auth

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const passwordHashCost = 10

var userValidate = validator.New(validator.WithRequiredStructEnabled())

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex"`
	Password string
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	err := userValidate.Struct(u)
	if err != nil {
		return fmt.Errorf("user not valid: %w", err)
	}
	return nil
}

func (u *User) checkPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

func GetUserByUsername(ctx context.Context, db *gorm.DB, username string) (User, error) {
	var user User
	result := db.First(&user, "username = ?", username)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func CreateUser(ctx context.Context, db *gorm.DB, username string, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), passwordHashCost)
	if err != nil {
		return err
	}

	user := User{
		Username: username,
		Password: string(passwordHash),
	}
	if err = db.Create(&user).Error; err != nil {
		return err
	}

	return nil
}
