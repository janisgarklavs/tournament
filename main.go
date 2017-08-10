package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

const addr = ":8080"
const dsn = "postgres://postgres:display@localhost/postgres?sslmode=disable"

func main() {
	log.Println("Server starting...")
	db, err := NewDB(dsn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Database started...")

	r := chi.NewRouter()
	h := Handlers{db}

	r.Get("/take", h.takeHandler)
	r.Get("/fund", h.fundHandler)
	r.Get("/announceTournament", h.announceHandler)
	r.Get("/joinTournament", h.joinHandler)
	r.Post("/resultTournament", h.resultHandler)
	r.Get("/balance", h.balanceHandler)
	r.Get("/reset", h.resetHandler)

	log.Println("All systems operational!")
	log.Fatal(http.ListenAndServe(addr, r))
}
