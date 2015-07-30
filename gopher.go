package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)


// Gopher item type characters, as defined in RFC1436, 3.8  Item type characters
// See: https://www.ietf.org/rfc/rfc1436.txt
type GopherType rune

const (
	FileItem      GopherType = '0'
	DirectoryItem GopherType = '1'
	ErrorItem     GopherType = '3'
	HTMLItem      GopherType = 'h'
	InfoItem      GopherType = 'i'
)


//"<Type><Title>\t<Selector>\t<Addr>\t<Port>\r\n", as RFC1436 specified.
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
	line, _ = reader.ReadString('\n') // read until \n
	line = strings.Trim(line, "\r\n ") // strip \r and \n, because Gopher standard specifies that lines will end with \r\n
	log.Printf("%s \"/%s\"", conn.RemoteAddr(), line)

	if line == "" { // empty line is like http request for /, so let's show the first page
		line = "page/1"
	}

	HandleRequest(conn, line) // defined in hackernews.go
}
