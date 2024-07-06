package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
	"github.io/sammers21/owl-esports/backend/dotabuff"
)

type Server struct {
	Storage *dotabuff.Storage
}

type Status struct {
	Ready bool `json:"ready"`
}

type PickWinrateRequest struct {
	Radiant []string `json:"radiant"`
	Dire []string `json:"dire"`
}

type PickWinrateResponse struct {
	RadiantWinrate float64 `json:"radiant_winrate"`
	DireWinrate float64 `json:"dire_winrate"`
}

func NewServer(storage *dotabuff.Storage) *Server {
	return &Server{
		Storage: storage,
	}
}

func (s *Server) Start(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"hello\": \"world\"}"))
	})
	// LGD vs IG
	// LGD: muerta, es, beastmaster, tiny, sd
	// IG: gyro, snapfire, underlord, hoodwink, cm
	// curl -X POST -H "Content-Type: application/json" -d '{"radiant": ["muerta", "es", "beastmaster", "tiny", "sd"], "dire": ["gyro", "snapfire", "underlord", "hoodwink", "cm"]}' http://localhost:8080/pick-winrate_v1
	mux.HandleFunc("/pick-winrate_v1", func(w http.ResponseWriter, r *http.Request) {
		if !s.Storage.Loaded() {
			http.Error(w, "Storage not loaded", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var req PickWinrateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Error().Err(err).Msg("Error decoding pick winrate request v1")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		radiantHeroes, err := s.Storage.FindHeroes(req.Radiant)
		if err != nil {
			log.Error().Err(err).Msg("Error fetching radiant heroes")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		direHeroes, err := s.Storage.FindHeroes(req.Dire)
		if err != nil {
			log.Error().Err(err).Msg("Error fetching dire heroes")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		radiantWinrate, direWinrate := s.Storage.PickWinRate(radiantHeroes, direHeroes)
		resp := PickWinrateResponse{
			RadiantWinrate: radiantWinrate,
			DireWinrate: direWinrate,
		}
		json, err := json.Marshal(resp)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling pick winrate response v1")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(json)
	})

	
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := Status{
			Ready: s.Storage.Loaded(),
		}
		json, err :=json.Marshal(status)
		if err != nil {
			log.Error().Err(err).Msg("Error marshalling status")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return 
		}
		w.Write(json)
	})
	handler := cors.Default().Handler(mux)
	log.Info().Msgf("Http server started on port %d", port)
	http.ListenAndServe(":"+fmt.Sprintf("%d", port), handler)
	return nil
}