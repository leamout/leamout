package freeswitch

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientReconnectsAndRestoresSubscriptions(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	var deliveries atomic.Int32
	events := make(chan struct{}, 4)
	if err := client.Subscribe(
		ctx,
		EventFormatPlain,
		[]string{"CHANNEL_CREATE", "CHANNEL_CREATE"},
		func(context.Context, Event) error {
			deliveries.Add(1)
			events <- struct{}{}
			return nil
		},
	); err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	waitFor(t, 500*time.Millisecond, func() bool {
		return server.SubscriptionCount() == 1 && client.Ready()
	})

	for cycle := int32(1); cycle <= 2; cycle++ {
		server.DropConnections()

		waitFor(t, time.Second, func() bool {
			return server.SubscriptionCount() >= int(cycle)+1 && client.Ready()
		})

		if err := server.SendEvent("CHANNEL_CREATE"); err != nil {
			t.Fatalf("SendEvent() error = %v", err)
		}

		select {
		case <-events:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for restored subscription event")
		}

		time.Sleep(20 * time.Millisecond)
		if got := deliveries.Load(); got != cycle {
			t.Fatalf("deliveries after reconnect cycle %d = %d, want %d", cycle, got, cycle)
		}
	}
}

func TestClientRecoversAfterFreeSWITCHIsUnavailable(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if err := client.Subscribe(ctx, EventFormatPlain, []string{"CHANNEL_ANSWER"}, func(context.Context, Event) error {
		return nil
	}); err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	server.SetAvailable(false)
	server.DropConnections()

	waitFor(t, 500*time.Millisecond, func() bool {
		return !client.Ready()
	})

	commandCtx, commandCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer commandCancel()
	if _, err := client.Command(commandCtx, "status"); err == nil {
		t.Fatal("Command() error = nil while FreeSWITCH is disconnected")
	}
	if err := client.HealthCheck(commandCtx); err == nil {
		t.Fatal("HealthCheck() error = nil while FreeSWITCH is disconnected")
	}

	server.SetAvailable(true)
	waitFor(t, time.Second, func() bool {
		return client.Ready() && server.SubscriptionCount() >= 2
	})

	healthCtx, healthCancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer healthCancel()
	if err := client.HealthCheck(healthCtx); err != nil {
		t.Fatalf("HealthCheck() after recovery error = %v", err)
	}
}

func TestClientCloseCancelsReconnectBackoff(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	client.reconnectMinDelay = time.Second
	client.reconnectMaxDelay = time.Second

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	server.SetAvailable(false)
	server.DropConnections()
	waitFor(t, 500*time.Millisecond, func() bool {
		client.mu.Lock()
		defer client.mu.Unlock()
		return client.reconnecting
	})

	started := time.Now()
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("Close() took %s during reconnect backoff", elapsed)
	}
}

func TestReconnectDelayIsBoundedAndJittered(t *testing.T) {
	client := &Client{
		reconnectMinDelay: 100 * time.Millisecond,
		reconnectMaxDelay: 500 * time.Millisecond,
	}

	first := client.reconnectDelay(0)
	second := client.reconnectDelay(1)
	late := client.reconnectDelay(20)

	if first < 80*time.Millisecond || first > 120*time.Millisecond {
		t.Fatalf("first reconnect delay = %s, want within jitter bounds", first)
	}
	if second == 200*time.Millisecond {
		t.Fatalf("second reconnect delay = %s, want jittered delay", second)
	}
	if late > 500*time.Millisecond {
		t.Fatalf("late reconnect delay = %s, exceeds maximum", late)
	}
}

func newTestClient(t *testing.T, address, password string) *Client {
	t.Helper()

	cfg := DefaultConfig(address, password)
	cfg.ConnectTimeout = 200 * time.Millisecond
	cfg.CommandTimeout = 200 * time.Millisecond
	cfg.ReconnectMinDelay = 10 * time.Millisecond
	cfg.ReconnectMaxDelay = 40 * time.Millisecond

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return client
}

