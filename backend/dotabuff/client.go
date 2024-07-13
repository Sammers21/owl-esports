package dotabuff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type DotabuffMatch struct {
	MatchID        int64
	Dire           []*Hero
	Radiant        []*Hero
	DireTeam       *Team
	RadiantTeam    *Team
	RadiantWon     bool
	TournamentLink string
}

type RadiantDireWinrate struct {
	Hero            *Hero
	RadiantWinrate  float64
	RadiantPickRate float64
	DireWinrate     float64
	DirePickRate    float64
}

type Team struct {
	Name string
	Link string
}

type Hero struct {
	Name string
	Link string
}

type Counter struct {
	Hero          *Hero
	Disadvantage  float64
	WinRate       float64
	MatchesPlayed int64
}

func (h *Hero) WinRateVsPick(pick []*Counter, radiant bool, SideMultiplier float64) float64 {
	var total float64 = 0
	for _, c := range pick {
		total += c.WinRate
	}
	averageWR := total / 5
	return averageWR * SideMultiplier
}

func (h *Hero) Counters() ([]*Counter, error) {
	// if file counters.json exists, return counters from it
	// if not, fetch counters from dotabuff and save them to counters.json
	_ = os.Mkdir("counters", 0755)
	info, err := os.Stat(fmt.Sprintf("counters/%s.json", h.Name))
	needToUpdate := err != nil || time.Since(info.ModTime()) > 24*time.Hour
	countersJson, err := os.ReadFile(fmt.Sprintf("counters/%s.json", h.Name))
	if err == nil && !needToUpdate {
		log.Info().Msg("Counters file found, parsing...")
		counters, err := ParseCounters(countersJson)
		if err != nil {
			return nil, err
		}
		return counters, nil
	} else {
		log.Info().Msg("Counters file not found, fetching from dotabuff...")
		parsed, err := getAndParse(h.Link + "/counters")
		if err != nil {
			return nil, err
		}
		res := make([]*Counter, 0)
		find := htmlquery.Find(parsed, "//table/tbody/tr[@data-link-to]")
		for _, n := range find {
			counter, err := CounterNodeToCounter(n)
			if err != nil {
				log.Printf("Error parsing counter: %v", err)
				continue
			}
			res = append(res, counter)
		}
		err = SaveCounters(res, h.Name)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func ParseCounters(countersJson []byte) ([]*Counter, error) {
	res := make([]*Counter, 0)
	err := json.Unmarshal(countersJson, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func SaveCounters(res []*Counter, name string) (err error) {
	b, err := json.Marshal(res)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("counters/%s.json", name), b, 0644)
	if err != nil {
		return err
	}
	return
}

func NodeToString(n *html.Node) string {
	var b bytes.Buffer
	err := html.Render(&b, n)
	if err != nil {
		log.Printf("Error rendering node: %v", err)
		return ""
	}
	return b.String()
}

func ExtractHerosFromDBLink(link string) (*DotabuffMatch, error) {
	parsed, err := getAndParse(link)
	if err != nil {
		return nil, err
	}
	radiant, rTeam, rWon := ParseSide("radiant", parsed)
	dire, dTeam, _ := ParseSide("dire", parsed)
	parsedTournamentLink := ParseTournamentLink(parsed)
	matchIdStr := strings.TrimPrefix(link, "https://www.dotabuff.com/matches/")
	matchId, err := strconv.ParseInt(matchIdStr, 10, 64)
	if err != nil {
		log.Printf("Error parsing match id: %v", err)
		return nil, err
	}
	return &DotabuffMatch{
		MatchID:        matchId,
		Dire:           dire,
		Radiant:        radiant,
		DireTeam:       dTeam,
		RadiantTeam:    rTeam,
		RadiantWon:     rWon,
		TournamentLink: parsedTournamentLink,
	}, nil
}

func ParseTournamentLink(root *html.Node) string {
	a := htmlquery.FindOne(root, "//dd/a[@class='esports-link']")
	link := htmlquery.SelectAttr(a, "href")
	return fmt.Sprintf("https://www.dotabuff.com%v", link)
}

func ExtractHerosFromDBMatch(id int64) (*DotabuffMatch, error) {
	link := fmt.Sprintf("https://www.dotabuff.com/matches/%d", id)
	return ExtractHerosFromDBLink(link)
}

func ParseSide(side string, root *html.Node) ([]*Hero, *Team, bool) {
	section := htmlquery.FindOne(root, fmt.Sprintf("//section[@class='%s']", side))
	sectionChilds := ChildArray(section)
	header := sectionChilds[0]
	headerChilds := ChildArray(header)
	a := headerChilds[0]
	aHref := htmlquery.SelectAttr(a, "href")
	aChilds := ChildArray(a)
	spanWTeamName := aChilds[1]
	teamName := htmlquery.InnerText(spanWTeamName)
	article := sectionChilds[1]
	articleChilds := ChildArray(article)
	table := articleChilds[0]
	tableChilds := ChildArray(table)
	tbodyTwo := tableChilds[1]
	tbodyTwoChilds := ChildArray(tbodyTwo)
	heroes := make([]*Hero, 0)
	for _, tr := range tbodyTwoChilds {
		hero := HeroFromTr(tr)
		heroes = append(heroes, hero)
	}
	team := &Team{
		Name: teamName,
		Link: fmt.Sprintf("https://www.dotabuff.com%v", aHref),
	}
	return heroes, team, len(headerChilds) > 1
}

func HeroFromTr(tr *html.Node) *Hero {
	trChilds := ChildArray(tr)
	td := trChilds[0]
	tdChilds := ChildArray(td)
	div := tdChilds[0]
	divChilds := ChildArray(div)
	div2 := divChilds[0]
	div2Childs := ChildArray(div2)
	div3 := div2Childs[0]
	div3Childs := ChildArray(div3)
	a := div3Childs[2]
	link := htmlquery.SelectAttr(a, "href")
	return DotaHeroFromLink(link)
}

func getAndParse(url string) (*html.Node, error) {
	return getAndParse0(url, 0)
}

func getAndParse0(url string, retry int) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching page")
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		log.Warn().Msg("Rate limited")
		sleepSec := 10 * (retry + 1)
		log.Info().Int("sleep-sec", sleepSec).Msg("Sleeping...")
		time.Sleep(time.Duration(sleepSec) * time.Second)
		return getAndParse0(url, retry+1)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error fetching page: %v", resp.Status)
	}
	parsed, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing body")
		return nil, err
	}
	return parsed, nil
}

