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