package agents

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

type Repository struct{
	queries *sqlc.Queries
	db *pgxpool.Pool
}

func NewRepository(queries *sqlc.Queries) *Repository { return &Repository{queries: queries} }

func (r *Repository) Create(ctx context.Context, organizationID uuid.UUID, req CreateRequest) (sqlc.VoiceAgent, error) {
	return r.queries.CreateVoiceAgent(ctx, sqlc.CreateVoiceAgentParams{
		OrganizationID:     organizationID,
		Name:               req.Name,
		Engine:             req.Engine,
		Instructions:       req.Instructions,
		Voice:              req.Voice,
		Language:           req.Language,
		EngineConfig:       []byte(req.EngineConfig),
		InterruptionPolicy: req.InterruptionPolicy,
		RecordingPolicy:    req.RecordingPolicy,
	})
}

func (r *Repository) List(ctx context.Context, organizationID uuid.UUID, req ListRequest) ([]sqlc.VoiceAgent, error) {
	return r.queries.ListVoiceAgentsByOrganizationID(ctx, sqlc.ListVoiceAgentsByOrganizationIDParams{
		OrganizationID: organizationID,
		Engine:         req.Engine,
		Language:       req.Language,
	})
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.VoiceAgent, error) {
	return r.queries.GetVoiceAgentByID(ctx, sqlc.GetVoiceAgentByIDParams{ID: id, OrganizationID: organizationID})
}

func (r *Repository) Update(ctx context.Context, organizationID, id uuid.UUID, req UpdateRequest) (sqlc.VoiceAgent, error) {
	var engineConfig []byte
	if req.EngineConfig != nil {
		engineConfig = []byte(*req.EngineConfig)
	}
	return r.queries.UpdateVoiceAgent(ctx, sqlc.UpdateVoiceAgentParams{
		Name:               req.Name,
		Engine:             req.Engine,
		Instructions:       req.Instructions,
		Voice:              req.Voice,
		Language:           req.Language,
		EngineConfig:       engineConfig,
		InterruptionPolicy: req.InterruptionPolicy,
		RecordingPolicy:    req.RecordingPolicy,
		ID:                 id,
		OrganizationID:     organizationID,
	})
}

func (r *Repository) Disable(ctx context.Context, organizationID, id uuid.UUID) error {
	return r.queries.DisableVoiceAgent(ctx, sqlc.DisableVoiceAgentParams{ID: id, OrganizationID: organizationID})
}

// SetDatabase enables serialized organization creation capacity checks.
func (r *Repository) SetDatabase(db *pgxpool.Pool) {
	r.db = db
}

// CreateWithLimit serializes creates per organization. The LIMIT applies to
// all Voice Agent records, including disabled ones.
func (r *Repository) CreateWithLimit(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateRequest,
	maxAgents int64,
) (sqlc.VoiceAgent, error) {
	if r.db == nil {
		return sqlc.VoiceAgent{}, apperror.NewServiceUnavailable("Voice Agent capacity store is unavailable", nil)
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.VoiceAgent{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	if _, err := queries.LockOrganizationForCapacity(ctx, organizationID); err != nil {
		if err == pgx.ErrNoRows {
			return sqlc.VoiceAgent{}, apperror.NewNotFound("organization not found")
		}
		return sqlc.VoiceAgent{}, err
	}
	count, err := queries.CountOrganizationVoiceAgents(ctx, organizationID)
	if err != nil {
		return sqlc.VoiceAgent{}, err
	}
	if count >= maxAgents {
		return sqlc.VoiceAgent{}, apperror.NewPaymentRequired("organization Voice Agent limit reached")
	}
	agent, err := queries.CreateVoiceAgent(ctx, sqlc.CreateVoiceAgentParams{
		OrganizationID:     organizationID,
		Name:               req.Name,
		Engine:             req.Engine,
		Instructions:       req.Instructions,
		Voice:              req.Voice,
		Language:           req.Language,
		EngineConfig:       []byte(req.EngineConfig),
		InterruptionPolicy: req.InterruptionPolicy,
		RecordingPolicy:    req.RecordingPolicy,
	})
	if err != nil {
		return sqlc.VoiceAgent{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return sqlc.VoiceAgent{}, err
	}
	return agent, nil
}
