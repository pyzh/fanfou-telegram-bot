package main

import (
	"net/http"
	"os"

	"github.com/dghubble/oauth1"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var oauthConfig = oauth1.Config{
	ConsumerKey:            os.Getenv("ConsumerKey"),
	ConsumerSecret:         os.Getenv("ConsumerSecret"),
	CallbackURL:            "https://fanfou-204619.appspot.com/callback",
	Endpoint:               FanfouEndpoint,
	DisableCallbackConfirm: true,
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Get("/_ah/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// fanfou callback
	// https://address/callback?telegram_id=1&oauth_token=e5be60f65bbd0d23b92d7abc705f3&request_secret=111
	r.Get("/callback", func(w http.ResponseWriter, r *http.Request) {
		requestSecret := r.URL.Query().Get("request_secret")
		requestToken, verifier, err := oauth1.ParseAuthorizationCallback(r)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		accessToken, accessSecret, err := oauthConfig.AccessToken(requestToken, requestSecret, verifier)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.Write([]byte(accessToken + accessSecret))
	})

	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		requestToken, requestSecret, err := oauthConfig.RequestToken()
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		// get authorization url
		authorizationURL, err := oauthConfig.AuthorizationURL(requestToken)
		q := authorizationURL.Query()
		telegramid := "1"
		callback := oauthConfig.CallbackURL + "?telegram_id=" + telegramid + "&request_secret=" + requestSecret //TODO
		q.Set("oauth_callback", callback)
		authorizationURL.RawQuery = q.Encode()

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		http.Redirect(w, r, authorizationURL.String(), http.StatusFound)
	})

	http.ListenAndServe(":8080", r)
}
