package packages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Parse decodes one manifest and rejects unknown fields or trailing JSON.
func Parse(data []byte) (Manifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode agent package: %w", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return Manifest{}, err
	}
	if err := Validate(manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func ensureEOF(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err == io.EOF {
		return nil
	} else if err != nil {
		return fmt.Errorf("decode trailing agent package data: %w", err)
	}
	return fmt.Errorf("agent package must contain exactly one JSON value")
}
