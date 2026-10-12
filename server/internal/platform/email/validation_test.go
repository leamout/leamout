package email

import "testing"

func TestValidURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{url: "https://console.example.com/forgot-password", want: true},
		{url: "http://localhost:3000/invitations/accept", want: true},
		{url: "http://127.0.0.1:3000/forgot-password", want: true},
		{url: "http://[::1]:3000/forgot-password", want: true},
		{url: "http://console.example.com/forgot-password", want: false},
		{url: "http://localhost.example.com", want: false},
		{url: "http://localhost@evil.example.com", want: false},
		{url: "https://user:password@console.example.com", want: false},
		{url: "javascript:alert(1)", want: false},
		{url: "//console.example.com", want: false},
		{url: "https://:3000", want: false},
	}

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			if got := ValidURL(test.url); got != test.want {
				t.Fatalf("ValidURL(%q) = %v, want %v", test.url, got, test.want)
			}
		})
	}
}
