package edge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/telecom/routing"
)

func TestHandlerReturnsForbiddenWhenRouteDoesNotAuthorize(t *testing.T) {
	service := NewService(&fakeStore{resolveErr: pgx.ErrNoRows}, &fakeState{})
	handler := NewHandler(service, nil, "edge-secret")
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/v1/sip-edge/authorize", strings.NewReader(`{
		"username":"managed-trunk",
		"realm":"sip.leamout.com",
		"from":"+14155550100",
		"to":"+14155550101"
	}`))
	request.Header.Set("Authorization", "Bearer edge-secret")
	recorder := httptest.NewRecorder()

	handler.Admit(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	var response struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Allowed {
		t.Fatal("denied response was marked allowed")
	}
}

type fakeInboundRouting struct {
	decision routing.ManagedInboundDeliveryDecision
	err      error
}

func (f fakeInboundRouting) ResolveManagedInboundDelivery(context.Context, routing.InboundRequest) (routing.ManagedInboundDeliveryDecision, error) {
	return f.decision, f.err
}

func TestHandlerResolvesSelfHostedManagedInboundRoute(t *testing.T) {
	organizationID := uuid.New()
	attachmentID := uuid.New()
	handler := NewHandler(nil, fakeInboundRouting{decision: routing.ManagedInboundDeliveryDecision{
		OrganizationID: organizationID, CarrierConnectionID: uuid.New(), PhoneNumberID: uuid.New(),
		CalledNumber: "+15551235001", RuntimeAttachmentID: attachmentID,
		DeploymentID: uuid.New(), DeploymentIdentity: "runtime-1",
		IngressHost: "runtime.example", IngressPort: 5061, IngressTransport: "tls",
	}}, "edge-secret")
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/internal/v1/sip-edge/resolve-inbound", strings.NewReader(`{
		"source_ip":"203.0.113.10","called_number":"+15551235001","caller_number":"+15557654321"
	}`))
	request.Header.Set("Authorization", "Bearer edge-secret")
	recorder := httptest.NewRecorder()

	handler.ResolveInbound(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response inboundResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if !response.Allowed || response.OrganizationID != organizationID || response.RuntimeAttachmentID != attachmentID {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.RouteURI != "sip:+15551235001@runtime.example:5061;transport=tls" {
		t.Fatalf("route URI = %q", response.RouteURI)
	}
}

func TestHandlerManagedInboundFailsClosedWithoutAttachment(t *testing.T) {
	handler := NewHandler(nil, fakeInboundRouting{err: routing.ErrNoRoute}, "edge-secret")
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/internal/v1/sip-edge/resolve-inbound", strings.NewReader(`{
		"source_ip":"203.0.113.10","called_number":"+15551235001","caller_number":"+15557654321"
	}`))
	request.Header.Set("Authorization", "Bearer edge-secret")
	recorder := httptest.NewRecorder()

	handler.ResolveInbound(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
