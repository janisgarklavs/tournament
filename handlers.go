package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Handlers struct {
	repo Datastore
}

func (h *Handlers) dumpHandler(w http.ResponseWriter, r *http.Request) {
	dump, err := h.repo.AllInfo()
	if err != nil {
		return
	}
	json.NewEncoder(w).Encode(dump)
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
	points, err := getPointsFromString(r.Form.Get("points"))

	if err != nil || playerID == "" || points < 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Println("bad input", playerID, points)
		return
	}

	player, err := h.repo.FindOrCreatePlayer(playerID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("something went wrong on creating player")
		return
	}

	if err := h.repo.AddFunds(player, points); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("cannot add funds")
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *Handlers) announceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tournamentID := r.Form.Get("tournamentId")
	deposit, err := getPointsFromString(r.Form.Get("deposit"))
	if err != nil || tournamentID == "" || deposit <= 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Println("bad input", tournamentID, deposit)
		return
	}

	if err := h.repo.CreateTournament(tournamentID, deposit); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("cannot create tournament")
		return
	}
	w.WriteHeader(http.StatusOK)

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
	r.ParseForm()

	player, err := h.repo.FindPlayer(r.Form.Get("playerId"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(player)
}
func (h *Handlers) resetHandler(w http.ResponseWriter, r *http.Request) {
	h.repo.ResetDatabase()
	log.Println("resetting")
	w.WriteHeader(http.StatusOK)
}
