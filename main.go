package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/subosito/gotenv"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	LoginURL     string
}

var config Config

func Login(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.LoginURL, http.StatusTemporaryRedirect)
}

func Callback(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()

	data := url.Values{
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"grant_type":    {"authorization_code"},
		"code":          {r.FormValue("code")},
		"redirect_uri":  {config.RedirectURL},
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.PostForm(
		"https://discord.com/api/oauth2/token",
		data,
	)

	if err != nil {
		w.Write([]byte("Internal Server Error!"))
	}

	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		w.Write([]byte("Internal Server Error!"))
	}

	w.Write([]byte(content))
}

func main() {

	gotenv.Load()

	config.ClientID = os.Getenv("CLIENT_ID")
	config.ClientSecret = os.Getenv("CLIENT_SECRET")
	config.RedirectURL = os.Getenv("REDIRECT_URL")
	config.LoginURL = "https://discord.com/api/oauth2/authorize?" +
		fmt.Sprintf("client_id=%v", config.ClientID) +
		fmt.Sprintf("&redirect_uri=%v", config.RedirectURL) +
		"&response_type=code&scope=identify%20email%20guilds"

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(httprate.LimitByIP(100, 10*time.Second))
	r.Use(middleware.StripSlashes)

	r.Get("/api/login", Login)
	r.Get("/api/callback", Callback)

	http.ListenAndServe(":8000", r)
}
