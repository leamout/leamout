package deployments

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateActivate(t *testing.T) {
	name := "primary"
	valid := ActivateRequest{
		OrganizationID: uuid.New(), LicenseID: uuid.New(), DeploymentID: "customer-prod",
		Name: &name, At: time.Now(),
	}
	if err := ValidateActivate(valid); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		mutate func(*ActivateRequest)
		want   error
	}{
		{"organization", func(value *ActivateRequest) { value.OrganizationID = uuid.Nil }, ErrOrganizationRequired},
		{"license", func(value *ActivateRequest) { value.LicenseID = uuid.Nil }, ErrLicenseRequired},
		{"deployment", func(value *ActivateRequest) { value.DeploymentID = " " }, ErrDeploymentIDRequired},
		{"name", func(value *ActivateRequest) { blank := " "; value.Name = &blank }, ErrInvalidDeploymentName},
		{"time", func(value *ActivateRequest) { value.At = time.Time{} }, ErrActivationTimeRequired},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := valid
			test.mutate(&request)
			if err := ValidateActivate(request); !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}
