package organization

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/pkg/apperror"
)

const adminRole = "admin"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, req CreateRequest) (sqlc.Organization, error) {
	if userID == uuid.Nil {
		return sqlc.Organization{}, apperror.NewUnauthorized("authentication required")
	}

	slug, err := normalizeSlug(req.Slug)
	if err != nil {
		return sqlc.Organization{}, err
	}

	name, err := normalizeRequiredString(req.Name, "name")
	if err != nil {
		return sqlc.Organization{}, err
	}

	org, err := s.repo.Create(ctx, sqlc.CreateOrganizationParams{Slug: slug, Name: name})
	if err != nil {
		if isUniqueViolation(err) {
			return sqlc.Organization{}, apperror.NewConflict("organization slug already exists")
		}
		return sqlc.Organization{}, err
	}

	_, err = s.repo.AddMember(ctx, sqlc.AddOrganizationMemberParams{OrganizationID: org.ID, UserID: userID, Role: adminRole})
	if err != nil {
		return sqlc.Organization{}, apperror.NewInternal("create organization membership", err)
	}

	return org, nil
}

func (s *Service) Get(ctx context.Context, userID, orgID uuid.UUID) (sqlc.Organization, error) {
	if err := s.requireMember(ctx, userID, orgID); err != nil {
		return sqlc.Organization{}, err
	}

	org, err := s.repo.GetByID(ctx, orgID)
	if err != nil {
		return sqlc.Organization{}, apperror.NewNotFound("organization not found")
	}

	return org, nil
}

func (s *Service) Update(ctx context.Context, userID, orgID uuid.UUID, req UpdateRequest) (sqlc.Organization, error) {
	if err := s.requireMember(ctx, userID, orgID); err != nil {
		return sqlc.Organization{}, err
	}

	if req.Name != nil {
		name, err := normalizeRequiredString(*req.Name, "name")
		if err != nil {
			return sqlc.Organization{}, err
		}
		req.Name = &name
	}

	if req.Slug != nil {
		slug, err := normalizeSlug(*req.Slug)
		if err != nil {
			return sqlc.Organization{}, err
		}
		req.Slug = &slug
	}

	org, err := s.repo.Update(ctx, sqlc.UpdateOrganizationParams{ID: orgID, Name: req.Name, Slug: req.Slug})
	if err != nil {
		if isUniqueViolation(err) {
			return sqlc.Organization{}, apperror.NewConflict("organization slug already exists")
		}
		return sqlc.Organization{}, apperror.NewNotFound("organization not found")
	}

	return org, nil
}

func (s *Service) Delete(ctx context.Context, userID, orgID uuid.UUID) error {
	if err := s.requireMember(ctx, userID, orgID); err != nil {
		return err
	}

	return s.repo.Delete(ctx, orgID)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]sqlc.Organization, error) {
	if userID == uuid.Nil {
		return nil, apperror.NewUnauthorized("authentication required")
	}

	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) requireMember(ctx context.Context, userID, orgID uuid.UUID) error {
	if userID == uuid.Nil {
		return apperror.NewUnauthorized("authentication required")
	}
	if err := validateOrganizationID(orgID); err != nil {
		return err
	}

	if _, err := s.repo.GetMember(ctx, sqlc.GetOrganizationMemberParams{OrganizationID: orgID, UserID: userID}); err != nil {
		return apperror.NewNotFound("organization not found")
	}

	return nil
}

func toResponse(org sqlc.Organization) Response {
	return Response{ID: org.ID, Slug: org.Slug, Name: org.Name, Status: org.Status, CreatedAt: pgconv.TimestamptzToTime(org.CreatedAt), UpdatedAt: pgconv.TimestamptzToTime(org.UpdatedAt)}
}

func toResponses(orgs []sqlc.Organization) []Response {
	responses := make([]Response, 0, len(orgs))
	for _, org := range orgs {
		responses = append(responses, toResponse(org))
	}
	return responses
}

func isUniqueViolation(err error) bool {
	pgErr, ok := err.(*pgconn.PgError)
	return ok && pgErr.Code == "23505"
}
