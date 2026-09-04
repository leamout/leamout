package managedvoice

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateCDR(t *testing.T) {
	valid := ProviderCDR{ProviderID: uuid.New(), CarrierConnectionID: uuid.New(), ProviderRecordID: "cdr-1", Direction: "termination", SIPCallID: "call-1", StartedAt: time.Now(), DurationSeconds: 30, Currency: "USD", CostMicros: 125000, Raw: json.RawMessage(`{"id":"cdr-1"}`)}
	if err := validateCDR(valid); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.Raw = json.RawMessage(`[]`)
	if err := validateCDR(invalid); err == nil {
		t.Fatal("array raw payload accepted")
	}
	invalid = valid
	invalid.CostMicros = -1
	if err := validateCDR(invalid); err == nil {
		t.Fatal("negative cost accepted")
	}
}
