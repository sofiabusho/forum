package database

import (
	"database/sql"
	"fmt"
	"os"

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

// InitializeDatabase sets up the database schema and default data
func InitializeDatabase() {
	db := CreateTable()
	defer db.Close()

	// Read and execute SQL schema
	sqlContent, err := os.ReadFile("internals/database/table.sql")
	if err != nil {
		fmt.Printf("Warning: Could not read table.sql: %v\n", err)
		return
	}

	// Makes reads/writes not block each other
	db.Exec("PRAGMA journal_mode=WAL;")
	// Waits a bit before returning 'database is locked
	db.Exec("PRAGMA busy_timeout=5000;")
	// Good balance of performance/safety for WAL
	db.Exec("PRAGMA synchronous=NORMAL;")

	// Execute SQL commands
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		fmt.Printf("Warning: Error executing SQL schema: %v\n", err)
	}

	// Insert default categories if they don't exist
	insertDefaultCategories(db)
}

// insertDefaultCategories adds the default forum categories
func insertDefaultCategories(db *sql.DB) {
	categories := []string{
		"Succulents",
		"Tropical Plants",
		"Herb Garden",
		"Indoor Plants",
		"Plant Care Tips",
		"Plant Diseases",
		"Propagation",
		"Flowering Plants",
	}

	for _, category := range categories {
		var exists int
		db.QueryRow("SELECT COUNT(*) FROM Categories WHERE name = ?", category).Scan(&exists)
		if exists == 0 {
			Insert(db, "Categories", "(name)", category)
		}
	}
}
