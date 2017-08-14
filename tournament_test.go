package main

import (
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
