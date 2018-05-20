package main

import (
  "net/http"
  "github.com/go-chi/chi"
  "github.com/go-chi/chi/middleware"
)

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
  http.ListenAndServe(":3000", r)
}