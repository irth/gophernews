package main

import (
	_ "errors"
	"fmt"
	"github.com/jmcvetta/napping"
	"log"
	"strconv"
	"strings"
)

const (
	api        = "https://hacker-news.firebaseio.com/v0/"
	topstories = api + "topstories.json"
	item_url   = api + "item/"
)

func gopherError(s string) []GopherItem {
	return []GopherItem{GopherItem{ErrorItem, s, "", "", 0}}
}

type HNItem struct {
	Author      string `json:"by"`
	Descendants int    `json:"descendants"`
	ID          int    `json:"id"`
	Children    []int  `json:"kids"`
	Score       int    `json:"score"`
	Text        string `json:"text"`
	Time        int    `json:"time"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	URL         string `json:"url"`
}

func GetItems(selector string) ([]GopherItem, error) { //GopherItem is defined in gopher.go
	log.Println(selector)
	if strings.HasPrefix(selector, "page/") {
		n, err := strconv.ParseInt(selector[5:], 10, 32)
		if err == nil {
			min := (n - 1) * 10
			max := min + 9
			if n < 1 {
				return gopherError("Invalid page number."), nil
			} else {

				var item_ids []int
				napping.Get(topstories, nil, &item_ids, nil)
				if int(max) > len(item_ids) {
					return gopherError("Invalid page number."), nil
				}
				var items []GopherItem
				var header = GopherItem{
					Type:     DirectoryItem,
					Addr:     *remoteaddr,
					Port:     *remoteport,
					Title:    fmt.Sprintf("*** GopherNews | PAGE %d | Data from Hacker News ***", n),
					Selector: fmt.Sprintf("page/%d", n),
				}
				items = append(items, header)

				if n > 1 {
					var prev = GopherItem{
						Type:     DirectoryItem,
						Addr:     *remoteaddr,
						Port:     *remoteport,
						Title:    "[Previous page...]",
						Selector: fmt.Sprintf("page/%d", n-1),
					}
					items = append(items, prev)
				}

				for _, id := range item_ids[min : max+1] {
					var hnitem HNItem
					napping.Get(fmt.Sprintf("%s%d.json", item_url, id), nil, &hnitem, nil)
					gopheritem := GopherItem{
						Type:     DirectoryItem,
						Title:    hnitem.Title,
						Selector: fmt.Sprintf("item/%d", hnitem.ID),
						Addr:     *remoteaddr,
						Port:     *remoteport,
					}
					items = append(items, gopheritem)
				}

				var next = GopherItem{
					Type:     DirectoryItem,
					Addr:     *remoteaddr,
					Port:     *remoteport,
					Title:    "[Next page...]",
					Selector: fmt.Sprintf("page/%d", n+1),
				}
				items = append(items, next)
				return items, nil

			}
		}
	}
	return []GopherItem{
		GopherItem{
			Type:     DirectoryItem,
			Title:    "yay",
			Selector: "page/-2",
			Addr:     "localhost",
			Port:     9876,
		},
	}, nil
}
