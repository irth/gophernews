package main

import (
	"flag"
	"log"
	"net"
)

var addr *string
var remoteaddr *string
var remoteport *int
var cache_time *int

func main() {
	addr = flag.String("listen", ":9876", "Address that the server will listen on (passed as is to net.Listen)")
	remoteaddr = flag.String("remoteaddr", "localhost", "IP or domain name of this server, to be used when generating links to items")
	remoteport = flag.Int("remoteport", 9876, "Port that this server will be available on, to be used when generating links to items")
	cache_time = flag.Int("cachetime", 1200, "Cached items' life span")
	flag.Parse()
	log.Println("Launching Gopher server...")

	var ln net.Listener
	var err error

	// listen on all interfaces
	ln, err = net.Listen("tcp", *addr)

	if err != nil {
		log.Fatalln(err) // an error occured, can't listen
	}

	// loop forever
	for {
		var conn net.Conn
		conn, err = ln.Accept()
		if err != nil {
			log.Println(err)
		} else {
			go HandleConnection(conn) // HandleConnection is defined in gopher.go
		}
	}
}