func CounterNodeToCounter(n *html.Node) (*Counter, error) {
	childs := ChildArray(n)
	if len(childs) != 5 {
		log.Error().Int("childs-len", len(childs)).Msg("Invalid amount of childs")
		return nil, fmt.Errorf("Invalid amount of childs")
	}
	heroName := htmlquery.SelectAttr(childs[0], "data-value")
	heroLink := htmlquery.SelectAttr(ChildArray(childs[1])[0], "href")
	disadvantage := htmlquery.SelectAttr(childs[2], "data-value")
	winRate := htmlquery.SelectAttr(childs[3], "data-value")
	matchesPlayed := htmlquery.SelectAttr(childs[4], "data-value")
	disParsed, err := strconv.ParseFloat(disadvantage, 64)
	if err != nil {
		log.Printf("Error parsing disadvantage: %v", err)
		return nil, err
	}
	winParsed, err := strconv.ParseFloat(winRate, 64)
	if err != nil {
		log.Printf("Error parsing win rate: %v", err)
		return nil, err
	}
	matchesParsed, err := strconv.ParseInt(matchesPlayed, 10, 64)
	if err != nil {
		log.Printf("Error parsing matches played: %v", err)
		return nil, err
	}
	return &Counter{
		Hero: &Hero{
			Name: heroName,
			Link: fmt.Sprintf("https://www.dotabuff.com%v", heroLink),
		},
		Disadvantage:  disParsed,
		WinRate:       winParsed,
		MatchesPlayed: matchesParsed,
	}, nil

}

func ChildArray(n *html.Node) []*html.Node {
	var res []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		res = append(res, c)
	}
	return res
}

