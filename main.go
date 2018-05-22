package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/dghubble/oauth1"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	tb "gopkg.in/tucnak/telebot.v2"
)

var oauthConfig = oauth1.Config{
	ConsumerKey:            os.Getenv("ConsumerKey"),
	ConsumerSecret:         os.Getenv("ConsumerSecret"),
	CallbackURL:            "https://fanfou-204818.appspot.com/callback",
	Endpoint:               FanfouEndpoint,
	DisableCallbackConfirm: true,
}

var datastoreClient *datastore.Client

type oauthInfo struct {
	Token  string
	Secret string
}

func main() {
	bot, err := tb.NewBot(tb.Settings{
		Token:  os.Getenv("TelegramToken"),
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Handle("/start", func(m *tb.Message) {
		authorizationURL, err := getAuthorizationURL(m.Sender.ID)
		if err != nil {
			bot.Send(m.Sender, err.Error())
		} else {
			text := fmt.Sprintf("**Authorization url** [click link](%s)", authorizationURL)
			bot.Send(m.Sender, text, tb.ModeMarkdown)
		}
	})

	bot.Handle(tb.OnText, func(m *tb.Message) {
		log.Println(m.Text)
	})

	go bot.Start()

	ctx := context.Background()
	projectID := os.Getenv("ProjectID")
	datastoreClient, err = datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}

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
		telegramIDParams := r.URL.Query().Get("telegram_id")
		telegramID, err := strconv.Atoi(telegramIDParams)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

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

		ctx := context.Background()
		k := getKey(telegramID)
		v := &oauthInfo{Token: accessToken, Secret: accessSecret}
		if _, err := datastoreClient.Put(ctx, k, v); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		bot.Send(&tb.User{ID: telegramID}, "Success Authorization")
		w.Write([]byte("it's ok"))
	})

	http.ListenAndServe(":8080", r)
}

func getAuthorizationURL(telegramID int) (url string, err error) {
	requestToken, requestSecret, err := oauthConfig.RequestToken()
	if err != nil {
		return
	}
	authorizationURL, err := oauthConfig.AuthorizationURL(requestToken)
	if err != nil {
		return
	}
	q := authorizationURL.Query()
	callback := fmt.Sprintf("%s?telegram_id=%d&request_secret=%s", oauthConfig.CallbackURL, telegramID, requestSecret)
	q.Set("oauth_callback", callback)
	authorizationURL.RawQuery = q.Encode()

	if err != nil {
		return
	}
	return authorizationURL.String(), nil
}

func getKey(telegramID int) *datastore.Key {
	return datastore.IDKey("fanfou_tokens", int64(telegramID), nil)

}
