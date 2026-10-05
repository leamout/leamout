package catalog

import (
	"fmt"
	"sort"
	"strings"

	"github.com/leamout/contracts/ai"
)

type Catalog struct {
	descriptors map[string]ai.Descriptor
	stt         map[string]ai.STT
	llm         map[string]ai.LLM
	tts         map[string]ai.TTS
	realtime    map[string]ai.Realtime
}

func New(providers ...ai.Provider) (*Catalog, error) {
	catalog := &Catalog{
		descriptors: make(map[string]ai.Descriptor, len(providers)),
		stt:         make(map[string]ai.STT),
		llm:         make(map[string]ai.LLM),
		tts:         make(map[string]ai.TTS),
		realtime:    make(map[string]ai.Realtime),
	}
	for _, provider := range providers {
		if err := catalog.Register(provider); err != nil {
			return nil, err
		}
	}
	return catalog, nil
}

func (c *Catalog) Register(provider ai.Provider) error {
	if c == nil {
		return fmt.Errorf("AI provider catalog is required")
	}
	if provider == nil {
		return fmt.Errorf("AI provider implementation is required")
	}

	descriptor := provider.Descriptor()
	if err := descriptor.Validate(); err != nil {
		return err
	}

	id := strings.TrimSpace(descriptor.ID)
	key := catalogKey(descriptor.Kind, id)
	if _, exists := c.descriptors[key]; exists {
		return fmt.Errorf(
			"AI provider %q already registered for %q",
			id,
			descriptor.Kind,
		)
	}

	switch descriptor.Kind {
	case ai.KindSTT:
		implementation, ok := provider.(ai.STT)
		if !ok {
			return fmt.Errorf("AI provider %q does not implement STT", id)
		}
		c.stt[id] = implementation
	case ai.KindLLM:
		implementation, ok := provider.(ai.LLM)
		if !ok {
			return fmt.Errorf("AI provider %q does not implement LLM", id)
		}
		c.llm[id] = implementation
	case ai.KindTTS:
		implementation, ok := provider.(ai.TTS)
		if !ok {
			return fmt.Errorf("AI provider %q does not implement TTS", id)
		}
		c.tts[id] = implementation
	case ai.KindRealtime:
		implementation, ok := provider.(ai.Realtime)
		if !ok {
			return fmt.Errorf("AI provider %q does not implement realtime", id)
		}
		c.realtime[id] = implementation
	default:
		return fmt.Errorf("unsupported AI provider kind %q", descriptor.Kind)
	}

	copyDescriptor := descriptor
	copyDescriptor.ID = id
	copyDescriptor.Capabilities = append(
		[]ai.Capability(nil),
		descriptor.Capabilities...,
	)
	c.descriptors[key] = copyDescriptor
	return nil
}

func (c *Catalog) Get(kind ai.Kind, id string) (ai.Descriptor, bool) {
	if c == nil {
		return ai.Descriptor{}, false
	}

	descriptor, ok := c.descriptors[catalogKey(kind, id)]
	if !ok {
		return ai.Descriptor{}, false
	}
	descriptor.Capabilities = append(
		[]ai.Capability(nil),
		descriptor.Capabilities...,
	)
	return descriptor, true
}

func (c *Catalog) STT(id string) (ai.STT, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.stt[strings.TrimSpace(id)]
	return provider, ok
}

func (c *Catalog) LLM(id string) (ai.LLM, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.llm[strings.TrimSpace(id)]
	return provider, ok
}

func (c *Catalog) TTS(id string) (ai.TTS, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.tts[strings.TrimSpace(id)]
	return provider, ok
}

func (c *Catalog) Realtime(id string) (ai.Realtime, bool) {
	if c == nil {
		return nil, false
	}
	provider, ok := c.realtime[strings.TrimSpace(id)]
	return provider, ok
}

func (c *Catalog) List(kind ai.Kind) []ai.Descriptor {
	if c == nil {
		return nil
	}

	result := make([]ai.Descriptor, 0)
	for _, descriptor := range c.descriptors {
		if descriptor.Kind != kind {
			continue
		}
		copyDescriptor := descriptor
		copyDescriptor.Capabilities = append(
			[]ai.Capability(nil),
			descriptor.Capabilities...,
		)
		result = append(result, copyDescriptor)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

func catalogKey(kind ai.Kind, id string) string {
	return string(kind) + ":" + strings.TrimSpace(id)
}
