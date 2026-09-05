package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/integrations/carriers/didww"
	"github.com/leamout/leamout/internal/platform/config"
)

const (
	didwwProviderSlug           = "didww"
	didwwConnectionName        = "DIDWW Managed Ingress"
	didwwTrunkName             = "DIDWW Managed Ingress"
	didwwProviderResourceType  = "voice_in_trunk"
)

type sipEndpoint struct {
	Host      string
	Port      int32
	Transport string
}

func DIDWWIngress(ctx context.Context, cfg config.Config) error {
	if strings.TrimSpace(cfg.DIDWW.APIKey) == "" {
		return nil
	}
	if strings.TrimSpace(cfg.DeploymentID) == "" {
		return fmt.Errorf("DIDWW ingress bootstrap requires LEAMOUT_DEPLOYMENT_ID")
	}
	if strings.TrimSpace(cfg.SIP.PublicHost) == "" {
		return fmt.Errorf("DIDWW ingress bootstrap requires SIP_PUBLIC_HOST")
	}
	if cfg.SIP.PublicPort <= 0 || cfg.SIP.PublicPort > 65535 {
		return fmt.Errorf("SIP_PUBLIC_PORT must be between 1 and 65535")
	}
	if _, err := didwwTransport(cfg.SIP.PublicTransport); err != nil {
		return err
	}

	sourceCIDRs, err := parseSourceCIDRs(cfg.DIDWW.SourceCIDRs)
	if err != nil {
		return err
	}
	if len(sourceCIDRs) == 0 {
		return fmt.Errorf("DIDWW ingress bootstrap requires DIDWW_SOURCE_CIDRS")
	}
	endpoints, err := parseSIPEndpoints(cfg.DIDWW.SIPEndpoints)
	if err != nil {
		return err
	}

	client, err := didww.NewClient(didww.Config{
		BaseURL: cfg.DIDWW.APIBaseURL,
		APIKey:  cfg.DIDWW.APIKey,
	})
	if err != nil {
		return err
	}
	providerTrunk, err := client.EnsureInboundTrunk(ctx, didww.EnsureInboundTrunkRequest{
		Name:                didwwTrunkName,
		ExternalReferenceID: didwwExternalReference(cfg.DeploymentID),
		Host:                cfg.SIP.PublicHost,
		Port:                cfg.SIP.PublicPort,
		Transport:           cfg.SIP.PublicTransport,
	})
	if err != nil {
		return fmt.Errorf("ensure DIDWW Voice IN trunk: %w", err)
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect bootstrap database: %w", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping bootstrap database: %w", err)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin DIDWW ingress bootstrap: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	providerID, err := activeProviderID(ctx, tx, didwwProviderSlug)
	if err != nil {
		return err
	}
	connectionID, err := ensurePlatformConnection(ctx, tx, providerID)
	if err != nil {
		return err
	}
	trunkID, err := ensurePlatformInboundTrunk(ctx, tx, connectionID)
	if err != nil {
		return err
	}
	if err := replaceSourceCIDRs(ctx, tx, connectionID, sourceCIDRs); err != nil {
		return err
	}
	if err := replaceInboundEndpoints(ctx, tx, trunkID, endpoints); err != nil {
		return err
	}
	if err := attachProviderResource(ctx, tx, connectionID, providerID, providerTrunk.ID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit DIDWW ingress bootstrap: %w", err)
	}
	return nil
}

func didwwExternalReference(deploymentID string) string {
	value := "leamout:" + strings.TrimSpace(deploymentID) + ":managed-ingress"
	if len(value) <= 100 {
		return value
	}
	return value[:100]
}

func activeProviderID(ctx context.Context, tx pgx.Tx, slug string) (uuid.UUID, error) {
	var id uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT id FROM carrier_providers WHERE slug=$1 AND status='active'`, slug).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("resolve active %s carrier provider: %w", slug, err)
	}
	return id, nil
}

func ensurePlatformConnection(ctx context.Context, tx pgx.Tx, providerID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
INSERT INTO carrier_connections (
    organization_id, provider_id, scope, name, status,
    outbound_auth_method, inbound_enabled, inbound_auth_method,
    max_cps, max_concurrent_calls, codecs, supports_video, supports_fax
)
VALUES (
    NULL, $1, 'platform', $2, 'active',
    'none', true, 'ip',
    100, 1000, ARRAY['PCMU','PCMA']::TEXT[], false, false
)
ON CONFLICT (name) WHERE scope = 'platform'
DO UPDATE SET
    status = 'active',
    inbound_enabled = true,
    inbound_auth_method = 'ip',
    inbound_username = NULL,
    inbound_secret_ciphertext = NULL,
    updated_at = now()
WHERE carrier_connections.provider_id = EXCLUDED.provider_id
RETURNING id`, providerID, didwwConnectionName).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ensure DIDWW platform carrier connection: %w", err)
	}
	return id, nil
}

