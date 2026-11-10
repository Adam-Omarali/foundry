package db

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID `json:"id"`
	UserEmail         string    `json:"user_email"`
	NumberOfDocuments int       `json:"num_documents"`
	Plan              string    `json:"plan"`
	CreatedAt         time.Time `json:"created_at"`
	LastSignin        time.Time `json:"last_signin"`
}

func GetNumberOfUsers() int {
	database := GetDB()
	defer database.Close()

	var count int
	err := database.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		panic(err)
	}
	return count
}

const userColumns = "id, user_email, num_documents, plan, created_at, last_signin"

func scanUser(row interface{ Scan(...any) error }) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.UserEmail, &u.NumberOfDocuments, &u.Plan, &u.CreatedAt, &u.LastSignin)
	return u, err
}

// UpsertUser creates the user on first sign-in (free plan) or bumps
// last_signin for an existing one, in a single atomic statement.
func UpsertUser(email string) (User, error) {
	database := GetDB()
	defer database.Close()

	query := `
	INSERT INTO users (user_email, plan)
	VALUES ($1, 'free')
	ON CONFLICT (user_email) DO UPDATE SET last_signin = NOW()
	RETURNING ` + userColumns

	user, err := scanUser(database.QueryRow(query, email))
	if err != nil {
		return User{}, fmt.Errorf("upserting user: %w", err)
	}
	return user, nil
}

// GetUserByID returns sql.ErrNoRows (wrapped) if the user doesn't exist.
func GetUserByID(id uuid.UUID) (User, error) {
	database := GetDB()
	defer database.Close()

	user, err := scanUser(database.QueryRow("SELECT "+userColumns+" FROM users WHERE id = $1", id))
	if err != nil {
		return User{}, fmt.Errorf("getting user %s: %w", id, err)
	}
	return user, nil
}

// CanAddDocument checks if a user can add more documents based on their plan
func CanAddDocument(userId uuid.UUID, plan string) error {
	database := GetDB()
	defer database.Close()

	database.Exec(`SET LOCAL request.user_id = $1`, userId)

	// Get current document count for the user
	var currentCount int
	err := database.QueryRow("SELECT num_documents FROM users WHERE id = $1", userId).Scan(&currentCount)
	if err != nil {
		return fmt.Errorf("error getting user document count: %v", err)
	}

	// Define plan limits
	var limit int
	switch plan {
	case "free":
		limit = 10
	case "premium":
		limit = 100
	case "unlimited":
		limit = -1 // No limit
	default:
		limit = 10 // Default to free plan limit
	}

	// Check if user has reached their limit
	if limit != -1 && currentCount >= limit {
		return fmt.Errorf("document limit reached for %s plan (%d documents). Please upgrade your plan to add more documents", plan, limit)
	}

	return nil
}
