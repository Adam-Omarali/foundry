package db

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	_ "github.com/lib/pq"
)

// connString builds the Postgres URL from DB_* parts. url.URL escapes the
// user and password, so a password containing @ / : # is safe.
func connString() (string, error) {
	get := func(key, fallback string) string {
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}

	user, password, name := os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME")
	if user == "" || password == "" || name == "" {
		return "", fmt.Errorf("DB_USER, DB_PASSWORD and DB_NAME must all be set")
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(get("DB_HOST", "localhost"), get("DB_PORT", "5432")),
		Path:     name,
		RawQuery: url.Values{"sslmode": {get("DB_SSLMODE", "disable")}}.Encode(),
	}
	return u.String(), nil
}

func GetDB() *sql.DB {
	connStr, err := connString()
	if err != nil {
		log.Fatal("DB config: ", err)
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening DB:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to DB:", err)
	}
	return db
}
