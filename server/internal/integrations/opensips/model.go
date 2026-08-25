package opensips

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	URL            string
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
}

func DefaultConfig(url string) Config {
	return Config{
		URL:            strings.TrimSpace(url),
		ConnectTimeout: 5 * time.Second,
		RequestTimeout: 5 * time.Second,
	}
}

type Command struct {
	Name   string            `json:"name"`
	Params map[string]string `json:"params,omitempty"`
}

func (c Command) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("OpenSIPS command name is required")
	}
	return nil
}

type Response struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Params  map[string]string `json:"params,omitempty"`
}

func (r Response) Validate() error {
	if r.Code < 0 {
		return fmt.Errorf("invalid OpenSIPS response code: %d", r.Code)
	}
	return nil
}

type Event struct {
	Name      string            `json:"name"`
	Timestamp time.Time         `json:"timestamp"`
	Params    map[string]string `json:"params,omitempty"`
}

func (e Event) Validate() error {
	if strings.TrimSpace(e.Name) == "" {
		return fmt.Errorf("OpenSIPS event name is required")
	}
	return nil
}
