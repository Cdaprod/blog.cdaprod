Continuing from where we left off in the `main` function:

```go
func main() {
    db := initDatabase()
    handlers.SetDatabase(db)

    minioClient := initMinioClient()
    handlers.SetMinioClient(minioClient)

    r := mux.NewRouter()
    r.HandleFunc("/", handlers.HomeHandler)
    r.HandleFunc("/about", handlers.AboutHandler)
    r.HandleFunc("/post", handlers.PostHandler)
    r.HandleFunc("/create", handlers.CreatePostHandler)
    r.HandleFunc("/auth/login", auth.HandleGoogleLogin)
    r.HandleFunc("/auth/callback", auth.HandleGoogleCallback)

    fmt.Println("Starting server at port 8080")
    if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}
```

### Complete Project Structure

To ensure clarity, here is a summary of the project structure and all code files:

#### Project Structure

```
blog.cdaprod/
├── auth/
│   ├── oauth.go
├── handlers/
│   ├── handlers.go
├── templates/
│   ├── index.html
│   ├── post.html
│   └── create.html
├── main.go
├── database.go
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

### Code Files

#### auth/oauth.go

```go
package auth

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/oauth2/v2"
)

var googleOauthConfig = &oauth2.Config{
    RedirectURL:  "http://localhost:8080/auth/callback",
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
    Endpoint:     google.Endpoint,
}

var oauthStateString = "random"

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
    url := googleOauthConfig.AuthCodeURL(oauthStateString)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
    if r.FormValue("state") != oauthStateString {
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    token, err := googleOauthConfig.Exchange(context.Background(), r.FormValue("code"))
    if err != nil {
        log.Printf("Could not get token: %s\n", err.Error())
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    client := googleOauthConfig.Client(context.Background(), token)
    service, err := oauth2.New(client)
    if err != nil {
        log.Printf("Could not create oauth2 service: %s\n", err.Error())
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    userinfo, err := service.Userinfo.Get().Do()
    if err != nil {
        log.Printf("Could not get user info: %s\n", err.Error())
        http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
        return
    }

    // Here you can handle the user's info (e.g., save it to a database)
    fmt.Fprintf(w, "UserInfo: %v", userinfo)
}
```

#### handlers/handlers.go

```go
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
```

#### main.go

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/Cdaprod/blog.cdaprod/auth"
    "github.com/Cdaprod/blog.cdaprod/handlers"
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
    r.HandleFunc("/create", handlers.CreatePostHandler)
    r.HandleFunc("/auth/login", auth.HandleGoogleLogin)
    r.HandleFunc("/auth/callback", auth.HandleGoogleCallback