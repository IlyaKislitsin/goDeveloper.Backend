package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

type client struct {
	nickname string
	messages chan string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cfg := net.ListenConfig{
		KeepAlive: time.Minute,
	}

	listener, err := cfg.Listen(ctx, "tcp", "localhost:8001")
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	log.Println("chat server started!")

	go broadcaster(ctx)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		wg.Add(1)
		go handleConn(ctx, conn, wg)
	}

	<-ctx.Done()

	log.Println("chat server stoping")
	listener.Close()
	wg.Wait()
	log.Println("chat server stoped")
}

func handleConn(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	nickname := ""
	fmt.Fprintln(conn, "hello, please write your nickname:")

	input := bufio.NewScanner(conn)

	for input.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			nickname = input.Text()
			break
		}
	}

	cli := client{
		nickname: nickname,
		messages: make(chan string),
	}

	go clientWriter(ctx, conn, cli)

	cli.messages <- "You are " + cli.nickname
	messages <- cli.nickname + " has arrived"
	entering <- cli

	log.Println(cli.nickname + " has arrived")

	for input.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			messages <- cli.nickname + ": " + input.Text()
		}
	}
	leaving <- cli
	messages <- cli.nickname + " has left"
}

func clientWriter(ctx context.Context, conn net.Conn, cli client) {
	for msg := range cli.messages {
		fmt.Fprintln(conn, msg)
	}
}

func broadcaster(ctx context.Context) {
	clients := make(map[client]bool)
	for {
		select {
		case <-ctx.Done():
			for cli := range clients {
				delete(clients, cli)
				close(cli.messages)
			}
			return
		case msg := <-messages:
			for cli := range clients {
				cli.messages <- msg
			}
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli.messages)
		}
	}
}
