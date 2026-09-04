package numbers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/modules/audit"
)

type fakeNumberRepository struct {
	createdBYOC      BYOCCreateRequest
	createdManaged   ManagedCreateRequest
	getNumber        sqlc.PhoneNumber
	releaseCalls     int
	setCarrierCalls  int
}

func (f *fakeNumberRepository) CreateBYOC(_ context.Context, _ uuid.UUID, req BYOCCreateRequest) (sqlc.PhoneNumber, error) {
	f.createdBYOC = req
	return sqlc.PhoneNumber{Number: req.Number, CountryCode: req.CountryCode, ProvisioningMode: string(ProvisioningModeBYOC)}, nil
}

func (f *fakeNumberRepository) CreateManaged(_ context.Context, _ uuid.UUID, req ManagedCreateRequest) (sqlc.PhoneNumber, error) {
	f.createdManaged = req
	return sqlc.PhoneNumber{Number: req.Number, CountryCode: req.CountryCode, ProvisioningMode: string(ProvisioningModeManaged)}, nil
}

func (f *fakeNumberRepository) List(context.Context, uuid.UUID) ([]sqlc.PhoneNumber, error) {
	return nil, nil
}

func (f *fakeNumberRepository) Get(context.Context, uuid.UUID, uuid.UUID) (sqlc.PhoneNumber, error) {
	return f.getNumber, nil
}

func (f *fakeNumberRepository) Update(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ UpdateRequest) (sqlc.PhoneNumber, error) {
	return f.getNumber, nil
}

func (f *fakeNumberRepository) ReleaseBYOC(context.Context, uuid.UUID, uuid.UUID) (sqlc.PhoneNumber, error) {
	f.releaseCalls++
	return f.getNumber, nil
}

func (f *fakeNumberRepository) SetCarrierConnection(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, audit.Event) (sqlc.PhoneNumber, error) {
	f.setCarrierCalls++
	return f.getNumber, nil
}

func TestCreateBYOCNormalizesIdentity(t *testing.T) {
	repo := &fakeNumberRepository{}
	service := NewService(repo)
	orgID := uuid.New()

	_, err := service.CreateBYOC(context.Background(), orgID, BYOCCreateRequest{
		Number:      " +233201234567 ",
		CountryCode: " gh ",
	})
	if err != nil {
		t.Fatalf("CreateBYOC() error = %v", err)
	}
	if repo.createdBYOC.Number != "+233201234567" {
		t.Fatalf("number = %q", repo.createdBYOC.Number)
	}
	if repo.createdBYOC.CountryCode != "GH" {
		t.Fatalf("country_code = %q", repo.createdBYOC.CountryCode)
	}
}

func TestCreateManagedRequiresProviderMetadata(t *testing.T) {
	service := NewService(&fakeNumberRepository{})
	orgID := uuid.New()

	if _, err := service.CreateManaged(context.Background(), orgID, ManagedCreateRequest{
		Number:             "+233201234567",
		CountryCode:        "GH",
		ProviderResourceID: "did-123",
	}); err == nil {
		t.Fatal("CreateManaged accepted missing provider_id")
	}

	if _, err := service.CreateManaged(context.Background(), orgID, ManagedCreateRequest{
		Number:      "+233201234567",
		CountryCode: "GH",
		ProviderID:  uuid.New(),
	}); err == nil {
		t.Fatal("CreateManaged accepted missing provider_resource_id")
	}
}

func TestReleaseBYOCRejectsManagedNumber(t *testing.T) {
	repo := &fakeNumberRepository{getNumber: sqlc.PhoneNumber{ProvisioningMode: string(ProvisioningModeManaged)}}
	service := NewService(repo)

	if err := service.ReleaseBYOC(context.Background(), uuid.New(), uuid.New()); err == nil {
		t.Fatal("ReleaseBYOC accepted managed number")
	}
	if repo.releaseCalls != 0 {
		t.Fatalf("managed release reached repository %d times", repo.releaseCalls)
	}
}

func TestReleaseBYOCReleasesBYOCNumber(t *testing.T) {
	repo := &fakeNumberRepository{getNumber: sqlc.PhoneNumber{ProvisioningMode: string(ProvisioningModeBYOC)}}
	service := NewService(repo)

	if err := service.ReleaseBYOC(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("ReleaseBYOC() error = %v", err)
	}
	if repo.releaseCalls != 1 {
		t.Fatalf("release calls = %d", repo.releaseCalls)
	}
}

func TestSetCarrierConnectionRejectsManagedNumber(t *testing.T) {
	repo := &fakeNumberRepository{getNumber: sqlc.PhoneNumber{ProvisioningMode: string(ProvisioningModeManaged)}}
	service := NewService(repo)

	_, err := service.SetCarrierConnection(context.Background(), uuid.New(), uuid.New(), CarrierConnectionRequest{CarrierConnectionID: uuid.New()})
	if err == nil {
		t.Fatal("SetCarrierConnection accepted managed number")
	}
	if repo.setCarrierCalls != 0 {
		t.Fatalf("managed carrier assignment reached repository %d times", repo.setCarrierCalls)
	}
}

func TestResponseHidesManagedCarrierConnection(t *testing.T) {
	connectionID := uuid.New()
	managed := response(sqlc.PhoneNumber{
		ProvisioningMode:    string(ProvisioningModeManaged),
		CarrierConnectionID: &connectionID,
	})
	if managed.ProvisioningMode != ProvisioningModeManaged {
		t.Fatalf("provisioning_mode = %q", managed.ProvisioningMode)
	}
	if managed.CarrierConnectionID != nil {
		t.Fatal("managed response exposed platform carrier_connection_id")
	}

	byoc := response(sqlc.PhoneNumber{
		ProvisioningMode:    string(ProvisioningModeBYOC),
		CarrierConnectionID: &connectionID,
	})
	if byoc.ProvisioningMode != ProvisioningModeBYOC {
		t.Fatalf("provisioning_mode = %q", byoc.ProvisioningMode)
	}
	if byoc.CarrierConnectionID == nil || *byoc.CarrierConnectionID != connectionID {
		t.Fatal("BYOC response hid organization carrier_connection_id")
	}
}
