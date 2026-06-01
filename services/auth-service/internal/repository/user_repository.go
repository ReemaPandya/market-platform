package repository

import (
	"auth-service/internal/config"
	"auth-service/internal/models"
)

func CreateUser(user models.User) error {

	query := `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
	`

	_, err := config.DB.Exec(
		query,
		user.Email,
		user.Password,
	)

	return err
}
func GetUserByEmail(email string) (*models.User, error) {

	query := `
		SELECT id, email, password
		FROM users
		WHERE email=$1
	`

	row := config.DB.QueryRow(query, email)

	var user models.User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
