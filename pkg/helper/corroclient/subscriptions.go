package corroclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type event struct {
	EOQ     *EOQ    `json:"eoq"`
	Columns Columns `json:"columns"`
	Row     []any   `json:"row"`
	Change  []any   `json:"change"`
}

func (c *CorroClient) request(req *http.Request) (*http.Response, error) {
	if c.bearer != "" {
		req.Header.Set("Authorization", c.bearer)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.c.Do(req)
}

func (c *CorroClient) subscribe(ctx context.Context, body io.ReadCloser) (<-chan Event, error) {
	reader := bufio.NewReader(body)
	eventChan := make(chan Event)

	go func() {
		defer body.Close()
		columns := []string{}
		for {
			eventData, _, err := reader.ReadLine()
			if err != nil {
				eventChan <- &Error{Message: err.Error()}
				break
			}

			var e event

			err = json.Unmarshal(eventData, &e)
			if err != nil {
				eventChan <- &Error{Message: err.Error()}
				break
			}

			if e.Columns != nil {
				columns = e.Columns
				continue
			}
			select {
			case <-ctx.Done():
				body.Close()
				close(eventChan)
				return
			default:
				eventChan <- readEvent(e, columns)
			}

		}
	}()

	return eventChan, nil
}

func (c *CorroClient) PostSubscription(ctx context.Context, statement Statement) (*Subscription, error) {
	data, err := json.Marshal(statement)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.getURL("/v1/subscriptions"), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := c.request(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("corrosubs: Invalid status code")
	}

	subscriptionId := resp.Header.Get("Corro-Query-Id")

	subCtx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	eventChan, err := c.subscribe(subCtx, resp.Body)
	if err != nil {
		return nil, err
	}

	return &Subscription{
		id:     subscriptionId,
		ctx:    subCtx,
		cancel: cancel,
		events: eventChan,
	}, nil
}

func (c *CorroClient) GetSubscription(ctx context.Context, subscriptionId string) (*Subscription, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.getURL("/v1/subscriptions/"+subscriptionId), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("corrosubs: Invalid status code")
	}

	subCtx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()

	eventChan, err := c.subscribe(subCtx, resp.Body)
	if err != nil {
		return nil, err
	}

	return &Subscription{
		id:     subscriptionId,
		ctx:    subCtx,
		cancel: cancel,
		events: eventChan,
	}, nil
}

func readEvent(event event, columns []string) Event {
	if event.EOQ != nil {
		return event.EOQ
	}
	if event.Columns != nil {
		return event.Columns
	}
	if event.Row != nil {
		row, err := readRow(event.Row)
		if err != nil {
			return &Error{Message: err.Error()}
		}

		row.columns = columns
		return row
	}

	if event.Change != nil {
		change, err := readChange(event.Change)
		if err != nil {
			return &Error{Message: err.Error()}
		}

		return change
	}
	return &Error{Message: "Unknown event type"}
}

type Subscription struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc
	events <-chan Event
}

func (s *Subscription) Close() {
	s.cancel()
}

func (s *Subscription) Events() <-chan Event {
	return s.events
}

func (s *Subscription) Id() string {
	return s.id
}
