// Package composable orchestrates provider-neutral speech-to-text, language
// model, and text-to-speech adapters as one realtime media session.
package composable

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	aicatalog "github.com/coffeyvidzro/monogo/internal/ai/catalog"
	"github.com/coffeyvidzro/monogo/internal/media/session"
	"github.com/leamout/contracts/ai"
)

type Engine struct {
	Catalog            *aicatalog.Catalog
	DefaultSTTProvider string
	DefaultLLMProvider string
	DefaultTTSProvider string
}

func (e Engine) Start(ctx context.Context, cfg session.Config) (session.Stream, error) {
	if ctx == nil {
		return nil, fmt.Errorf("composable engine context is required")
	}
	if cfg.Engine != session.EngineComposable {
		return nil, fmt.Errorf("composable engine cannot start session engine %q", cfg.Engine)
	}
	if e.Catalog == nil {
		return nil, fmt.Errorf("AI provider catalog is required")
	}

	sttID, sttRuntime := providerRuntime(cfg, "stt", e.DefaultSTTProvider, "deepgram")
	llmID, llmRuntime := providerRuntime(cfg, "llm", e.DefaultLLMProvider, "groq")
	ttsID, ttsRuntime := providerRuntime(cfg, "tts", e.DefaultTTSProvider, "cartesia")

	stt, ok := e.Catalog.STT(sttID)
	if !ok {
		return nil, fmt.Errorf("STT provider %q is not registered", sttID)
	}
	llm, ok := e.Catalog.LLM(llmID)
	if !ok {
		return nil, fmt.Errorf("LLM provider %q is not registered", llmID)
	}
	tts, ok := e.Catalog.TTS(ttsID)
	if !ok {
		return nil, fmt.Errorf("TTS provider %q is not registered", ttsID)
	}

	transcriber, err := stt.StartSTT(ctx, ai.STTRequest{
		Runtime:  sttRuntime,
		Format:   providerAudioFormat(cfg.InputFormat),
		Language: cfg.Language,
	})
	if err != nil {
		return nil, fmt.Errorf("start transcription: %w", err)
	}

	streamCtx, cancel := context.WithCancel(ctx)
	s := &stream{
		ctx:              streamCtx,
		cancel:           cancel,
		transcriber:      transcriber,
		llm:              llm,
		tts:              tts,
		llmRuntime:       llmRuntime,
		ttsRuntime:       ttsRuntime,
		sttProviderID:    sttID,
		llmProviderID:    llmID,
		ttsProviderID:    ttsID,
		config:           cfg,
		audio:            make(chan session.AudioFrame),
		events:           make(chan session.Event, 32),
		done:             make(chan struct{}),
		pendingToolCalls: make(map[string]string),
	}
	if instructions := strings.TrimSpace(cfg.Instructions); instructions != "" {
		s.messages = append(s.messages, ai.Message{Role: ai.RoleSystem, Content: instructions})
	}
	go s.run()
	return s, nil
}

func providerRuntime(
	cfg session.Config,
	role string,
	configuredDefault string,
	fallback string,
) (string, ai.Runtime) {
	providerID := strings.TrimSpace(configuredDefault)
	if providerID == "" {
		providerID = fallback
	}
	if configured, ok := cfg.Provider(role); ok {
		providerID = strings.TrimSpace(configured.Provider)
		return providerID, ai.Runtime{
			Credential: configured.APIKey,
			Config:     append(json.RawMessage(nil), configured.Config...),
		}
	}
	return providerID, ai.Runtime{}
}

type stream struct {
	ctx         context.Context
	cancel      context.CancelFunc
	transcriber ai.STTStream
	llm         ai.LLM
	tts         ai.TTS
	llmRuntime  ai.Runtime
	ttsRuntime  ai.Runtime

	sttProviderID string
	llmProviderID string
	ttsProviderID string
	config        session.Config
	audio         chan session.AudioFrame
	events        chan session.Event

	mu               sync.Mutex
	messages         []ai.Message
	generation       uint64
	responseCancel   context.CancelFunc
	responseActive   bool
	pendingToolCalls map[string]string
	responses        sync.WaitGroup
	closeOnce        sync.Once
	done             chan struct{}
}

func (s *stream) SendAudio(ctx context.Context, frame session.AudioFrame) error {
	return s.transcriber.SendAudio(ctx, providerAudioFrame(frame))
}

func (s *stream) Interrupt(context.Context) error {
	s.cancelResponse()
	return nil
}

