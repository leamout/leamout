package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/leamout/leamout/server/internal/commercial/licensing"
)

func TestIssueProducesVerifiableLicenseAndNeverOverwrites(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "private.key")
	claimsPath := filepath.Join(dir, "claims.json")
	output := filepath.Join(dir, "license.json")
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	if err := os.WriteFile(keyPath, []byte(base64.StdEncoding.EncodeToString(private)), 0600); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	claims := licensing.Claims{Version: 1, LicenseID: "id", CustomerID: "customer", DeploymentID: "deployment", ValidFrom: now, ExpiresAt: now.Add(time.Hour), GraceUntil: now.Add(time.Hour), MaxConcurrentCalls: 5, MaxNodes: 2}
	data, _ := json.Marshal(claims)
	if err := os.WriteFile(claimsPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := issue(claimsPath, keyPath, "key", output); err != nil {
		t.Fatal(err)
	}
	signed, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := licensing.Verify(signed, map[string]string{"key": base64.StdEncoding.EncodeToString(public)}, "deployment"); err != nil {
		t.Fatal(err)
	}
	if err := issue(claimsPath, keyPath, "key", output); err == nil {
		t.Fatal("overwrote existing license")
	}
}
