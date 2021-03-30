package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
	"github.com/joho/godotenv"
	"cloud.google.com/go/compute/metadata"
)

func main() {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Server is starting...")
	// Initialize template parameters.
	service := os.Getenv("K_SERVICE")
	if service == "" {
		service = "???"
	}

	revision := os.Getenv("K_REVISION")
	if revision == "" {
		revision = "???"
	}

	project := os.Getenv("GOOGLE_CLOUD_PROJECT")

	// Environment variable GOOGLE_CLOUD_PROJECT is only set locally.
	// On Cloud Run, strip the timestamp prefix from log entries.
	if project == "" {
		log.SetFlags(0)
	}

	// Only attempt to check the Cloud Run metadata server if it looks like
	// the service is deployed to Cloud Run or GOOGLE_CLOUD_PROJECT not already set.
	if project == "" || service != "???" {
		var err error
		if project, err = metadata.ProjectID(); err != nil {
			logger.Printf("metadata.ProjectID: Cloud Run metadata server: %v", err)
		}
	}
	if project == "" {
		project = "???"
	}

	// PORT environment variable is provided by Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	loadEnv()

	binance := Binance{}
	if err := binance.init(); err != nil{
		logger.Fatalf(err.Error())
	}

	ctrl := &controller{
		logger:        logger,
		nextRequestID: func() string { return strconv.FormatInt(time.Now().UnixNano(), 36) },
		client:        binance,
	}

	router := http.NewServeMux()
	router.HandleFunc("/", ctrl.index)
	router.HandleFunc("/health", ctrl.health)
	router.HandleFunc("/marketOrder", ctrl.marketOrder)
	router.HandleFunc("/cancelOrder", ctrl.cancelOrder)
	router.HandleFunc("/orderStatus", ctrl.orderStatus)
	router.HandleFunc("/candles", ctrl.candles)
	router.HandleFunc("/tickerPrices", ctrl.tickerPrices)
	router.HandleFunc("/accountBalance", ctrl.accountBalance)
	router.HandleFunc("/symbolDepth", ctrl.symbolDepth)

	server := &http.Server{
		Addr:         ":"+port,
		Handler:      (middlewares{ctrl.tracing, ctrl.logging}).apply(router),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	ctx := ctrl.shutdown(context.Background(), server)

	logger.Printf("Server is ready to handle requests at %q\n", port)
	atomic.StoreInt64(&ctrl.healthy, time.Now().UnixNano())

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %q: %s\n", port, err)
	}
	<-ctx.Done()
	logger.Printf("Server stopped\n")
}

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

