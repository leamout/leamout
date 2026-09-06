package numbers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/modules/audit"
)

type fakeNumberRepository struct {
	createdBYOC CreateRequest
	managedSelectionID string
	getNumber sqlc.PhoneNumber
	getForRelease sqlc.PhoneNumber
	releaseCalls int
	selectionOrgID uuid.UUID
	selectionCandidate ManagedNumberCandidate
	selectionID string
}

func (f *fakeNumberRepository) CreateBYOC(_ context.Context, _ uuid.UUID, req CreateRequest) (sqlc.PhoneNumber, error) {
	f.createdBYOC = req
	return sqlc.PhoneNumber{Number:req.Number, CountryCode:req.CountryCode, ProvisioningMode:string(ProvisioningModeBYOC), Status:"active"}, nil
}
func (f *fakeNumberRepository) CreateManaged(_ context.Context, _ uuid.UUID, selectionID string) (sqlc.PhoneNumber, error) {
	f.managedSelectionID = selectionID
	return sqlc.PhoneNumber{ProvisioningMode:string(ProvisioningModeManaged), Status:"provisioning"}, nil
}
func (f *fakeNumberRepository) List(context.Context, uuid.UUID) ([]sqlc.PhoneNumber,error){return nil,nil}
func (f *fakeNumberRepository) Get(context.Context,uuid.UUID,uuid.UUID)(sqlc.PhoneNumber,error){return f.getNumber,nil}
func (f *fakeNumberRepository) GetForRelease(context.Context,uuid.UUID,uuid.UUID)(sqlc.PhoneNumber,error){return f.getForRelease,nil}
func (f *fakeNumberRepository) Update(context.Context,uuid.UUID,uuid.UUID,UpdateRequest)(sqlc.PhoneNumber,error){return f.getNumber,nil}
func (f *fakeNumberRepository) ReleaseBYOC(context.Context,uuid.UUID,uuid.UUID)(sqlc.PhoneNumber,error){f.releaseCalls++; return f.getForRelease,nil}
func (f *fakeNumberRepository) SetCarrierConnection(context.Context,uuid.UUID,uuid.UUID,uuid.UUID,audit.Event)(sqlc.PhoneNumber,error){return f.getNumber,nil}
func (f *fakeNumberRepository) SaveManagedSelection(_ context.Context, organizationID uuid.UUID, candidate ManagedNumberCandidate)(string,error){
	f.selectionOrgID=organizationID; f.selectionCandidate=candidate; return f.selectionID,nil
}

type fakeManagedInventory struct { request AvailableSearchRequest; candidates []ManagedNumberCandidate }
func (f *fakeManagedInventory) SearchAvailable(_ context.Context, request AvailableSearchRequest)([]ManagedNumberCandidate,error){f.request=request;return f.candidates,nil}

func TestCreateDispatchesBYOC(t *testing.T){
	repo:=&fakeNumberRepository{}; service:=NewService(repo)
	_,err:=service.Create(context.Background(),uuid.New(),CreateRequest{Type:ProvisioningModeBYOC,Number:" +233201234567 ",CountryCode:" gh "})
	if err!=nil{t.Fatal(err)}
	if repo.createdBYOC.Number!="+233201234567"||repo.createdBYOC.CountryCode!="GH"{t.Fatalf("created=%+v",repo.createdBYOC)}
}

func TestCreateDispatchesManagedSelection(t *testing.T){
	repo:=&fakeNumberRepository{}; service:=NewService(repo)
	result,err:=service.Create(context.Background(),uuid.New(),CreateRequest{Type:ProvisioningModeManaged,SelectionID:" sel_test "})
	if err!=nil{t.Fatal(err)}
	if repo.managedSelectionID!="sel_test"{t.Fatalf("selection=%q",repo.managedSelectionID)}
	if result.Status!="provisioning"{t.Fatalf("status=%q",result.Status)}
}

func TestManagedCreateRejectsBYOCFields(t *testing.T){
	service:=NewService(&fakeNumberRepository{})
	_,err:=service.Create(context.Background(),uuid.New(),CreateRequest{Type:ProvisioningModeManaged,SelectionID:"sel_test",Number:"+233201234567"})
	if err==nil{t.Fatal("managed create accepted caller-supplied number")}
}

func TestSearchAvailableReturnsOpaqueCustomerResponse(t *testing.T){
	repo:=&fakeNumberRepository{selectionID:"sel_test"}
	inventory:=&fakeManagedInventory{candidates:[]ManagedNumberCandidate{{Provider:"didww",ProviderInventoryID:"available-1",ProviderProductID:"sku-1",Number:"+12125550100",CountryCode:"US",ChannelsIncludedCount:2}}}
	service:=NewService(repo); service.SetManagedAcquisition(inventory)
	result,err:=service.SearchAvailable(context.Background(),uuid.New(),AvailableSearchRequest{CountryCode:" us ",Contains:"+212"})
	if err!=nil{t.Fatal(err)}
	if len(result)!=1||result[0].SelectionID!="sel_test"{t.Fatalf("result=%+v",result)}
}

func TestReleaseRejectsManagedNumber(t *testing.T){
	repo:=&fakeNumberRepository{getForRelease:sqlc.PhoneNumber{ProvisioningMode:string(ProvisioningModeManaged),Status:"active"}}
	service:=NewService(repo)
	if err:=service.Release(context.Background(),uuid.New(),uuid.New());err==nil{t.Fatal("release accepted managed number")}
	if repo.releaseCalls!=0{t.Fatalf("release calls=%d",repo.releaseCalls)}
}

func TestResponseUsesTypeAndHidesManagedCarrier(t *testing.T){
	connectionID:=uuid.New()
	managed:=response(sqlc.PhoneNumber{ProvisioningMode:string(ProvisioningModeManaged),CarrierConnectionID:&connectionID})
	if managed.Type!=ProvisioningModeManaged{t.Fatalf("type=%q",managed.Type)}
	if managed.CarrierConnectionID!=nil{t.Fatal("managed response exposed platform carrier")}
	byoc:=response(sqlc.PhoneNumber{ProvisioningMode:string(ProvisioningModeBYOC),CarrierConnectionID:&connectionID})
	if byoc.Type!=ProvisioningModeBYOC||byoc.CarrierConnectionID==nil{t.Fatalf("response=%+v",byoc)}
}