func TestCommandClearsWriteDeadline(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	client.commandTimeout = 30 * time.Millisecond
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	if _, err := client.Command(context.Background(), "status"); err != nil {
		t.Fatalf("first Command() error = %v", err)
	}
	time.Sleep(2 * client.commandTimeout)
	if _, err := client.Command(context.Background(), "status"); err != nil {
		t.Fatalf("Command() after the previous deadline error = %v", err)
	}
}

func TestCommandTimeoutReconnectsBeforeNextCommand(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	client.commandTimeout = 30 * time.Millisecond
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	started := time.Now()
	if _, err := client.Command(context.Background(), "slow"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("slow Command() error = %v, want context deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > 150*time.Millisecond {
		t.Fatalf("slow Command() took %s, want command timeout to bound it", elapsed)
	}

	waitFor(t, time.Second, client.Ready)
	if _, err := client.Command(context.Background(), "status"); err != nil {
		t.Fatalf("Command() after timeout recovery error = %v", err)
	}
}

func TestUnholdUsesFreeSWITCHArgumentOrder(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Unhold(context.Background(), "call-id"); err != nil {
		t.Fatalf("Unhold() error = %v", err)
	}
	if got, want := server.LastCommand(), "api uuid_hold off call-id"; got != want {
		t.Fatalf("Unhold() command = %q, want %q", got, want)
	}
}

func TestBreakPreservesQueuedChannelApplications(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	if err := client.Break(context.Background(), "call-id"); err != nil {
		t.Fatalf("Break() error = %v", err)
	}
	if got, want := server.LastCommand(), "bgapi uuid_break call-id"; got != want {
		t.Fatalf("Break() command = %q, want %q", got, want)
	}
}

func TestBGAPIParsesJobUUIDFromReplyText(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	job, err := client.BGAPI(context.Background(), "uuid_break call-id")
	if err != nil {
		t.Fatalf("BGAPI() error = %v", err)
	}
	if job.ID != "test-job" {
		t.Fatalf("BGAPI() job ID = %q, want test-job", job.ID)
	}
}

func TestPlaybackUsesDisplaceWithoutBreakingPark(t *testing.T) {
	server := newFakeESLServer(t, "secret")
	defer server.Close()

	client := newTestClient(t, server.Address(), "secret")
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	const path = "tone_stream://%(30000,0,440)"
	if err := client.PlayAudio(context.Background(), "call-id", path); err != nil {
		t.Fatalf("PlayAudio() error = %v", err)
	}
	if err := client.StopAudio(context.Background(), "call-id"); err != nil {
		t.Fatalf("StopAudio() error = %v", err)
	}

	want := []string{
		"api uuid_setvar call-id " + displaceHangupOnErrorVar + " false",
		"api uuid_displace call-id start " + path + " 0 mux",
		"api uuid_setvar call-id " + playbackPathVariable + " " + path,
		"api uuid_getvar call-id " + playbackPathVariable,
		"api uuid_displace call-id stop " + path,
		"api uuid_setvar call-id " + playbackPathVariable,
	}
	if got := server.Commands(); !slices.Equal(got, want) {
		t.Fatalf("playback commands = %#v, want %#v", got, want)
	}
}

func waitFor(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}

type fakeESLServer struct {
	t             *testing.T
	password      string
	listener      net.Listener
	available     atomic.Bool
	accepts       atomic.Int32
	subscriptions atomic.Int32

	mu          sync.Mutex
	connections map[net.Conn]struct{}
	latest      net.Conn
	lastCommand string
	commands    []string

	wg sync.WaitGroup
}

func newFakeESLServer(t *testing.T, password string) *fakeESLServer {
	t.Helper()

	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen fake ESL server: %v", err)
	}

	server := &fakeESLServer{
		t:           t,
		password:    password,
		listener:    listener,
		connections: make(map[net.Conn]struct{}),
	}
	server.available.Store(true)
	server.wg.Add(1)
	go server.acceptLoop()

	return server
}

func (s *fakeESLServer) Address() string {
	return s.listener.Addr().String()
}

