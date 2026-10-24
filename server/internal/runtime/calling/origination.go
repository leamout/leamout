package calling

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"unicode"
)

const (
	openSIPSEgressHost       = "opensips"
	openSIPSEgressPort       = 5060
	leamoutCallIDVar         = "leamout_call_id"
	routeURIHeaderVar        = "sip_h_X-Leamout-Route-URI"
	pinnedRouteHeaderVar     = "sip_h_X-Leamout-Pinned-Route-URI"
	trunkHeaderVar           = "sip_h_X-Leamout-Trunk-ID"
	privacyHeaderVar         = "sip_h_X-Leamout-Privacy"
	dtmfTypeVar              = "dtmf_type"
	mediaEncryptionHeaderVar = "sip_h_X-Leamout-Media-Encryption"
)

func (c *Controller) Originate(
	ctx context.Context,
	req OriginateRequest,
) (OriginateResult, error) {
	endpoint, routeURI, err := freeSWITCHEgress(req)
	if err != nil {
		return OriginateResult{}, &OriginateError{Class: OriginateFailureValidation, Err: err}
	}

	variables, err := egressVariables(req, routeURI)
	if err != nil {
		return OriginateResult{}, &OriginateError{Class: OriginateFailureValidation, Err: err}
	}

	call, err := c.client.Originate(ctx, freeswitch.OriginateRequest{
		Endpoint:    endpoint,
		Destination: req.Destination,
		CallerID:    req.CallerID,
		Variables:   variables,
	})
	if err != nil {
		if errors.Is(err, freeswitch.ErrOriginateUncertain) {
			return OriginateResult{ChannelID: call.UUID}, &OriginateError{
				Class: OriginateFailureUncertain, Err: err,
			}
		}
		return OriginateResult{}, classifyClientOriginateError(err)
	}
	if strings.TrimSpace(call.UUID) == "" {
		return OriginateResult{}, fmt.Errorf("FreeSWITCH returned empty channel UUID")
	}

	return OriginateResult{ChannelID: call.UUID}, nil
}

func classifyClientOriginateError(err error) error {
	class := OriginateFailureInternal
	if errors.Is(err, context.DeadlineExceeded) {
		class = OriginateFailureTimeout
	} else if errors.Is(err, context.Canceled) {
		class = OriginateFailureInternal
	} else {
		var networkError net.Error
		if errors.As(err, &networkError) {
			if networkError.Timeout() {
				class = OriginateFailureTimeout
			} else {
				class = OriginateFailureTransport
			}
		}
	}
	return &OriginateError{Class: class, Err: fmt.Errorf("originate call: %w", err)}
}

func egressVariables(req OriginateRequest, routeURI string) (map[string]string, error) {
	if req.CallID == uuid.Nil {
		return nil, fmt.Errorf("call id is required")
	}
	if req.TrunkID == uuid.Nil {
		return nil, fmt.Errorf("resolved trunk id is required")
	}

	if req.MaxDurationSeconds < 0 {
		return nil, fmt.Errorf("maximum duration cannot be negative")
	}

	pinnedIP, err := netip.ParseAddr(req.PinnedIP)
	if err != nil || pinnedIP.Zone() != "" {
		return nil, fmt.Errorf("pinned SIP egress address must be an IP literal")
	}
	if req.Port == 0 {
		return nil, fmt.Errorf("pinned SIP egress port is required")
	}
	transport := strings.ToLower(strings.TrimSpace(req.Transport))
	if transport != "udp" && transport != "tcp" && transport != "tls" {
		return nil, fmt.Errorf("pinned SIP egress transport is invalid")
	}
	target := net.JoinHostPort(pinnedIP.Unmap().String(), strconv.Itoa(int(req.Port)))
	pinnedURI := fmt.Sprintf("sip:%s;transport=%s", target, transport)

	variables := map[string]string{
		leamoutCallIDVar:     req.CallID.String(),
		routeURIHeaderVar:    routeURI,
		pinnedRouteHeaderVar: pinnedURI,
		trunkHeaderVar:       req.TrunkID.String(),
	}

	if req.MaxDurationSeconds > 0 {
		variables["execute_on_answer"] = fmt.Sprintf(
			"sched_hangup +%d ALLOTTED_TIMEOUT",
			req.MaxDurationSeconds,
		)
	}

	if req.Privacy {
		variables[privacyHeaderVar] = "id"
	}

	if mode := strings.ToLower(strings.TrimSpace(req.DTMFMode)); mode != "" {
		switch mode {
		case "rfc2833", "info", "none":
			variables[dtmfTypeVar] = mode
		default:
			return nil, fmt.Errorf("DTMF mode is invalid: %q", req.DTMFMode)
		}
	}

	if encryption := strings.ToLower(strings.TrimSpace(req.MediaEncryption)); encryption != "" {
		switch encryption {
		case "none", "sdes_srtp":
			variables[mediaEncryptionHeaderVar] = encryption
		default:
			return nil, fmt.Errorf("media encryption is invalid: %q", req.MediaEncryption)
		}
	}

	return variables, nil
}

func freeSWITCHEgress(req OriginateRequest) (string, string, error) {
	host := strings.TrimSpace(req.Host)
	if host == "" {
		return "", "", fmt.Errorf("resolved route host is required")
	}
	if strings.ContainsAny(host, " \t\r\n,{}[]") {
		return "", "", fmt.Errorf("resolved route host is invalid")
	}
	if req.Port == 0 {
		return "", "", fmt.Errorf("resolved route port is required")
	}

	transport := strings.ToLower(strings.TrimSpace(req.Transport))
	switch transport {
	case "udp", "tcp", "tls":
	default:
		return "", "", fmt.Errorf("resolved route transport is invalid: %q", req.Transport)
	}

	destination := strings.TrimSpace(req.Destination)
	if !validPSTNAddress(destination) {
		return "", "", fmt.Errorf("resolved route destination is invalid")
	}

	callerID := strings.TrimSpace(req.CallerID)
	if callerID != "" && !validPSTNAddress(callerID) {
		return "", "", fmt.Errorf("caller id is invalid")
	}

	trunkTarget := net.JoinHostPort(host, strconv.Itoa(int(req.Port)))
	routeURI := fmt.Sprintf("sip:%s;transport=%s", trunkTarget, transport)

	openSIPSTarget := net.JoinHostPort(openSIPSEgressHost, strconv.Itoa(openSIPSEgressPort))
	endpoint := fmt.Sprintf("sofia/internal/%s@%s;transport=udp", destination, openSIPSTarget)

	return endpoint, routeURI, nil
}

func validPSTNAddress(value string) bool {
	if value == "" || len(value) > 32 {
		return false
	}
	for i, r := range value {
		if i == 0 && r == '+' {
			continue
		}
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return value != "+"
}
