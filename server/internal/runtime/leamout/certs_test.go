package leamout

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInstallAndValidateCertificates(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	cert, key := certificateFixture(t, now, "sip.example.test")
	root := t.TempDir()
	fullchainPath := filepath.Join(root, "source.crt")
	keyPath := filepath.Join(root, "source.key")
	caPath := filepath.Join(root, "carrier-ca.pem")
	for path, content := range map[string][]byte{fullchainPath: cert, keyPath: key, caPath: cert} {
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	certDir := filepath.Join(root, "installed")
	if err := installCertificates(certDir, fullchainPath, keyPath, caPath, "sip.example.test", now); err != nil {
		t.Fatal(err)
	}
	if err := validateCertificates(certDir, "sip.example.test", now); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(filepath.Join(certDir, "privkey.pem")); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("private key permissions = %v, %v", info, err)
	}
}

func TestCertificateValidationRejectsMismatchAndHostname(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	cert, _ := certificateFixture(t, now, "sip.example.test")
	_, otherKey := certificateFixture(t, now, "sip.example.test")
	if err := validateCertificateBytes(cert, otherKey, cert, "sip.example.test", now); err == nil {
		t.Fatal("mismatched private key accepted")
	}
	_, key := certificateFixture(t, now, "sip.example.test")
	if err := validateCertificateBytes(cert, key, cert, "other.example.test", now); err == nil {
		t.Fatal("wrong hostname accepted")
	}
}

func certificateFixture(t *testing.T, now time.Time, hostname string) ([]byte, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: hostname},
		DNSNames:     []string{hostname},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		IsCA:         true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	privateKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if bytes.Contains(privateKey, []byte(hostname)) {
		t.Fatal("unexpected hostname in private key fixture")
	}
	return cert, privateKey
}
