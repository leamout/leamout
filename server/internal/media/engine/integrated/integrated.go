// Package integrated implements end-to-end realtime voice engines.
package integrated

import (
	"context"
	"fmt"
	"strings"

	aicatalog "github.com/coffeyvidzro/monogo/internal/ai/catalog"
	"github.com/coffeyvidzro/monogo/internal/media/session"
	"github.com/leamout/contracts/ai"
)

type Engine struct {
	Catalog         *aicatalog.Catalog
	DefaultProvider string
}

func (e Engine) Start(ctx context.Context, cfg session.Config) (session.Stream, error) {
	if ctx == nil {
		return nil, fmt.Errorf("integrated engine context is required")
	}
	if cfg.Engine != session.EngineIntegrated {
		return nil, fmt.Errorf("integrated engine cannot start session engine %q", cfg.Engine)
	}
	if e.Catalog == nil {
		return nil, fmt.Errorf("provider catalog is required")
	}

	providerID := strings.TrimSpace(e.DefaultProvider)
	if providerID == "" {
		providerID = "openai"
	}
	runtime := ai.Runtime{}
	if configured, ok := cfg.Provider("realtime"); ok {
		providerID = strings.TrimSpace(configured.Provider)
		runtime = ai.Runtime{
			Credential: configured.APIKey,
			Config:     append([]byte(nil), configured.Config...),
		}
	}

	provider, ok := e.Catalog.Realtime(providerID)
	if !ok {
		return nil, fmt.Errorf("realtime provider %q is not registered", providerID)
	}
	stream, err := provider.StartRealtime(ctx, ai.RealtimeRequest{
		Runtime:      runtime,
		InputFormat:  contractFormat(cfg.InputFormat),
		OutputFormat: contractFormat(cfg.OutputFormat),
		Language:     cfg.Language,
		Instructions: cfg.Instructions,
		Voice:        cfg.Voice,
		Tools:        contractTools(cfg.Tools),
	})
	if err != nil {
		return nil, err
	}
	return newStream(ctx, stream), nil
}

func contractFormat(format session.AudioFormat) ai.AudioFormat {
	return ai.AudioFormat{
		Encoding:     ai.AudioEncodingPCM16LE,
		SampleRateHz: format.SampleRateHz,
		Channels:     format.Channels,
	}
}

func contractTools(tools []session.ToolDefinition) []ai.ToolDefinition {
	result := make([]ai.ToolDefinition, 0, len(tools))
	for _, tool := range tools {
		result = append(result, ai.ToolDefinition{
			ID:          tool.ID.String(),
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  append([]byte(nil), tool.Parameters...),
		})
	}
	return result
}

type stream struct {
	provider ai.RealtimeStream
	events   chan session.Event
	audio    chan session.AudioFrame
	cancel   context.CancelFunc
}

func newStream(parent context.Context, provider ai.RealtimeStream) *stream {
	ctx, cancel := context.WithCancel(parent)
	result := &stream{
		provider: provider,
		events:   make(chan session.Event, 64),
		audio:    make(chan session.AudioFrame, 32),
		cancel:   cancel,
	}
	go result.forwardEvents(ctx)
	go result.forwardAudio(ctx)
	return result
}

func (s *stream) SendAudio(ctx context.Context, frame session.AudioFrame) error {
	return s.provider.SendAudio(ctx, ai.AudioFrame{
		Data:       append([]byte(nil), frame.Data...),
		Format:     contractFormat(frame.Format),
		CapturedAt: frame.CapturedAt,
	})
}

func (s *stream) Interrupt(ctx context.Context) error {
	return s.provider.Interrupt(ctx)
}

func (s *stream) SubmitToolResult(ctx context.Context, result session.ToolResult) error {
	return s.provider.SubmitToolResult(ctx, ai.ToolResult{
		ToolCallID: result.ToolCallID,
		Name:       result.Name,
		Content:    result.Content,
		IsError:    result.IsError,
	})
}

func (s *stream) Audio() <-chan session.AudioFrame { return s.audio }
func (s *stream) Events() <-chan session.Event     { return s.events }

func (s *stream) Close(ctx context.Context) error {
	s.cancel()
	return s.provider.Close(ctx)
}

func (s *stream) forwardAudio(ctx context.Context) {
	defer close(s.audio)
	for {
		select {
		case <-ctx.Done():
			return
		case frame, ok := <-s.provider.Audio():
			if !ok {
				return
			}
			converted := session.AudioFrame{
				Data: append([]byte(nil), frame.Data...),
				Format: session.AudioFormat{
					SampleRateHz: frame.Format.SampleRateHz,
					Channels:     frame.Format.Channels,
				},
				CapturedAt: frame.CapturedAt,
			}
			select {
			case s.audio <- converted:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *stream) forwardEvents(ctx context.Context) {
	defer close(s.events)
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-s.provider.Events():
			if !ok {
				return
			}
			converted := session.Event{
				Type:       session.EventType(event.Type),
				ProviderID: event.ProviderID,
				OccurredAt: event.OccurredAt,
			}
			if event.Transcript != nil {
				converted.Transcript = &session.TranscriptEvent{Text: event.Transcript.Text}
			}
			if event.Response != nil {
				converted.Response = &session.ResponseEvent{Text: event.Response.Text}
			}
			if event.ToolCall != nil {
				converted.ToolCall = &session.ToolCallEvent{
					ID:        event.ToolCall.ID,
					Name:      event.ToolCall.Name,
					Arguments: append([]byte(nil), event.ToolCall.Arguments...),
				}
			}
			if event.Usage != nil {
				converted.Usage = &session.UsageEvent{
					InputTokens:  event.Usage.InputTokens,
					OutputTokens: event.Usage.OutputTokens,
					TotalTokens:  event.Usage.TotalTokens,
				}
			}
			if event.Failure != nil {
				converted.Failure = &session.FailureEvent{
					Source:   event.Failure.Source,
					Code:     event.Failure.Code,
					Message:  event.Failure.Message,
					Terminal: event.Failure.Terminal,
				}
			}
			select {
			case s.events <- converted:
			case <-ctx.Done():
				return
			}
		}
	}
}

var _ session.Stream = (*stream)(nil)
