package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Tournament struct {
	ID       string `db:"id"`
	Deposit  int    `db:"deposit"`
	Finished bool   `db:"finished"`
}

type Player struct {
	ID      string `json:"playerId" db:"id"`
	Balance int    `json:"balance" db:"balance"`
}

type Dump struct {
	Tournaments []*Tournament `json:"tournaments"`
	Players     []*Player     `json:"players"`
}

func (p *Player) MarshalJSON() ([]byte, error) {
	balance, _ := strconv.ParseFloat(fmt.Sprintf("%d.%d", p.Balance/100, p.Balance%100), 64)
	return json.Marshal(&struct {
		ID      string  `json:"playerId"`
		Balance float64 `json:"balance"`
	}{
		ID:      p.ID,
		Balance: balance,
	})
}

type Datastore interface {
	FindPlayer(playerID string) (*Player, error)
	FindOrCreatePlayer(playerID string) (*Player, error)
	TakeFunds(player *Player, points int) error
	AddFunds(player *Player, points int) error
	CreateTournament(tournamentID string, deposit int) error
	FindTournament(tournamentID string) (*Tournament, error)
	AllInfo() (*Dump, error)
	ResetDatabase()
}

type DB struct {
	*sqlx.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) FindPlayer(playerID string) (*Player, error) {
	var player Player
	err := db.Get(&player, "SELECT id, balance FROM player WHERE id = $1;", playerID)
	if err != nil {
		return nil, err
	}

	return &player, nil
}

func (db *DB) FindOrCreatePlayer(playerID string) (*Player, error) {
	player, err := db.FindPlayer(playerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.CreatePlayer(playerID)
		}
		return nil, err
	}
	return player, nil
}

func (db *DB) CreatePlayer(playerID string) (*Player, error) {
	var player Player
	err := db.Get(&player, "INSERT INTO player (id) VALUES ($1) RETURNING id, balance;", playerID)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

func (db *DB) TakeFunds(player *Player, points int) error {
	_, err := db.Exec("UPDATE player SET balance = balance - $1 WHERE id = $2;", points, player.ID)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) AddFunds(player *Player, points int) error {
	_, err := db.Exec("UPDATE player SET balance = balance + $1 WHERE id = $2;", points, player.ID)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateTournament(tournamentID string, deposit int) error {
	if _, err := db.Exec("INSERT INTO tournament (id, deposit) VALUES ($1, $2);", tournamentID, deposit); err != nil {
		return err
	}
	return nil
}

func (db *DB) FindTournament(tournamentID string) (*Tournament, error) {
	var tournament Tournament
	if err := db.Get(&tournament, "SELECT id, deposit FROM tournament WHERE !finished AND id = $1", tournamentID); err != nil {
		return nil, err
	}
	return &tournament, nil
}

func (db *DB) ResetDatabase() {
	db.Exec("TRUNCATE tournament_entries, tournament, player;")
}

func (db *DB) AllInfo() (*Dump, error) {
	var players []*Player
	var tournaments []*Tournament

	if err := db.Select(&players, "SELECT id, balance FROM player;"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	}
	if err := db.Select(&tournaments, "SELECT id, deposit, finished FROM tournament;"); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}

	}
	dump := &Dump{
		Players:     players,
		Tournaments: tournaments,
	}
	return dump, nil
}
