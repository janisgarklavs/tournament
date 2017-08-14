package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//Tournament is structure that represent tournament table entry in database
type Tournament struct {
	ID       string `db:"id"`
	Deposit  int    `db:"deposit"`
	Finished bool   `db:"finished"`
}

//Player is structure that represent player table entry in database
type Player struct {
	ID      string `json:"playerId" db:"id"`
	Balance int    `json:"balance" db:"balance"`
}

//MarshalJSON is custom json marshaler to present points in float format (points / 100 and remainder)
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

//Datastore is interface that holds all methods for data access layer
type Datastore interface {
	FindPlayer(playerID string) (*Player, error)
	FindOrCreatePlayer(playerID string) (*Player, error)
	TakeFunds(player *Player, points int) error
	AddFunds(player *Player, points int) error
	CreateTournament(tournamentID string, deposit int) error
	FindTournament(tournamentID string) (*Tournament, error)
	TournamentJoinPlayers(tournament *Tournament, playerID string, backers []string) error
	FinishTournament(tournament *Tournament, winners []Winner) error
	ResetDatabase()
}

//DB holds sqlx database handle
type DB struct {
	*sqlx.DB
}

//NewDB creates new database handle with provided dsn
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

//FindPlayer returns player by its id
func (db *DB) FindPlayer(playerID string) (*Player, error) {
	var player Player
	err := db.Get(&player, "SELECT id, balance FROM player WHERE id = $1;", playerID)
	if err != nil {
		return nil, err
	}

	return &player, nil
}

//FindOrCreatePlayer returns player by its id or creates new one if not found
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

//CreatePlayer creates new player with given id
func (db *DB) CreatePlayer(playerID string) (*Player, error) {
	var player Player
	err := db.Get(&player, "INSERT INTO player (id) VALUES ($1) RETURNING id, balance;", playerID)
	if err != nil {
		return nil, err
	}
	return &player, nil
}

//TakeFunds takes player and deducts given points away from its balance
func (db *DB) TakeFunds(player *Player, points int) error {
	_, err := db.Exec("UPDATE player SET balance = balance - $1 WHERE id = $2;", points, player.ID)
	if err != nil {
		return err
	}
	return nil
}

//AddFunds takes player and adds given points to its balance
func (db *DB) AddFunds(player *Player, points int) error {
	_, err := db.Exec("UPDATE player SET balance = balance + $1 WHERE id = $2;", points, player.ID)
	if err != nil {
		return err
	}
	return nil
}

//CreateTournament creates new tournament entry with it's deposit
func (db *DB) CreateTournament(tournamentID string, deposit int) error {
	if _, err := db.Exec("INSERT INTO tournament (id, deposit) VALUES ($1, $2);", tournamentID, deposit); err != nil {
		return err
	}
	return nil
}

//FindTournament returns tournament which is not finished or error
func (db *DB) FindTournament(tournamentID string) (*Tournament, error) {
	var tournament Tournament
	if err := db.Get(&tournament, "SELECT id, deposit FROM tournament WHERE finished = false AND id = $1", tournamentID); err != nil {
		return nil, err
	}
	return &tournament, nil
}

//TournamentJoinPlayers takes tournament and takes points for players and adds them to tournament entries
func (db *DB) TournamentJoinPlayers(tournament *Tournament, playerID string, backers []string) error {
	log.Println(tournament.Deposit, len(backers))
	if len(backers) == 0 {
		tx := db.MustBegin()
		var err error
		_, err = tx.Exec("UPDATE player SET balance = balance - $1 WHERE id = $2;", tournament.Deposit, playerID)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec("INSERT INTO tournament_entries (tournament_id, user_id, backing_id) VALUES ($1, $2, null)", tournament.ID, playerID)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}
		return nil
	}
	parts := splitEvenly(tournament.Deposit, len(backers)+1)

	tx := db.MustBegin()
	var err error
	_, err = tx.Exec("UPDATE player SET balance = balance - $1 WHERE id = $2;", parts[0], playerID)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("INSERT INTO tournament_entries (tournament_id, user_id, backing_id) VALUES ($1, $2, null)", tournament.ID, playerID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for i, v := range backers {
		_, err = tx.Exec("UPDATE player SET balance = balance - $1 WHERE id = $2;", parts[i+1], v)
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = tx.Exec("INSERT INTO tournament_entries (tournament_id, user_id, backing_id) VALUES ($1, $2, $3)", tournament.ID, v, playerID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

//FinishTournament takes tournament and winners, and correspondingly gives out points to winning entries and their backers
func (db *DB) FinishTournament(tournament *Tournament, winners []Winner) error {
	tx := db.MustBegin()
	defer tx.Rollback()
	for _, v := range winners {
		players, err := db.findPlayersWithBackers(tournament.ID, v.PlayerID)
		if err != nil {
			return err
		}
		var rewards []int
		if len(players) == 1 {
			rewards = []int{v.Prize * 100}
		} else {
			rewards = splitEvenly(v.Prize*100, len(players))
		}
		for i, v := range players {
			_, err := tx.Exec("UPDATE player SET balance = balance + $1 WHERE id = $2;", rewards[i], v)
			if err != nil {
				return err
			}
		}
	}
	_, err := tx.Exec("UPDATE tournament SET finished = true WHERE id = $1;", tournament.ID)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (db *DB) findPlayersWithBackers(tournamentID string, playerID string) ([]string, error) {
	var players []string
	if err := db.Select(&players, "SELECT user_id FROM tournament_entries WHERE tournament_id = $1 AND user_id = $2 OR backing_id = $2", tournamentID, playerID); err != nil {
		return nil, err
	}
	return players, nil
}

// ResetDatabase truncates all tables for clean database
func (db *DB) ResetDatabase() {
	db.Exec("TRUNCATE tournament_entries, tournament, player;")
}
