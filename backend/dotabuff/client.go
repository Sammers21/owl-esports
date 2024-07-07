package dotabuff

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

func (h *Hero) WinRateVsPick(pick []*Counter) float64 {
	var total float64 = 0
	for _, c := range pick {
		total += c.WinRate
	}
	return total / 5
}

func (h *Hero) Counters() ([]*Counter, error) {
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
	return res, nil
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
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching page")
		return nil, err
	}
	defer resp.Body.Close()
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

func Heroes() ([]*Hero, error) {
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
	}
	split := strings.Split(name, "-")
	caser := cases.Title(language.English)
	for i, s := range split {
		split[i] = caser.String(s)
	}
	return strings.Join(split, " ")
}
