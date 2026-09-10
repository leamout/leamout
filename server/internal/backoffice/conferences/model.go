package conferences

import "github.com/leamout/leamout/internal/database/sqlc"

type Conference = sqlc.ListBackofficeConferencesRow
type Detail = sqlc.GetBackofficeConferenceRow
