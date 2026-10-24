package calling

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"net"
	"strings"
	"testing"
)

func TestFreeSWITCHEgress(t *testing.T) {
	endpoint, routeURI, err := freeSWITCHEgress(OriginateRequest{
		Destination: "+14155550100",
		CallerID:    "+14155550199",
		Host:        "carrier.example.com",
		Port:        5061,
		Transport:   "tls",
	})
	if err != nil {
		t.Fatal(err)
	}
	if endpoint != "sofia/internal/+14155550100@opensips:5060;transport=udp" {
		t.Fatalf("endpoint = %q", endpoint)
	}
	if routeURI != "sip:carrier.example.com:5061;transport=tls" {
		t.Fatalf("route URI = %q", routeURI)
	}
}

func TestFreeSWITCHEgressRejectsUnsafeHost(t *testing.T) {
	_, _, err := freeSWITCHEgress(OriginateRequest{
		Destination: "+14155550100",
		Host:        "carrier.example.com\ninvalid",
		Port:        5060,
		Transport:   "udp",
	})
	if err == nil {
		t.Fatal("expected invalid host error")
	}
}

func TestEgressVariablesCarryLeamoutIdentity(t *testing.T) {
	callID := uuid.New()
	trunkID := uuid.New()

	variables, err := egressVariables(OriginateRequest{
		CallID:             callID,
		TrunkID:            trunkID,
		Port:               5061,
		Transport:          "tls",
		PinnedIP:           "1.1.1.1",
		Privacy:            true,
		DTMFMode:           "rfc2833",
		MediaEncryption:    "sdes_srtp",
		MaxDurationSeconds: 60,
	}, "sip:carrier.example.com:5061;transport=tls")
	if err != nil {
		t.Fatal(err)
	}

	if variables[leamoutCallIDVar] != callID.String() {
		t.Fatalf("call id variable = %q", variables[leamoutCallIDVar])
	}
	if variables[pinnedRouteHeaderVar] != "sip:1.1.1.1:5061;transport=tls" {
		t.Fatalf("pinned route = %q", variables[pinnedRouteHeaderVar])
	}
	if variables[trunkHeaderVar] != trunkID.String() {
		t.Fatalf("trunk id variable = %q", variables[trunkHeaderVar])
	}
	if !strings.Contains(variables[routeURIHeaderVar], "carrier.example.com") {
		t.Fatalf("route URI variable = %q", variables[routeURIHeaderVar])
	}
	if variables["execute_on_answer"] != "sched_hangup +60 ALLOTTED_TIMEOUT" {
		t.Fatalf(
			"execute_on_answer variable = %q",
			variables["execute_on_answer"],
		)
	}
}

func TestEgressVariablesRejectHostnamesAsPinnedIP(t *testing.T) {
	_, err := egressVariables(
		OriginateRequest{CallID: uuid.New(), TrunkID: uuid.New(), PinnedIP: "carrier.example.com", Port: 5060, Transport: "udp"},
		"sip:carrier.example.com:5060;transport=udp",
	)
	if err == nil {
		t.Fatal("expected pinned egress IP literal validation")
	}
}

func TestEgressVariablesRejectNegativeMaximumDuration(t *testing.T) {
	_, err := egressVariables(
		OriginateRequest{
			CallID:             uuid.New(),
			TrunkID:            uuid.New(),
			MaxDurationSeconds: -1,
		},
		"sip:carrier.example.com:5061;transport=tls",
	)
	if err == nil {
		t.Fatal("expected maximum duration validation error")
	}
}
func TestClassifyClientOriginateError(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		class OriginateFailureClass
	}{
		{name: "deadline", err: context.DeadlineExceeded, class: OriginateFailureTimeout},
		{name: "transport", err: &net.OpError{Op: "write", Net: "tcp", Err: errors.New("reset")}, class: OriginateFailureTransport},
		{name: "canceled", err: context.Canceled, class: OriginateFailureInternal},
		{name: "internal", err: errors.New("rejected"), class: OriginateFailureInternal},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got *OriginateError
			if !errors.As(classifyClientOriginateError(test.err), &got) || got.Class != test.class {
				t.Fatalf("classifyClientOriginateError() = %+v, want %s", got, test.class)
			}
		})
	}
}
