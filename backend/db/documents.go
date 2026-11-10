package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"sort"
	"strings"

	"github.com/google/uuid"
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
	ID      uuid.UUID `json:"id"`
	Title   string    `json:"title"`
	URL     string    `json:"url"`
	Content string    `json:"content"`
	Read    bool      `json:"read"`
}

func InsertDocument(userId uuid.UUID, title string, url string, summary string, embedding []float32) {
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

func GetDocuments(userId uuid.UUID, embedding []float32, userQuery string) []Document {
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

	slog.Debug("documents found", "count", len(documents))

	return documents
}

// ErrDocumentNotFound means no document with that id belongs to the user.
var ErrDocumentNotFound = errors.New("document not found")

// updatableField describes one client-updatable document field.
type updatableField struct {
	column string                             // SQL column. Only ever taken from this map, never from the request.
	parse  func(json.RawMessage) (any, error) // decodes + validates the JSON value
}

// documentUpdatableFields is the allowlist for PATCH /api/documents/{id}.
// To make another field updatable, add one entry, e.g.
//
//	"title": {column: "title", parse: parseJSON[string]},
var documentUpdatableFields = map[string]updatableField{
	"read": {column: "read", parse: parseJSON[bool]},
}

func parseJSON[T any](raw json.RawMessage) (any, error) {
	// json.Unmarshal turns null into the zero value without an error, so a
	// {"read": null} would silently become false. Reject it explicitly.
	if string(raw) == "null" {
		return nil, errors.New("must not be null")
	}
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, fmt.Errorf("wrong type: %v", err)
	}
	return v, nil
}

type columnValue struct {
	column string
	value  any
}

// DocumentPatch is a validated set of column updates.
type DocumentPatch []columnValue

// ParseDocumentPatch validates a client's {field: newValue} object: every field
// must be in the allowlist and have the right type. Errors are client errors.
func ParseDocumentPatch(body map[string]json.RawMessage) (DocumentPatch, error) {
	if len(body) == 0 {
		return nil, errors.New("no fields to update")
	}
	names := make([]string, 0, len(body))
	for name := range body {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic SQL for the same set of fields

	patch := make(DocumentPatch, 0, len(names))
	for _, name := range names {
		field, ok := documentUpdatableFields[name]
		if !ok {
			return nil, fmt.Errorf("field %q is not updatable", name)
		}
		value, err := field.parse(body[name])
		if err != nil {
			return nil, fmt.Errorf("field %q: %v", name, err)
		}
		patch = append(patch, columnValue{field.column, value})
	}
	return patch, nil
}

// UpdateDocument applies the patch to one of the user's documents and returns
// the updated document, or ErrDocumentNotFound.
func UpdateDocument(userId uuid.UUID, documentId uuid.UUID, patch DocumentPatch) (Document, error) {
	database := GetDB()
	defer database.Close()

	sets := make([]string, len(patch))
	args := make([]any, 0, len(patch)+2)
	for i, cv := range patch {
		sets[i] = fmt.Sprintf("%s = $%d", cv.column, i+1)
		args = append(args, cv.value)
	}
	args = append(args, documentId, userId)

	query := fmt.Sprintf(`
	UPDATE documents
	SET %s
	WHERE id = $%d AND user_id = $%d
	RETURNING id, title, url, content, read
	`, strings.Join(sets, ", "), len(patch)+1, len(patch)+2)

	var doc Document
	err := database.QueryRow(query, args...).Scan(&doc.ID, &doc.Title, &doc.URL, &doc.Content, &doc.Read)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, ErrDocumentNotFound
	}
	if err != nil {
		return Document{}, fmt.Errorf("updating document: %w", err)
	}
	return doc, nil
}
