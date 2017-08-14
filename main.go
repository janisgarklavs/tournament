package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

const addr = ":8080"
const dsn = "postgres://postgres:example@database/postgres?sslmode=disable"

func init() {
	schema := `
		create table if not exists player (
			id varchar(64) not null primary key,
			balance integer not null default 0 check (balance >= 0)
		);

		create table if not exists tournament (
			id varchar(64) not null primary key,
			deposit integer not null,
			finished boolean not null default false
		);

		create table if not exists tournament_entries (
			id serial not null primary key,
			tournament_id varchar(64) not null references tournament (id),
			user_id varchar(64) not null references player (id),
			backing_id varchar(64) references player (id)
		);
	`
	db, err := NewDB(dsn)
	if err != nil {
		log.Fatal(err)
	}

	db.MustExec(schema)
	db.Close()
}

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