func (s *stream) SubmitToolResult(ctx context.Context, result session.ToolResult) error {
	if ctx == nil {
		return fmt.Errorf("tool result context is required")
	}
	s.mu.Lock()
	name, exists := s.pendingToolCalls[result.ToolCallID]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("unknown pending tool call %q", result.ToolCallID)
	}
	if result.Name != "" && result.Name != name {
		s.mu.Unlock()
		return fmt.Errorf("tool result name does not match pending call")
	}
	delete(s.pendingToolCalls, result.ToolCallID)
	s.messages = append(s.messages, ai.Message{
		Role:       ai.RoleTool,
		Content:    result.Content,
		ToolCallID: result.ToolCallID,
	})
	if len(s.pendingToolCalls) != 0 {
		s.mu.Unlock()
		return nil
	}
	s.generation++
	generation := s.generation
	responseCtx, cancel := context.WithCancel(s.ctx)
	s.responseCancel = cancel
	s.responseActive = true
	messages := append([]ai.Message(nil), s.messages...)
	s.responses.Add(1)
	s.mu.Unlock()
	s.emit(session.Event{Type: session.EventResponseStarted, OccurredAt: time.Now().UTC()})
	go s.generate(responseCtx, generation, messages)
	return nil
}

func (s *stream) Audio() <-chan session.AudioFrame { return s.audio }
func (s *stream) Events() <-chan session.Event     { return s.events }

func (s *stream) Close(ctx context.Context) error {
	s.closeOnce.Do(func() { s.cancel() })
	if err := s.transcriber.Close(ctx); err != nil && ctx.Err() == nil {
		return err
	}
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *stream) run() {
	defer close(s.done)
	defer close(s.audio)
	defer close(s.events)
	for {
		select {
		case event, ok := <-s.transcriber.Events():
			if !ok {
				s.cancelResponse()
				s.responses.Wait()
				return
			}
			if event.Err != nil || event.Type == ai.STTEventError {
				message := "speech-to-text provider failed"
				if event.Err != nil {
					message = event.Err.Error()
				}
				s.emit(session.Event{
					Type: session.EventError,
					Failure: &session.FailureEvent{
						Source: s.sttProviderID, Message: message, Terminal: true,
					},
					OccurredAt: time.Now().UTC(),
				})
				s.cancelResponse()
				s.responses.Wait()
				return
			}
			s.handleTurn(event)
		case <-s.ctx.Done():
			s.cancelResponse()
			s.responses.Wait()
			return
		}
	}
}

func (s *stream) handleTurn(event ai.STTEvent) {
	switch event.Type {
	case ai.STTEventSpeechStarted:
		s.emit(session.Event{
			Type:       session.EventSpeechStarted,
			ProviderID: event.ProviderID,
			OccurredAt: time.Now().UTC(),
		})
		// Publish speech first so the transport clears already-buffered playback;
		// generation cancellation below fences all subsequent provider audio.
		s.cancelResponse()
	case ai.STTEventSpeechStopped:
		text := strings.TrimSpace(event.Text)
		s.emit(session.Event{
			Type:       session.EventSpeechStopped,
			ProviderID: event.ProviderID,
			OccurredAt: time.Now().UTC(),
		})
		if text == "" {
			return
		}
		s.emit(session.Event{
			Type:       session.EventTranscriptFinal,
			Transcript: &session.TranscriptEvent{Text: text},
			ProviderID: event.ProviderID,
			OccurredAt: time.Now().UTC(),
		})
		s.startResponse(text)
	}
}

func (s *stream) startResponse(text string) {
	s.mu.Lock()
	if s.responseCancel != nil {
		s.responseCancel()
	}
	s.generation++
	generation := s.generation
	responseCtx, cancel := context.WithCancel(s.ctx)
	s.responseCancel = cancel
	s.responseActive = true
	s.messages = append(s.messages, ai.Message{Role: ai.RoleUser, Content: text})
	messages := append([]ai.Message(nil), s.messages...)
	s.responses.Add(1)
	s.mu.Unlock()
	s.emit(session.Event{Type: session.EventResponseStarted, OccurredAt: time.Now().UTC()})
	go s.generate(responseCtx, generation, messages)
}

