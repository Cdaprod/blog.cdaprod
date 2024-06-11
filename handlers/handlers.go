package handlers

import (
    "context"
    "database/sql"
    "fmt"
    "html/template"
    "io"
    "log"
    "net/http"
    "strings"

    "github.com/gorilla/mux"
    "github.com/minio/minio-go/v7"
)

type Post struct {
    ID              int
    Title           string
    Content         string
    Code            string
    MinioObjectName string
}

var db *sql.DB
var minioClient *minio.Client

func SetDatabase(database *sql.DB) {
    db = database
}

func SetMinioClient(client *minio.Client) {
    minioClient = client
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
    tmpl, err := template.ParseFiles("templates/index.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    data := struct {
        Title   string
        Content string
    }{
        Title:   "Welcome to my blog!",
        Content: "This is the home page of my awesome blog.",
    }
    tmpl.Execute(w, data)
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "About my blog")
}

func PostHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    var post Post
    row := db.QueryRow("SELECT id, title, content, code, minio_object_name FROM posts WHERE id = ?", id)
    err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Code, &post.MinioObjectName)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Fetch content from MinIO
    obj, err := minioClient.GetObject(context.Background(), "blog-posts", post.MinioObjectName, minio.GetObjectOptions{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer obj.Close()
    content, err := io.ReadAll(obj)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    post.Content = string(content)

    tmpl, err := template.ParseFiles("templates/post.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, post)
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        tmpl, err := template.ParseFiles("templates/create.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        tmpl.Execute(w, nil)
    } else if r.Method == "POST" {
        title := r.FormValue("title")
        content := r.FormValue("content")
        code := r.FormValue("code")
        minioObjectName := title + ".txt"
        
        // Convert the content to a reader
        contentReader := strings.NewReader(content)
        
        // Upload content to MinIO
        _, err := minioClient.PutObject(context.Background(), "blog-posts", minioObjectName, contentReader, int64(contentReader.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // Insert post details into the database
        _, err = db.Exec("INSERT INTO posts (title, content, code, minio_object_name) VALUES (?, ?, ?, ?)", title, content, code, minioObjectName)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}