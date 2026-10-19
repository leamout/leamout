package s3

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/url"
	"strings"
	"testing"
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		host   string
		secure bool
		ok     bool
	}{
		{name: "https", value: "https://s3.example.com", host: "s3.example.com", secure: true, ok: true},
		{name: "http", value: "http://minio:9000", host: "minio:9000", secure: false, ok: true},
		{name: "path", value: "https://s3.example.com/path", ok: false},
		{name: "credentials", value: "https://user:secret@s3.example.com", ok: false},
		{name: "query", value: "https://s3.example.com?x=1", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, secure, err := parseEndpoint(tt.value)
			if !tt.ok {
				if err == nil {
					t.Fatal("parseEndpoint() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseEndpoint() error = %v", err)
			}
			if host != tt.host || secure != tt.secure {
				t.Fatalf("parseEndpoint() = (%q, %v), want (%q, %v)", host, secure, tt.host, tt.secure)
			}
		})
	}
}

func TestPublicEndpointIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{ip: "8.8.8.8", want: true},
		{ip: "2606:4700:4700::1111", want: true},
		{ip: "127.0.0.1", want: false},
		{ip: "10.0.0.1", want: false},
		{ip: "169.254.169.254", want: false},
		{ip: "100.64.0.1", want: false},
		{ip: "198.18.0.1", want: false},
		{ip: "::1", want: false},
		{ip: "fc00::1", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			if got := publicEndpointIP(net.ParseIP(tt.ip)); got != tt.want {
				t.Fatalf("publicEndpointIP(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestValidateKey(t *testing.T) {
	for _, key := range []string{"", "/recording.wav", "../recording.wav", "recordings/../recording.wav"} {
		if err := validateKey(key); err == nil {
			t.Fatalf("validateKey(%q) error = nil, want error", key)
		}
	}
	if err := validateKey("recordings/org/recording.wav"); err != nil {
		t.Fatalf("validateKey() error = %v", err)
	}
}

func TestNewRequiresHTTPS(t *testing.T) {
	for _, endpoint := range []string{
		"http://s3.example.com",
		"http://127.0.0.1:9000",
	} {
		if _, err := New(context.Background(), Config{
			Endpoint: endpoint,
			Bucket: "recordings",
			AccessKey: "test-key",
			SecretKey: "test-secret",
		}); err == nil {
			t.Fatalf("insecure endpoint %q accepted", endpoint)
		}
	}
}

func TestNewUsesAWSClientForPresigning(t *testing.T) {
	client, err := New(context.Background(), Config{
		Endpoint: "https://storage.example.com",
		Region: "us-east-1",
		Bucket: "recordings",
		AccessKey: "test-key",
		SecretKey: "test-secret",
		UsePathStyle: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	value, expires, err := client.PlaybackURL(
		context.Background(), "organizations/tenant/recordings/test.wav",
	)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Scheme != "https" || parsed.Host != "storage.example.com" ||
		!strings.HasPrefix(parsed.Path, "/recordings/organizations/") {
		t.Fatalf("presigned URL does not use customer's path-style endpoint: %s", value)
	}
	if parsed.Query().Get("X-Amz-Signature") == "" ||
		parsed.Query().Get("X-Amz-Expires") != "900" {
		t.Fatalf("missing SigV4 signature or TTL in playback URL: %s", value)
	}
	if expires.IsZero() {
		t.Fatal("presigned URL expiration missing")
	}
}

func TestClientPutRejectsNonSeekableReader(t *testing.T) {
	client, err := New(context.Background(), Config{
		Endpoint: "https://storage.example.com",
		Bucket: "recordings",
		AccessKey: "test-key",
		SecretKey: "test-secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, ok := interface{}(bytes.NewBufferString("recording")).(io.ReadSeeker)
	if ok {
		t.Fatal("test reader unexpectedly seekable")
	}
	if err := client.Put(context.Background(), "recording.wav", "audio/wav",
		bytes.NewBufferString("recording"), 9); err == nil {
		t.Fatal("non-seekable upload accepted")
	}
}

func TestS3TransportNeverUsesProxy(t *testing.T) {
	if publicTransport().Proxy != nil {
		t.Fatal("customer S3 transport must not use environment proxies")
	}
}
