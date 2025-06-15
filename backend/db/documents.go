package db

import (
	"log"

	"github.com/pgvector/pgvector-go"
)

type Document struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

func InsertDocument(userId int, title string, url string, summary string, embedding []float32) {
	database := GetDB()
	defer database.Close()

	vector := pgvector.NewVector(embedding)

	query := `
	INSERT INTO documents (user_id, title, url, content, embedding)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := database.Exec(query, userId, title, url, summary, vector)
	if err != nil {
		log.Fatal("Error inserting document:", err)
	}
}

func GetDocuments(userId int, embedding []float32) []Document {
	database := GetDB()
	defer database.Close()

	vector := pgvector.NewVector(embedding)

	query := `
	SELECT id, title, url, content
	FROM documents
	WHERE user_id = $1
	AND embedding <=> $2 < 0.3
	ORDER BY embedding <=> $2
	LIMIT 10;
	`
	rows, err := database.Query(query, userId, vector)
	if err != nil {
		log.Fatal("Error getting documents:", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.ID, &doc.Title, &doc.URL, &doc.Content)
		if err != nil {
			log.Fatal("Error scanning document:", err)
		}
		documents = append(documents, doc)
	}

	if err = rows.Err(); err != nil {
		log.Fatal("Error iterating documents:", err)
	}

	return documents
}
