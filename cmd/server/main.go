// the main server module provides server (metric storage and update) function
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"internal/app"
	"internal/domain"
	"internal/ports/storage"

	"internal/adapters/cryptor"
	"internal/adapters/firewall"
	"internal/adapters/logger"
	"internal/adapters/signer"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
)

// statically linked variables (YP iter20 requirement)
var buildVersion string
var buildDate string
var buildCommit string

func handleGZIPRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(rw, r)
			return
		}

		logger.Info("srv-gzip: handling gzipped request")

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			errText := fmt.Sprintf("error reading compressed msg body: %s", err.Error())
			logger.Error("handleGZIPRequests: " + errText)
			http.Error(rw, errText, http.StatusInternalServerError)
			return
		}

		defer gz.Close()
		body, err := io.ReadAll(gz)
		if err != nil {
			errText := fmt.Sprintf("error extracting msg body: %s", err.Error())
			logger.Error("handleGZIPRequests: " + errText)
			http.Error(rw, errText, http.StatusInternalServerError)
			return
		}

		r.Body.Close()
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
	app.Sc = app.InitServerConfig()

	app.Stor = storage.NewUniStorage(&app.Sc)
	defer app.Stor.Close()

	//post-init unistorage actions
	err := app.Stor.Bootstrap()
	if err != nil {
		logger.Fatalf("srv: post-init bootstrap failed, error: %s", err)
	}

	// run `server` in its own goroutine
	wg.Add(1)
	go server(ctx, &wg)

	// if err := run(); err != nil {
	// 	//logger.Error("Server error", zap.Error(err))
	// 	log.Fatal(err)
	// }

	// listen for ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	logger.Info("srv: received ^C - shutting down")

	// tell the goroutines to stop
	logger.Info("srv: telling goroutines to stop")
	cancel()

	// and wait for them to reply back
	wg.Wait()

	logger.Info("srv: shutdown")
}

func server(ctx context.Context, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	// statically linked variables (YP iter20 requirement)
	logger.Infof("Build version: %s", naIfEmpty(buildVersion))
	logger.Infof("Build date: %s", naIfEmpty(buildDate))
	logger.Infof("Build commit: %s", naIfEmpty(buildCommit))

	logger.Infof("srv: transport mode %s", app.Sc.TransportMode)
	logger.Infof("srv: using endpoint %s", app.Sc.Endpoint)
	logger.Infof("srv: storage mode = %v", app.Sc.StorageMode)
	logger.Infof("srv: compress replies = %v %v", app.Sc.CompressReplies, app.Sc.CompressibleContentTypes)
	logger.Infof("srv: signed messaging = %v", signer.IsSignedMessagingEnabled())
	logger.Infof("srv: encryption=%v", cryptor.CanDecrypt())
	logger.Infof("srv: firewall=%v", firewall.IsFirewallEnabled())

	switch app.Sc.StorageMode {
	case domain.Database:
		//remove password from log output
		// //old mode
		// var safeDSN = strings.Split(sc.DatabaseDSN, " ")
		// for i, v := range safeDSN {
		// 	if strings.Contains(v, "password=") {
		// 		safeDSN[i] = "password=***"
		// 	}
		// }
		// logger.Infof("srv: DSN %s", strings.Join(safeDSN, " "))

		//nu mode
		re := regexp.MustCompile(`(password)=(?P<password>\S*)`)
		s := re.ReplaceAllLiteralString(app.Sc.DatabaseDSN, "password=***")
		logger.Infof("srv: DSN %s", s)

	case domain.File:
		logger.Infof("srv: datafile %s", app.Sc.FileStoragePath)
	}

	//read server state on start
	if app.Sc.StorageMode == domain.File && app.Sc.RestoreMetrics {
		err := app.Stor.LoadState(app.Sc.FileStoragePath)
		if err != nil {
			logger.Errorf("srv: failed to load server state from [%s], error: %s", app.Sc.FileStoragePath, err.Error())
		}
	}

	//regular dumper
	wg.Add(1)
	go stateDumper(ctx, app.Sc, wg)

	var srv *http.Server
	var grpcSrv *grpc.Server

	switch app.Sc.TransportMode {
	case domain.TRANSPORT_HTTP:
		srv = serveHTTP()
	case domain.TRANSPORT_GRPC:
		grpcSrv = serveGRPC()
	}

	<-ctx.Done()
	logger.Info("srv: shutdown requested")

	// shut down gracefully with timeout of 5 seconds max
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// graceful server shutdown
	switch app.Sc.TransportMode {
	case domain.TRANSPORT_HTTP:
		srv.Shutdown(shutdownCtx) // ignore server error "Err shutting down server : context canceled"
	case domain.TRANSPORT_GRPC:
		grpcSrv.GracefulStop()
	}

	//save server state on shutdown
	if app.Sc.StorageMode == domain.File {
		err := app.Stor.SaveState(app.Sc.FileStoragePath)
		if err != nil {
			logger.Errorf("srv: failed to save server state to [%s], error: %s", app.Sc.FileStoragePath, err)
		}
	}

	logger.Info("srv: server stopped")
}

// http server part
func serveHTTP() *http.Server {
	mux := chi.NewRouter()
	mux.Use(logger.LoggerMiddleware)
	mux.Use(firewall.HandleTrustedNetworkRequests)
	mux.Use(signer.HandleSignedRequests)
	mux.Use(handleGZIPRequests)
	mux.Use(cryptor.HandleEncryptedRequests)
	if app.Sc.CompressReplies {
		mux.Use(middleware.Compress(5, app.Sc.CompressibleContentTypes...))
	}

	mux.Get("/", handleIndex)
	mux.Get("/ping", handlePingDBServer)
	mux.Post("/value/", handleRequestMetricV2)
	mux.Get("/value/{type}/{name}", handleRequestMetricV1)
	mux.Post("/update/", handleUpdateMetricV2)
	mux.Post("/update/{type}/{name}/{value}", handleUpdateMetricV1)
	mux.Post("/updates/", handleUpdateMetrics)

	mux.Mount("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	// create a server
	srv := &http.Server{Addr: app.Sc.Endpoint, Handler: mux}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			logger.Errorf("Listen: %s", err.Error())
		}
	}()

	return srv
}

func stateDumper(ctx context.Context, sc domain.ServerConfig, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	//save dump is disabled or set to immediate mode
	if (sc.StorageMode != domain.File) || (sc.StoreInterval == 0) {
		return
	}

	ticker := time.NewTicker(time.Duration(sc.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			logger.Infof("TRACE: dump state [%s]\n", now.Format("2006-01-02 15:04:05"))

			err := app.Stor.SaveState(sc.FileStoragePath)
			if err != nil {
				logger.Errorf("srv-dumper: failed to save server state to [%s], error: %s", sc.FileStoragePath, err)
			}
		case <-ctx.Done():
			logger.Info("srv-dumper: stop requested")
			return
		}
	}

}

func naIfEmpty(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
