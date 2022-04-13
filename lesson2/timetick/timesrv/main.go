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

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	cfg := net.ListenConfig{
		KeepAlive: time.Minute,
	}

	l, err := cfg.Listen(ctx, "tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	log.Println("im started!")

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println(err)
				return
			} else {
				wg.Add(1)
				go handleConn(ctx, conn, wg)
			}
		}
	}()

	<-ctx.Done()

	log.Println("done")
	l.Close()
	wg.Wait()
	log.Println("exit")
}

func handleConn(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	localWg := &sync.WaitGroup{}
	localWg.Add(1)
	go serverWriter(ctx, conn, localWg)

	// каждую 1 секунду отправлять клиентам текущее время сервера
	tck := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			localWg.Wait()
			return
		case t := <-tck.C:
			fmt.Fprintf(conn, "now: %s\n", t)
		}
	}
}

func serverWriter(ctx context.Context, conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	messages := make(chan string)

	wg.Add(1)
	go handleServerWriter(ctx, conn, messages, wg)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			for scanner.Scan() {
				messages <- scanner.Text()
			}
		}
	}
}

func handleServerWriter(ctx context.Context, conn net.Conn, messages <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case m := <-messages:
			fmt.Fprintf(conn, "message from server: %s\n", m)
		}
	}
}