func RaidantAndDireWR() ([]*RadiantDireWinrate, error) {
	// if file radiant_dire_winrate.json exists, return heroes from it
	// if not, fetch heroes from dotabuff and save them to radiant_dire_winrate.json
	radiantDireWinrateJson, err := os.ReadFile("radiant_dire_winrate.json")
	if err == nil {
		log.Info().Msg("RadiantDireWinrate file found, parsing...")
		radiantDireWinrate, err := ParseRadiantDireWinrate(radiantDireWinrateJson)
		if err != nil {

			return nil, err
		}
		return radiantDireWinrate, nil
	} else {
		log.Info().Msg("RadiantDireWinrate file not found, fetching from dotabuff...")
		parsed, err := getAndParse("https://www.dotabuff.com/heroes/meta?view=played&metric=faction")
		if err != nil {
			return nil, err
		}
		tbody := htmlquery.FindOne(parsed, "//section/footer/article/table/tbody")
		tbodyChilds := ChildArray(tbody)
		res := make([]*RadiantDireWinrate, 0)
		for _, tr := range tbodyChilds {
			trChilds := ChildArray(tr)
			tdOne := trChilds[1]
			a := ChildArray(tdOne)[0]
			href := htmlquery.SelectAttr(a, "href")
			hero := DotaHeroFromLink(href)
			tdTwo := trChilds[2]
			tdThree := trChilds[3]
			tdFour := trChilds[4]
			tdFive := trChilds[5]
			rdWinrate := htmlquery.SelectAttr(tdThree, "data-value")
			// parse rdWinrate to float
			rdWinrateParsed, err := strconv.ParseFloat(rdWinrate, 64)
			if err != nil {
				log.Printf("Error parsing rdWinrate: %v for hero %v", err, hero.Name)
				return nil, err
			}
			rdPickRate := htmlquery.SelectAttr(tdTwo, "data-value")
			rdPickRateParsed, err := strconv.ParseFloat(rdPickRate, 64)
			if err != nil {
				log.Printf("Error parsing rdPickRate: %v for hero %v", err, hero.Name)
				return nil, err
			}
			direWinrate := htmlquery.SelectAttr(tdFive, "data-value")
			direWinrateParsed, err := strconv.ParseFloat(direWinrate, 64)
			if err != nil {
				log.Printf("Error parsing direWinrate: %v for hero %v", err, hero.Name)
				return nil, err
			}
			direPickRate := htmlquery.SelectAttr(tdFour, "data-value")
			direPickRateParsed, err := strconv.ParseFloat(direPickRate, 64)
			if err != nil {
				log.Printf("Error parsing direPickRate: %v for hero %v", err, hero.Name)
				return nil, err
			}
			res = append(res, &RadiantDireWinrate{
				Hero:            hero,
				RadiantWinrate:  rdWinrateParsed,
				RadiantPickRate: rdPickRateParsed,
				DireWinrate:     direWinrateParsed,
				DirePickRate:    direPickRateParsed,
			})
		}
		err = SaveRadiantDireWinrate(res)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func SaveRadiantDireWinrate(res []*RadiantDireWinrate) (err error) {
	log.Info().Msg("Saving radiant dire winrate...")
	b, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error marshalling radiant dire winrate: %v", err)
		return
	}
	err = os.WriteFile("radiant_dire_winrate.json", b, 0644)
	if err != nil {
		log.Printf("Error saving radiant dire winrate: %v", err)
		return
	}
	return
}

func ParseRadiantDireWinrate(radiantDireWinrateJson []byte) ([]*RadiantDireWinrate, error) {
	res := make([]*RadiantDireWinrate, 0)
	err := json.Unmarshal(radiantDireWinrateJson, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func Heroes() ([]*Hero, error) {
	// if file heros.json exists, return heroes from it
	// if not, fetch heroes from dotabuff and save them to heros.json
	heroesJson, err := os.ReadFile("heroes.json")
	if err == nil {
		log.Info().Msg("Heroes file found, parsing...")
		heroes, err := ParseHeroes(heroesJson)
		if err != nil {
			return nil, err
		}
		return heroes, nil
	} else {
		log.Info().Msg("Heroes file not found, fetching from dotabuff...")
		parsed, err := getAndParse("https://www.dotabuff.com/heroes")
		if err != nil {
			return nil, err
		}
		heroes := make(map[string]*Hero)
		find := htmlquery.Find(parsed, "//table/tbody/*//a[@href]")
		for _, n := range find {
			href := htmlquery.SelectAttr(n, "href")
			// strings.Starts
			if len(href) > 8 && strings.HasPrefix(href, "/heroes/") {
				hero := DotaHeroFromLink(href)
				heroes[hero.Name] = hero
			}
		}
		res := make([]*Hero, 0, len(heroes))
		for _, hero := range heroes {
			res = append(res, hero)
		}
		err = SaveHeroes(res)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func SaveHeroes(heroes []*Hero) error {
	b, err := json.Marshal(heroes)
	if err != nil {
		return err
	}
	err = os.WriteFile("heroes.json", b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func ParseHeroes(b []byte) ([]*Hero, error) {
	res := make([]*Hero, 0)
	err := json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DotaHeroFromLink(link string) *Hero {
	name := prettyName(strings.TrimPrefix(link, "/heroes/"))
	return &Hero{
		Name: name,
		Link: fmt.Sprintf("https://www.dotabuff.com%v", link),
	}
}

func prettyName(name string) string {
	if name == "keeper-of-the-light" {
		return "Keeper of the Light"
	} else if name == "anti-mage" {
		return "Anti-Mage"
	}
	split := strings.Split(name, "-")
	caser := cases.Title(language.English)
	for i, s := range split {
		split[i] = caser.String(s)
	}
	return strings.Join(split, " ")
}
