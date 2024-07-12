package dotabuff

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Engine struct {
	// original fields
	Heroes   []*Hero
	Counters map[string][]*Counter
	SideWR   []*RadiantDireWinrate

	// aggregated fields
	HeroShortNames map[string]*Hero
	CountersMap    map[string]map[string]*Counter
	HeroSideWR     map[string]*RadiantDireWinrate

	// internal fields
	lock  sync.Mutex
	mysql *MySQL
}

func NewEngine(mysql *MySQL) *Engine {
	return &Engine{
		Heroes:         make([]*Hero, 0),
		HeroShortNames: make(map[string]*Hero),
		SideWR:         make([]*RadiantDireWinrate, 0),
		Counters:       make(map[string][]*Counter),
		CountersMap:    make(map[string]map[string]*Counter),
		HeroSideWR:     make(map[string]*RadiantDireWinrate),
		lock:           sync.Mutex{},
		mysql:          mysql,
	}
}

func (s *Engine) Loaded() bool {
	status := s.lock.TryLock()
	if status {
		defer s.lock.Unlock()
	}
	return status
}

func (s *Engine) LoadHeroes() error {
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
	log.Info().Msg("Loading radiant and dire winrates...")
	wrs, err := RaidantAndDireWR()
	if err != nil {
		log.Error().Err(err).Msg("Error loading radiant and dire winrates")
		return err
	}
	s.SideWR = wrs
	for _, wr := range wrs {
		s.HeroSideWR[wr.Hero.Name] = wr
	}
	log.Info().Msg("Radiant and dire winrates has been loaded")
	return nil
}

func (e *Engine) SideMultiplier(hero *Hero, radiant bool) float64 {
	return 1
	// wr := e.HeroSideWR[hero.Name]
	// var multiplier float64
	// if radiant {
	// 	multiplier = wr.RadiantWinrate / wr.DireWinrate
	// } else {
	// 	multiplier = wr.DireWinrate / wr.RadiantWinrate
	// }
	// return multiplier
}

func (e *Engine) FindHero(name string) (*Hero, bool) {
	hero, ok := e.HeroShortNames[name]
	return hero, ok
}

func (e *Engine) FindHeroes(names []string) ([]*Hero, error) {
	heroes := make([]*Hero, 0, len(names))
	for _, name := range names {
		hero, ok := e.FindHero(name)
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
	} else if len(dash) == 2 {
		addArr(dash)
	} else if len(hero.Name) >= 4 {
		shortName := string(hero.Name[:4])
		shortNameLower := strings.ToLower(shortName)
		(*mp)[shortName] = hero
		(*mp)[shortNameLower] = hero
	}
}

func (e *Engine) LoadCounters() error {
	e.lock.Lock()
	defer e.lock.Unlock()
	tick := time.Now()
	for i, hero := range e.Heroes {
		counters, err := hero.Counters()
		if err != nil {
			log.Printf("Error fetching counters for %s: %v", hero.Name, err)
			continue
		}
		log.Info().Msgf("%d/%d: %s has %d counters", i+1, len(e.Heroes), hero.Name, len(counters))
		e.Counters[hero.Name] = counters
		for _, c := range counters {
			if _, ok := e.CountersMap[c.Hero.Name]; !ok {
				e.CountersMap[c.Hero.Name] = make(map[string]*Counter, 0)
			}
			e.CountersMap[c.Hero.Name][hero.Name] = c
		}
	}
	log.Info().Msgf("Counters has been loaded in %0.2f seconds", time.Since(tick).Seconds())
	return nil
}

func (e *Engine) PickWinRate(radiant, dire []*Hero) (float64, float64) {
	var radiantWinRate, direWinRate float64
	for _, hero := range radiant {
		counterArr := make([]*Counter, 0)
		for _, enemy := range dire {
			countersOfHero := e.CountersMap[hero.Name]
			counterArr = append(counterArr, countersOfHero[enemy.Name])
		}

		radiantWinRate += hero.WinRateVsPick(counterArr, true, e.SideMultiplier(hero, true))
	}
	for _, hero := range dire {
		counterArr := make([]*Counter, 0)
		for _, enemy := range radiant {
			countersOfHeroMap := e.CountersMap[hero.Name]
			if countersOfHeroMap == nil {
				panic(fmt.Sprintf("Counters not found for %s", hero.Name))
			}
			heroCounterList := countersOfHeroMap[enemy.Name]
			if heroCounterList == nil {
				panic(fmt.Sprintf("Counter not found for %s vs %s", hero.Name, enemy.Name))
			}
			counterArr = append(counterArr, heroCounterList)
		}
		direWinRate += hero.WinRateVsPick(counterArr, false, e.SideMultiplier(hero, false))
	}
	totalR := radiantWinRate / 5
	totalD := direWinRate / 5
	return totalR, totalD
}

