package main

import (
	"flag"
	"log"
	"net"
)

// Arguments
var addr *string // this will be passed to net.Listen
var remoteaddr *string // this will be sent do Gopher clients as this servers's address
var remoteport *int // see above
var cache_time *int // how long items will be stored in cache 

func main() {
	// parse command line args
	addr = flag.String("listen", ":9876", "Address that the server will listen on (passed as is to net.Listen)")
	remoteaddr = flag.String("remoteaddr", "localhost", "IP or domain name of this server, to be used when generating links to items")
	remoteport = flag.Int("remoteport", 9876, "Port that this server will be available on, to be used when generating links to items")
	cache_time = flag.Int("cachetime", 1200, "Cached items' life span")
	flag.Parse()

	log.Println("Launching Gopher server...")

	var ln net.Listener
	var err error

	// listen on specified address
	ln, err = net.Listen("tcp", *addr)

	if err != nil {
		log.Fatalln(err) // an error occured, can't listen
	}
	log.Printf("Listening at %s", ln.Addr())

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
