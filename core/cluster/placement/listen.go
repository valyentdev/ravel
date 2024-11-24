package placement

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

func getPlacementSubject(region string) string {
	return fmt.Sprintf("placement.%s", region)
}

type Listener struct {
	nats         *nats.Conn
	subscription *nats.Subscription
}

func NewListener(nats *nats.Conn) *Listener {
	return &Listener{
		nats: nats,
	}
}

type AnswerFunc func(*PlacementResponse) error

func (l Listener) Stop() {
	if l.subscription != nil {
		err := l.subscription.Unsubscribe()
		if err != nil {
			slog.Warn("failed to unsubscribe from placement requests", "error", err)
		}
	}
}

func (l *Listener) HandleMachinePlacementRequest(ctx context.Context, region string, handler func(msg *PlacementRequest) *PlacementResponse) error {
	sub, err := l.nats.SubscribeSync(getPlacementSubject(region))
	if err != nil {
		return err
	}

	l.subscription = sub

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := sub.NextMsgWithContext(ctx)
				if err != nil {
					return
				}
				var request PlacementRequest
				err = json.Unmarshal(msg.Data, &request)
				if err != nil {
					slog.Warn("failed to unmarshal placement request", "error", err)
					continue
				}

				response := handler(&request)
				if response == nil {
					continue
				}

				bytes, _ := json.Marshal(response)

				err = l.nats.Publish(msg.Reply, bytes)
				if err != nil {
					slog.Warn("failed to publish placement response", "error", err)
				}
			}
		}
	}()

	return nil
}
