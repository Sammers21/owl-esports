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

func getAndParse(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching page: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("Error fetching page: %v", resp.Status)
		return nil, err
	}
	parsed, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Printf("Error parsing body: %v", err)
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
	split := strings.Split(name, "-")
	caser := cases.Title(language.English)
	for i, s := range split {
		split[i] = caser.String(s)
	}
	return strings.Join(split, " ")
}
