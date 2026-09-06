package trunks

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestCreateManagedTrunkFailsClosedUntilProvisioningExists(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(context.Background(), uuid.New(), CreateRequest{
		Type: ProvisioningModeManaged,
		Name: "Leamout managed",
	})
	if err == nil {
		t.Fatal("managed trunk creation succeeded without a managed provisioning boundary")
	}
}

func TestCreateManagedTrunkRejectsCarrierConnection(t *testing.T) {
	service := NewService(nil)
	connectionID := uuid.New()

	_, err := service.Create(context.Background(), uuid.New(), CreateRequest{
		Type:                ProvisioningModeManaged,
		CarrierConnectionID: &connectionID,
		Name:                "Leamout managed",
	})
	if err == nil {
		t.Fatal("managed trunk accepted a customer carrier_connection_id")
	}
}
