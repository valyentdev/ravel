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
	nats *nats.Conn
}

func NewListener(nats *nats.Conn) *Listener {
	return &Listener{
		nats: nats,
	}
}

type AnswerFunc func(*MachinePlacementResponse) error

func (l *Listener) HandleMachinePlacementRequest(ctx context.Context, region string, handler func(msg *MachinePlacementRequest, answer AnswerFunc)) error {
	sub, err := l.nats.SubscribeSync(getPlacementSubject(region))
	if err != nil {
		return err
	}

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
				var request MachinePlacementRequest
				err = json.Unmarshal(msg.Data, &request)
				if err != nil {
					slog.Warn("failed to unmarshal placement request", "error", err)
					continue
				}

				handler(&request, func(mpr *MachinePlacementResponse) error {
					bytes, err := json.Marshal(mpr)
					if err != nil {
						return err
					}
					slog.Info("Sending placement response", "response", mpr)
					return l.nats.Publish(msg.Reply, bytes)
				})
			}
		}
	}()

	return nil
}
