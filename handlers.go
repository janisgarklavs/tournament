package main

import (
	"log"
	"net/http"
	"strconv"
)

type Handlers struct {
	repo Datastore
}

func (h *Handlers) takeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	playerID := r.Form.Get("playerId")
	points, err := getPointsFromString(r.Form.Get("points"))
	if err != nil || playerID == "" || points < 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Println("bad input", playerID, points, err)
		return
	}

	player, err := h.repo.FindPlayer(playerID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("did not find player")
		return
	}
	if err := h.repo.TakeFunds(player, points); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("cannot take enough from player")
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *Handlers) fundHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	playerID := r.Form.Get("playerId")
	points, err := strconv.ParseFloat(r.Form.Get("points"), 64)
	if err != nil || playerID == "" || points < 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Println("bad input", playerID, points)
		return
	}
	points = points * 100
	player, err := h.repo.FindOrCreatePlayer(playerID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("something went wrong on creating player")
		return
	}
	if err := h.repo.AddFunds(player, int(points)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("cannot add funds")
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *Handlers) announceHandler(w http.ResponseWriter, r *http.Request) {
	// create new tournament
	r.ParseForm()
	log.Println(r.Form.Get("tournamentId"), r.Form.Get("deposit"))
	//repo.CreateTournament(tournamentID, deposit)
}
func (h *Handlers) joinHandler(w http.ResponseWriter, r *http.Request) {
	// user join tournament
	r.ParseForm()
	log.Println(r.Form.Get("tournamentId"), r.Form.Get("playerId"), r.Form["backerId"])
	//repo.FindTournament(tournamentID)
	// tournamnet.Join(playerId, backers...)
	//this withdraws money
}
func (h *Handlers) resultHandler(w http.ResponseWriter, r *http.Request) {
	//finish tournament
	//json decode

	// repo.FindTournament(tournamentID)
	// tournamnet.Result([]results)

}
func (h *Handlers) balanceHandler(w http.ResponseWriter, r *http.Request) {
	// show balance for user
	r.ParseForm()
	log.Println(r.Form.Get("playerId"))
	// repo.FindPlayer(playerID)
}
func (h *Handlers) resetHandler(w http.ResponseWriter, r *http.Request) {
	// reset db to intial values
	log.Println("resetting")
}
