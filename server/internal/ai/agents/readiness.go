package agents

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/ai/providers"
	"github.com/leamout/leamout/server/pkg/apperror"
)

func (s *Service) Readiness(
	ctx context.Context,
	organizationID uuid.UUID,
	agentID uuid.UUID,
) (ReadinessReport, error) {
	agent, err := s.Get(ctx, organizationID, agentID)
	if err != nil {
		return ReadinessReport{}, err
	}
	report := ReadinessReport{
		ConfigurationRevision: agent.ConfigurationRevision,
		Engine:                agent.Engine,
		Bindings:              []BindingDiagnostic{},
		Issues:                []ReadinessIssue{},
	}
	if s.providers == nil {
		report.Issues = append(report.Issues, readinessIssue(
			"provider_service_unavailable",
			"bindings",
			"AI provider readiness is unavailable.",
			"Restore the AI provider service before starting the agent.",
		))
		return report, nil
	}
	statuses, err := s.providers.BindingStatuses(ctx, organizationID, agentID)
	if err != nil {
		return ReadinessReport{}, err
	}
	byRole := make(map[string]providers.BindingStatus, len(statuses))
	for _, status := range statuses {
		byRole[status.Role] = status
		report.Bindings = append(report.Bindings, BindingDiagnostic{
			Role:             status.Role,
			Provider:         status.Provider,
			IntegrationID:   status.IntegrationID,
			ConnectionState: status.ConnectionState,
			FailureCode:      status.FailureCode,
			Config:           append([]byte(nil), status.Config...),
		})
	}
	expected := expectedRoles(agent.Engine)
	for _, status := range statuses {
		if _, ok := expected[status.Role]; !ok {
			report.Issues = append(report.Issues, readinessIssue(
				"unexpected_provider_binding",
				"bindings."+status.Role,
				fmt.Sprintf("The %s role is not valid for the selected engine.", status.Role),
				"Remove the binding or select the matching engine mode.",
			))
		}
	}
	for _, role := range []string{
		providers.RoleRealtime,
		providers.RoleSTT,
		providers.RoleLLM,
		providers.RoleTTS,
	} {
		if _, expectedRole := expected[role]; !expectedRole {
			continue
		}
		status, ok := byRole[role]
		if !ok {
			report.Issues = append(report.Issues, readinessIssue(
				"missing_provider_binding",
				"bindings."+role,
				fmt.Sprintf("The %s role requires a provider binding.", role),
				fmt.Sprintf("Connect a ready provider integration to the %s role.", role),
			))
			continue
		}
		if status.ConnectionState != providers.ConnectionReady {
			code := "integration_not_ready"
			field := "bindings." + role + ".integration_id"
			message := fmt.Sprintf(
				"The %s integration is not ready.",
				status.Provider,
			)
			remediation := "Verify or rotate the integration, then run readiness again."
			if status.IntegrationID == nil {
				code = "platform_credential_unavailable"
				field = "bindings." + role + ".provider"
				message = fmt.Sprintf(
					"The platform credential for %s is unavailable.",
					status.Provider,
				)
				remediation = "Configure the provider API key for this deployment or bind an organization integration."
			}
			report.Issues = append(report.Issues, readinessIssue(
				code,
				field,
				message,
				remediation,
			))
		}
	}
	report.Ready = len(report.Issues) == 0
	return report, nil
}

func expectedRoles(engine string) map[string]struct{} {
	if canonicalEngine(engine) == EngineRealtime {
		return map[string]struct{}{
			providers.RoleRealtime: {},
		}
	}
	return map[string]struct{}{
		providers.RoleSTT: {},
		providers.RoleLLM: {},
		providers.RoleTTS: {},
	}
}

func readinessIssue(
	code string,
	field string,
	message string,
	remediation string,
) ReadinessIssue {
	return ReadinessIssue{
		Code:        code,
		Field:       field,
		Message:     message,
		Remediation: remediation,
	}
}

func (s *Service) RequireReady(
	ctx context.Context,
	organizationID uuid.UUID,
	agentID uuid.UUID,
) error {
	report, err := s.Readiness(ctx, organizationID, agentID)
	if err != nil {
		return err
	}
	if report.Ready {
		return nil
	}

	if len(report.Issues) == 0 {
		return apperror.NewConflict("voice agent is not ready")
	}

	issue := report.Issues[0]
	message := fmt.Sprintf(
		"voice agent is not ready: %s: %s",
		issue.Code,
		issue.Message,
	)
	if issue.Code == "provider_service_unavailable" {
		return apperror.NewServiceUnavailable(message, nil)
	}
	return apperror.NewConflict(message)
}
