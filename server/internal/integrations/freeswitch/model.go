package freeswitch

import (
	"fmt"
	"strings"
	"time"
)

const (
	ContentTypeAuthRequest  = "auth/request"
	ContentTypeCommandReply = "command/reply"
	ContentTypeAPIResponse  = "api/response"
	ContentTypeEventPlain   = "text/event-plain"
)

type Config struct {
	Address        string
	Password       string
	ConnectTimeout time.Duration
	CommandTimeout time.Duration
}

func DefaultConfig(address, password string) Config {
	return Config{
		Address:        strings.TrimSpace(address),
		Password:       password,
		ConnectTimeout: 5 * time.Second,
		CommandTimeout: 5 * time.Second,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		return fmt.Errorf("FreeSWITCH address is required")
	}
	if c.Password == "" {
		return fmt.Errorf("FreeSWITCH password is required")
	}
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("FreeSWITCH connect timeout must be positive")
	}
	if c.CommandTimeout <= 0 {
		return fmt.Errorf("FreeSWITCH command timeout must be positive")
	}
	return nil
}

type Frame struct {
	ContentType string
	Headers     map[string]string
	Body        string
}

func (f Frame) Header(name string) string {
	return f.Headers[name]
}

func (f Frame) ReplyText() string {
	return f.Header("Reply-Text")
}

func (f Frame) OK() bool {
	return strings.HasPrefix(strings.ToUpper(f.ReplyText()), "+OK")
}

type Reply struct {
	Text string
	Body string
}

type Job struct {
	ID string
}

type Event struct {
	Name    string
	Headers map[string]string
	Body    string
}

func (e Event) Header(name string) string {
	return e.Headers[name]
}

type EventFormat string

const (
	EventFormatPlain EventFormat = "plain"
	EventFormatJSON  EventFormat = "json"
)

type OriginateRequest struct {
	Endpoint    string
	Destination string
	CallerID    string
	Variables   map[string]string
}

type TransferRequest struct {
	CallID      string
	Destination string
	Dialplan    string
	Context     string
}

type RecordRequest struct {
	CallID string
	Path   string
	Action string
}

type Channel struct {
	UUID  string
	Name  string
	State string
}

type Call struct {
	UUID         string
	CallerName   string
	CallerNumber string
	Destination  string
	State        string
}

type Endpoint struct {
	Name string
	Type string
	Data string
}

type SIPProfileStatus struct {
	Profile string
	Raw     string
}

type ConferenceRequest struct {
	Name      string
	Command   string
	Arguments []string
}

type ConferenceResult struct {
	Text string
	Body string
}

type ConferenceMember struct {
	ID       string
	CallerID string
	Muted    bool
	Deaf     bool
}

type ConferenceMembers struct {
	Conference string
	Members    []ConferenceMember
}
