package deployments
	"github.com/jackc/pgx/v5/pgtype"
	created, err := activate(ctx, transactionStore{queries: sqlc.New(tx)}, request)
	queries *sqlc.Queries
	maxDeployments, err := s.queries.LockLicenseForDeploymentActivation(ctx, sqlc.LockLicenseForDeploymentActivationParams{
		LicenseID:      request.LicenseID,
		OrganizationID: request.OrganizationID,
		ActivationTime: pgtype.Timestamptz{Time: request.At.UTC(), Valid: true},
	})
	return s.queries.CountActiveDeploymentsByLicense(ctx, sqlc.CountActiveDeploymentsByLicenseParams{
		LicenseID: request.LicenseID, OrganizationID: request.OrganizationID,
	})
	value, err := s.queries.GetDeploymentForActivation(ctx, sqlc.GetDeploymentForActivationParams{
		LicenseID: request.LicenseID, DeploymentID: request.DeploymentID,
	})
	return fromRow(value), err
	value, err := s.queries.CreateDeploymentAt(ctx, sqlc.CreateDeploymentAtParams{
		LicenseID: request.LicenseID, DeploymentID: request.DeploymentID, Name: request.Name,
		ActivatedAt: pgtype.Timestamptz{Time: request.At.UTC(), Valid: true},
	})
	return fromRow(value), err
func fromRow(value sqlc.Deployment) Deployment {
	}
