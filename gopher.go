package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type GopherType rune

const (
	FileItem      GopherType = '0'
	DirectoryItem GopherType = '1'
	ErrorItem     GopherType = '3'
)

type GopherItem struct {
	Type     GopherType
	Title    string
	Selector string
	Addr     string
	Port     int
}

func (g GopherItem) String() string {
	return fmt.Sprintf("%s%s\t%s\t%s\t%d\r\n", string(g.Type), g.Title, g.Selector, g.Addr, g.Port)
}

func (g GopherItem) Bytes() []byte {
	return []byte(g.String())
}

func HandleConnection(conn net.Conn) {
	defer conn.Close() // make sure that the connection will close when the function exits

	reader := bufio.NewReader(conn) // create a buffered reader for connection
	var line string
	line, _ = reader.ReadString('\n')
	log.Printf(line)
	line = strings.Trim(line, "\r\n ") // strip unnecessary characters
	log.Printf("%s requested \"/%s\"", conn.RemoteAddr(), line)

	if line == "" { // empty line is like http request for /
		line = "page/1"
	}

	items, err := GetItems(line) // defined in hackernews.go
	if err != nil {
		conn.Write(GopherItem{
			Type:     ErrorItem,
			Title:    err.Error(),
			Selector: "",
			Addr:     "",
			Port:     0,
		}.Bytes())
	} else {
		for _, item := range items {
			conn.Write(item.Bytes())
		}
	}

	// dot means bye in Gopher language
	conn.Write([]byte(".\r\n"))
}
