package rtpengine

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	Address        string
	ConnectTimeout time.Duration
	CommandTimeout time.Duration
}

func DefaultConfig(address string) Config {
	return Config{Address: strings.TrimSpace(address), ConnectTimeout: 5 * time.Second, CommandTimeout: 5 * time.Second}
}

func (c Config) Validate() error {
	if c.Address == "" { return fmt.Errorf("RTPEngine address is required") }
	if c.ConnectTimeout <= 0 { return fmt.Errorf("RTPEngine connect timeout must be positive") }
	if c.CommandTimeout <= 0 { return fmt.Errorf("RTPEngine command timeout must be positive") }
	return nil
}

type Command string

const (
	CommandOffer Command = "offer"
	CommandAnswer Command = "answer"
	CommandDelete Command = "delete"
	CommandQuery Command = "query"
)

type Session struct { CallID, FromTag, ToTag, Branch string; CreatedAt time.Time }

func (s Session) Validate() error {
	if strings.TrimSpace(s.CallID) == "" { return fmt.Errorf("RTPEngine call ID is required") }
	if strings.TrimSpace(s.FromTag) == "" { return fmt.Errorf("RTPEngine from tag is required") }
	if strings.TrimSpace(s.Branch) == "" { return fmt.Errorf("RTPEngine branch is required") }
	return nil
}

type OfferRequest struct { Session Session; SDP string; Flags []string }
type AnswerRequest struct { Session Session; SDP string; Flags []string }
type DeleteRequest struct { Session Session; Flags []string }

type OfferResponse struct { SDP string; Data map[string]any }
type AnswerResponse struct { SDP string; Data map[string]any }
type QueryResponse struct { Data map[string]any }

type Response struct { Result string; Error string; Data map[string]any }
func (r Response) OK() bool { return strings.EqualFold(r.Result, "ok") }
func (r Response) String(key string) string { value, _ := r.Data[key].(string); return value }
