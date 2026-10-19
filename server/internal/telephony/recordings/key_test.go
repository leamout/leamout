package recordings

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRecordingObjectKeyUsesOrganizationNamespace(t *testing.T) {
	organizationID := uuid.New()
	recordingID := uuid.New()
	key, err := recordingObjectKey(
		organizationID,
		recordingID,
		time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC),
		"wav",
	)
	if err != nil {
		t.Fatalf("recordingObjectKey() error = %v", err)
	}
	want := "organizations/" + organizationID.String() +
		"/recordings/2026/10/02/" + recordingID.String() + ".wav"
	if key != want {
		t.Fatalf("recordingObjectKey() = %q, want %q", key, want)
	}
}

func TestValidateRecordingObjectKeyRejectsOtherOrganization(t *testing.T) {
	organizationID := uuid.New()
	otherOrganizationID := uuid.New()
	key := "organizations/" + otherOrganizationID.String() + "/recordings/file.wav"

	err := validateRecordingObjectKey(organizationID, key)
	if err == nil || !strings.Contains(err.Error(), "outside organization prefix") {
		t.Fatalf("validateRecordingObjectKey() error = %v", err)
	}
}

func TestValidateRecordingObjectKeyRejectsTraversal(t *testing.T) {
	organizationID := uuid.New()
	key := "organizations/" + organizationID.String() + "/recordings/../other/file.wav"

	if err := validateRecordingObjectKey(organizationID, key); err == nil {
		t.Fatal("validateRecordingObjectKey() error = nil")
	}
}
