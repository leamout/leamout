package voiceapplications

import "github.com/leamout/leamout/internal/database/sqlc"

type VoiceApplication = sqlc.ListBackofficeVoiceApplicationsRow
type Detail = sqlc.GetBackofficeVoiceApplicationRow
