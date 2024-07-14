package dotabuff

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Engine *Engine
	Tg    *TelegramBot
}

type Status struct {
	Ready bool `json:"ready"`
}

type PickWinrateRequest struct {
	Radiant []string `json:"radiant"`
	Dire    []string `json:"dire"`
}

type PickWinrateResponse struct {
	RadiantWinrate float64 `json:"radiant_winrate"`
	DireWinrate    float64 `json:"dire_winrate"`
}

func NewServer(engine *Engine, tg *TelegramBot) *Server {
	return &Server{
		Engine: engine,
		Tg:    tg,
	}
}

func (s *Server) Start(port int) error {
	mux := http.NewServeMux()
	// curl -X GET http://localhost:8080/owl-esports/pickline -F line="troll warlord,vengeful spirit,sand king,jakiro,phoenix,elder titan,rubick,keeper of the light,doom,templar assassin" -F tg="77107633"
	mux.HandleFunc("/owl-esports/pickline", func(w http.ResponseWriter, r *http.Request) {		
		w.Header().Set("Content-Type", "application/json")
		// get param line
		line := r.URL.Query().Get("line")
		tg := r.URL.Query().Get("tg")
		if line == "" {
			http.Error(w, "line is required", http.StatusBadRequest)
			return
		}
		if tg == "" {
			http.Error(w, "tg is required", http.StatusBadRequest)
			return
		}
		// prase string to int64
		chatId, err := strconv.ParseInt(tg, 10, 64)
		if err != nil {
			http.Error(w, "tg is invalid", http.StatusBadRequest)
			return
		}
		lines := strings.Split(line, ",")
		if len(lines) != 10 {
			http.Error(w, "line is invalid", http.StatusBadRequest)
			return
		}
		s.Tg.SendPickWinRatesToUser(chatId, 0, lines)
		w.Write([]byte(`{"message": "Sending pick winrates to user"}`))
	})
	// LGD vs IG
	// LGD: muerta, es, beastmaster, tiny, sd
	// IG: gyro, snapfire, underlord, hoodwink, cm
	// curl -X POST -H "Content-Type: application/json" -d '{"radiant": ["muerta", "es", "beastmaster", "tiny", "sd"], "dire": ["gyro", "snapfire", "underlord", "hoodwink", "cm"]}' http://localhost:8080/pick-winrate_v1
	mux.HandleFunc("/pick-winrate_v1", func(w http.ResponseWriter, r *http.Request) {
		if !s.Engine.Loaded() {
			http.Error(w, "Data has not been loaded yet", http.StatusServiceUnavailable)
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
		radiantHeroes, err := s.Engine.FindHeroes(req.Radiant)
		if err != nil {
			log.Error().Err(err).Msg("Error fetching radiant heroes")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		direHeroes, err := s.Engine.FindHeroes(req.Dire)
		if err != nil {
			log.Error().Err(err).Msg("Error fetching dire heroes")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		radiantWinrate, direWinrate := s.Engine.PickWinRate(radiantHeroes, direHeroes)
		resp := PickWinrateResponse{
			RadiantWinrate: radiantWinrate,
			DireWinrate:    direWinrate,
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
			Ready: s.Engine.Loaded(),
		}
		json, err := json.Marshal(status)
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