func (s *stream) generate(
	ctx context.Context,
	generation uint64,
	messages []ai.Message,
) {
	defer s.responses.Done()
	completion, err := s.llm.Generate(ctx, ai.LLMRequest{
		Runtime:      s.llmRuntime,
		Messages:     messages,
		Tools:        providerTools(s.config.Tools),
		Instructions: s.config.Instructions,
	})
	if err != nil {
		s.failResponse(ctx, generation, err)
		return
	}
	defer func() { _ = completion.Close() }()

	voice, err := s.tts.StartTTS(ctx, ai.TTSRequest{
		Runtime:  s.ttsRuntime,
		Format:   providerAudioFormat(s.config.OutputFormat),
		Voice:    s.config.Voice,
		Language: s.config.Language,
	})
	if err != nil {
		s.failResponse(ctx, generation, err)
		return
	}
	defer func() { _ = voice.Close() }()

	var text strings.Builder
	var pending strings.Builder
	var heldChunk string
	toolCalls := make(map[int]*session.ToolCallEvent)
	toolOrder := make([]int, 0, 1)
	completionEvents := completion.Events()
	for {
		select {
		case event, ok := <-completionEvents:
			if !ok || event.Done {
				completeCalls := make([]*session.ToolCallEvent, 0, len(toolOrder))
				for _, index := range toolOrder {
					call := toolCalls[index]
					if call == nil || call.ID == "" || call.Name == "" || !json.Valid(call.Arguments) {
						s.emitCurrent(ctx, generation, session.Event{
							Type: session.EventError,
							Failure: &session.FailureEvent{
								Source: s.llmProviderID, Code: "invalid_tool_call",
								Message: "LLM provider returned an incomplete tool call", Terminal: false,
							},
							OccurredAt: time.Now().UTC(),
						})
						continue
					}
					completeCalls = append(completeCalls, call)
					s.emitCurrent(ctx, generation, session.Event{
						Type:       session.EventToolCall,
						ToolCall:   call,
						OccurredAt: time.Now().UTC(),
					})
				}
				if len(completeCalls) > 0 {
					s.awaitToolResults(generation, completeCalls)
					return
				}
				finalChunk := strings.TrimSpace(pending.String())
				switch {
				case heldChunk != "" && finalChunk != "":
					if err := voice.SendText(ctx, ai.TextChunk{Text: heldChunk}); err != nil {
						s.failResponse(ctx, generation, err)
						return
					}
					if err := voice.SendText(ctx, ai.TextChunk{Text: finalChunk, Final: true}); err != nil {
						s.failResponse(ctx, generation, err)
						return
					}
				case heldChunk != "":
					if err := voice.SendText(ctx, ai.TextChunk{Text: heldChunk, Final: true}); err != nil {
						s.failResponse(ctx, generation, err)
						return
					}
				case finalChunk != "":
					if err := voice.SendText(ctx, ai.TextChunk{Text: finalChunk, Final: true}); err != nil {
						s.failResponse(ctx, generation, err)
						return
					}
				default:
					s.stopResponse(generation, "")
					return
				}
				pending.Reset()
				heldChunk = ""
				completionEvents = nil
				continue
			}
			if event.Err != nil {
				s.failResponse(ctx, generation, event.Err)
				return
			}
			if event.InputTokens != 0 || event.OutputTokens != 0 || event.TotalTokens != 0 {
				s.emitCurrent(ctx, generation, session.Event{
					Type: session.EventUsage,
					Usage: &session.UsageEvent{
						InputTokens:  event.InputTokens,
						OutputTokens: event.OutputTokens,
						TotalTokens:  event.TotalTokens,
					},
					ProviderID: event.ResponseID,
					OccurredAt: time.Now().UTC(),
				})
			}
			if event.TextDelta != "" {
				text.WriteString(event.TextDelta)
				s.emitCurrent(ctx, generation, session.Event{
					Type:       session.EventResponseDelta,
					Response:   &session.ResponseEvent{Text: event.TextDelta},
					ProviderID: event.ResponseID,
					OccurredAt: time.Now().UTC(),
				})
				pending.WriteString(event.TextDelta)
				if shouldFlushSpeechChunk(pending.String()) {
					chunk := strings.TrimSpace(pending.String())
					pending.Reset()
					if heldChunk != "" {
						if err := voice.SendText(ctx, ai.TextChunk{Text: heldChunk}); err != nil {
							s.failResponse(ctx, generation, err)
							return
						}
					}
					heldChunk = chunk
				}
			}
			if event.ToolCallID != "" || event.ToolName != "" || len(event.ToolArguments) != 0 {
				call, exists := toolCalls[event.ToolIndex]
				if !exists {
					call = &session.ToolCallEvent{}
					toolCalls[event.ToolIndex] = call
					toolOrder = append(toolOrder, event.ToolIndex)
				}
				if event.ToolCallID != "" {
					call.ID = event.ToolCallID
				}
				if event.ToolName != "" {
					call.Name = event.ToolName
				}
				call.Arguments = append(call.Arguments, event.ToolArguments...)
			}
		case <-ctx.Done():
			return
		case event, ok := <-voice.Events():
			if !ok || event.Done {
				s.stopResponse(generation, strings.TrimSpace(text.String()))
				return
			}
			if event.Err != nil {
				s.failResponse(ctx, generation, event.Err)
				return
			}
			if len(event.Audio.Data) != 0 && s.isCurrent(generation) {
				frame, err := sessionAudioFrame(event.Audio)
				if err != nil {
					s.failResponse(ctx, generation, err)
					return
				}
				select {
				case s.audio <- frame:
				case <-ctx.Done():
					return
				case <-s.ctx.Done():
					return
				}
			}
		}
	}
}

