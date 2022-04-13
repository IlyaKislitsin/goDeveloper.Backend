package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	go func() {
		io.Copy(os.Stdout, conn)
	}()
	_, err = io.Copy(conn, os.Stdin) // until you send ^Z
	if err != nil {
		fmt.Printf("io.Copy error: %s\n", err)
	}
	fmt.Printf("%s: exit", conn.LocalAddr())
}
