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
