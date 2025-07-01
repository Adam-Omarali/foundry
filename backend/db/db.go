package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func GetDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL") // or use a hardcoded string for local dev
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening DB:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to DB:", err)
	}
	return db
}