func (s *stream) awaitToolResults(generation uint64, calls []*session.ToolCallEvent) {
	s.mu.Lock()
	if generation != s.generation || !s.responseActive {
		s.mu.Unlock()
		return
	}
	s.responseActive = false
	s.responseCancel = nil
	message := ai.Message{Role: ai.RoleAssistant}
	for _, call := range calls {
		message.ToolCalls = append(message.ToolCalls, ai.ToolCall{
			ID:        call.ID,
			Name:      call.Name,
			Arguments: append(json.RawMessage(nil), call.Arguments...),
		})
		s.pendingToolCalls[call.ID] = call.Name
	}
	s.messages = append(s.messages, message)
	s.mu.Unlock()
	s.emit(session.Event{Type: session.EventResponseStopped, OccurredAt: time.Now().UTC()})
}

func (s *stream) cancelResponse() {
	s.mu.Lock()
	s.generation++
	cancel, active := s.responseCancel, s.responseActive
	s.responseCancel = nil
	s.responseActive = false
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if active {
		s.emit(session.Event{Type: session.EventResponseStopped, OccurredAt: time.Now().UTC()})
	}
}

func (s *stream) stopResponse(generation uint64, assistantText string) {
	s.mu.Lock()
	if generation != s.generation || !s.responseActive {
		s.mu.Unlock()
		return
	}
	s.responseActive = false
	s.responseCancel = nil
	if assistantText != "" {
		s.messages = append(s.messages, ai.Message{Role: ai.RoleAssistant, Content: assistantText})
	}
	s.mu.Unlock()
	s.emit(session.Event{Type: session.EventResponseStopped, OccurredAt: time.Now().UTC()})
}

func (s *stream) failResponse(ctx context.Context, generation uint64, err error) {
	if ctx.Err() != nil || !s.isCurrent(generation) {
		return
	}
	s.emit(session.Event{
		Type: session.EventError,
		Failure: &session.FailureEvent{
			Source: "composable", Message: err.Error(), Terminal: false,
		},
		OccurredAt: time.Now().UTC(),
	})
	s.stopResponse(generation, "")
}

func (s *stream) isCurrent(generation uint64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return generation == s.generation && s.responseActive
}

func (s *stream) emitCurrent(ctx context.Context, generation uint64, event session.Event) {
	if s.isCurrent(generation) {
		select {
		case s.events <- event:
		case <-ctx.Done():
		case <-s.ctx.Done():
		}
	}
}

func (s *stream) emit(event session.Event) {
	select {
	case s.events <- event:
	case <-s.ctx.Done():
	}
}

func providerAudioFormat(format session.AudioFormat) ai.AudioFormat {
	return ai.AudioFormat{
		Encoding:     ai.AudioEncodingPCM16LE,
		SampleRateHz: format.SampleRateHz,
		Channels:     format.Channels,
	}
}

func providerAudioFrame(frame session.AudioFrame) ai.AudioFrame {
	return ai.AudioFrame{
		Data:       frame.Data,
		Format:     providerAudioFormat(frame.Format),
		CapturedAt: frame.CapturedAt,
	}
}

func sessionAudioFrame(frame ai.AudioFrame) (session.AudioFrame, error) {
	if frame.Format.Encoding != ai.AudioEncodingPCM16LE {
		return session.AudioFrame{}, fmt.Errorf(
			"composable engine received unsupported provider audio encoding %q",
			frame.Format.Encoding,
		)
	}
	return session.AudioFrame{
		Data: frame.Data,
		Format: session.AudioFormat{
			SampleRateHz: frame.Format.SampleRateHz,
			Channels:     frame.Format.Channels,
		},
		CapturedAt: frame.CapturedAt,
	}, nil
}

func providerTools(tools []session.ToolDefinition) []ai.ToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	result := make([]ai.ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		result = append(result, ai.ToolDefinition{
			ID:          tool.ID.String(),
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  append(json.RawMessage(nil), tool.Parameters...),
		})
	}
	return result
}

func shouldFlushSpeechChunk(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	last := trimmed[len(trimmed)-1]
	if last == '.' || last == '!' || last == '?' || last == ';' || last == ':' || last == '\n' {
		return true
	}
	return len(trimmed) >= 120 && len(text) > 0 && (text[len(text)-1] == ' ' || text[len(text)-1] == '\n')
}
