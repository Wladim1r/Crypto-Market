package connmanager

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Wladim1r/profile/internal/models"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

type client struct {
	Conn     *websocket.Conn
	Profile  *models.User
	Prices   map[string]decimal.Decimal
	SendChan chan []byte
}

type ConnectionManager struct {
	clients map[int]*client
	mu      sync.RWMutex
}

func NewConnManager() ConnectionManager {
	return ConnectionManager{clients: make(map[int]*client)}
}

func (c *client) writer() {
	defer c.Conn.Close()

	for msg := range c.SendChan {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			slog.Error("could not write message to websocket connection", "error", err.Error())
			continue
		}
	}
}

func (cm *ConnectionManager) Register(userID int, profile *models.User, conn *websocket.Conn) {
	cm.mu.Lock()
	sendChan := make(chan []byte, 100)
	client := &client{
		Conn:     conn,
		Profile:  profile,
		Prices:   make(map[string]decimal.Decimal),
		SendChan: sendChan,
	}

	cm.clients[userID] = client
	cm.mu.Unlock()

	go client.writer()
}

func (cm *ConnectionManager) WriteToUser(userID int, msg models.SecondStat) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, ok := cm.clients[userID]
	if !ok {
		return fmt.Errorf("Connection for user %d does not exist", userID)
	}
	client.Prices[msg.Symbol] = decimal.NewFromFloat32(msg.Price)

	profile := models.Profile{
		ID:   uint(userID),
		Name: client.Profile.Name,
		Coins: models.CoinsProfile{
			Quantities: make(map[string]decimal.Decimal),
			Prices:     make(map[string]decimal.Decimal),
			Totals:     make(map[string]decimal.Decimal),
		},
	}

	for _, coin := range client.Profile.Coins {
		if price, ok := client.Prices[coin.Symbol]; ok {
			profile.Coins.Quantities[coin.Symbol] = coin.Quantity
			profile.Coins.Prices[coin.Symbol] = price
			profile.Coins.Totals[coin.Symbol] = price.Mul(coin.Quantity)
		}
	}

	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("could not parse 'profile' into JSON: %w", err)
	}

	client.SendChan <- profileJSON

	return nil
}

func (cm *ConnectionManager) Unregister(userID int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, ok := cm.clients[userID]
	if ok {
		delete(cm.clients, userID)
		client.Conn.Close()
	}
}
