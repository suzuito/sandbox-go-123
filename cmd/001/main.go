package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 参考 https://castaneai.hatenablog.com/entry/2020/04/28/210002
	// Graceful shutdown in Go

	// SIGINT,SIGTERMをキャッチする
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop() // おまじない(TODO なんのための記述なの？調べる)

	// Sample server
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/hoge", func(w http.ResponseWriter, r *http.Request) {
		for i := range 5 {
			fmt.Println(i)
			time.Sleep(time.Second)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "hello")
	})
	server := http.Server{
		Handler: serveMux,
		Addr:    ":3000",
	}

	chGracefulShutdown := make(chan error)
	defer close(chGracefulShutdown)
	go func() {
		// Signalのハンドラー
		// SIGINT,SIGTERMをキャッチした後、ctx.Doneが制御を返す
		<-ctx.Done()
		fmt.Println("start graceful shut down")
		ctxSignalHandler, cancel := context.WithTimeout(context.Background(), time.Second*100) // 100秒待ってもserver.Shutdown(ctx)が返ってこない場合、強制的にシャットダウンする
		defer cancel()
		// Graceful Shutdownをスタートする。
		// Graceful Shutdownが成功したら、server.Shutdown(ctxSignalHandler)はnilを返す。
		// Graceful Shutdownが失敗したら、server.Shutdown(ctxSignalHandler)は非nilを返す。
		chGracefulShutdown <- server.Shutdown(ctxSignalHandler)
	}()
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(http.ErrServerClosed, err) {
			fmt.Printf("server.ListenAndServe() is failed : %+v\n", err)
			os.Exit(1)
		}
	}
	if err := <-chGracefulShutdown; err != nil {
		fmt.Printf("graceful shut down is failed (server.Shutdown(ctx) is failed) : %+v\n", err)
		os.Exit(2)
	}
	fmt.Println("graceful shut down is complete")
}
