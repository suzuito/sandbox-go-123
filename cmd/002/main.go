package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for i := 1; i <= 5; i++ {
			fmt.Println(i)
			time.Sleep(1 * time.Second)
		}
		fmt.Println("Hello, World!")
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	done := make(chan error, 1)
	go func() {
		done <- server.ListenAndServe()
	}()

	select {
	case err := <-done:
		if err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	case <-ctx.Done():
		fmt.Println("Server stopping")
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(c); err != nil {
			log.Fatalf("HTTP server Shutdown: %v", err)
		}
		fmt.Println("Server gracefully stopped")
	}
}
