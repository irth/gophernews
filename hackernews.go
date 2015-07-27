package main

import (
	_ "errors"
	"fmt"
	"github.com/jmcvetta/napping"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
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
	Parent      int    `json:"parent"`
	Score       int    `json:"score"`
	Text        string `json:"text"`
	Time        int    `json:"time"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	RequestTime int
}

var items_cache map[int]HNItem

func GetItem(id int, lifespan int) HNItem {
	if items_cache == nil {
		items_cache = make(map[int]HNItem)
	}
	if item, ok := items_cache[id]; ok {
		if int(time.Now().Unix())-item.RequestTime <= lifespan {
			return item
		}
	}
	var item HNItem
	napping.Get(fmt.Sprintf("%s%d.json", item_url, id), nil, &item, nil)
	item.RequestTime = int(time.Now().Unix())
	items_cache[id] = item
	return item
}

func WriteMenu(conn net.Conn, items []GopherItem) {
	for _, item := range items {
		conn.Write(item.Bytes())
	}
	conn.Write([]byte(".\r\n"))
}

func HandleRequest(conn net.Conn, selector string) { //GopherItem is defined in gopher.go
	log.Println(selector)
	if strings.HasPrefix(selector, "page/") {
		n, err := strconv.ParseInt(selector[5:], 10, 32)
		if err == nil {
			min := (n - 1) * 10
			max := min + 9
			if n < 1 {
				WriteMenu(conn, gopherError("Invalid page number."))
			} else {
				var item_ids []int
				napping.Get(topstories, nil, &item_ids, nil)
				if int(max) > len(item_ids) {
					WriteMenu(conn, gopherError("Invalid page number."))
					return
				}
				var items []GopherItem
				var header = GopherItem{
					Type:     InfoItem,
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
					hnitem := GetItem(id, 300)
					gopheritem := GopherItem{
						Type:     DirectoryItem,
						Title:    fmt.Sprintf("[Score: %d] %s", hnitem.Score, hnitem.Title),
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
				WriteMenu(conn, items)
			}
		} else {
			WriteMenu(conn, gopherError("Invalid page number."))
		}
	} else if strings.HasPrefix(selector, "item/") {
		n, err := strconv.ParseInt(selector[5:], 10, 32)
		if err == nil {
			if n < 0 {
				WriteMenu(conn, gopherError("Invalid item number."))
				return
			} else {
				item := GetItem(int(n), 300)
				var menu []GopherItem
				if item.Type == "story" {
					link := GopherItem{
						Type:     HTMLItem,
						Title:    item.Title,
						Addr:     *remoteaddr,
						Port:     *remoteport,
						Selector: fmt.Sprintf("URL:%s", item.URL),
					}
					info := GopherItem{
						Type:     InfoItem,
						Title:    fmt.Sprintf("Author: %s, score: %d, %d comment(s).", item.Author, item.Score, item.Descendants),
						Addr:     *remoteaddr,
						Port:     *remoteport,
						Selector: fmt.Sprintf("item/%d", n),
					}

					menu = append(menu, link)
					menu = append(menu, info)

					if len(item.Text) > 0 {
						text := GopherItem{
							Type:     HTMLItem,
							Title:    "[Click here to see the text...]",
							Addr:     *remoteaddr,
							Port:     *remoteport,
							Selector: fmt.Sprintf("text/%d", item.ID),
						}
						menu = append(menu, text)
					}
				} else if item.Type == "comment" {
					info := GopherItem{
						Type:     InfoItem,
						Title:    fmt.Sprintf("Author: %s, score: %d.", item.Author, item.Score)
						Addr:     *remoteaddr,
						Port:     *remoteport,
						Selector: fmt.Sprintf("item/%d", n),
					}
					text := GopherItem{
						Type:     HTMLItem,
						Title:    "[Click here to see the comment...]",
						Addr:     *remoteaddr,
						Port:     *remoteport,
						Selector: fmt.Sprintf("text/%d", item.ID),
					}
					parent := GopherItem{
						Type:     DirectoryItem,
						Title:    "[Click here to go to the parent...]",
						Addr:     *remoteaddr,
						Port:     *remoteport,
						Selector: fmt.Sprintf("item/%d", item.Parent),
					}

					menu = append(menu, text)
					menu = append(menu, info)
					menu = append(menu, parent)
				}

				for _, child_id := range item.Children {
					child := GetItem(child_id, 300)
					shorttext := strings.Replace(child.Text, "\t", " ", -1)
					if len(child.Text) > 68 {
						shorttext = shorttext[:55] + "..."
					}
					shorttext = fmt.Sprintf("[Score: %d] %s", child.Score, shorttext)
					child_item := GopherItem{
						Type:     DirectoryItem,
						Title:    shorttext,
						Addr:     *remoteaddr,
						Port:     *remoteport,
						Selector: fmt.Sprintf("item/%d", child.ID),
					}
					menu = append(menu, child_item)
				}

				WriteMenu(conn, menu)
			}
		}
	} else if strings.HasPrefix(selector, "URL:") {
		fmt.Fprintf(conn, "<meta http-equiv=\"refresh\" content=\"0; url=%s\"><a href=\"%s\">Click here if automatic redirect does not work.</a>", selector[4:], selector[4:])
	} else if strings.HasPrefix(selector, "text/") {
		n, err := strconv.ParseInt(selector[5:], 10, 32)
		if err == nil {
			if n < 0 {
				fmt.Fprintf(conn, "Invalid item number.\r\n")
			} else {
				item := GetItem(int(n), 300)
				fmt.Fprintln(conn, item.Text)
			}
		} else {
			fmt.Fprintf(conn, "Invalid item number.\r\n")
		}
	}

}
