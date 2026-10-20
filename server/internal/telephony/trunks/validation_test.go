package trunks

import "testing"

func TestNormalizeHostEnforcesPublicSIPDestinations(t *testing.T) {
	tests := []struct {
		name   string
		host   string
		want   string
		reject bool
	}{
		{name: "public IPv4", host: "1.1.1.1", want: "1.1.1.1"},
		{name: "public IPv6", host: "2606:4700:4700::1111", want: "2606:4700:4700::1111"},
		{name: "public hostname", host: "SIP.Example.COM.", want: "sip.example.com"},
		{name: "loopback", host: "127.0.0.1", reject: true},
		{name: "private IPv4", host: "10.0.0.5", reject: true},
		{name: "carrier grade NAT", host: "100.64.1.2", reject: true},
		{name: "metadata IP", host: "169.254.169.254", reject: true},
		{name: "IPv6 private", host: "fd00::1", reject: true},
		{name: "mapped private IPv4", host: "::ffff:127.0.0.1", reject: true},
		{name: "localhost", host: "localhost", reject: true},
		{name: "local DNS", host: "gateway.internal", reject: true},
		{name: "invalid syntax", host: "bad host", reject: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeHost(tt.host)
			if tt.reject {
				if err == nil {
					t.Fatalf("normalizeHost(%q) = %q, expected rejection", tt.host, got)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("normalizeHost(%q) = %q, %v; want %q", tt.host, got, err, tt.want)
			}
		})
	}
}
