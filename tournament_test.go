package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTournamentIntegration(t *testing.T) {
	db, _ := NewDB(dsn)
	db.ResetDatabase()
	defer func() {
		db.ResetDatabase()
	}()

	Convey("Given there is open database connection", t, func() {
		Convey("Given I add 100 points to player P1", func() {
			fundPlayer("P1", 100, db)
			Convey("It should create new player with id of P1 and balance of 100", func() {
				user, err := db.FindPlayer("P1")
				So(err, ShouldBeNil)
				So(user.Balance, ShouldEqual, 10000)
			})
		})
		Convey("Given I add 100 more points to a same player P1", func() {
			fundPlayer("P1", 100, db)
			Convey("It should have now 200 points", func() {
				user, err := db.FindPlayer("P1")
				So(err, ShouldBeNil)
				So(user.Balance, ShouldEqual, 20000)
			})
		})
		Convey("Given I take 300 points from the same player P1", func() {
			w := takeFundsFromPlayer("P1", 300, db)
			Convey("It should have error and still have 20000 points on it", func() {
				user, err := db.FindPlayer("P1")
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				So(err, ShouldBeNil)
				So(user.Balance, ShouldEqual, 20000)
			})
		})
		Convey("Given i try to take points from not existing user P2", func() {
			w := takeFundsFromPlayer("P2", 100, db)
			Convey("It should have error that user is not found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
		Convey("Given i create new tournament with id of 1 and deposit of 50", func() {
			w := createTournament("1", 50, db)
			Convey("It should have created new tournament with correspoding deposit and id", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				tournament, err := db.FindTournament("1")
				So(err, ShouldBeNil)
				So(tournament.ID, ShouldEqual, "1")
				So(tournament.Deposit, ShouldEqual, 5000)
			})
		})
		Convey("Given i try to join tournament which doesnt exists", func() {
			w := joinTournament("2", "P3", nil, db)
			Convey("it should result in error not found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
		Convey("Given i try to join tournament with user that doesnt exists", func() {
			w := joinTournament("1", "P3", nil, db)
			Convey("it should result in error bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
		Convey("Given i try to join tournament with user that doesnt have enough balance", func() {
			fundPlayer("P2", 20, db)
			w := joinTournament("1", "P2", nil, db)
			Convey("it should result in error bad request", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})
		})
		Convey("Given i try to join tournament with user and backer that doesnt have enough balance", func() {
			w := joinTournament("1", "P1", []string{"P2"}, db)
			Convey("it should result in error bad request and first user balance should be unchanged", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				player, _ := db.FindPlayer("P1")
				player2, _ := db.FindPlayer("P2")
				So(player.Balance, ShouldEqual, 20000)
				So(player2.Balance, ShouldEqual, 2000)
			})
		})
		Convey("Given i try to join tournament with user and 1 backer", func() {
			fundPlayer("P2", 180, db)
			w := joinTournament("1", "P1", []string{"P2"}, db)
			Convey("it should result in sucesful request and both user balances should be deducted by same amount", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				player, _ := db.FindPlayer("P1")
				player2, _ := db.FindPlayer("P2")
				So(player.Balance, ShouldEqual, 17500)
				So(player2.Balance, ShouldEqual, 17500)
			})
		})
		Convey("Given i join 1 more player P3 with 2 backers P4 and P5", func() {
			fundPlayer("P3", 100, db)
			fundPlayer("P4", 100, db)
			fundPlayer("P5", 100, db)
			w := joinTournament("1", "P3", []string{"P4", "P5"}, db)
			Convey("Thier balance also should be split accordingly", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				player3, _ := db.FindPlayer("P3")
				player4, _ := db.FindPlayer("P4")
				player5, _ := db.FindPlayer("P5")
				So(player3.Balance, ShouldEqual, 10000-1667)
				So(player4.Balance, ShouldEqual, 10000-1667)
				So(player5.Balance, ShouldEqual, 10000-1666)
			})
		})
		Convey("Given i result tournament which doesnt exists", func() {
			w := finishTournament("2", nil, db)
			Convey("it should result in error not found", func() {
				So(w.Code, ShouldEqual, http.StatusNotFound)
			})
		})
		Convey("Given i result tournament which exists but user doesnt exists", func() {
			w := finishTournament("1", map[string]int{"P2": 100}, db)
			Convey("It should result in error bad request and tournament and p2 user values unchanged", func() {
				So(w.Code, ShouldEqual, http.StatusBadRequest)
				user, _ := db.FindPlayer("P2")
				_, err := db.FindTournament("1")
				So(user.Balance, ShouldEqual, 17500)
				So(err, ShouldBeNil)
			})
		})
		Convey("Given i result tournament with prize where are backers", func() {
			w := finishTournament("1", map[string]int{"P1": 100}, db)
			Convey("It should result in succesful request and p1 and p2 both should have prize splited", func() {
				So(w.Code, ShouldEqual, http.StatusOK)
				player1, _ := db.FindPlayer("P1")
				player2, _ := db.FindPlayer("P2")
				So(player1.Balance, ShouldEqual, 22500)
				So(player2.Balance, ShouldEqual, 22500)
			})
		})

		Convey("Given i request all players balance for P1, P2 , P3, P4, P5", func() {
			w1 := playerBalance("P1", db)
			w2 := playerBalance("P2", db)
			w3 := playerBalance("P3", db)
			w4 := playerBalance("P4", db)
			w5 := playerBalance("P5", db)
			Convey("They all should show correct values for each player", func() {
				type playerResponse struct {
					Balance float64 `json:"balance"`
				}
				var player1, player2, player3, player4, player5 playerResponse
				json.NewDecoder(w1.Body).Decode(&player1)
				json.NewDecoder(w2.Body).Decode(&player2)
				json.NewDecoder(w3.Body).Decode(&player3)
				json.NewDecoder(w4.Body).Decode(&player4)
				json.NewDecoder(w5.Body).Decode(&player5)
				So(player1.Balance, ShouldEqual, 225)
				So(player2.Balance, ShouldEqual, 225)
				So(player3.Balance, ShouldEqual, 83.33)
				So(player4.Balance, ShouldEqual, 83.33)
				So(player5.Balance, ShouldEqual, 83.34)

			})
		})

	})
}

func fundPlayer(id string, points int, db *DB) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/fund?playerId=%v&points=%d", id, points), nil)
	w := httptest.NewRecorder()
	h := Handlers{db}
	handler := http.HandlerFunc(h.fundHandler)
	handler.ServeHTTP(w, req)
	return w
}

func takeFundsFromPlayer(id string, points int, db *DB) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/take?playerId=%v&points=%d", id, points), nil)
	w := httptest.NewRecorder()
	h := Handlers{db}
	handler := http.HandlerFunc(h.takeHandler)
	handler.ServeHTTP(w, req)
	return w
}

func createTournament(id string, deposit int, db *DB) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/announceTournament?tournamentId=%v&deposit=%d", id, deposit), nil)
	w := httptest.NewRecorder()
	h := Handlers{db}
	handler := http.HandlerFunc(h.announceHandler)
	handler.ServeHTTP(w, req)
	return w
}

func joinTournament(tournamentID string, playerID string, backers []string, db *DB) *httptest.ResponseRecorder {
	backerURL := "&backerId=%v"
	var backerConcatString string
	for _, v := range backers {
		backerConcatString += fmt.Sprintf(backerURL, v)
	}
	url := "/joinTournament?tournamentId=%v&playerId=%v"
	if len(backers) > 0 {
		url += backerConcatString
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf(url, tournamentID, playerID), nil)

	w := httptest.NewRecorder()
	h := Handlers{db}
	handler := http.HandlerFunc(h.joinHandler)
	handler.ServeHTTP(w, req)
	return w
}

func finishTournament(tournamentID string, winners map[string]int, db *DB) *httptest.ResponseRecorder {
	result := &ResultsRequest{
		TournamentID: tournamentID,
		Winners:      make([]Winner, 0),
	}

	for id, prize := range winners {
		winner := &Winner{PlayerID: id, Prize: prize}
		result.Winners = append(result.Winners, *winner)
	}
	data, _ := json.Marshal(result)
	b := bytes.NewBuffer(data)
	req, _ := http.NewRequest("POST", "/resultTournament", b)
	w := httptest.NewRecorder()
	h := Handlers{db}
	handler := http.HandlerFunc(h.resultHandler)
	handler.ServeHTTP(w, req)
	return w
}

func playerBalance(id string, db *DB) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", fmt.Sprintf("/balance?playerId=%v", id), nil)
	w := httptest.NewRecorder()
	h := Handlers{db}
	handler := http.HandlerFunc(h.balanceHandler)
	handler.ServeHTTP(w, req)
	return w
}
