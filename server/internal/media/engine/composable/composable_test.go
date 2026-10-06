package composable

import (
	"context"
	"sync"
	"testing"
	"time"

	aicatalog "github.com/coffeyvidzro/monogo/internal/ai/catalog"
	"github.com/coffeyvidzro/monogo/internal/media/session"
	"github.com/google/uuid"
	"github.com/leamout/contracts/ai"
)

func TestComposableRunsTurnThroughRegisteredProviders(t *testing.T) {
	transcriber := newFakeTranscriber()
	generator := &fakeGenerator{deltas: []string{"hello. ", "caller"}}
	synthesizer := &fakeSynthesizer{audio: []byte{1, 0, 2, 0}}
	stream := startTestStream(t, transcriber, generator, synthesizer)

	input := session.AudioFrame{Data: []byte{9, 0}, Format: testFormat()}
	if err := stream.SendAudio(context.Background(), input); err != nil {
		t.Fatalf("SendAudio() error = %v", err)
	}
	if got := <-transcriber.audio; string(got.Data) != string(input.Data) {
		t.Fatalf("transcriber audio = %v", got.Data)
	}
	transcriber.events <- ai.STTEvent{
		Type:       ai.STTEventSpeechStarted,
		ProviderID: "stt-1",
	}
	transcriber.events <- ai.STTEvent{
		Type:       ai.STTEventSpeechStopped,
		ProviderID: "stt-1",
		Text:       "hi",
	}

	select {
	case frame := <-stream.Audio():
		if string(frame.Data) != string(synthesizer.audio) {
			t.Fatalf("output audio = %v", frame.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for synthesized audio")
	}
	if generator.lastUser() != "hi" {
		t.Fatalf("LLM user message = %q", generator.lastUser())
	}
	if synthesizer.lastText() != "hello.caller" {
		t.Fatalf("TTS text = %q", synthesizer.lastText())
	}
	if got := synthesizer.finalFlags(); len(got) != 2 || got[0] || !got[1] {
		t.Fatalf("TTS final flags = %v", got)
	}

	foundResponseDelta := false
	for !foundResponseDelta {
		select {
		case event := <-stream.Events():
			if event.Type == session.EventResponseDelta {
				if event.Response == nil || event.Response.Text == "" {
					t.Fatalf("response delta payload = %+v", event.Response)
				}
				foundResponseDelta = true
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for response delta event")
		}
	}
	closeTestStream(t, stream)
}

func TestComposableBargeInCancelsResponseAndDropsStaleAudio(t *testing.T) {
	transcriber := newFakeTranscriber()
	synthesizer := &fakeSynthesizer{audio: []byte{7, 0}, waitForCancel: true}
	stream := startTestStream(t, transcriber, &fakeGenerator{deltas: []string{"first response"}}, synthesizer)
	transcriber.events <- ai.STTEvent{
		Type: ai.STTEventSpeechStopped,
		Text: "first",
	}

	select {
	case <-synthesizer.started:
	case <-time.After(time.Second):
		t.Fatal("TTS did not start")
	}
	transcriber.events <- ai.STTEvent{Type: ai.STTEventSpeechStarted}
	select {
	case <-synthesizer.cancelled:
	case <-time.After(time.Second):
		t.Fatal("barge-in did not cancel TTS")
	}
	select {
	case frame := <-stream.Audio():
		t.Fatalf("stale audio reached output: %v", frame.Data)
	case <-time.After(50 * time.Millisecond):
	}
	closeTestStream(t, stream)
}

func TestComposableRejectsUnregisteredProvider(t *testing.T) {
	catalog, err := aicatalog.New(newFakeTranscriber(), &fakeGenerator{}, &fakeSynthesizer{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	cfg := testSessionConfig()
	cfg.Providers = []session.ProviderRuntime{{
		Role: "llm", Provider: "missing", APIKey: "secret",
	}}
	_, err = (Engine{Catalog: catalog}).Start(context.Background(), cfg)
	if err == nil {
		t.Fatal("Start() error = nil")
	}
}

func startTestStream(
	t *testing.T,
	transcriber *fakeTranscriber,
	generator *fakeGenerator,
	synthesizer *fakeSynthesizer,
) session.Stream {
	t.Helper()
	if synthesizer.started == nil {
		synthesizer.started = make(chan struct{})
	}
	if synthesizer.cancelled == nil {
		synthesizer.cancelled = make(chan struct{})
	}
	catalog, err := aicatalog.New(transcriber, generator, synthesizer)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got, err := (Engine{Catalog: catalog}).Start(context.Background(), testSessionConfig())
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	return got
}

func testSessionConfig() session.Config {
	return session.Config{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		CallID:         uuid.New(),
		ChannelID:      uuid.New(),
		Engine:         session.EngineComposable,
		InputFormat:    testFormat(),
		OutputFormat:   testFormat(),
		Instructions:   "be helpful",
		Voice:          "voice-override",
	}
}

func closeTestStream(t *testing.T, stream session.Stream) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := stream.Close(ctx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func testFormat() session.AudioFormat {
	return session.AudioFormat{SampleRateHz: 16000, Channels: 1}
}

type fakeTranscriber struct {
	audio     chan ai.AudioFrame
	events    chan ai.STTEvent
	closeOnce sync.Once
}

func newFakeTranscriber() *fakeTranscriber {
	return &fakeTranscriber{
		audio:  make(chan ai.AudioFrame, 1),
		events: make(chan ai.STTEvent, 8),
	}
}

func (f *fakeTranscriber) Descriptor() ai.Descriptor {
	return ai.Descriptor{ID: "deepgram", Name: "Deepgram", Kind: ai.KindSTT}
}

func (f *fakeTranscriber) StartSTT(
	context.Context,
	ai.STTRequest,
) (ai.STTStream, error) {
	return f, nil
}

func (f *fakeTranscriber) SendAudio(ctx context.Context, frame ai.AudioFrame) error {
	select {
	case f.audio <- frame:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (f *fakeTranscriber) Finalize(context.Context) error { return nil }
func (f *fakeTranscriber) Events() <-chan ai.STTEvent     { return f.events }
func (f *fakeTranscriber) Close(context.Context) error {
	f.closeOnce.Do(func() { close(f.events) })
	return nil
}

type fakeGenerator struct {
	deltas   []string
	mu       sync.Mutex
	messages []ai.Message
}

func (f *fakeGenerator) Descriptor() ai.Descriptor {
	return ai.Descriptor{ID: "groq", Name: "Groq", Kind: ai.KindLLM}
}

func (f *fakeGenerator) Generate(
	_ context.Context,
	req ai.LLMRequest,
) (ai.LLMStream, error) {
	f.mu.Lock()
	f.messages = append([]ai.Message(nil), req.Messages...)
	f.mu.Unlock()
	events := make(chan ai.LLMEvent, len(f.deltas)+1)
	for _, delta := range f.deltas {
		events <- ai.LLMEvent{ResponseID: "llm-1", TextDelta: delta}
	}
	events <- ai.LLMEvent{Done: true}
	close(events)
	return &fakeLLMStream{events: events}, nil
}

func (f *fakeGenerator) lastUser() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.messages) - 1; i >= 0; i-- {
		if f.messages[i].Role == ai.RoleUser {
			return f.messages[i].Content
		}
	}
	return ""
}

type fakeLLMStream struct {
	events chan ai.LLMEvent
}

func (f *fakeLLMStream) Events() <-chan ai.LLMEvent { return f.events }
func (f *fakeLLMStream) Close() error               { return nil }

type fakeSynthesizer struct {
	audio         []byte
	waitForCancel bool
	started       chan struct{}
	cancelled     chan struct{}
	mu            sync.Mutex
	text          string
	final         []bool
}

func (f *fakeSynthesizer) Descriptor() ai.Descriptor {
	return ai.Descriptor{ID: "cartesia", Name: "Cartesia", Kind: ai.KindTTS}
}

func (f *fakeSynthesizer) StartTTS(
	ctx context.Context,
	req ai.TTSRequest,
) (ai.TTSStream, error) {
	f.mu.Lock()
	if f.started == nil {
		f.started = make(chan struct{})
	}
	if f.cancelled == nil {
		f.cancelled = make(chan struct{})
	}
	started, cancelled := f.started, f.cancelled
	f.mu.Unlock()
	events := make(chan ai.TTSEvent, 2)
	close(started)
	result := &fakeTTSStream{events: events}
	result.onText = func(chunk ai.TextChunk) {
		f.mu.Lock()
		f.text += chunk.Text
		f.final = append(f.final, chunk.Final)
		f.mu.Unlock()
		if chunk.Final && !f.waitForCancel {
			events <- ai.TTSEvent{
				Audio: ai.AudioFrame{Data: f.audio, Format: req.Format},
			}
			events <- ai.TTSEvent{Done: true}
			close(events)
		}
	}
	if f.waitForCancel {
		go func() {
			<-ctx.Done()
			close(cancelled)
			// Deliberately publish after cancellation to verify generation fencing.
			events <- ai.TTSEvent{
				Audio: ai.AudioFrame{Data: f.audio, Format: req.Format},
			}
			close(events)
		}()
	}
	return result, nil
}

func (f *fakeSynthesizer) finalFlags() []bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]bool(nil), f.final...)
}

func (f *fakeSynthesizer) lastText() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.text
}

type fakeTTSStream struct {
	events chan ai.TTSEvent
	onText func(ai.TextChunk)
}

func (f *fakeTTSStream) Events() <-chan ai.TTSEvent { return f.events }
func (f *fakeTTSStream) Close() error               { return nil }
func (f *fakeTTSStream) SendText(_ context.Context, chunk ai.TextChunk) error {
	f.onText(chunk)
	return nil
}
