package models

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	Db *pgxpool.Pool
}

type UserRepoI interface {
	CreateUser(user *User) error
	GetUserByID(id uuid.UUID) (*User, error)
	GetUserByEmail(email string) (*User, error)
}

func NewUserRepo(db *pgxpool.Pool) UserRepoI {
	return &UserRepository{
		Db: db,
	}
}

// CreateUser creates a new user in the database
func (usr *UserRepository) CreateUser(user *User) error {
	// Prepare the SQL query to insert a new user
	query := `INSERT INTO users (username, email, password, created_at, updated_at)
			  VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`

	// Execute the query
	err := usr.Db.QueryRow(context.Background(), query, user.Username, user.Email, user.Password).Scan(&user.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("no rows affected while creating user")
		}
		log.Println("Error creating user:", err)
		return err
	}

	log.Println("User created successfully with ID:", user.ID)
	return nil
}

// GetUserByID retrieves a user by ID from the database
func (usr *UserRepository) GetUserByID(id uuid.UUID) (*User, error) {
	// Prepare the SQL query to get a user by ID
	query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = $1`

	// Declare a User struct to hold the result
	var user User

	// Execute the query
	err := usr.Db.QueryRow(context.Background(), query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no user found with ID %d", id)
		}
		log.Println("Error retrieving user:", err)
		return nil, err
	}

	log.Println("User retrieved successfully:", user.Username)
	return &user, nil
}

// GetUserByID retrieves a user by email from the database
func (usr *UserRepository) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = $1`

	var user User

	// Execute the query
	err := usr.Db.QueryRow(context.Background(), query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Error(err)
			return nil, fmt.Errorf("no user found with email %s", email)
		}
		logger.Error(err)
		return nil, errors.New("internal server error")
	}

	return &user, nil
}
