package recordings

import "github.com/leamout/leamout/internal/database/sqlc"

type Recording = sqlc.ListBackofficeRecordingsRow
type Detail = sqlc.GetBackofficeRecordingRow
