package db

import (
	"fmt"
	"log"
	"time"
)

type User struct {
	ID                int       `json:"id"`
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

func InsertUser(email string, plan string) User {
	database := GetDB()
	defer database.Close()

	database.Exec(`SET LOCAL request.user_email = $1`, email)

	query := `
	INSERT INTO users (user_email, plan)
	VALUES ($1, $2)
	ON CONFLICT (user_email) DO NOTHING
	RETURNING id
	`

	var id int
	err := database.QueryRow(query, email, plan).Scan(&id)
	if err != nil {
		log.Fatal("Error inserting/updating user:", err)
	}

	var user User
	err = database.QueryRow("SELECT id, user_email, num_documents, plan, created_at, last_signin FROM users WHERE user_email = $1", email).Scan(
		&user.ID,
		&user.UserEmail,
		&user.NumberOfDocuments,
		&user.Plan,
		&user.CreatedAt,
		&user.LastSignin,
	)
	if err != nil {
		log.Fatal("Error getting user:", err)
	}
	return user
}

func SignInUser(email string) int {
	database := GetDB()
	defer database.Close()

	database.Exec(`SET LOCAL request.user_email = $1`, email)

	query := `
	UPDATE users 
	SET last_signin = NOW()
	WHERE user_email = $1
	RETURNING id
	`

	var id int
	err := database.QueryRow(query, email).Scan(&id)
	if err != nil {
		log.Fatal("Error signing in user:", err)
	}

	return id
}

func GetUserId(email string) int {
	database := GetDB()
	defer database.Close()

	database.Exec(`SET LOCAL request.user_email = $1`, email)

	query := `
	SELECT id FROM users WHERE user_email = $1
	`
	var id int
	err := database.QueryRow(query, email).Scan(&id)
	if err != nil {
		log.Fatal("Error getting user id:", err)
	}
	return id
}

// GetUserByEmail retrieves a user by their email address
func GetUserByEmail(email string) User {
	database := GetDB()
	defer database.Close()

	database.Exec(`SET LOCAL request.user_email = $1`, email)

	var user User
	err := database.QueryRow("SELECT id, user_email, num_documents, plan, created_at, last_signin FROM users WHERE user_email = $1", email).Scan(
		&user.ID,
		&user.UserEmail,
		&user.NumberOfDocuments,
		&user.Plan,
		&user.CreatedAt,
		&user.LastSignin,
	)
	if err != nil {
		log.Fatal("Error getting user by email:", err)
	}
	return user
}

// CanAddDocument checks if a user can add more documents based on their plan
func CanAddDocument(userId int, plan string) error {
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
