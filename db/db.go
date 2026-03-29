package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func InitDB() *sql.DB {
	var db *sql.DB
	// get DB_URL from env
	dsn := os.Getenv("DB_URL")

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		log.Fatal("Unable to connect to Database")
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Database unreachable")
	}
	return db
}
