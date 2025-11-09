package models

import (
	"log/slog"
	"strconv"
)

type AggTrade struct {
	EventType        string `json:"e"` // "aggTrade"
	EventTime        int64  `json:"E"` // Время когда сервер отправил
	Symbol           string `json:"s"` // Торговая пара
	AggregateTradeID int64  `json:"a"` // Уникальный ID
	Price            string `json:"p"` // Цена сделки
	Quantity         string `json:"q"` // Объем сделки
	FirstTradeID     int64  `json:"f"` // ID первой микросделки
	LastTradeID      int64  `json:"l"` // ID последней микросделки
	TradeTime        int64  `json:"T"` // Время самой сделки
	IsBuyer          bool   `json:"m"` // Направление
	Ignore           bool   `json:"M"` // Игнорировать (всегда true)
}

func (at *AggTrade) PriceFloat() float64 {
	pf, err := strconv.ParseFloat(at.Price, 64)
	if err != nil {
		slog.Error("Could not parse Price into float64", "error", err)
		return -1
	}
	return pf
}
