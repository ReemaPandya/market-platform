package services

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func Signup(user models.User) error {

	_, err := repository.GetUserByEmail(user.Email)

	if err == nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	return repository.CreateUser(user)
}

func Login(email string, password string) (*models.User, error) {

	user, err := repository.GetUserByEmail(email)

	if err != nil {
		return nil, errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	)

	if err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}
