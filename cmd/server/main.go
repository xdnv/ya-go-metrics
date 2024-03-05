package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"internal/app"
	"internal/ports/storage"

	"internal/adapters/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var stor = storage.NewMemStorage()
var sc app.ServerConfig

func handleGZIPRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(rw, r)
			return
		}

		logger.Info("srv-gzip: handling gzipped request")

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		defer gz.Close()
		body, err := io.ReadAll(gz)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))

		next.ServeHTTP(rw, r)
	})
}

func main() {
	//sync internal/logger upon exit
	defer logger.Sync()

	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// a WaitGroup for the goroutines to tell us they've stopped
	wg := sync.WaitGroup{}

	//Warning! do not run outside function, it will break tests due to flag.Parse()
	sc = app.InitServerConfig()

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

	// logger, err := zap.NewDevelopment()
	// if err != nil {
	// 	panic("srv: cannot initialize zap logger")
	// }
	// defer logger.Sync()

	// sugar := logger.Sugar()

	//sugar.Infof("Failed to fetch URL: %s", url)
	//sugar.Errorf("Failed to fetch URL: %s", url)

	//fmt.Printf("using endpoint: %s\n", sc.Endpoint)
	//sugar.Infof("srv: using endpoint %s", sc.Endpoint)
	//sugar.Infof("srv: datafile %s", sc.FileStoragePath)
	logger.Info(fmt.Sprintf("srv: using endpoint %s", sc.Endpoint))
	logger.Info(fmt.Sprintf("srv: datafile %s", sc.FileStoragePath))

	//read server state on start
	if (sc.FileStoragePath != "") && sc.RestoreMetrics {
		err := stor.LoadState(sc.FileStoragePath)
		if err != nil {
			fmt.Printf("srv: failed to load server state from [%s], error: %s\n", sc.FileStoragePath, err)
		}
	}

	//regular dumper
	wg.Add(1)
	go stateDumper(ctx, sc, wg)

	mux := chi.NewRouter()
	//mux.Use(middleware.Logger)
	mux.Use(logger.LoggerMiddleware)
	mux.Use(handleGZIPRequests)
	mux.Use(middleware.Compress(5, sc.CompressibleContentTypes...))

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
	//fmt.Println("srv: shutdown requested")
	logger.Info("srv: shutdown requested")

	// shut down gracefully with timeout of 5 seconds max
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ignore server error "Err shutting down server : context canceled"
	srv.Shutdown(shutdownCtx)

	//save server state on shutdown
	if sc.FileStoragePath != "" {
		err := stor.SaveState(sc.FileStoragePath)
		if err != nil {
			//fmt.Printf("srv: failed to save server state to [%s], error: %s\n", sc.FileStoragePath, err)
			logger.Info(fmt.Sprintf("srv: failed to save server state to [%s], error: %s\n", sc.FileStoragePath, err))
		}
	}

	fmt.Println("srv: server stopped")
}

func stateDumper(ctx context.Context, sc app.ServerConfig, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	//save dump is disabled or set to immediate mode
	if (sc.FileStoragePath == "") || (sc.StoreInterval == 0) {
		return
	}

	ticker := time.NewTicker(time.Duration(sc.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			fmt.Printf("TRACE: dump state [%s]\n", now.Format("2006-01-02 15:04:05"))

			err := stor.SaveState(sc.FileStoragePath)
			if err != nil {
				fmt.Printf("srv-dumper: failed to save server state to [%s], error: %s\n", sc.FileStoragePath, err)
			}
		case <-ctx.Done():
			fmt.Println("srv-dumper: stop requested")
			return
		}
	}

}
