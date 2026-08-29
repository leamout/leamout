package calls

import (
	"testing"

	"github.com/google/uuid"
)

func TestValidateCreateRequest(t *testing.T) {
	t.Parallel()

	trunkID := uuid.New()
	tests := []struct {
		name string
		req  CreateCallRequest
		want bool
	}{
		{name: "valid request", req: CreateCallRequest{TrunkID: trunkID, From: "1000", To: "1001"}, want: true},
		{name: "missing originator", req: CreateCallRequest{TrunkID: trunkID, To: "1001"}},
		{name: "missing trunk", req: CreateCallRequest{From: "1000", To: "1001"}},
		{name: "nil application rejected", req: CreateCallRequest{ApplicationID: ptr(uuid.Nil), TrunkID: trunkID, From: "1000", To: "1001"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateCreateRequest(&tt.req); (err == nil) != tt.want {
				t.Fatalf("validateCreateRequest() error = %v, want success %t", err, tt.want)
			}
		})
	}
}

func TestValidateRecordRequest(t *testing.T) {
	t.Parallel()

	request := RecordRequest{Path: "/recordings/call.wav"}
	if err := validateRecordRequest(&request); err != nil {
		t.Fatalf("validateRecordRequest() error = %v", err)
	}
	if request.Action != "start" {
		t.Fatalf("action = %q, want start", request.Action)
	}
	if err := validateRecordRequest(&RecordRequest{Action: "stop"}); err != nil {
		t.Fatalf("stop recording should not require a path: %v", err)
	}
}

func ptr(value uuid.UUID) *uuid.UUID { return &value }
