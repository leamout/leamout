package opensips

import "time"

type Config struct {
	URL            string
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
}

func DefaultConfig(url string) Config {
	return Config{
		URL:            url,
		ConnectTimeout: 5 * time.Second,
		RequestTimeout: 5 * time.Second,
	}
}

type Command struct {
	Name   string
	Params map[string]string
}

type Response struct {
	Code    int
	Message string
	Params  map[string]string
}

type Event struct {
	Name      string
	Timestamp time.Time
	Params    map[string]string
}
