package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	retry "github.com/hashicorp/go-retryablehttp"
)

var port string
var url string
var bodytype = "application/x-www-form-urlencoded"

// Pre test setup
func init(){
	port = os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	url = os.Getenv("SERVICE_URL")
	if url == "" {
		url = "http://localhost:" + port
	}
}

func TestService(t *testing.T) {
	resp, err := retry.Get(url + "/")
	if err != nil {
		t.Fatalf("retry.Get: %v", err)
	}

	if got := resp.StatusCode; got != http.StatusOK {
		t.Errorf("HTTP Response: got %q, want %q", got, http.StatusOK)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll: %v", err)
	}
}
func TestOpenOrders(t *testing.T) {
	resp, err := retry.Get(url + "/openOrders")
	if err != nil {
		t.Fatalf("retry.Get: %v", err)
	}

	if got := resp.StatusCode; got != http.StatusOK {
		t.Errorf("HTTP Response: got %q, want %q", got, http.StatusOK)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll: %v", err)
	}
}
func TestCandles(t *testing.T) {
	values := make(map[string][]string)
	values["symbol"] = []string{"BTCUSDT"}
	values["interval"] = []string{"1h"}

	resp, err := retry.PostForm(url + "/candles", values)
	if err != nil {
		t.Fatalf("retry.Get: %v", err)
	}

	if got := resp.StatusCode; got != http.StatusOK {
		t.Errorf("HTTP Response: got %v, want %v", got, http.StatusOK)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll: %v", err)
	}
	if data != nil {
		t.Fatalf("ioutil.ReadAll: %v", err)
	}
}


func TestAccountBalance(t *testing.T) {

	resp, err := retry.Get(url + "/accountBalance")
	if err != nil {
		t.Fatalf("retry.Get: %v", err)
	}

	if got := resp.StatusCode; got != http.StatusOK {
		t.Errorf("HTTP Response: got %v, want %v", got, http.StatusOK)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll: %v", err)
	}
	if data != nil {
		t.Fatalf("ioutil.ReadAll: %v", err)
	}
}