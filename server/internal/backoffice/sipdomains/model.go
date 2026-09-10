package sipdomains

import "github.com/leamout/leamout/internal/database/sqlc"

type SIPDomain = sqlc.ListBackofficeSIPDomainsRow
type Detail = sqlc.GetBackofficeSIPDomainRow
