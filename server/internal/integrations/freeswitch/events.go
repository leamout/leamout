package freeswitch

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"net/url"
	"strings"
	"time"
)

type EventHandler func(context.Context, Event) error

type subscription struct {
	format  EventFormat
	events  []string
	handler EventHandler
}

func (s subscription) command() string {
	command := "event " + string(s.format)
	if len(s.events) > 0 {
		command += " " + strings.Join(s.events, " ")
	}
	return command
}

func (s subscription) matches(eventName string) bool {
	if len(s.events) == 0 {
		return true
	}

	for _, event := range s.events {
		if event == eventName {
			return true
		}
	}
	return false
}

func (c *Client) Subscribe(
	ctx context.Context,
	format EventFormat,
	events []string,
	handler EventHandler,
) error {
	subscription, err := newSubscription(format, events, handler)
	if err != nil {
		return err
	}

	c.subscriptionsMu.Lock()
	defer c.subscriptionsMu.Unlock()

	if err := c.sendSubscription(ctx, subscription); err != nil {
		return err
	}
	c.subscriptions = append(c.subscriptions, subscription)

	return nil
}

func newSubscription(
	format EventFormat,
	events []string,
	handler EventHandler,
) (subscription, error) {
	if handler == nil {
		return subscription{}, fmt.Errorf("FreeSWITCH event handler is required")
	}
	if format != EventFormatPlain {
		return subscription{}, fmt.Errorf("unsupported FreeSWITCH event format: %q", format)
	}

	clean := make([]string, 0, len(events))
	seen := make(map[string]struct{}, len(events))
	for _, event := range events {
		event = strings.TrimSpace(event)
		if event == "" {
			continue
		}
		if _, ok := seen[event]; ok {
			continue
		}
		seen[event] = struct{}{}
		clean = append(clean, event)
	}

	return subscription{
		format:  format,
		events:  clean,
		handler: handler,
	}, nil
}

func (c *Client) sendSubscription(ctx context.Context, subscription subscription) error {
	frame, err := c.command(ctx, subscription.command())
	if err != nil {
		return err
	}
	if !frame.OK() {
		return fmt.Errorf("FreeSWITCH event subscription failed: %s", frame.ReplyText())
	}
	return nil
}

func (c *Client) restoreSubscriptions(ctx context.Context) error {
	for _, subscription := range c.subscriptionSnapshot() {
		if err := c.sendSubscription(ctx, subscription); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) subscriptionSnapshot() []subscription {
	c.subscriptionsMu.RLock()
	defer c.subscriptionsMu.RUnlock()

	return append([]subscription(nil), c.subscriptions...)
}

func plainEventHeader(body, name string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimRight(line, "\r")
		key, value, ok := strings.Cut(line, ": ")
		if !ok || key != name {
			continue
		}
		decoded, err := url.PathUnescape(value)
		if err == nil {
			return decoded
		}
		return value
	}
	return ""
}

// A fixed number of worker queues bounds the number of concurrently executing
// ESL callbacks. Each channel is assigned to exactly one FIFO queue, retaining
// the order of its events while allowing independent channels to progress.
const (
	eventDispatchShards = 16
	eventQueueDepth     = 128
	eventHandlerTimeout = 5 * time.Second
)

type queuedEvent struct {
	event    Event
	handlers []EventHandler
}

func (c *Client) startEventDispatch() {
	for index := range c.eventQueues {
		queue := make(chan queuedEvent, eventQueueDepth)
		c.eventQueues[index] = queue
		c.wg.Add(1)
		go c.consumeEventQueue(queue)
	}
}

func (c *Client) consumeEventQueue(queue <-chan queuedEvent) {
	defer c.wg.Done()

	for {
		select {
		case <-c.lifecycleCtx.Done():
			return
		case item := <-queue:
			for _, handler := range item.handlers {
				ctx, cancel := context.WithTimeout(c.lifecycleCtx, eventHandlerTimeout)
				err := handler(ctx, item.event)
				cancel()
				if err != nil {
					log.Printf("FreeSWITCH event handler failed: event=%s channel=%s error=%v",
						item.event.Name, item.event.Header("Unique-ID"), err)
				}
				if c.lifecycleCtx.Err() != nil {
					return
				}
			}
		}
	}
}

func (c *Client) dispatchEvent(frame Frame) error {
	event := Event{
		Headers: frame.Headers,
		Body:    frame.Body,
		Name:    frame.Header("Event-Name"),
	}

	subscriptions := c.subscriptionSnapshot()
	handlers := make([]EventHandler, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		if subscription.matches(event.Name) {
			handlers = append(handlers, subscription.handler)
		}
	}
	if len(handlers) == 0 {
		return nil
	}
	select {
	case <-c.lifecycleCtx.Done():
		return context.Canceled
	default:
	}

	channel := strings.TrimSpace(event.Header("Unique-ID"))
	key := channel
	if key == "" {
		key = strings.TrimSpace(event.Header("variable_leamout_call_id"))
	}
	if key == "" {
		key = "unkeyed"
	}

	queue := c.eventQueues[eventDispatchShard(key)]
	select {
	case queue <- queuedEvent{event: event, handlers: handlers}:
		return nil
	default:
		return fmt.Errorf("FreeSWITCH event queue full: event=%s channel=%s", event.Name, channel)
	}
}

func eventDispatchShard(key string) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))
	return int(hash.Sum32() % eventDispatchShards)
}
