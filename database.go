package main

import (
    "database/sql"
    "log"
    _ "github.com/mattn/go-sqlite3"
)

func initDatabase() *sql.DB {
    db, err := sql.Open("sqlite3", "./blog.db")
    if err != nil {
        log.Fatal(err)
    }

    createTableSQL := `CREATE TABLE IF NOT EXISTS posts (
        "id" INTEGER PRIMARY KEY AUTOINCREMENT,
        "title" TEXT,
        "content" TEXT,
        "code" TEXT,
        "minio_object_name" TEXT
    );`
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatal(err)
    }

    return db
}