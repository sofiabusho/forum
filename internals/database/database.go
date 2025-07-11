package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var dbPath = "./forum.db" 

// Connect with SQLite DB
func CreateTable() *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	return db
}

// imports the database schema from a file
func Insert(db *sql.DB, table string, columns string, values ...any) {
	placeholders := ""
	for i := range values {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
	}
	query := fmt.Sprintf("INSERT INTO %s %s VALUES (%s)", table, columns, placeholders)
	_, err := db.Exec(query, values...)
	if err != nil {
		fmt.Println("Insert error:", err)
	}
}
