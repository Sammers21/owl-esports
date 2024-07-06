package dotabuff

import (
	"fmt"
	"net/http"

	"github.com/antchfx/htmlquery"
	"github.com/rs/zerolog/log"
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
	// print all the response headers
	for k, v := range resp.Header {
		log.Printf("Header: %v: %v", k, v)
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
	heroes := []*Hero{}
	for _, n := range htmlquery.Find(parsed, "//a[@href]") {
		href := htmlquery.SelectAttr(n, "href")
		if len(href) > 7 && href[:7] == "/heroes" {
			log.Printf("Found hero: %v", htmlquery.InnerText(n))
			heroes = append(heroes, &Hero{
				Name: htmlquery.InnerText(n),
				Link: fmt.Sprintf("https://www.dotabuff.com%v", htmlquery.SelectAttr(n, "href")),
			})
		}
	}
	return heroes, nil
}