func (s *fakeESLServer) SubscriptionCount() int {
	return int(s.subscriptions.Load())
}

func (s *fakeESLServer) LastCommand() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastCommand
}

func (s *fakeESLServer) Commands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.commands...)
}

func (s *fakeESLServer) SetAvailable(available bool) {
	s.available.Store(available)
}

func (s *fakeESLServer) DropConnections() {
	s.mu.Lock()
	connections := make([]net.Conn, 0, len(s.connections))
	for conn := range s.connections {
		connections = append(connections, conn)
	}
	s.mu.Unlock()

	for _, conn := range connections {
		_ = conn.Close()
	}
}

func (s *fakeESLServer) SendEvent(name string) error {
	s.mu.Lock()
	conn := s.latest
	s.mu.Unlock()
	if conn == nil {
		return fmt.Errorf("no active ESL connection")
	}

	body := "Event-Name: " + name + "\nUnique-ID: test-channel\n"
	_, err := fmt.Fprintf(
		conn,
		"Content-Type: %s\nContent-Length: %d\n\n%s",
		ContentTypeEventPlain,
		len(body),
		body,
	)
	return err
}

func (s *fakeESLServer) Close() {
	_ = s.listener.Close()
	s.DropConnections()
	s.wg.Wait()
}

func (s *fakeESLServer) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}

		s.accepts.Add(1)
		s.wg.Add(1)
		go s.handle(conn)
	}
}

func (s *fakeESLServer) handle(conn net.Conn) {
	defer s.wg.Done()
	defer func() {
		s.mu.Lock()
		delete(s.connections, conn)
		if s.latest == conn {
			s.latest = nil
		}
		s.mu.Unlock()
		_ = conn.Close()
	}()

	if !s.available.Load() {
		return
	}

	s.mu.Lock()
	s.connections[conn] = struct{}{}
	s.latest = conn
	s.mu.Unlock()

	if _, err := fmt.Fprint(conn, "Content-Type: auth/request\n\n"); err != nil {
		return
	}

	reader := bufio.NewReader(conn)
	command, err := readESLCommand(reader)
	if err != nil || command != "auth "+s.password {
		return
	}
	if _, err := fmt.Fprint(conn, "Content-Type: command/reply\nReply-Text: +OK accepted\n\n"); err != nil {
		return
	}

	for {
		command, err := readESLCommand(reader)
		if err != nil {
			return
		}
		s.mu.Lock()
		s.lastCommand = command
		s.commands = append(s.commands, command)
		s.mu.Unlock()

		switch {
		case strings.HasPrefix(command, "event plain"):
			s.subscriptions.Add(1)
			if _, err := fmt.Fprint(conn, "Content-Type: command/reply\nReply-Text: +OK event listener enabled\n\n"); err != nil {
				return
			}
		case command == "api status":
			body := "UP 0 years, 0 days, 0 hours, 0 minutes, 1 second"
			if _, err := fmt.Fprintf(conn, "Content-Type: api/response\nContent-Length: %d\n\n%s", len(body), body); err != nil {
				return
			}
		case command == "api slow":
			time.Sleep(100 * time.Millisecond)
			if _, err := fmt.Fprint(conn, "Content-Type: api/response\nContent-Length: 3\n\n+OK"); err != nil {
				return
			}
		case command == "api uuid_getvar call-id "+playbackPathVariable:
			body := "tone_stream://%(30000,0,440)"
			if _, err := fmt.Fprintf(conn, "Content-Type: api/response\nContent-Length: %d\n\n%s", len(body), body); err != nil {
				return
			}
		case strings.HasPrefix(command, "bgapi "):
			if _, err := fmt.Fprint(conn, "Content-Type: command/reply\nReply-Text: +OK Job-UUID: test-job\n\n"); err != nil {
				return
			}
		default:
			if _, err := fmt.Fprint(conn, "Content-Type: api/response\nContent-Length: 3\n\n+OK"); err != nil {
				return
			}
		}
	}
}

func readESLCommand(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	command := strings.TrimSpace(line)

	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(line) == "" {
			return command, nil
		}
	}
}
