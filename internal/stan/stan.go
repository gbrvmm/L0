package stan

import (
	"context"
	"encoding/json"
	"log"
	"time"

	stan "github.com/nats-io/stan.go"

	"github.com/gbrvmm/L0/internal/cache"
	"github.com/gbrvmm/L0/internal/config"
	"github.com/gbrvmm/L0/internal/db"
	"github.com/gbrvmm/L0/internal/model"
)

type Subscriber struct {
	sc  stan.Conn
	sub stan.Subscription
}

func Start(ctx context.Context, cfg config.Config, c *cache.Cache, repo *db.Repository) (*Subscriber, error) {
	sc, err := stan.Connect(cfg.StanClusterID, cfg.StanClientID, stan.NatsURL(cfg.StanURL))
	if err != nil {
		return nil, err
	}

	options := []stan.SubscriptionOption{
		stan.DurableName(cfg.Durable),
		stan.SetManualAckMode(),
		stan.AckWait(30 * time.Second),
		stan.MaxInflight(64),
		stan.DeliverAllAvailable(),
	}

	h := func(m *stan.Msg) {
		var o model.Order
		if err := json.Unmarshal(m.Data, &o); err != nil {
			log.Printf("[stan] bad json: %v", err)
			_ = repo.SaveBadMessage(ctx, m.Data, "invalid json")
			_ = m.Ack() // не зацикливаем плохие сообщения
			return
		}
		if err := o.Validate(); err != nil {
			log.Printf("[stan] validation error: %v", err)
			_ = repo.SaveBadMessage(ctx, m.Data, err.Error())
			_ = m.Ack()
			return
		}
		if err := repo.SaveOrder(ctx, o); err != nil {
			log.Printf("[stan] db save error for %s: %v", o.OrderUID, err)
			return // не ack — получим повтор и попробуем снова
		}
		c.Set(o.OrderUID, o)
		if err := m.Ack(); err != nil {
			log.Printf("[stan] ack error: %v", err)
		}
	}

	sub, err := sc.QueueSubscribe(cfg.Channel, "orders-group", h, options...)
	if err != nil {
		sc.Close()
		return nil, err
	}

	go func() {
		<-ctx.Done()
		if sub != nil {
			_ = sub.Close()
		}
		if sc != nil {
			sc.Close()
		}
	}()

	return &Subscriber{sc: sc, sub: sub}, nil
}
