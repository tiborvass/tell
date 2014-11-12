package main

import (
	"log"
	"net"
	"github.com/tiborvass/tell/functions"
	"github.com/docker/libchan/spdy"
)

func main() {
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		t, err := spdy.NewServerTransport(conn)
		if err != nil {
			log.Fatal(err)
		}
		receiver, err := t.WaitReceiveChannel()
		if err != nil {
			log.Fatal(err)
		}

		ch := make(chan error)
		go func() {
			select {
			case err := <-ch:
				log.Fatal(err)
			}
		}()

		if err := functions.Export(PublicFunctions{}, receiver, ch); err != nil {
			log.Fatal(err)
		}
	}
}