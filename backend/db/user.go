package db

import (
	"log"
	"time"
)

type User struct {
	id                   int
	user_email           string
	number_of_docsuments int
	plan                 string
	created_at           time.Time
	last_signin          time.Time
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

func InsetUser(email string, plan string) User {
	database := GetDB()
	defer database.Close()

	query := `
	INSERT INTO users (user_email, plan)
	VALUES ($1, $2)
	ON CONFLICT (user_email) DO UPDATE SET
		plan = EXCLUDED.plan
	RETURNING id
	`

	var id int
	err := database.QueryRow(query, "adamomarali37@gmail.com", "free").Scan(&id)
	if err != nil {
		log.Fatal("Error inserting/updating user:", err)
	}

	var user User
	err = database.QueryRow("SELECT * FROM users WHERE user_email = $1", "adamomarali37@gmail.com").Scan(&user.id, &user.user_email, &user.number_of_docsuments, &user.plan, &user.created_at, &user.last_signin)
	if err != nil {
		log.Fatal("Error getting user:", err)
	}
	return user

}

func SignInUser(email string) int {
	database := GetDB()
	defer database.Close()

	query := `
	UPDATE users 
	SET last_signin = NOW()
	WHERE user_email = $1
	`

	_, err := database.Exec(query, email)
	if err != nil {
		log.Fatal("Error signing in user:", err)
	}

	var user User
	err = database.QueryRow("SELECT * FROM users WHERE user_email = $1", email).Scan(&user.id, &user.user_email, &user.number_of_docsuments, &user.plan, &user.created_at, &user.last_signin)
	if err != nil {
		log.Fatal("Error getting user:", err)
	}
	return user.id
}
