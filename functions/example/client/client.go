package main

import (
	"fmt"
	"log"
	"net"
	"github.com/tiborvass/tell/functions"
	"github.com/tiborvass/tell/functions/example/client/rpc"
	"github.com/docker/libchan/spdy"
)

func main() {
	conn, err := net.Dial("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	t, err := spdy.NewClientTransport(conn)
	if err != nil {
		log.Fatal(err)
	}
	sender, err := t.NewSendChannel()
	if err != nil {
		log.Fatal(err)
	}

	rpc := new(rpc.Functions)
	if err := functions.PairFromStruct(rpc, sender); err != nil {
		log.Fatal(err)
	}

	s, n := rpc.F1([]byte("hello world"), false, 24)

	fmt.Printf("client: results = %q, %d\n", s, n)
	rpc.F2()
	fmt.Println("client: exiting")
}