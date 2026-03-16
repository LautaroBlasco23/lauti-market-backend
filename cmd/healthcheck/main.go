package main

import (
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.DialTimeout("tcp", "localhost:8080", 2*time.Second)
	if err != nil {
		os.Exit(1)
	}
	conn.Close()
}
