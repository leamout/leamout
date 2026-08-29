package calls

import "testing"

func TestFreeSWITCHEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     OriginateRequest
		want    string
		wantErr bool
	}{
		{
			name: "udp route",
			req: OriginateRequest{
				Host:        "sip.carrier.example",
				Port:        5060,
				Transport:   "udp",
				Destination: "+14155550100",
			},
			want: "sofia/external/+14155550100@sip.carrier.example:5060;transport=udp",
		},
		{
			name: "ipv6 route",
			req: OriginateRequest{
				Host:        "2001:db8::10",
				Port:        5061,
				Transport:   "TLS",
				Destination: "+233200000000",
			},
			want: "sofia/external/+233200000000@[2001:db8::10]:5061;transport=tls",
		},
		{
			name: "invalid transport",
			req: OriginateRequest{
				Host:        "sip.carrier.example",
				Port:        5060,
				Transport:   "ws",
				Destination: "+14155550100",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := freeSWITCHEndpoint(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("freeSWITCHEndpoint() error = %v, wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("freeSWITCHEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}
