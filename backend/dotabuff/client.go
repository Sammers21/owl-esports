package dotabuff

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Hero struct {
	Name string
	Link string
}

func Heroes() ([]*Hero, error) {
	resp, err := http.Get("https://www.dotabuff.com/heroes")
	if err != nil {
		log.Printf("Error fetching page: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("Error fetching page: %v", resp.Status)
		return nil, err
	}
	if err != nil {
		log.Printf("Error reading body: %v", err)
		return nil, err
	}
	parsed, err := htmlquery.Parse(resp.Body)
	if err != nil {
		log.Printf("Error parsing body: %v", err)
		return nil, err
	}
	heroes := make(map[string]*Hero)
	// /html/body/div[2]/div[2]/div[3]/div[5]/div[1]/div/div/div[2]/div[1]/div[2]/div/div[2]/table/tbody/tr[1]/td[1]/div/a
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