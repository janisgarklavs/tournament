package main

import (
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

type Datastore interface {
	FindPlayer(playerID string) (*Player, error)
	FindOrCreatePlayer(playerID string) (*Player, error)
	TakeFunds(player *Player, points int) error
	AddFunds(player *Player, points int) error
	// CreateTournament()
	// FindTournament()
}

type DB struct {
	*sqlx.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) FindPlayer(playerID string) (*Player, error) {
	return &Player{ID: playerID, Balance: 0}, nil
}

func (db *DB) FindOrCreatePlayer(playerID string) (*Player, error) {
	return db.FindPlayer(playerID)
}

func (db *DB) TakeFunds(player *Player, points int) error {
	return nil
}

func (db *DB) AddFunds(player *Player, points int) error {
	return nil
}
