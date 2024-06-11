package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/Cdaprod/blog.cdaprod/auth"
    "github.com/Cdaprod/blog.cdaprod/handlers"
    "github.com/Cdaprod/blog.cdaprod/middleware"
    "github.com/gorilla/mux"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
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

func initMinioClient() *minio.Client {
    minioClient, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
        Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
        Secure: true,
    })
    if err != nil {
        log.Fatalln(err)
    }

    return minioClient
}

func main() {
    db := initDatabase()
    handlers.SetDatabase(db)

    minioClient := initMinioClient()
    handlers.SetMinioClient(minioClient)

    r := mux.NewRouter()
    r.HandleFunc("/", handlers.HomeHandler)
    r.HandleFunc("/about", handlers.AboutHandler)
    r.HandleFunc("/post", handlers.PostHandler)
    r.HandleFunc("/create", handlers.CreatePostHandler).Methods("GET")
    r.HandleFunc("/create", handlers.CreatePostHandler).Methods("POST").Handler(middleware.AuthMiddleware(http.HandlerFunc(handlers.CreatePostHandler)))
    r.HandleFunc("/auth/login", auth.HandleGoogleLogin)
    r.HandleFunc("/auth/callback", auth.HandleGoogleCallback)

    fmt.Println("Starting server at port 8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}