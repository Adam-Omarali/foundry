package db

import "log"

func InsertDocument(userId int, title string, url string, summary string, embedding []float32) {
	database := GetDB()
	defer database.Close()

	query := `
	INSERT INTO documents (user_id, title, url, summary, embedding)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := database.Exec(query, userId, title, url, summary, embedding)
	if err != nil {
		log.Fatal("Error inserting document:", err)
	}
}
