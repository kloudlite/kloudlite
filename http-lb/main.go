package main

import (
	"flag"
	"log"
	"net"
)

func handleConn(conn net.Conn, svc string) {
	
}

func main() {
	var localAddr string
	var remoteAddr string

	flag.StringVar(&localAddr, "local-addr", "0.0.0.0:80", "--local-addr <host:port>")
	flag.StringVar(&remoteAddr, "remote-addr", "", "--remote-addr <host:port>")
	flag.Parse()

	if remoteAddr == "" {
		panic("flag remoteAddr should not be empty")
	}

	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn, "port-80-svc")
	}
}
