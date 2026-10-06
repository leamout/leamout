package email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/coffeyvidzro/monogo/internal/security/encryption"
	"github.com/google/uuid"
)

type fakeRepo struct {
	d         Delivery
	completed bool
	failed    string
	permanent bool
	notReady  bool
}

func (r *fakeRepo) Claim(context.Context, uuid.UUID) (Delivery, error) { return r.d, nil }
func (r *fakeRepo) Complete(context.Context, uuid.UUID, uuid.UUID, Result) error {
	r.completed = true
	return nil
}
func (r *fakeRepo) Fail(_ context.Context, _ Delivery, _ uuid.UUID, code string, p bool) error {
	r.failed = code
	r.permanent = p
	return nil
}

type fakeSender struct {
	message Message
	calls   int
	err     error
}

func (s *fakeSender) Send(_ context.Context, m Message) (Result, error) {
	s.message = m
	s.calls++
	return Result{MessageID: "id"}, s.err
}
func fixture(t *testing.T) (*fakeRepo, *fakeSender, *encryption.Cipher) {
	t.Helper()
	cipher, err := encryption.New(base64.RawURLEncoding.EncodeToString(make([]byte, 32)))
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	expiry := time.Now().Add(time.Minute)
	data, _ := json.Marshal(Data{Code: "012345", ExpiresAt: expiry})
	payload, err := cipher.EncryptForScope("email:"+id.String(), string(data))
	if err != nil {
		t.Fatal(err)
	}
	return &fakeRepo{d: Delivery{ID: id, To: "user@example.com", Template: "otp", Payload: payload, ExpiresAt: expiry}}, &fakeSender{}, cipher
}
func TestWorkerSendsEncryptedPayload(t *testing.T) {
	r, s, c := fixture(t)
	if strings.Contains(r.d.Payload, "012345") {
		t.Fatal("plaintext code")
	}
	if err := NewWorker(r, s, c).ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !r.completed || s.calls != 1 || s.message.To != "user@example.com" || !strings.Contains(s.message.Text, "012345") {
		t.Fatal("delivery not completed")
	}
}
func TestWorkerRejectsExpiredAndTamperedPayload(t *testing.T) {
	for _, expired := range []bool{true, false} {
		r, s, c := fixture(t)
		if expired {
			r.d.ExpiresAt = time.Now().Add(-time.Minute)
		} else {
			r.d.ID = uuid.New()
		}
		if err := NewWorker(r, s, c).ProcessOne(context.Background()); err != nil {
			t.Fatal(err)
		}
		if s.calls != 0 || !r.permanent {
			t.Fatal("unsafe payload sent")
		}
	}
}
func TestWorkerRetriesWithoutPersistingSecrets(t *testing.T) {
	r, s, c := fixture(t)
	s.err = errors.New("secret@example.com 012345")
	if err := NewWorker(r, s, c).ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if r.completed || r.permanent || r.failed != "send_failed" {
		t.Fatal("incorrect retry")
	}
	s.err = &SendError{Code: "MessageRejected", Permanent: true}
	if err := NewWorker(r, s, c).ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !r.permanent || r.failed != "MessageRejected" {
		t.Fatal("permanent failure retried")
	}
}

func (r *fakeRepo) Ready(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return !r.notReady, nil
}

func TestWorkerSkipsSupersededDelivery(t *testing.T) {
	r, s, c := fixture(t)
	r.notReady = true
	if err := NewWorker(r, s, c).ProcessOne(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.calls != 0 || !r.permanent {
		t.Fatal("superseded delivery was sent")
	}
}
