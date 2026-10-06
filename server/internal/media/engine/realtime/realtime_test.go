package realtime

import (
	"context"
	"encoding/json"
	"testing"

	aicatalog "github.com/coffeyvidzro/monogo/internal/ai/catalog"
	"github.com/coffeyvidzro/monogo/internal/media/session"
	"github.com/google/uuid"
	"github.com/leamout/contracts/ai"
)

func TestEngineStartsRealtimeProviderWithSessionConfig(t *testing.T) {
	provider := &fakeRealtimeProvider{}
	catalog, err := aicatalog.New(provider)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	format := session.AudioFormat{SampleRateHz: 24000, Channels: 1}
	toolID := uuid.New()
	cfg := session.Config{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		CallID:         uuid.New(),
		ChannelID:      uuid.New(),
		Engine:         session.EngineRealtime,
		InputFormat:    format,
		OutputFormat:   format,
		Instructions:   "Be concise.",
		Voice:          "marin",
		Tools: []session.ToolDefinition{{
			ID:         toolID,
			Name:       "lookup",
			Parameters: json.RawMessage(`{"type":"object"}`),
		}},
		Providers: []session.ProviderRuntime{{
			Role:     "realtime",
			Provider: "openai",
			APIKey:   "secret",
			Config:   json.RawMessage(`{"model":"gpt-realtime-2.1"}`),
		}},
	}
	engine := Engine{Catalog: catalog, DefaultProvider: "openai"}

	stream, err := engine.Start(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer func() { _ = stream.Close(context.Background()) }()

	request := provider.request
	if request.Runtime.Credential != "secret" {
		t.Fatalf("credential = %q", request.Runtime.Credential)
	}
	if request.Instructions != cfg.Instructions {
		t.Fatalf("instructions = %q", request.Instructions)
	}
	if request.Voice != cfg.Voice {
		t.Fatalf("voice = %q", request.Voice)
	}
	if request.InputFormat.Encoding != ai.AudioEncodingPCM16LE || request.InputFormat.SampleRateHz != 24000 {
		t.Fatalf("input format = %+v", request.InputFormat)
	}
	if len(request.Tools) != 1 || request.Tools[0].ID != toolID.String() {
		t.Fatalf("tools = %+v", request.Tools)
	}
}

func TestEngineRejectsNonRealtimeSession(t *testing.T) {
	format := session.AudioFormat{SampleRateHz: 24000, Channels: 1}
	_, err := (Engine{}).Start(context.Background(), session.Config{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		CallID:         uuid.New(),
		ChannelID:      uuid.New(),
		Engine:         session.EngineEcho,
		InputFormat:    format,
		OutputFormat:   format,
	})
	if err == nil {
		t.Fatal("Start() error = nil")
	}
}

type fakeRealtimeProvider struct {
	request ai.RealtimeRequest
}

func (*fakeRealtimeProvider) Descriptor() ai.Descriptor {
	return ai.Descriptor{ID: "openai", Name: "OpenAI", Kind: ai.KindRealtime}
}

func (p *fakeRealtimeProvider) StartRealtime(_ context.Context, request ai.RealtimeRequest) (ai.RealtimeStream, error) {
	p.request = request
	return newFakeRealtimeStream(), nil
}

type fakeRealtimeStream struct {
	audio  chan ai.AudioFrame
	events chan ai.RealtimeEvent
}

func newFakeRealtimeStream() *fakeRealtimeStream {
	return &fakeRealtimeStream{
		audio:  make(chan ai.AudioFrame),
		events: make(chan ai.RealtimeEvent),
	}
}

func (*fakeRealtimeStream) SendAudio(context.Context, ai.AudioFrame) error        { return nil }
func (*fakeRealtimeStream) Interrupt(context.Context) error                       { return nil }
func (*fakeRealtimeStream) SubmitToolResult(context.Context, ai.ToolResult) error { return nil }
func (s *fakeRealtimeStream) Audio() <-chan ai.AudioFrame                         { return s.audio }
func (s *fakeRealtimeStream) Events() <-chan ai.RealtimeEvent                     { return s.events }
func (s *fakeRealtimeStream) Close(context.Context) error {
	close(s.audio)
	close(s.events)
	return nil
}

var _ ai.Realtime = (*fakeRealtimeProvider)(nil)
var _ ai.RealtimeStream = (*fakeRealtimeStream)(nil)