func (e *Engine) PickWinRateFromLines(all []string) (float64, float64, error) {
	if !e.Loaded() {
		return 0, 0, fmt.Errorf("Data has not been loaded yet. Please try again in like 30 seconds")
	}
	if len(all) != 10 {
		return 0, 0, fmt.Errorf("Invalid number of heroes: %d", len(all))
	}
	radiantHeroes, direHeroes, err := e.SplitToDireAndRadiant(all)
	if err != nil {
		return 0, 0, err
	}
	radiantWinRate, direWinRate := e.PickWinRate(radiantHeroes, direHeroes)
	return radiantWinRate, direWinRate, nil
}

func (e *Engine) SplitToDireAndRadiant(all []string) ([]*Hero, []*Hero, error) {
	radiant := all[:5]
	dire := all[5:]
	radiantHeroes, err := e.FindHeroes(radiant)
	if err != nil {
		return nil, nil, fmt.Errorf("Error finding radiant heroes: %v", err)
	}
	direHeroes, err := e.FindHeroes(dire)
	if err != nil {
		return nil, nil, fmt.Errorf("Error finding dire heroes: %v", err)
	}
	return radiantHeroes, direHeroes, nil
}

func (e *Engine) PickWinRateFromDBMatch(match *DotabuffMatch) (float64, float64, error) {
	if !e.Loaded() {
		return 0, 0, fmt.Errorf("Data has not been loaded yet. Please try again in like 30 seconds")
	}
	rw, dw := e.PickWinRate(match.Radiant, match.Dire)
	if e.mysql != nil {
		go e.mysql.InsertDotabuffMatch(match, "v1.1", rw, dw)
	}
	return rw, dw, nil
}

// GenerateHeatMap generates a heatmap of the winrate of the heroes
// you selected and returns the path to the image.
func (e *Engine) GenerateHeatMap(all []string) (string, error) {
	radiantHeroes, direHeroes, err := e.SplitToDireAndRadiant(all)
	if err != nil {
		return "", err
	}
	heroesFullNames := make([]string, 0, len(all))
	for _, name := range radiantHeroes {
		heroesFullNames = append(heroesFullNames, name.Name)
	}
	for _, name := range direHeroes {
		heroesFullNames = append(heroesFullNames, name.Name)
	}
	resCmdArg := ""
	heroesStr := strings.Join(heroesFullNames, ",")
	resCmdArg += fmt.Sprintf("--heroes=\"%s\"", heroesStr)
	winRatesArg := ""
	for _, rHero := range radiantHeroes {
		for _, dHero := range direHeroes {
			winRatesArg += fmt.Sprintf("%.2f,", e.CountersMap[dHero.Name][rHero.Name].WinRate)
		}
		winRatesArg = strings.TrimSuffix(winRatesArg, ",")
		if rHero != radiantHeroes[len(radiantHeroes)-1] {
			winRatesArg += ";"
		}
	}
	resCmdArg += fmt.Sprintf(" --winrates=\"%s\"", winRatesArg)
	cmd := exec.Command("/Users/sammers/miniforge3/bin/python3", "heatmap.py", fmt.Sprintf("--heroes=%s", heroesStr), fmt.Sprintf("--winrates=%s", winRatesArg))
	execAndPrint(cmd)
	return resCmdArg, nil
}

func execAndPrint(cmd *exec.Cmd) {
	log.Info().Msgf("Executing command: %v", cmd)
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	cmd.Stdout = mw
	cmd.Stderr = mw
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msg("Error executing command")
	}
	log.Info().Msg(stdBuffer.String())
}
