// Command license issues signed licenses offline. Never ship private keys to customers.
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/leamout/leamout/server/internal/commercial/licensing"
)

func main() {
	claimsPath := flag.String("claims", "", "JSON claims file")
	keyPath := flag.String("private-key", "", "file containing base64 Ed25519 private key")
	keyID := flag.String("key-id", "", "signing key ID")
	output := flag.String("output", "", "new signed license file (must not exist)")
	generate := flag.Bool("generate-key", false, "generate a new issuer private key at -output; prints public key only")
	flag.Parse()
	if *generate {
		if err := generateKey(*output); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if err := issue(*claimsPath, *keyPath, *keyID, *output); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func issue(claimsPath, keyPath, keyID, output string) error {
	if claimsPath == "" || keyPath == "" || keyID == "" || output == "" {
		return fmt.Errorf("claims, private-key, key-id and output are required")
	}
	payload, err := os.ReadFile(claimsPath)
	if err != nil {
		return err
	}
	var claims licensing.Claims
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("trailing license claims data")
	}
	if err := licensing.ValidateClaims(claims); err != nil {
		return err
	}
	// Serialize known fields into a canonical issued payload.
	payload, err = json.Marshal(claims)
	if err != nil {
		return err
	}
	encoded, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	key, err := base64.StdEncoding.DecodeString(string(bytes.TrimSpace(encoded)))
	if err != nil || len(key) != ed25519.PrivateKeySize {
		return fmt.Errorf("private key must be a base64 Ed25519 private key")
	}
	doc := licensing.Document{
		KeyID:     keyID,
		Payload:   base64.StdEncoding.EncodeToString(payload),
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(ed25519.PrivateKey(key), payload)),
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = file.Write(append(data, '\n'))
	closeErr := file.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func generateKey(path string) error {
	if path == "" {
		return fmt.Errorf("output is required")
	}
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = file.WriteString(base64.StdEncoding.EncodeToString(private) + "\n")
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	_, _ = fmt.Println(base64.StdEncoding.EncodeToString(public))
	return nil
}
