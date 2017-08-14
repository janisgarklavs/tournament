package main

import (
	"encoding/json"
	"net/http"
)

// ResultsRequest is request for /POST resultTournament call body decoding
type ResultsRequest struct {
	TournamentID string   `json:"tournamentId"`
	Winners      []Winner `json:"winners"`
}

// Winner holds winning entries in for ResultsRequest
type Winner struct {
	PlayerID string `json:"playerId"`
	Prize    int    `json:"prize"`
}

//Handlers structure holds our handlers and access to datastore interface
type Handlers struct {
	repo Datastore
}

/**
* GET /take
**/
func (h *Handlers) takeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	playerID := r.Form.Get("playerId")
	points, err := getPointsFromString(r.Form.Get("points"))
	if err != nil || playerID == "" || points < 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	player, err := h.repo.FindPlayer(playerID)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := h.repo.TakeFunds(player, points); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

/**
* GET /fund
**/
func (h *Handlers) fundHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	playerID := r.Form.Get("playerId")
	points, err := getPointsFromString(r.Form.Get("points"))

	if err != nil || playerID == "" || points < 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	player, err := h.repo.FindOrCreatePlayer(playerID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.repo.AddFunds(player, points); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

/**
* GET /announceTournament
**/
func (h *Handlers) announceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tournamentID := r.Form.Get("tournamentId")
	deposit, err := getPointsFromString(r.Form.Get("deposit"))
	if err != nil || tournamentID == "" || deposit <= 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if err := h.repo.CreateTournament(tournamentID, deposit); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)

}

/**
* GET /joinTournament
**/
func (h *Handlers) joinHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	tournamentID := r.Form.Get("tournamentId")
	playerID := r.Form.Get("playerId")
	if tournamentID == "" || playerID == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	tournament, err := h.repo.FindTournament(tournamentID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := h.repo.TournamentJoinPlayers(tournament, playerID, r.Form["backerId"]); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)

}

/**
* POST /resultTournament
**/
func (h *Handlers) resultHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var results ResultsRequest
	if err := decoder.Decode(&results); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	tournament, err := h.repo.FindTournament(results.TournamentID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := h.repo.FinishTournament(tournament, results.Winners); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

}

/**
* GET /balance
**/
func (h *Handlers) balanceHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	player, err := h.repo.FindPlayer(r.Form.Get("playerId"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(player)
}

/**
* GET /reset
**/
func (h *Handlers) resetHandler(w http.ResponseWriter, r *http.Request) {
	h.repo.ResetDatabase()
	w.WriteHeader(http.StatusOK)
}
