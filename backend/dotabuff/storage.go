package dotabuff

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Storage struct {
	// original fields
	Heroes []*Hero
	Counters map[string][]*Counter

	// aggregated fields
	HeroShortNames map[string]*Hero
	CountersMap map[string]map[string]*Counter

	lock sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{
		Heroes: make([]*Hero, 0),
		HeroShortNames: make(map[string]*Hero),
		Counters: make(map[string][]*Counter),
		CountersMap: make(map[string]map[string]*Counter),
		lock: sync.Mutex{},
	}
}

func (s *Storage) Loaded() bool {
	status := s.lock.TryLock()
	if status {
		defer s.lock.Unlock()
	} 
	return status
}

func (s *Storage) LoadHeroes() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Info().Msg("Loading heroes...")
	heroes, err := Heroes()
	if err != nil {
		return err
	}
	log.Info().
		Int("count", len(heroes)).
		Msg("Heroes has been loaded")
	s.Heroes = heroes
	for _, hero := range heroes {
		s.HeroShortNames[hero.Name] = hero
		s.HeroShortNames[strings.ToLower(hero.Name)] = hero
		addShortNames(&s.HeroShortNames, hero)
	}
	return nil
}

func (s *Storage) FindHero(name string) (*Hero, bool) {
	hero, ok := s.HeroShortNames[name]
	return hero, ok
}

func (s *Storage) FindHeroes(names []string) ([]*Hero, error) {
	heroes := make([]*Hero, 0, len(names))
	for _, name := range names {
		hero, ok := s.FindHero(name)
		if !ok {
			return nil, fmt.Errorf("Hero %s not found", name)
		}
		heroes = append(heroes, hero)
	}
	return heroes, nil
}

	

func addShortNames(mp *map[string]*Hero, hero *Hero) {
	addArr := func(arrWords []string) {
		firstLetter := string(arrWords[0][0])
		secondLetter := string(arrWords[1][0])
		combo := firstLetter + secondLetter
		comboLower := strings.ToLower(combo)
		(*mp)[combo] = hero
		(*mp)[comboLower] = hero
	}
	space := strings.Split(hero.Name, " ")
	dash := strings.Split(hero.Name, "-")
	if len(space) == 2 {
		addArr(space)
	}else if len(dash) == 2 {
		addArr(dash)
	} else if len(hero.Name) >= 4 {
		shortName := string(hero.Name[:4])
		shortNameLower := strings.ToLower(shortName)
		(*mp)[shortName] = hero
		(*mp)[shortNameLower] = hero
	}
}

func (s *Storage) LoadCounters() error {
	s.lock.Lock()
	defer s.lock.Unlock()
	tick := time.Now()
	for i, hero := range s.Heroes {
		counters, err := hero.Counters()
		if err != nil {
			log.Printf("Error fetching counters for %s: %v", hero.Name, err)
			continue
		}
		log.Info().Msgf("%d/%d: %s has %d counters", i+1, len(s.Heroes), hero.Name, len(counters))
		s.Counters[hero.Name] = counters
		for _, c := range counters {
			if _, ok := s.CountersMap[c.Hero.Name]; !ok {
				s.CountersMap[c.Hero.Name] = make(map[string]*Counter)
			}
			s.CountersMap[c.Hero.Name][hero.Name] = c
		}
	}
	log.Info().Msgf("Counters has been loaded in %0.2f seconds", time.Since(tick).Seconds())
	return nil
}

func (s *Storage) PickWinRate(radiant, dire []*Hero) (float64, float64) {
	var radiantWinRate, direWinRate float64
	for _, hero := range radiant {
		counterArr := make([]*Counter, 0)
		for _, enemy := range dire {
			countersOfHero := s.CountersMap[hero.Name]
			counterArr = append(counterArr, countersOfHero[enemy.Name])
		}
		radiantWinRate += hero.WinRateVsPick(counterArr)
	}
	for _, hero := range dire {
		counterArr := make([]*Counter, 0)
		for _, enemy := range radiant {
			countersOfHero := s.CountersMap[hero.Name]
			counterArr = append(counterArr, countersOfHero[enemy.Name])
		}
		direWinRate += hero.WinRateVsPick(counterArr)
	}
	return radiantWinRate / 5, direWinRate / 5
}


