# GopherNews

A proxy that allows you to browse Hacker News using your favourite Gopher client, written in Go!

# Usage

    go get github.com/irth/gophernews
    gophernews [--args]

To connect:

    lynx gopher://localhost:9876

# Arguments

    $ gophernews --help
    Usage of gophernews:
      -cachetime int
        	Cached items' life span (default 1200)
      -listen string
        	Address that the server will listen on (passed as is to net.Listen) (default ":9876")
      -remoteaddr string
        	IP or domain name of this server, to be used when generating links to items (default "localhost")
      -remoteport int
        	Port that this server will be available on, to be used when generating links to items (default 9876)

# Demo server
Check it out at [gopher://hn.irth.pl/](gopher://hn.irth.pl).
