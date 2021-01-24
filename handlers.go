package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

type middleware func(http.Handler) http.Handler
type middlewares []middleware

// recursive function that adds the Middleware flow to the different endpoints?
func (mws middlewares) apply(hdlr http.Handler) http.Handler {
	if len(mws) == 0 {
		return hdlr
	}
	return mws[1:].apply(mws[0](hdlr))
}

// Ensure the handlers implement the required interfaces/types at compile time
var (
	_ http.Handler = http.HandlerFunc((&controller{}).index)
	_ http.Handler = http.HandlerFunc((&controller{}).health)
	_ middleware   = (&controller{}).logging
	_ middleware   = (&controller{}).tracing
)

type controller struct {
	logger        *log.Logger
	nextRequestID func() string
	healthy       int64
	client        binance.Client
}

func (c *controller) shutdown(ctx context.Context, server *http.Server) context.Context {
	ctx, done := context.WithCancel(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer done() //

		<-quit
		signal.Stop(quit)
		close(quit)

		atomic.StoreInt64(&c.healthy, 0)
		server.ErrorLog.Printf("Server is shutting down...\n")

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			server.ErrorLog.Fatalf("Could not gracefully shutdown the server: %s\n", err)
		}
	}()

	return ctx
}

func (c *controller) index(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
}

func (c *controller) health(w http.ResponseWriter, req *http.Request) {
	if h := atomic.LoadInt64(&c.healthy); h == 0 {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		fmt.Fprintf(w, "uptime: %s\n", time.Since(time.Unix(0, h)))
	}
}

func (c *controller) logging(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func(start time.Time) {
			requestID := w.Header().Get("X-Request-Id")
			if requestID == "" {
				requestID = "unknown"
			}
			c.logger.Println(requestID, req.Method, req.URL.Path, req.RemoteAddr, req.UserAgent(), time.Since(start))
		}(time.Now())
		hdlr.ServeHTTP(w, req)
	})
}

func (c *controller) tracing(hdlr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestID := req.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = c.nextRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)
		hdlr.ServeHTTP(w, req)
	})
}

// TODO: move these to exchange, create interface with these functions?
// TODO: standardize the json convertion?

func (c *controller) marketOrder(w http.ResponseWriter, req *http.Request) {
	if !POST(w, req){
		return
	}

	symbol := req.Form.Get("symbol")
	order := req.Form.Get("order")
	quantity := req.Form.Get("quantity")
	if isEmpty(symbol, order, quantity) {
		http.Error(w, "Invalid Input, symbol, order, quantity must be provided", http.StatusBadRequest)
		return
	}

	marketOrder, err := c.client.marketOrder(symbol, order, quantity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeToJson(w, marketOrder)
}

func (c *controller) orderStatus(w http.ResponseWriter, req *http.Request) {
	if !POST(w, req){
		return
	}

	symbol := req.Form.Get("symbol")
	orderInput := req.Form.Get("orderId")
	if isEmpty(symbol, orderInput) {
		http.Error(w, "Invalid Input, symbol, order, quantity must be provided", http.StatusBadRequest)
	}

	orderId, err := strconv.Atoi(orderInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status, err := c.client.orderStatus(int64(orderId), symbol)

	writeToJson(w, status)

}


func (c *controller) cancelOrder(w http.ResponseWriter, req *http.Request) {
	if !POST(w, req){
		return
	}
	writeToJson(w, order)
}

func (c *controller) tickerPrices(w http.ResponseWriter, req *http.Request) {
	if !POST(w, req){
		return
	}

	writeToJson(w, prices)
}

func (c *controller) symbolCandles(w http.ResponseWriter, req *http.Request) {
	if !POST(w, req){
		return
	}
	writeToJson(w, candles)
}

func (c *controller) symbolDepth(w http.ResponseWriter, req *http.Request) {
	if !POST(w, req){
		return
	}
	writeToJson(w, depth)a
}


func (c *controller) accountBalance(w http.ResponseWriter, req *http.Request) {
	balance, err  := c.client.accountBalance()
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	writeToJson(w, balance)
}

func isEmpty(ss ...string) bool {
	for _, s := range ss {
		if s == "" {
			return true
		}
	}
	return false
}


func POST(w http.ResponseWriter, req *http.Request) bool {
	if req.Method != "POST" {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return false
	}
	if err := req.ParseForm();err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	return true
}

func writeToJson(w http.ResponseWriter, thing interface{}) {
	response, err := json.Marshal(thing)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
