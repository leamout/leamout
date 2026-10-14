package freeswitchexporter

import (
	"context"
	"errors"
	"fmt"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func Run(ctx context.Context) error {
	password := os.Getenv("FREESWITCH_ESL_PASSWORD")
	if password == "" {
		return fmt.Errorf("FREESWITCH_ESL_PASSWORD required")
	}
	addr := os.Getenv("FREESWITCH_ESL_ADDRESS")
	if addr == "" {
		addr = "freeswitch:8021"
	}
	client, err := freeswitch.New(freeswitch.DefaultConfig(addr, password))
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect ESL: %w", err)
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", ":9282")
	if err != nil {
		return err
	}
	server := &http.Server{Handler: handler(client), ReadHeaderTimeout: 5 * time.Second}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-done; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func handler(client *freeswitch.Client) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := client.HealthCheck(ctx); err != nil {
			http.Error(w, "FreeSWITCH unavailable", 503)
			return
		}
		calls, err := count(ctx, client, "show calls count")
		if err != nil {
			http.Error(w, "FreeSWITCH call statistics unavailable", 503)
			return
		}
		channels, err := count(ctx, client, "show channels count")
		if err != nil {
			http.Error(w, "FreeSWITCH channel statistics unavailable", 503)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = fmt.Fprintf(w, "# TYPE leamout_freeswitch_up gauge\nleamout_freeswitch_up 1\n# TYPE leamout_freeswitch_calls gauge\nleamout_freeswitch_calls %d\n# TYPE leamout_freeswitch_channels gauge\nleamout_freeswitch_channels %d\n", calls, channels)
	})
	return mux
}
func count(ctx context.Context, client *freeswitch.Client, command string) (uint64, error) {
	reply, err := client.Command(ctx, command)
	if err != nil {
		return 0, err
	}
	return parseCount(reply.Body)
}
func parseCount(value string) (uint64, error) {
	fields := strings.Fields(value)
	if len(fields) < 2 || strings.TrimSuffix(fields[1], ".") != "total" {
		return 0, fmt.Errorf("invalid FreeSWITCH count response")
	}
	return strconv.ParseUint(fields[0], 10, 64)
}
