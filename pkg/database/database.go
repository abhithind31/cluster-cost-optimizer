package database

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() error {
    // Get connection parameters from environment variables
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    dbname := os.Getenv("DB_NAME")
    host := os.Getenv("DB_HOST")
    port := os.Getenv("DB_PORT")

    if user == "" || password == "" || dbname == "" || host == "" || port == "" {
        return fmt.Errorf("database environment variables not set")
    }

    connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
        user, password, dbname, host, port)
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        return err
    }
    return db.Ping()
}

func GetDB() *sql.DB {
    return db
}