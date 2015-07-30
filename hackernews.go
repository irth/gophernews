package main

import (
	"fmt"
	"github.com/jmcvetta/napping"
	"github.com/kennygrant/sanitize"
	"html"
	"net"
	"strconv"
	"strings"
	"time"
)

// urls for Hacker News's Firebase API endpoints
const (
	api        = "https://hacker-news.firebaseio.com/v0/"
	topstories = api + "topstories.json"
	item_url   = api + "item/"
)

// shorthand for showing an error
func gopherError(s string) []GopherItem {
	return []GopherItem{GopherItem{ErrorItem, s, "", "", 0}}
}

// https://github.com/HackerNews/API#items
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
	Deleted     bool   `json:"deleted"`
	Dead        bool   `json:"dead"`
	Parts       []int  `json:"parts"`
	RequestTime int
}

// there we will store items to be cached
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

// shorthand for sending menu to the client
func WriteMenu(conn net.Conn, items []GopherItem) {
	for _, item := range items {
		conn.Write(item.Bytes())
	}
	conn.Write([]byte(".\r\n")) // "." means bye in Gopher language
}

// shorthand for getting gopheritems that are on the same server (to avoid typing *remoteaddr, *remoteport every time. DRY)
func LocalItem(itemtype GopherType, itemtitle, itemselector string) GopherItem{
	return GopherItem{
					Type:     itemtype,
					Addr:     *remoteaddr,
					Port:     *remoteport,
					Title:  itemtitle,
					Selector: itemselector,
				}
}

// function called in gopher.go
func HandleRequest(conn net.Conn, selector string) {
	// this is a simple "router"
    if strings.HasPrefix(selector, "page/") { // top stories
		n, err := strconv.ParseInt(selector[5:], 10, 32) // parse page number
		if err == nil {
			min := (n - 1) * 10 // determine first and last item ID for given page
			max := min + 9
			if n < 1 { // well, negative page number?
				WriteMenu(conn, gopherError("Invalid page number."))
			} else {
				var item_ids []int
				napping.Get(topstories, nil, &item_ids, nil) // get top stories from hacker news API
				if int(max) > len(item_ids) { 
					WriteMenu(conn, gopherError("Invalid page number."))
					return
				}
				var items []GopherItem
				var header = LocalItem(
                    InfoItem,
                    fmt.Sprintf("*** GopherNews | PAGE %d | Data from Hacker News ***", n),
                    fmt.Sprintf("page/%d", n))
				items = append(items, header)

				if n > 1 {
					var prev = LocalItem(DirectoryItem, "[Previous page...]", fmt.Sprintf("page/%d", n-1))
					items = append(items, prev)
				}

				for _, id := range item_ids[min : max+1] {
					hnitem := GetItem(id, *cache_time)
					gopheritem := LocalItem(
                        DirectoryItem,
                        fmt.Sprintf("[Score: %d] %s", hnitem.Score, hnitem.Title),
                        fmt.Sprintf("item/%d", hnitem.ID))
					items = append(items, gopheritem)
				}

				var next = LocalItem(
                    DirectoryItem,
                    "[Next page...]",
					fmt.Sprintf("page/%d", n+1))
				items = append(items, next)

                WriteMenu(conn, items)
			}
		} else {
			WriteMenu(conn, gopherError("Invalid page number."))
		}
	} else if strings.HasPrefix(selector, "item/") { // specific items (in hacker news everything is an item, and has an unique ID. comments, stories, everything
		n, err := strconv.ParseInt(selector[5:], 10, 32) // parse item id
		if err == nil {
			if n < 0 {
				WriteMenu(conn, gopherError("Invalid item number."))
				return
			} else {
				item := GetItem(int(n), *cache_time)
				var menu []GopherItem
                // we display stories and comments a little different, comments do not have score, or simple means to determine the count of descendants
                // (we don't want to traverse recursively and count, because that would take too much time)
				if item.Type == "story" {
					if len(item.URL) > 0 {
						link := LocalItem(HTMLItem, item.Title, fmt.Sprintf("URL:%s", item.URL))
						menu = append(menu, link)
					}

					if len(item.Text) > 0 {
						text := LocalItem(HTMLItem, "[Click here to see the text...]", fmt.Sprintf("text/%d", item.ID))
						menu = append(menu, text)
					}

					info := LocalItem(
                        InfoItem,
						fmt.Sprintf("Author: %s, score: %d, %d comment(s).", item.Author, item.Score, item.Descendants),
						fmt.Sprintf("item/%d", n));
					menu = append(menu, info)
				} else if item.Type == "comment" {
					info := LocalItem(InfoItem, fmt.Sprintf("Author: %s.", item.Author), fmt.Sprintf("item/%d", n))
					text := LocalItem(HTMLItem, "[Click here to see the text...]", fmt.Sprintf("text/%d", item.ID))

					if item.Dead {
						text.Title = "[dead]"
						text.Type = InfoItem
					}
					if item.Deleted {
						text.Title = "[deleted]"
						text.Type = InfoItem
					}

					parent := LocalItem(DirectoryItem, "[Click here to go to the parent...]", fmt.Sprintf("item/%d", item.Parent))

					menu = append(menu, text)
					menu = append(menu, parent)

                    if !(item.Dead || item.Deleted) {
						menu = append(menu, info)
					}
				}

				menu = append(menu, LocalItem(InfoItem, "", fmt.Sprintf("item/%d", item.ID))) // separator

				for _, child_id := range item.Children {
					child := GetItem(child_id, *cache_time)
					shorttext := strings.Replace(strings.Replace(html.UnescapeString(sanitize.HTML(child.Text)), "\t", " ", -1), "\n", " ", -1)
					if len(shorttext) > 120 {
						shorttext = shorttext[:117] + "..."
					}
					if child.Dead {
						shorttext = "[dead]"
					}
					if child.Deleted {
						shorttext = "[deleted]"
					}

					child_item := LocalItem(DirectoryItem, shorttext, fmt.Sprintf("item/%d", child.ID))
					menu = append(menu, child_item)
				}

				WriteMenu(conn, menu)
			}
		}
	} else if strings.HasPrefix(selector, "URL:") { // for clients that dont support http links
		fmt.Fprintf(conn, "<meta http-equiv=\"refresh\" content=\"0; url=%s\"><a href=\"%s\">Click here if automatic redirect does not work.</a>", selector[4:], selector[4:])
	} else if strings.HasPrefix(selector, "text/") { // get html code for item (for example to see comment)
		n, err := strconv.ParseInt(selector[5:], 10, 32)
		if err == nil {
			if n < 0 {
				fmt.Fprintf(conn, "Invalid item number.\r\n")
			} else {
				item := GetItem(int(n), *cache_time)
				fmt.Fprintln(conn, item.Text)
			}
		} else {
			fmt.Fprintf(conn, "Invalid item number.\r\n")
		}
	}

}
