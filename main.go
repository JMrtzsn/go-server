package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {
	listenAddr := setAddress()
	logger := initLogger()

	ctrl := &controller{
		logger:        logger,
		nextRequestID: func() string { return strconv.FormatInt(time.Now().UnixNano(), 36) },
	}

	router := http.NewServeMux()
	router.HandleFunc("/", ctrl.index)
	router.HandleFunc("/health", ctrl.health)

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      (middlewares{ctrl.tracing, ctrl.logging}).apply(router),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	ctx := ctrl.shutdown(context.Background(), server)

	logger.Printf("Server is ready to handle requests at %q\n", listenAddr)
	atomic.StoreInt64(&ctrl.healthy, time.Now().UnixNano())

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %q: %s\n", listenAddr, err)
	}
	<-ctx.Done()
	logger.Printf("Server stopped\n")
}

func setAddress() string {
	if len(os.Args) == 2 {
		return os.Args[1]
	}
	return ":8080"
}

func initLogger() *log.Logger {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Server is starting...")
	return logger
}


