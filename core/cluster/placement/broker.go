package placement

import (
	"encoding/json"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/nats-io/nats.go"
)

var ErrPlacementFailed = errors.New("placement failed")

type Broker struct {
	nc *nats.Conn
}

func NewBroker(nc *nats.Conn) *Broker {
	return &Broker{
		nc: nc,
	}
}

func (b *Broker) GetAvailableWorkers(req PlacementRequest) ([]PlacementResponse, error) {
	inbox := nats.NewInbox()

	bytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	offers := []PlacementResponse{}

	sub, err := b.nc.Subscribe(inbox, func(msg *nats.Msg) {
		var response PlacementResponse
		err := json.Unmarshal(msg.Data, &response)
		if err != nil {
			return
		}

		offers = append(offers, response)
	})

	if err != nil {
		return nil, err
	}
	err = b.nc.PublishRequest(getPlacementSubject(req.Region), inbox, bytes)
	if err != nil {
		return nil, err
	}

	time.Sleep(1 * time.Second)
	err = sub.Unsubscribe()
	if err != nil {
		slog.Error("failed to unsubscribe", "error", err)
		err = nil // Try to continue anyway
	}

	if len(offers) == 0 {
		return nil, ErrPlacementFailed
	}

	return sortCandidates(offers), nil
}

func sortCandidates(candidates []PlacementResponse) []PlacementResponse {
	var sorted []PlacementResponse
	sorted = append(sorted, candidates...)

	slices.SortStableFunc(sorted, func(a PlacementResponse, b PlacementResponse) int {
		aScore := a.GetScore()
		bScore := b.GetScore()

		if aScore > bScore {
			return -1
		}

		return 1
	})
	return sorted
}
