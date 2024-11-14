package models

import (
	"database/sql"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db}
}
func (ur *UserRepository) GetUserByEmail(email string) (User, error) {
	rows := ur.db.QueryRow("SELECT id, email, password FROM users where email = :1", email)

	var user User
	err := rows.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		return User{}, err
	}
	return user, nil
}
func (ur *UserRepository) GetUserById(id string) (User, error) {
	rows := ur.db.QueryRow("SELECT id, id, password FROM users where id = :1", id)

	var user User
	err := rows.Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		return User{}, err
	}
	return user, nil
}
