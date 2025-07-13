package users

import (
	"bufio"
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//go:embed ncsc_common_passwords_8_chars_up.txt
var commonPasswordsFS embed.FS

const (
	commonPasswordsFile = "ncsc_common_passwords_8_chars_up.txt"
	passwordHashCost    = 10
)

var userValidate = validator.New(validator.WithRequiredStructEnabled())

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex" validate:"required,gte=1,lte=50"`
	Password string `validate:"required,gte=1"`
}

func (u *User) BeforeSave(tx *gorm.DB) error {
	err := userValidate.Struct(u)
	if err != nil {
		return fmt.Errorf("user not valid: %w", err)
	}
	return nil
}

func (u *User) CheckPassword(password string) error {
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
	if err := validatePasswordStrength(password); err != nil {
		return err
	}

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

func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(password) > 64 {
		return errors.New("password must be 64 characters or less")
	}
	if err := validatePasswordNotCommon(password); err != nil {
		return err
	}
	return nil
}

func validatePasswordNotCommon(password string) error {
	f, err := commonPasswordsFS.Open(commonPasswordsFile)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if password == scanner.Text() {
			return errors.New("password must not be in list of most common passwords")
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
