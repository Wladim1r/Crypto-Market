package reddis

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/Wladim1r/profile/connmanager"
	"github.com/Wladim1r/profile/internal/models"
	"github.com/redis/go-redis/v9"
)

type rDB struct {
	rdb *redis.Client
	cm  *connmanager.ConnectionManager
}

func NewClient(cm *connmanager.ConnectionManager) *rDB {
	return &rDB{
		rdb: redis.NewClient(&redis.Options{
			Addr:     "redis:6379",
			Password: "",
			DB:       0,
		}),
		cm: cm,
	}
}

func (rdb *rDB) Start(wg *sync.WaitGroup, ctx context.Context, inChan chan models.SecondStat) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			slog.Info("got interruption signal")
			return
		case msg := <-inChan:
			slog.Info("GOT message from redis stream", "msg", msg)

			rdb.cm.WriteToUser(int(msg.UserID), msg)
		}
	}
}

func (r *rDB) Subscribe(wg *sync.WaitGroup, ctx context.Context, outChan chan models.SecondStat) {
	defer wg.Done()

	subscriber := r.rdb.Subscribe(ctx, "stream")

	secondStat := new(models.SecondStat)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := subscriber.ReceiveMessage(ctx)
			if err != nil {
				slog.Error("Could not receive message from redis stream channel", "error", err)
				continue
			}

			if err := json.Unmarshal([]byte(msg.Payload), secondStat); err != nil {
				slog.Error("Failed to parsing bytes into SecondStat struct", "error", err)
				continue
			}
			select {
			case <-ctx.Done():
				return
			case outChan <- *secondStat:
				slog.Info("Msg from redis stream has just send to channel")
			}
		}
	}
}
