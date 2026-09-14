package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/licensing"
)

func main() {
	deploymentPath := flag.String("deployment", "", "deployment.json path")
	outputDir := flag.String("out", "", "fixture output directory")
	flag.Parse()
	if *deploymentPath == "" || *outputDir == "" {
		fmt.Fprintln(os.Stderr, "--deployment and --out are required")
		os.Exit(2)
	}
	content, err := os.ReadFile(*deploymentPath)
	must(err)
	var deployment struct {
		DeploymentID string `json:"deployment_id"`
		PublicKey    string `json:"public_key"`
	}
	must(json.Unmarshal(content, &deployment))
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	must(err)
	signer, err := licensing.NewSigner("acceptance-fixture", privateKey)
	must(err)
	now := time.Now().UTC()
	artifact, err := signer.SignV1(licensing.LicenseClaimsV1{
		LicenseID:           uuid.New(),
		OrganizationID:      uuid.New(),
		DeploymentID:        deployment.DeploymentID,
		DeploymentPublicKey: deployment.PublicKey,
		IssuedAt:            now.Add(-time.Minute),
		ExpiresAt:           now.Add(time.Hour),
	})
	must(err)
	keyring, err := json.Marshal(map[string]any{
		"version": 1,
		"keys": map[string]string{
			"acceptance-fixture": base64.RawURLEncoding.EncodeToString(privateKey.Public().(ed25519.PublicKey)),
		},
	})
	must(err)
	must(os.MkdirAll(*outputDir, 0o700))
	must(os.WriteFile(filepath.Join(*outputDir, "license.json"), artifact, 0o600))
	must(os.WriteFile(filepath.Join(*outputDir, "keyring.json"), append(keyring, '\n'), 0o600))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
