package dotabuff

import (
	"database/sql"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

type MySQL struct {
	db *sql.DB
}

func NewMySQL(conn string) (*MySQL, error) {
	if conn == "" {
		log.Info().Msg("MySQL is not configured")
		return nil, nil
	}
	db, err := sql.Open("mysql", conn)
	if err != nil {
		return nil, err
	}
	log.Info().Str("conn", conn).Msg("Connected to MySQL")
	return &MySQL{db: db}, nil
}

func (m *MySQL) InsertDotabuffMatch(match *DotabuffMatch, algorithmVersion string, radiantWinPrediction float64, direWinPrediction float64) error {
	if m.db == nil {
		return nil
	}
	radiantWinPredictionBool := radiantWinPrediction > direWinPrediction
	queryMultiline:= `INSERT IGNORE INTO dotabuff_match (
		id,
		dire_heroes,
		radiant_heroes,
		radiant_won,
		radiant_won_prediction,
		tournament_link,
		dire_team_link,
		radiant_team_link,
		radiant_win_prediction,
		dire_win_prediction,
		algorithm_version
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	query := strings.ReplaceAll(queryMultiline, "\n", "")
	_, err := m.db.Exec(query,
		match.MatchID,
		heroesToString(match.Dire),
		heroesToString(match.Radiant),
		match.RadiantWon,
		radiantWinPredictionBool,
		match.TournamentLink,
		match.DireTeam.Link,
		match.RadiantTeam.Link,
		radiantWinPrediction,
		direWinPrediction,
		algorithmVersion,
	)
	if err != nil {
		log.Error().Err(err).Msg("Error inserting dotabuff match")
	}
	return err
}

func heroesToString(heroes []*Hero) string {
	names := make([]string, 0)
	for _, hero := range heroes {
		names = append(names, hero.Name)
	}
	return strings.Join(names, " ")
}