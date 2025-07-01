package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/pgvector/pgvector-go"
)

// Common English stop words to filter out
var stopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true, "be": true, "by": true,
	"for": true, "from": true, "has": true, "he": true, "in": true, "is": true, "it": true, "its": true,
	"of": true, "on": true, "that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
	"i": true, "you": true, "your": true, "we": true, "they": true, "them": true, "this": true, "these": true,
	"those": true, "but": true, "or": true, "if": true, "then": true, "else": true, "when": true, "where": true,
	"why": true, "how": true, "all": true, "any": true, "both": true, "each": true, "few": true, "more": true,
	"most": true, "other": true, "some": true, "such": true, "no": true, "nor": true, "not": true, "only": true,
	"own": true, "same": true, "so": true, "than": true, "too": true, "very": true, "can": true,
	"just": true, "should": true, "now": true, "up": true, "down": true, "out": true, "off": true, "over": true,
	"under": true, "again": true, "further": true, "once": true, "here": true, "there": true,
}

// filterStopWords removes common stop words from a query string
func filterStopWords(query string) string {
	words := strings.Fields(strings.ToLower(query))
	var filteredWords []string

	for _, word := range words {
		// Remove punctuation and clean the word
		cleanWord := strings.Trim(word, ".,!?;:()[]{}'\"")
		// Only include words that are not stop words and have meaningful length
		if !stopWords[cleanWord] && len(cleanWord) > 2 {
			filteredWords = append(filteredWords, cleanWord)
		}
	}

	return strings.Join(filteredWords, " ")
}

type Document struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
	Read    bool   `json:"read"`
}

func InsertDocument(userId int, title string, url string, summary string, embedding []float32) {
	database := GetDB()
	defer database.Close()

	database.Exec(`SET LOCAL request.user_id = $1`, userId)

	vector := pgvector.NewVector(embedding)

	query := `
	INSERT INTO documents (user_id, title, url, content, embedding, read)
	VALUES ($1, $2, $3, $4, $5, false)
	`

	_, err := database.Exec(query, userId, title, url, summary, vector)
	if err != nil {
		log.Fatal("Error inserting document:", err)
	}

	query = `
	UPDATE users
	SET num_documents = num_documents + 1
	WHERE id = $1
	`
	_, err = database.Exec(query, userId)
	if err != nil {
		log.Fatal("Error updating user documents:", err)
	}
}

func GetDocuments(userId int, embedding []float32, userQuery string) []Document {
	database := GetDB()
	defer database.Close()

	database.Exec(`SET LOCAL request.user_id = $1`, userId)

	vector := pgvector.NewVector(embedding)

	query := `
	SELECT id, title, url, content, read
	FROM documents
	WHERE user_id = $1
	AND embedding <=> $2 < 0.4
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
		err := rows.Scan(&doc.ID, &doc.Title, &doc.URL, &doc.Content, &doc.Read)
		if err != nil {
			log.Fatal("Error scanning document:", err)
		}
		documents = append(documents, doc)
	}

	if len(documents) == 0 && userQuery != "" {
		// Filter out stop words from the query
		filteredQuery := filterStopWords(userQuery)

		// Only perform keyword search if we have meaningful terms after filtering
		if filteredQuery != "" {
			// If no semantic matches, try keyword search for each word
			query = `
			SELECT id, title, url, content, read
			FROM documents
			WHERE user_id = $1
			AND (
				to_tsvector('english', title) @@ to_tsquery('english', $2)
				OR to_tsvector('english', content) @@ to_tsquery('english', $2)
			)
			LIMIT 10;
			`

			// Convert spaces to OR operators for tsquery
			searchTerms := strings.Replace(filteredQuery, " ", " | ", -1)

			searchRows, err := database.Query(query, userId, searchTerms)
			if err != nil {
				log.Fatal("Error performing keyword search:", err)
			}
			defer searchRows.Close()

			for searchRows.Next() {
				var doc Document
				err := searchRows.Scan(&doc.ID, &doc.Title, &doc.URL, &doc.Content, &doc.Read)
				if err != nil {
					log.Fatal("Error scanning keyword search result:", err)
				}
				documents = append(documents, doc)
			}

			if err = searchRows.Err(); err != nil {
				log.Fatal("Error iterating keyword search results:", err)
			}
		}
	}

	if err = rows.Err(); err != nil {
		log.Fatal("Error iterating documents:", err)
	}

	fmt.Println("Documents:", documents)

	return documents
}

func MarkDocumentAsRead(userId int, documentId int) error {
	database := GetDB()
	defer database.Close()

	database.Exec(`SET LOCAL request.user_id = $1`, userId)

	query := `
	UPDATE documents
	SET read = true
	WHERE id = $1 AND user_id = $2
	`

	result, err := database.Exec(query, documentId, userId)
	if err != nil {
		return fmt.Errorf("error updating document read status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found or not owned by user")
	}

	return nil
}
