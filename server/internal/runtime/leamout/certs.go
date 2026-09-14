package leamout

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func runCerts(stdout, stderr io.Writer, args []string) int {
	return runCertsAt(stdout, stderr, args, "/etc/leamout/certs", time.Now().UTC())
}

func runCertsAt(stdout, stderr io.Writer, args []string, certDir string, now time.Time) int {
	if len(args) == 0 || (args[0] != "install" && args[0] != "verify") {
		writeln(stderr, "usage: leamout certs <install|verify> [--fullchain <path> --private-key <path> --carrier-ca <path>] [--hostname <name>]")
		return 2
	}
	action := args[0]
	values := map[string]string{}
	for index := 1; index < len(args); index += 2 {
		if index+1 >= len(args) {
			writef(stderr, "%s requires a value\n", args[index])
			return 2
		}
		switch args[index] {
		case "--fullchain", "--private-key", "--carrier-ca", "--hostname":
			values[args[index]] = args[index+1]
		default:
			writef(stderr, "unknown certificate option: %s\n", args[index])
			return 2
		}
	}
	if action == "install" {
		for _, option := range []string{"--fullchain", "--private-key", "--carrier-ca"} {
			if values[option] == "" {
				writef(stderr, "%s is required for certificate installation\n", option)
				return 2
			}
		}
		if err := installCertificates(certDir, values["--fullchain"], values["--private-key"], values["--carrier-ca"], values["--hostname"], now); err != nil {
			writef(stderr, "install certificates: %v\n", err)
			return 1
		}
		writeln(stdout, "✓ TLS certificates installed")
	} else if err := validateCertificates(certDir, values["--hostname"], now); err != nil {
		writef(stderr, "verify certificates: %v\n", err)
		return 1
	}
	writeln(stdout, "✓ TLS certificates valid")
	return 0
}

func installCertificates(certDir, fullchainPath, privateKeyPath, carrierCAPath, hostname string, now time.Time) error {
	fullchain, err := os.ReadFile(fullchainPath)
	if err != nil {
		return fmt.Errorf("read full chain: %w", err)
	}
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("read private key: %w", err)
	}
	carrierCA, err := os.ReadFile(carrierCAPath)
	if err != nil {
		return fmt.Errorf("read carrier CA: %w", err)
	}
	if err := validateCertificateBytes(fullchain, privateKey, carrierCA, hostname, now); err != nil {
		return err
	}
	if err := os.MkdirAll(certDir, 0o750); err != nil {
		return err
	}
	for name, file := range map[string]struct {
		content []byte
		mode    os.FileMode
	}{
		"fullchain.pem":  {fullchain, 0o640},
		"privkey.pem":    {privateKey, 0o600},
		"carrier-ca.pem": {carrierCA, 0o640},
	} {
		if err := writeAtomicFile(filepath.Join(certDir, name), file.content, file.mode); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	return nil
}

func validateCertificates(certDir, hostname string, now time.Time) error {
	fullchain, err := os.ReadFile(filepath.Join(certDir, "fullchain.pem"))
	if err != nil {
		return err
	}
	privateKey, err := os.ReadFile(filepath.Join(certDir, "privkey.pem"))
	if err != nil {
		return err
	}
	carrierCA, err := os.ReadFile(filepath.Join(certDir, "carrier-ca.pem"))
	if err != nil {
		return err
	}
	return validateCertificateBytes(fullchain, privateKey, carrierCA, hostname, now)
}

func validateCertificateBytes(fullchain, privateKey, carrierCA []byte, hostname string, now time.Time) error {
	certBlock, _ := pem.Decode(fullchain)
	if certBlock == nil || certBlock.Type != "CERTIFICATE" {
		return errors.New("full chain does not start with a certificate")
	}
	certificate, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("parse leaf certificate: %w", err)
	}
	if now.Before(certificate.NotBefore) || !now.Before(certificate.NotAfter) {
		return errors.New("leaf certificate is not currently valid")
	}
	if hostname != "" {
		if err := certificate.VerifyHostname(hostname); err != nil {
			return fmt.Errorf("verify certificate hostname: %w", err)
		}
	}
	keyBlock, _ := pem.Decode(privateKey)
	if keyBlock == nil {
		return errors.New("private key is not PEM encoded")
	}
	key, err := parsePrivateKey(keyBlock.Bytes)
	if err != nil {
		return err
	}
	certificatePublic, err := x509.MarshalPKIXPublicKey(certificate.PublicKey)
	if err != nil {
		return err
	}
	keyPublic, err := x509.MarshalPKIXPublicKey(key.Public())
	if err != nil {
		return err
	}
	if !bytes.Equal(certificatePublic, keyPublic) {
		return errors.New("private key does not match leaf certificate")
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(carrierCA) {
		return errors.New("carrier CA contains no certificates")
	}
	return nil
}

func parsePrivateKey(der []byte) (crypto.Signer, error) {
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		if signer, ok := key.(crypto.Signer); ok {
			return signer, nil
		}
	}
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, errors.New("unsupported private key")
}
