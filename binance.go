package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"os"
)

type Binance struct {
	Client binance.Client
}
//TODO create struct

func (exchange *Binance) init() error {
	apiKey := os.Getenv("BINANCE_KEY")
	if apiKey == "" {
		return errors.New("failed to load BINANCE_KEY from .env")
	}
	apiSecret := os.Getenv("BINANCE_SECRET")
	if apiSecret == "" {
		return errors.New("failed to load BINANCE_SECRET from .env")
	}
	client := *binance.NewClient(apiKey, apiSecret)
	exchange.Client = client
	return nil
}

// Create a Marketorder: marketOrder("BTCUSD", "BUY", "5")
// Main usage point for algorithm execution
func (exchange *Binance) marketOrder(symbol string, order string, quantity string) (*binance.CreateOrderResponse, error) {
	side, err := setSideType(order)
	if err != nil {
		return nil, err
	}
	result, err := exchange.Client.NewCreateOrderService().
		Symbol(symbol).
		Side(side).
		Type(binance.OrderTypeMarket).
		Quantity(quantity).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func setSideType(side string) (binance.SideType, error) {
	var sideType binance.SideType
	if side == "BUY" {
		sideType = binance.SideTypeBuy
	} else if side == "SELL" {
		sideType = binance.SideTypeSell
	} else {
		return sideType, fmt.Errorf("received invalid order type%v", side)
	}
	return sideType, nil
}

func (exchange *Binance) orderStatus(orderId int64, symbol string) (*binance.Order, error) {
	order, err := exchange.Client.NewGetOrderService().Symbol(symbol).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (exchange *Binance) cancelOrder(orderId int64, symbol string) (*binance.CancelOrderResponse, error) {
	order, err := exchange.Client.NewCancelOrderService().Symbol(symbol).
		OrderID(orderId).Do(context.Background())
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (exchange *Binance) openOrders(symbol string) ([]*binance.Order, error) {
	order, err := exchange.Client.NewListOpenOrdersService().Symbol(symbol).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (exchange *Binance) tickerPrices(symbol string) ([]*binance.SymbolPrice, error) {
	order, err := exchange.Client.NewListPricesService().Symbol(symbol).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	return order, nil
}


func (exchange *Binance) candles(symbol string, interval string) ([]*binance.Kline, error) {
	candles, err := exchange.Client.NewKlinesService().Symbol(symbol).
		Interval(interval).Do(context.Background())
	if err != nil {

		return nil, err
	}
	return candles, nil
}

func (exchange *Binance) accountBalance() ([]binance.Balance, error) {
	res, err := exchange.Client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return nil, err
	}
	return res.Balances, nil
}


func (exchange *Binance) symbolDepth(symbol string) (*binance.DepthResponse, error) {
	order, err := exchange.Client.NewDepthService().Symbol(symbol).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	return order, nil
}