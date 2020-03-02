package main

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"os"
	"strings"
)

type Item struct {
	Id     string
	Rating string
	Perks  []string
}

func main() {
	startParser(
		"light.gg.txt",
		"https://www.light.gg/db/items/%v",
		"https://www.light.gg/db/category/1?page=%v&f=4(6;5)", // Weapon: Exotic, Legendary
		1,
		11)
}

func startParser(outputFileName string, itemUrlFormat string, catalogUrlFormat string, catalogStartPage int, catalogEndPage int) {
	// file
	file, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", outputFileName, err)
		return
	}
	defer file.Close()

	c := colly.NewCollector(
		colly.Async(false),
		colly.MaxDepth(3),
		colly.CacheDir("./cache"),
	)

	// Call all links from page
	c.OnHTML("#item-list", func(e *colly.HTMLElement) {
		itemIds := e.ChildAttrs("div.item", "data-id")
		for _, itemId := range itemIds {
			log.Printf("Item Id:%v", itemId)
			err = e.Request.Visit(fmt.Sprintf(itemUrlFormat, itemId))
			if err != nil {
				log.Printf("Cannot call item URL %s\n", err)
				return
			}
		}
	})

	c.OnHTML("#item-container", func(e *colly.HTMLElement) {
		item := Item{
			Id: strings.TrimLeft(e.ChildText("#item-details li:last-child"), "API ID: "),
			Rating: e.ChildText("#community-rarity span strong") +
				" PVE:" + e.ChildText("#review-container div:nth-child(2) span") +
				" PVP:" + e.ChildText("#review-container div:nth-child(3) span"),
			Perks: make([]string, 0),
		}

		e.ForEach(".perks li.pref", func(i int, e *colly.HTMLElement) {
			item.Perks = append(item.Perks, e.ChildAttr(".item", "data-id"))
		})

		if len(item.Perks) == 0 {
			log.Printf("Skip because no perks: %v\n", item.Id)
			return
		}

		line := fmt.Sprintf("dimwishlist:item=%s&perks=%s#notes:%s\n", item.Id, strings.Join(item.Perks, ","), item.Rating)

		file.WriteString(line)
	})

	for page := catalogStartPage; page <= catalogEndPage; page++ {
		url := fmt.Sprintf(catalogUrlFormat, page)
		log.Println("Catalog URL:" + url)
		if err := c.Visit(url); err != nil {
			log.Fatalf("Cannot call collection URL %s\n", err)
			break
		}
	}
}