func ensurePlatformInboundTrunk(ctx context.Context, tx pgx.Tx, connectionID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO trunks (organization_id, carrier_connection_id, name, direction, status, managed_default)
VALUES (NULL, $1, $2, 'inbound', 'active', false)
ON CONFLICT (name) WHERE organization_id IS NULL
DO UPDATE SET
    carrier_connection_id = EXCLUDED.carrier_connection_id,
    direction = 'inbound',
    status = 'active',
    managed_default = false,
    updated_at = now()
RETURNING id`, connectionID, didwwTrunkName).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("ensure DIDWW platform inbound trunk: %w", err)
	}
	return id, nil
}

func replaceSourceCIDRs(ctx context.Context, tx pgx.Tx, connectionID uuid.UUID, cidrs []netip.Prefix) error {
	if _, err := tx.Exec(ctx, `DELETE FROM carrier_connection_source_ips WHERE carrier_connection_id=$1`, connectionID); err != nil {
		return fmt.Errorf("clear DIDWW ingress source CIDRs: %w", err)
	}
	for _, cidr := range cidrs {
		if _, err := tx.Exec(ctx, `INSERT INTO carrier_connection_source_ips (organization_id, carrier_connection_id, cidr) VALUES (NULL,$1,$2)`, connectionID, cidr); err != nil {
			return fmt.Errorf("store DIDWW ingress source CIDR %s: %w", cidr, err)
		}
	}
	return nil
}

func replaceInboundEndpoints(ctx context.Context, tx pgx.Tx, trunkID uuid.UUID, endpoints []sipEndpoint) error {
	if _, err := tx.Exec(ctx, `DELETE FROM trunk_endpoints WHERE trunk_id=$1 AND direction='inbound'`, trunkID); err != nil {
		return fmt.Errorf("clear DIDWW inbound trunk endpoints: %w", err)
	}
	for _, endpoint := range endpoints {
		if _, err := tx.Exec(ctx, `
INSERT INTO trunk_endpoints (
    organization_id, trunk_id, host, port, transport, direction, priority, weight, enabled
) VALUES (NULL,$1,$2,$3,$4,'inbound',10,100,true)`, trunkID, endpoint.Host, endpoint.Port, endpoint.Transport); err != nil {
			return fmt.Errorf("store DIDWW inbound SIP endpoint %s: %w", endpoint.Host, err)
		}
	}
	return nil
}

func attachProviderResource(ctx context.Context, tx pgx.Tx, connectionID, providerID uuid.UUID, resourceID string) error {
	if strings.TrimSpace(resourceID) == "" {
		return fmt.Errorf("DIDWW Voice IN trunk id is required")
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO carrier_connection_provider_resources (
    carrier_connection_id, provider_id, resource_type, provider_resource_id
) VALUES ($1,$2,$3,$4)
ON CONFLICT (carrier_connection_id, resource_type)
DO UPDATE SET
    provider_id = EXCLUDED.provider_id,
    provider_resource_id = EXCLUDED.provider_resource_id,
    updated_at = now()`, connectionID, providerID, didwwProviderResourceType, resourceID); err != nil {
		return fmt.Errorf("attach DIDWW Voice IN trunk resource: %w", err)
	}
	return nil
}

func parseSourceCIDRs(values []string) ([]netip.Prefix, error) {
	result := make([]netip.Prefix, 0, len(values))
	seen := make(map[netip.Prefix]struct{}, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(value))
		if err != nil {
			return nil, fmt.Errorf("invalid DIDWW source CIDR %q: %w", value, err)
		}
		prefix = prefix.Masked()
		if _, ok := seen[prefix]; ok {
			continue
		}
		seen[prefix] = struct{}{}
		result = append(result, prefix)
	}
	return result, nil
}

func parseSIPEndpoints(values []string) ([]sipEndpoint, error) {
	result := make([]sipEndpoint, 0, len(values))
	for _, value := range values {
		endpoint, err := parseSIPEndpoint(value)
		if err != nil {
			return nil, err
		}
		result = append(result, endpoint)
	}
	return result, nil
}

func parseSIPEndpoint(value string) (sipEndpoint, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return sipEndpoint{}, fmt.Errorf("DIDWW SIP endpoint is empty")
	}
	transport := "udp"
	if before, after, ok := strings.Cut(value, "/"); ok {
		value, transport = before, strings.ToLower(strings.TrimSpace(after))
	}
	if _, err := didwwTransport(transport); err != nil {
		return sipEndpoint{}, err
	}

	host := value
	port := 5060
	if strings.Contains(value, ":") {
		parsedHost, parsedPort, err := net.SplitHostPort(value)
		if err != nil {
			return sipEndpoint{}, fmt.Errorf("invalid DIDWW SIP endpoint %q; use host:port/transport: %w", value, err)
		}
		host = parsedHost
		port, err = strconv.Atoi(parsedPort)
		if err != nil {
			return sipEndpoint{}, fmt.Errorf("invalid DIDWW SIP endpoint port %q: %w", parsedPort, err)
		}
	}
	host = strings.TrimSpace(host)
	if host == "" || port <= 0 || port > 65535 {
		return sipEndpoint{}, fmt.Errorf("invalid DIDWW SIP endpoint %q", value)
	}
	return sipEndpoint{Host: host, Port: int32(port), Transport: transport}, nil
}

func didwwTransport(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "udp", "tcp", "tls":
		return value, nil
	default:
		return "", fmt.Errorf("SIP transport must be udp, tcp or tls")
	}
}
