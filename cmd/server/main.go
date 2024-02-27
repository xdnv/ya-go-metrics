package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"internal/app"
	"internal/ports"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var storage = ports.NewMemStorage()

func main() {
	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// a WaitGroup for the goroutines to tell us they've stopped
	wg := sync.WaitGroup{}

	// run `server` in it's own goroutine
	wg.Add(1)
	go server(ctx, &wg)

	// if err := run(); err != nil {
	// 	//logger.Error("Server error", zap.Error(err))
	// 	log.Fatal(err)
	// }

	// listen for ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("srv: received ^C - shutting down")

	// tell the goroutines to stop
	fmt.Println("srv: telling goroutines to stop")
	cancel()

	// and wait for them to reply back
	wg.Wait()
	fmt.Println("srv: shutdown")
}

func server(ctx context.Context, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	sc := app.InitServerConfig()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("srv: cannot initialize zap logger")
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	//sugar.Infof("Failed to fetch URL: %s", url)
	//sugar.Errorf("Failed to fetch URL: %s", url)

	//fmt.Printf("using endpoint: %s\n", sc.Endpoint)
	sugar.Infof("srv: using endpoint %s", sc.Endpoint)
	sugar.Infof("srv: datafile %s", sc.FileStoragePath)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Compress(5))

	mux.Get("/", index)
	mux.Post("/value/", requestMetricV2)
	mux.Get("/value/{type}/{name}", requestMetricV1)
	mux.Post("/update/", updateMetricV2)
	mux.Post("/update/{type}/{name}/{value}", updateMetricV1)

	// create a server
	srv := &http.Server{Addr: sc.Endpoint, Handler: mux}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("Listen: %s\n", err)
			//log.Fatal(err)
		}
	}()

	<-ctx.Done()
	fmt.Println("srv: shutdown requested")

	// shut down gracefully with timeout of 5 seconds max
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ignore server error "Err shutting down server : context canceled"
	srv.Shutdown(shutdownCtx)

	fmt.Println("srv: server stopped")
}
