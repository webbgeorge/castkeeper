package users

import (
	"bufio"
	"context"
	"embed"
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

func GetUserByID(ctx context.Context, db *gorm.DB, id uint) (User, error) {
	var user User
	result := db.First(&user, "id = ?", id)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func ListUsers(ctx context.Context, db *gorm.DB) ([]User, error) {
	var users []User
	result := db.
		Order("id asc").
		Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
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

func UpdateUsername(ctx context.Context, db *gorm.DB, id uint, newUsername string) error {
	user, err := GetUserByID(ctx, db, id)
	if err != nil {
		return fmt.Errorf("failed to GetUserByID: %w", err)
	}

	user.Username = newUsername
	if err = db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func UpdatePassword(ctx context.Context, db *gorm.DB, id uint, newPassword string) error {
	user, err := GetUserByID(ctx, db, id)
	if err != nil {
		return fmt.Errorf("failed to GetUserByID: %w", err)
	}

	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), passwordHashCost)
	if err != nil {
		return err
	}

	user.Password = string(newPasswordHash)
	if err = db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

func DeleteUser(ctx context.Context, db *gorm.DB, id uint) error {
	user := User{Model: gorm.Model{ID: id}}
	if err := db.Delete(&user).Error; err != nil {
		return err
	}
	return nil
}

type PasswordStrengthError struct {
	Message string
}

func (e PasswordStrengthError) Error() string {
	return e.Message
}

func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return PasswordStrengthError{"password must be at least 8 characters"}
	}
	if len(password) > 64 {
		return PasswordStrengthError{"password must be 64 characters or less"}
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
			return PasswordStrengthError{"password is too easy to guess"}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
